package build

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunEcho(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "out.log")
	var got []string
	name := "go"
	args := []string{"version"}
	// go must exist in this dev environment; if not, skip.
	res := Run(name, args, Options{Dir: dir, LogPath: logPath, OnLine: func(s string, e bool) { got = append(got, s) }})
	if !res.Success {
		t.Fatalf("go version failed: %+v", res)
	}
	if _, err := os.Stat(logPath); err != nil {
		t.Fatalf("log not written: %v", err)
	}
	if len(got) == 0 {
		t.Fatal("expected streamed lines")
	}
}

func TestExtractErrors(t *testing.T) {
	lines := []string{"ok", "e: file.kt:10: error: something broke", "BUILD SUCCESSFUL"}
	errs := ExtractErrors(lines, 5)
	if len(errs) != 1 || !strings.Contains(errs[0], "e:") {
		t.Fatalf("got %v", errs)
	}
}
