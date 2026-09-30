// Package build runs Gradle / Wails / CMake / generic commands with live log capture.
package build

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	rflog "github.com/ojilon/releaseforge/internal/log"
)

// Result is the structured outcome of a runner invocation.
type Result struct {
	Success  bool
	ExitCode int
	LogPath  string
	Errors   []string
	Command  string
}

// Options controls a Run invocation.
type Options struct {
	// Dir is the working directory.
	Dir string
	// LogPath is the file to persist output to. If empty, output is not persisted.
	LogPath string
	// OnLine is called for every output line (stdout and stderr).
	OnLine func(text string, isErr bool)
	// Env extra environment entries.
	Env []string
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

// Run executes name args in opts.Dir, streams output, persists to LogPath.
func Run(name string, args []string, opts Options) Result {
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
	emit := func(text string, isErr bool) {
		if stream != nil {
			stream.Write(text, isErr)
		}
		if opts.OnLine != nil {
			opts.OnLine(text, isErr)
		}
	}

	cmd := exec.Command(name, args...)
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
		emit(fmt.Sprintf("failed to start: %v", err), true)
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
			emit(t, false)
		}
		done <- struct{}{}
	}()
	go func() {
		sc := bufio.NewScanner(stderr)
		sc.Buffer(make([]byte, 1024*1024), 1024*1024)
		for sc.Scan() {
			t := sc.Text()
			appendLine(t)
			emit(t, true)
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
	return Result{Success: code == 0, ExitCode: code, LogPath: opts.LogPath, Command: display, Errors: errs}
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

// GradleArgs builds a gradle invocation for a task list.
func GradleArgs(tasks ...string) []string {
	return tasks
}
