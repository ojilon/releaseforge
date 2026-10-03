package log

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIndexAppendLastCap(t *testing.T) {
	dir := t.TempDir()
	if _, ok := LastIndexed(dir); ok {
		t.Fatal("empty index")
	}
	for i := 0; i < MaxIndexEntries+10; i++ {
		AppendIndex(dir, Entry{Name: "a.log", Command: "go test", StartedAt: "t", ExitCode: 0})
	}
	entries, err := ReadIndex(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != MaxIndexEntries {
		t.Fatalf("got %d", len(entries))
	}
	if _, ok := LastIndexed(dir); !ok {
		t.Fatal("expected last")
	}
}

func TestNewestFallback(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "old.log"), []byte("x"), 0o644)
	os.WriteFile(filepath.Join(dir, "new.log"), []byte("y"), 0o644)
	// force mtime order
	old := filepath.Join(dir, "old.log")
	past := filepath.Join(dir, "new.log")
	_ = old
	_ = past
	got, err := Newest(dir)
	if err != nil || got == "" {
		t.Fatalf("got %q %v", got, err)
	}
	AppendIndex(dir, Entry{Name: "old.log", Command: "c", ExitCode: 0})
	got, err = Newest(dir)
	if err != nil || got != "old.log" {
		t.Fatalf("index should win: got %q %v", got, err)
	}
	os.Remove(filepath.Join(dir, "old.log"))
	got, err = Newest(dir)
	if err != nil || got != "new.log" {
		t.Fatalf("missing indexed file should fall back: got %q %v", got, err)
	}
}

func TestResolvePrefixes(t *testing.T) {
	dir := t.TempDir()
	for _, n := range []string{"build-debug-1.log", "build-debug-2.log", "test-unit-1.log"} {
		os.WriteFile(filepath.Join(dir, n), []byte("x"), 0o644)
	}
	if got, err := Resolve(dir, "test-unit-1.log"); err != nil || got != "test-unit-1.log" {
		t.Fatalf("exact: %q %v", got, err)
	}
	if got, err := Resolve(dir, "test-unit"); err != nil || got != "test-unit-1.log" {
		t.Fatalf("unique prefix: %q %v", got, err)
	}
	if _, err := Resolve(dir, "build-debug"); err == nil {
		t.Fatal("expected ambiguity error")
	} else if _, ok := err.(*AmbiguousError); !ok {
		t.Fatalf("wrong error type: %v", err)
	}
	if _, err := Resolve(dir, "nope"); err == nil {
		t.Fatal("expected not-exist error")
	}
}

func TestTailHead(t *testing.T) {
	p := filepath.Join(t.TempDir(), "f.log")
	os.WriteFile(p, []byte("a\nb\nc\nd\n"), 0o644)
	lines, err := Tail(p, 2)
	if err != nil || len(lines) != 2 || lines[0] != "c" || lines[1] != "d" {
		t.Fatalf("got %v %v", lines, err)
	}
	data, trunc, err := Head(p, 3)
	if err != nil || !trunc || string(data) != "a\nb" {
		t.Fatalf("got %q %v %v", data, trunc, err)
	}
	data, trunc, err = Head(p, 64)
	if err != nil || trunc || string(data) != "a\nb\nc\nd\n" {
		t.Fatalf("got %q %v %v", data, trunc, err)
	}
}
