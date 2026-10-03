package tui

import (
	"strings"
	"testing"
	"time"
)

func TestPill(t *testing.T) {
	for phase, want := range map[string]string{
		"ok": "ok", "failed": "failed", "building": "building",
		"testing": "testing", "idle": "idle", "bogus": "idle",
	} {
		if got := Pill(phase); !strings.Contains(got, want) {
			t.Fatalf("Pill(%q)=%q", phase, got)
		}
	}
}

func TestRule(t *testing.T) {
	if got := Rule(10); len([]rune(stripANSI(got))) != 8 {
		t.Fatalf("got %q", got)
	}
	if got := Rule(0); got == "" {
		t.Fatal("expected fallback rule")
	}
}

func TestSparkline(t *testing.T) {
	if got := Sparkline([]int{0, 0, 0}); got != "" {
		t.Fatalf("all-zero should hide: %q", got)
	}
	got := Sparkline([]int{0, 5, 10})
	if len([]rune(got)) != 3 {
		t.Fatalf("got %q", got)
	}
	if !strings.HasSuffix(got, "█") {
		t.Fatalf("max should be full block: %q", got)
	}
	long := make([]int, 20)
	long[19] = 7
	if got := Sparkline(long); len([]rune(got)) != 14 {
		t.Fatalf("got %d runes", len([]rune(got)))
	}
}

func TestBucketCommits(t *testing.T) {
	today := time.Date(2026, 10, 3, 12, 0, 0, 0, time.UTC)
	counts := BucketCommits([]string{"2026-10-03", "2026-10-03", "2026-09-20", "bogus", "2026-10-04"}, today)
	if counts[13] != 2 {
		t.Fatalf("today: %v", counts)
	}
	if counts[0] != 1 {
		t.Fatalf("13 days ago: %v", counts)
	}
	sum := 0
	for _, c := range counts {
		sum += c
	}
	if sum != 3 {
		t.Fatalf("future/bogus must be ignored: %v", counts)
	}
}

func stripANSI(s string) string {
	var b strings.Builder
	esc := false
	for _, r := range s {
		if r == 0x1b {
			esc = true
		}
		if !esc {
			b.WriteRune(r)
		}
		if esc && r == 'm' {
			esc = false
		}
	}
	return b.String()
}
