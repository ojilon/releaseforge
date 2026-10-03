// Package build runs Gradle / Wails / CMake / generic commands with live log capture.
package build

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	rflog "github.com/ojilon/releaseforge/internal/log"
)

// Result is the structured outcome of a runner invocation.
type Result struct {
	Success  bool
	ExitCode int
	LogPath  string
	Errors   []string
	Command  string
	// Report holds structured errors when extraction found any.
	Report *Report
}

// Options controls a Run/Start invocation.
type Options struct {
	// Dir is the working directory.
	Dir string
	// LogPath is the file to persist output to. If empty, output is not persisted.
	LogPath string
	// OnLine is called for every output line (stdout and stderr).
	// Only Run calls it; Start delivers lines on Handle.Lines instead.
	OnLine func(text string, isErr bool)
	// Env extra environment entries.
	Env []string
	// Timeout kills the process after the duration (0 = none).
	Timeout time.Duration
}

// Handle is a running process started by Start.
type Handle struct {
	// Lines carries every output line until closed. Always drained by the
	// owner (Run drains; the TUI chains one-line reads).
	Lines <-chan rflog.Line
	// Done receives the final Result exactly once.
	Done <-chan Result
	// Cancel kills the process; the Result reports "cancelled".
	Cancel context.CancelFunc
}

// errorPatterns are matched case-insensitively to extract headline errors.
var errorPatterns = []string{
	"FAILED", "FAILURE", "BUILD FAILED",
	"e: ", "error:", "Error:", "ERROR:",
	"Execution failed", "What went wrong",
	"UnsatisfiedLinkError",
}

// ExtractErrors returns up to max lines matching error patterns.
func ExtractErrors(lines []string, max int) []string {
	var out []string
	for _, l := range lines {
		ll := strings.ToLower(l)
		for _, p := range errorPatterns {
			if strings.Contains(l, p) || strings.Contains(ll, strings.ToLower(p)) {
				t := strings.TrimSpace(l)
				if t != "" {
					out = append(out, t)
				}
				break
			}
		}
		if len(out) >= max {
			break
		}
	}
	return out
}

// Start executes name args in opts.Dir without blocking, streaming output on
// Handle.Lines and the final Result on Handle.Done. The owner must drain
// Lines to completion; Cancel aborts the process.
func Start(name string, args []string, opts Options) *Handle {
	ctx, cancel := context.WithCancel(context.Background())
	lines := make(chan rflog.Line, 1024)
	done := make(chan Result, 1)
	var timer *time.Timer
	var timedOut atomic.Bool
	if opts.Timeout > 0 {
		d := opts.Timeout
		timer = time.AfterFunc(d, func() { timedOut.Store(true); cancel() })
	}
	go func() {
		defer close(lines)
		if timer != nil {
			defer timer.Stop()
		}
		res := execute(ctx, name, args, opts, func(text string, isErr bool) {
			lines <- rflog.Line{Text: text, IsErr: isErr}
		})
		if ctx.Err() != nil {
			res.Success = false
			if timedOut.Load() {
				res.Errors = append([]string{fmt.Sprintf("timed out after %s", opts.Timeout)}, res.Errors...)
			} else {
				res.Errors = append([]string{"cancelled"}, res.Errors...)
			}
		}
		done <- res
	}()
	return &Handle{Lines: lines, Done: done, Cancel: cancel}
}

// Run executes name args in opts.Dir, streams output via OnLine, persists to
// LogPath, and blocks until the process exits.
func Run(name string, args []string, opts Options) Result {
	h := Start(name, args, opts)
	for ln := range h.Lines {
		if opts.OnLine != nil {
			opts.OnLine(ln.Text, ln.IsErr)
		}
	}
	return <-h.Done
}

// execute runs the process once; emit receives every output line.
func execute(ctx context.Context, name string, args []string, opts Options, emit func(string, bool)) Result {
	display := name + " " + strings.Join(args, " ")
	var stream *rflog.Stream
	if opts.LogPath != "" {
		s, err := rflog.New(opts.LogPath)
		if err == nil {
			stream = s
			defer s.Close()
			s.Write("$ "+display+" (dir="+opts.Dir+")", false)
		}
	}
	sink := func(text string, isErr bool) {
		if stream != nil {
			stream.Write(text, isErr)
		}
		emit(text, isErr)
	}

	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = opts.Dir
	if len(opts.Env) > 0 {
		cmd.Env = append(os.Environ(), opts.Env...)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return Result{ExitCode: -1, LogPath: opts.LogPath, Command: display, Errors: []string{err.Error()}}
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return Result{ExitCode: -1, LogPath: opts.LogPath, Command: display, Errors: []string{err.Error()}}
	}
	if err := cmd.Start(); err != nil {
		sink(fmt.Sprintf("failed to start: %v", err), true)
		return Result{ExitCode: -1, LogPath: opts.LogPath, Command: display, Errors: []string{err.Error()}}
	}
	var lines []string
	var linesMu sync.Mutex
	appendLine := func(t string) {
		linesMu.Lock()
		lines = append(lines, t)
		linesMu.Unlock()
	}
	done := make(chan struct{}, 2)
	// bufio.Reader wrapper not needed; use pipes directly.
	go func() {
		sc := bufio.NewScanner(stdout)
		sc.Buffer(make([]byte, 1024*1024), 1024*1024)
		for sc.Scan() {
			t := sc.Text()
			appendLine(t)
			sink(t, false)
		}
		done <- struct{}{}
	}()
	go func() {
		sc := bufio.NewScanner(stderr)
		sc.Buffer(make([]byte, 1024*1024), 1024*1024)
		for sc.Scan() {
			t := sc.Text()
			appendLine(t)
			sink(t, true)
		}
		done <- struct{}{}
	}()
	<-done
	<-done
	waitErr := cmd.Wait()
	code := 0
	if waitErr != nil {
		if ee, ok := waitErr.(*exec.ExitError); ok {
			code = ee.ExitCode()
		} else {
			code = 1
		}
	}
	errs := ExtractErrors(lines, 15)
	res := Result{Success: code == 0, ExitCode: code, LogPath: opts.LogPath, Command: display, Errors: errs}
	if rep := ParseErrors(lines); len(rep.Items) > 0 {
		rep.LogPath = opts.LogPath
		res.Report = &rep
	}
	return res
}

// GradleWrapper returns the gradle wrapper script for dir.
func GradleWrapper(dir string) string {
	if runtime.GOOS == "windows" {
		if _, err := os.Stat(filepath.Join(dir, "gradlew.bat")); err == nil {
			return filepath.Join(dir, "gradlew.bat")
		}
	} else {
		if _, err := os.Stat(filepath.Join(dir, "gradlew")); err == nil {
			return filepath.Join(dir, "gradlew")
		}
	}
	return "gradle"
}
