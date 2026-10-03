package metrics

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestAppendLastCap(t *testing.T) {
	root := t.TempDir()
	if _, ok := Last(root, "p", "build"); ok {
		t.Fatal("expected none")
	}
	for i := 0; i < MaxRecords+10; i++ {
		Append(root, Record{Project: "p", Kind: "build", Variant: "debug", Success: true})
	}
	data, err := os.ReadFile(filepath.Join(root, "projects", "p", "metrics.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	if n := len(strings.Split(strings.TrimSpace(string(data)), "\n")); n != MaxRecords {
		t.Fatalf("got %d lines", n)
	}
	if _, ok := Last(root, "p", "test"); ok {
		t.Fatal("kind filter broken")
	}
	rec, ok := Last(root, "p", "build")
	if !ok || !rec.Success || rec.Variant != "debug" {
		t.Fatalf("got %+v %v", rec, ok)
	}
}

func TestCorruptLinesSkipped(t *testing.T) {
	root := t.TempDir()
	p := filepath.Join(root, "projects", "p")
	os.MkdirAll(p, 0o755)
	os.WriteFile(filepath.Join(p, "metrics.jsonl"),
		[]byte("{\"at\":\"x\",\"project\":\"p\",\"kind\":\"build\"}\nnot json\n"), 0o644)
	rec, ok := Last(root, "p", "build")
	if !ok || rec.At != "x" {
		t.Fatalf("got %+v %v", rec, ok)
	}
	Append(root, Record{Project: "p", Kind: "build"})
	if _, ok := Last(root, "p", "build"); !ok {
		t.Fatal("append after corrupt failed")
	}
}

func TestAgoAndDuration(t *testing.T) {
	if Ago(time.Now().UTC().Format(time.RFC3339)) != "just now" {
		t.Fatal("now")
	}
	if Ago("bogus") != "unknown time" {
		t.Fatal("bogus")
	}
	if DurationText(850) != "850ms" || DurationText(12345) != "12.3s" {
		t.Fatal("duration")
	}
}
