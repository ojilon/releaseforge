package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestComputeBreakpoints(t *testing.T) {
	cases := []struct {
		w, h int
		want Mode
	}{
		{50, 12, ModeTiny},
		{70, 24, ModeCompact},
		{100, 30, ModeNormal},
		{130, 40, ModeWide},
		{200, 60, ModeUltra},
		{0, 0, ModeTiny},
	}
	for _, c := range cases {
		l := Compute(c.w, c.h)
		if l.Mode != c.want {
			t.Errorf("Compute(%d,%d) mode=%s want %s", c.w, c.h, l.Mode, c.want)
		}
		if !l.Content.Valid() || !l.BarRect().Valid() || !l.HeaderRect().Valid() {
			t.Errorf("Compute(%d,%d) produced an invalid rect: %+v", c.w, c.h, l)
		}
		total := l.HeaderH + l.TabsH + l.BodyH + l.BarH + l.FooterH
		if c.h >= 3 && total != c.h {
			t.Errorf("Compute(%d,%d) chrome sums to %d, want %d", c.w, c.h, total, c.h)
		}
	}
}

func TestSplit(t *testing.T) {
	l := Compute(70, 30)
	if p, s := l.Split(); s.Valid() || p != l.Content {
		t.Fatalf("narrow layout must not split: %+v %+v", p, s)
	}
	wide := Compute(160, 40)
	p, s := wide.Split()
	if !s.Valid() || p.Width+s.Width+2 != wide.Content.Width {
		t.Fatalf("wide split: %+v %+v (content %+v)", p, s, wide.Content)
	}
}

func TestClampAndFrameExact(t *testing.T) {
	content := "short\n" + strings.Repeat("x", 200) + "\n日本語とemoji🙂"
	c := Clamp(content, 40, 5)
	lines := strings.Split(c, "\n")
	if len(lines) != 5 {
		t.Fatalf("clamp height: %d", len(lines))
	}
	for _, l := range lines {
		if w := ansi.StringWidth(l); w != 40 {
			t.Fatalf("clamp width %d: %q", w, l)
		}
	}
	f := Frame(content, 60, 10, FrameBorder)
	if got := strings.Count(f, "\n") + 1; got != 10 {
		t.Fatalf("frame height %d", got)
	}
	for _, l := range strings.Split(f, "\n") {
		if w := ansi.StringWidth(l); w != 60 {
			t.Fatalf("frame width %d: %q", w, l)
		}
	}
}

func TestTruncateWideRunes(t *testing.T) {
	if got := Truncate("日本語テキスト", 4); ansi.StringWidth(got) > 4 || got == "" {
		t.Fatalf("wide rune truncate: %q (width %d)", got, ansi.StringWidth(got))
	}
	if got := Truncate("abc🙂def", 5); ansi.StringWidth(got) > 5 || !strings.Contains(got, "…") {
		t.Fatalf("emoji truncate: %q", got)
	}
	if got := Truncate("plain", 100); got != "plain" {
		t.Fatalf("no-op truncate: %q", got)
	}
}

func TestTabsCompactAndWide(t *testing.T) {
	labels := []string{"Status", "Tests", "Git"}
	wide := Tabs(labels, 1, 80, false)
	if !strings.Contains(wide, "Tests") {
		t.Fatalf("wide tabs: %q", wide)
	}
	compact := Tabs(labels, 1, 20, true)
	if strings.Contains(compact, "Status") {
		t.Fatalf("compact tabs show inactive labels: %q", compact)
	}
	if !strings.Contains(compact, "Tests") {
		t.Fatalf("compact tabs missing active label: %q", compact)
	}
}

func TestPlaceCenterNeverOverflows(t *testing.T) {
	l := Compute(100, 30)
	out := PlaceCenter(l, strings.Repeat("x", 5000), 80, 20)
	if got := strings.Count(out, "\n") + 1; got != 30 {
		t.Fatalf("overlay screen height %d", got)
	}
	for _, line := range strings.Split(out, "\n") {
		if w := ansi.StringWidth(line); w != 100 {
			t.Fatalf("overlay screen width %d", w)
		}
	}
}
