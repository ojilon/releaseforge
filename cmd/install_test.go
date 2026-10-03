package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewestReleaseDir(t *testing.T) {
	root := t.TempDir()
	for _, d := range []string{"0.0.1", "0.0.2", "scratch"} {
		if err := os.MkdirAll(filepath.Join(root, d), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(root, "junk.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := newestReleaseDir(root); got != "0.0.2" {
		t.Fatalf("got %q want 0.0.2", got)
	}
	plain := t.TempDir()
	if err := os.MkdirAll(filepath.Join(plain, "nodots"), 0o755); err != nil {
		t.Fatal(err)
	}
	if got := newestReleaseDir(plain); got != "nodots" {
		t.Fatalf("fallback got %q want nodots", got)
	}
	if got := newestReleaseDir(filepath.Join(t.TempDir(), "missing")); got != "" {
		t.Fatalf("missing dir got %q want empty", got)
	}
}
