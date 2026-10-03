package build

import (
	"runtime"
	"strings"
	"testing"
	"time"
)

// sleeper returns a ~5s portable sleep command.
func sleeper() (string, []string) {
	if runtime.GOOS == "windows" {
		return "ping", []string{"-n", "6", "127.0.0.1"}
	}
	return "sleep", []string{"5"}
}

func TestStartStreamsBeforeExit(t *testing.T) {
	h := Start("go", []string{"version"}, Options{Dir: t.TempDir()})
	n := 0
	for range h.Lines {
		n++
	}
	res := <-h.Done
	if !res.Success || n == 0 {
		t.Fatalf("success=%v lines=%d errs=%v", res.Success, n, res.Errors)
	}
}

func TestStartCancel(t *testing.T) {
	prog, args := sleeper()
	h := Start(prog, args, Options{Dir: t.TempDir()})
	time.Sleep(300 * time.Millisecond)
	h.Cancel()
	for range h.Lines {
	}
	res := <-h.Done
	if res.Success {
		t.Fatal("expected failure after cancel")
	}
	if len(res.Errors) == 0 || res.Errors[0] != "cancelled" {
		t.Fatalf("got %v", res.Errors)
	}
}

func TestStartTimeout(t *testing.T) {
	prog, args := sleeper()
	h := Start(prog, args, Options{Dir: t.TempDir(), Timeout: 200 * time.Millisecond})
	for range h.Lines {
	}
	res := <-h.Done
	if res.Success {
		t.Fatal("expected timeout failure")
	}
	if len(res.Errors) == 0 || !strings.HasPrefix(res.Errors[0], "timed out after ") {
		t.Fatalf("got %v", res.Errors)
	}
}

func TestRunStillCallsOnLine(t *testing.T) {
	n := 0
	res := Run("go", []string{"version"}, Options{Dir: t.TempDir(), OnLine: func(string, bool) { n++ }})
	if !res.Success || n == 0 {
		t.Fatalf("success=%v lines=%d", res.Success, n)
	}
}
