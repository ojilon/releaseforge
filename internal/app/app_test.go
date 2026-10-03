package app

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ojilon/releaseforge/internal/build"
)

func windowSizeMsg(w, h int) tea.Msg {
	return tea.WindowSizeMsg{Width: w, Height: h}
}

func TestParse(t *testing.T) {
	cases := map[string][2]string{
		"":                {"", ""},
		"  ":              {"", ""},
		"quit":            {"quit", ""},
		"Q":               {"q", ""},
		"?":               {"?", ""},
		"-h":              {"-h", ""},
		"scan D:\\x":      {"scan", "D:\\x"},
		"OPEN":            {"open", ""},
		"rescan":          {"rescan", ""},
		"build release":   {"build", "release"},
		"test all":        {"test", "all"},
		"version --set X": {"version", "--set"},
		"notes 0.1":       {"notes", "0.1"},
		"frobnicate":      {"frobnicate", ""},
	}
	for in, want := range cases {
		verb, args := parse(in)
		if verb != want[0] {
			t.Errorf("parse(%q) verb=%q want %q", in, verb, want[0])
		}
		if want[1] != "" && (len(args) == 0 || args[0] != want[1]) {
			t.Errorf("parse(%q) args=%v want first %q", in, args, want[1])
		}
	}
}

func drainAll(t *testing.T, m *Model) (lines int, done doneMsg) {
	t.Helper()
	for i := 0; i < 1000; i++ {
		msg := m.waitLine()()
		switch v := msg.(type) {
		case linesMsg:
			lines += len(v.lines)
			for _, ln := range v.lines {
				m.ring.Append(ln.Text)
			}
		case doneMsg:
			return lines, v
		default:
			t.Fatalf("unexpected message %T", msg)
		}
	}
	t.Fatal("drain did not terminate")
	return 0, doneMsg{}
}

func TestWaitLineDrains(t *testing.T) {
	dir := t.TempDir()
	m := New(dir, dir)
	h := build.Start("go", []string{"version"}, build.Options{Dir: dir})
	m.handle = h
	m.pendingOK, m.pendingErr = "ok-label", "fail-label"
	n, done := drainAll(t, &m)
	if n == 0 {
		t.Fatal("no lines streamed")
	}
	if done.err != nil || done.label != "ok-label" {
		t.Fatalf("done=%+v", done)
	}
}

func TestCancelTerminatesDrain(t *testing.T) {
	dir := t.TempDir()
	m := New(dir, dir)
	h := build.Start("go", []string{"version"}, build.Options{Dir: dir})
	m.handle = h
	m.pendingOK, m.pendingErr = "ok-label", "fail-label"
	h.Cancel()
	n, done := drainAll(t, &m)
	_ = n
	if done.err == nil {
		// Finished before the kill landed; either outcome is valid as long
		// as the drain terminated instead of hanging.
		t.Logf("process beat the cancel (label %q)", done.label)
	} else if done.label != "fail-label" {
		t.Fatalf("done=%+v", done)
	}
}

func TestViewWideAndNarrow(t *testing.T) {
	dir := t.TempDir()
	m := New(dir, dir)
	wide, _ := m.Update(windowSizeMsg(100, 30))
	mw := wide.(Model)
	out := mw.View()
	for _, want := range []string{"ReleaseForge", "● idle", "help · status"} {
		if !strings.Contains(out, want) {
			t.Fatalf("wide missing %q:\n%s", want, out)
		}
	}
	if !strings.Contains(out, "╭") {
		t.Fatalf("wide should be bordered:\n%s", out)
	}
	narrow, _ := mw.Update(windowSizeMsg(70, 30))
	out = narrow.(Model).View()
	if strings.Contains(out, "recent · version") {
		t.Fatalf("narrow should hide long hints:\n%s", out)
	}
	if !strings.Contains(out, "help|quit") {
		t.Fatalf("narrow footer missing:\n%s", out)
	}
}

func TestViewPhasePill(t *testing.T) {
	dir := t.TempDir()
	m := New(dir, dir)
	m.phase = "failed"
	mw, _ := m.Update(windowSizeMsg(100, 30))
	if out := mw.(Model).View(); !strings.Contains(out, "● failed") {
		t.Fatalf("missing failed pill:\n%s", out)
	}
}
