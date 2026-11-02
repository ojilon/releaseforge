package app

import (
	"errors"
	"runtime"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/ansi"
	"github.com/ojilon/releaseforge/internal/build"
	"github.com/ojilon/releaseforge/internal/history"
	"github.com/ojilon/releaseforge/internal/storage"
	"github.com/ojilon/releaseforge/internal/tui"
)

func testModel(t *testing.T) (Model, string) {
	t.Helper()
	root := t.TempDir()
	if err := storage.EnsureLayout(root); err != nil {
		t.Fatal(err)
	}
	project := t.TempDir()
	return New(root, project), root
}

func sized(t *testing.T, m Model, w, h int) Model {
	t.Helper()
	next, _ := m.Update(tea.WindowSizeMsg{Width: w, Height: h})
	return next.(Model)
}

func keyRune(s string) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
}

func consoleText(m Model) string {
	return strings.Join(m.console().ring.Snapshot(), "\n")
}

// The shell must never wrap: every rendered line has exactly the terminal
// width and there are exactly height lines, at every breakpoint.
func TestViewFitsTerminal(t *testing.T) {
	for _, size := range [][2]int{{50, 12}, {70, 20}, {100, 30}, {160, 40}, {200, 50}} {
		m, _ := testModel(t)
		m = sized(t, m, size[0], size[1])
		out := m.View()
		lines := strings.Split(out, "\n")
		if len(lines) != size[1] {
			t.Fatalf("%dx%d: got %d lines, want %d:\n%s", size[0], size[1], len(lines), size[1], out)
		}
		for i, l := range lines {
			if w := ansi.StringWidth(l); w != size[0] {
				t.Fatalf("%dx%d line %d: width %d (want %d): %q", size[0], size[1], i, w, size[0], l)
			}
		}
	}
}

func TestViewChrome(t *testing.T) {
	m, _ := testModel(t)
	m = sized(t, m, 120, 30)
	out := m.View()
	for _, want := range []string{"ReleaseForge", "Status", "Console", "● idle", "❯"} {
		if !strings.Contains(out, want) {
			t.Fatalf("wide view missing %q:\n%s", want, out)
		}
	}
	// Compact view keeps the header and a short footer, but no tab labels.
	compact := sized(t, m, 70, 20).View()
	if !strings.Contains(compact, "ReleaseForge") || strings.Contains(compact, "Tests") {
		t.Fatalf("compact view wrong:\n%s", compact)
	}
}

func TestSingleConsumerDispatch(t *testing.T) {
	m, _ := testModel(t)
	m = sized(t, m, 120, 30)
	m.execCommand("window 6") // Help always has command rows to navigate
	hw := m.window().(*helpWindow)
	before := hw.table.Cursor()

	// Typing in the bar must not move the window's table cursor.
	next, _ := m.Update(keyRune("j"))
	m = next.(Model)
	if got := m.input.Value(); got != "j" {
		t.Fatalf("bar did not receive the key: %q", got)
	}
	if got := m.window().(*helpWindow).table.Cursor(); got != before {
		t.Fatalf("window also received the key: cursor %d -> %d", before, got)
	}

	// With the window focused, nav keys go to the window and never type.
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlW})
	m = next.(Model)
	if m.focus != FocusWindow {
		t.Fatalf("ctrl+w did not focus window: %s", m.focus)
	}
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = next.(Model)
	if m.input.Value() != "j" {
		t.Fatalf("window focus leaked into the bar: %q", m.input.Value())
	}
	if m.window().(*helpWindow).table.Cursor() != before+1 {
		t.Fatal("window did not receive the nav key")
	}
	// esc returns focus to the bar.
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = next.(Model)
	if m.focus != FocusBar {
		t.Fatal("esc did not return focus to the bar")
	}
}

func TestWindowSwitchKeepsScroll(t *testing.T) {
	m, _ := testModel(t)
	m = sized(t, m, 120, 30)
	m.console().Append(strings.Split(strings.Repeat("line\n", 100), "\n")...)
	m.switchTo(WinBuild)
	m.console().view.SetYOffset(7)
	offset := m.console().view.YOffset
	m.switchTo(WinGit)
	m.switchTo(WinBuild)
	if got := m.console().view.YOffset; got != offset {
		t.Fatalf("scroll lost across window switch: %d -> %d", offset, got)
	}
}

func TestThemeCommandWithoutArgsDoesNotPanic(t *testing.T) {
	m, _ := testModel(t)
	m = sized(t, m, 100, 30)
	m.execCommand("theme") // used to panic: args[0] index out of range
	if !strings.Contains(consoleText(m), "theme:") {
		t.Fatalf("no theme output:\n%s", consoleText(m))
	}
	m.execCommand("theme bogus")
	if !strings.Contains(consoleText(m), "unknown") {
		t.Fatal("unknown theme must be reported, not silently ignored")
	}
	m.execCommand("theme light")
	if m.themeName != "light" {
		t.Fatalf("theme not applied: %s", m.themeName)
	}
	m.execCommand("theme dark") // restore for other tests
}

func TestUnknownCommandSuggests(t *testing.T) {
	m, _ := testModel(t)
	m = sized(t, m, 100, 30)
	m.execCommand("buld")
	if !strings.Contains(consoleText(m), "did you mean `build`") {
		t.Fatalf("no suggestion:\n%s", consoleText(m))
	}
}

func TestHistoryDownClearsInput(t *testing.T) {
	m, _ := testModel(t)
	m.history = []string{"status"}
	m.histIndex = 0
	m.historyUp()
	if m.input.Value() != "status" {
		t.Fatalf("historyUp: %q", m.input.Value())
	}
	m.historyDown()
	if m.input.Value() != "" {
		t.Fatalf("historyDown past the end must clear, got %q", m.input.Value())
	}
}

func TestPaletteRanksCommands(t *testing.T) {
	m, _ := testModel(t)
	m = sized(t, m, 120, 30)
	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlP})
	m = next.(Model)
	if m.focus != FocusPalette {
		t.Fatal("ctrl+p did not open palette")
	}
	for _, r := range "build" {
		next, _ = m.Update(keyRune(string(r)))
		m = next.(Model)
	}
	if len(m.palette.shown) == 0 || !strings.Contains(m.palette.shown[0].title, "build") {
		t.Fatalf("bad ranking: %+v", m.palette.shown)
	}
	out := m.View()
	if !strings.Contains(out, "command palette") || !strings.Contains(out, "matches") {
		t.Fatalf("overlay not rendered:\n%s", out)
	}
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = next.(Model)
	if m.focus != FocusBar {
		t.Fatal("esc did not close palette")
	}
}

func TestCompletionAndQuotedPath(t *testing.T) {
	m, _ := testModel(t)
	m = sized(t, m, 120, 30)
	m.input.SetValue("bui")
	m.complete()
	if got := m.input.Value(); got != "build " {
		t.Fatalf("tab completion: %q", got)
	}
	m.input.SetValue("open \"C:/Dev/My App/repo\"")
	intent, err := tui.ParseIntent(m.input.Value())
	if err != nil || len(intent.Args) != 1 || intent.Args[0] != "C:/Dev/My App/repo" {
		t.Fatalf("quoted path: %+v %v", intent, err)
	}
}

func TestOpenRecentByIndex(t *testing.T) {
	m, root := testModel(t)
	m = sized(t, m, 120, 30)
	other := t.TempDir()
	if err := history.TouchRecent(root, other, "OtherProj", "generic", true); err != nil {
		t.Fatal(err)
	}
	m.execCommand("open 1")
	if m.projectDir != other {
		t.Fatalf("recent index did not open: %q want %q", m.projectDir, other)
	}
}

func TestCancelYieldsCancelledPhase(t *testing.T) {
	m, _ := testModel(t)
	m = sized(t, m, 120, 30)
	m.run = &RunState{Kind: "build", Variant: "debug", Started: time.Now(),
		OKLabel: "ok", FailLabel: "fail"}
	m.setPhase(PhaseBuilding)
	m.cancelRun() // no handle: should be harmless
	m.onDone(doneMsg{label: "cancelled fail", cancelled: true,
		err: errors.New("stopped by user")})
	if m.phase != PhaseCancelled {
		t.Fatalf("late doneMsg after cancel -> %s", m.phase)
	}
	if !strings.Contains(consoleText(m), "■ cancelled") {
		t.Fatalf("no cancel marker:\n%s", consoleText(m))
	}
}

// drainAll exercises the real async handle + waitLine plumbing with a
// portable command (no adb/gradle/go required).
func drainAll(t *testing.T, m *Model) (int, doneMsg) {
	t.Helper()
	for i := 0; i < 2000; i++ {
		cmd := m.waitLine()
		if cmd == nil {
			t.Fatal("waitLine returned nil while a run was active")
		}
		switch v := cmd().(type) {
		case linesMsg:
			for _, ln := range v.lines {
				m.console().ring.Append(ln.Text)
			}
		case doneMsg:
			return i + 1, v
		default:
			t.Fatalf("unexpected message %T", v)
		}
	}
	t.Fatal("drain did not terminate")
	return 0, doneMsg{}
}

func TestWaitLineDrainsAndCancel(t *testing.T) {
	program, args := "/bin/sh", []string{"-c", "echo hello; echo world"}
	if runtime.GOOS == "windows" {
		program, args = "cmd.exe", []string{"/C", "echo hello& echo world"}
	}
	dir := t.TempDir()
	m, _ := testModel(t)
	m = sized(t, m, 100, 30)
	h := build.Start(program, args, build.Options{Dir: dir})
	m.run = &RunState{Kind: "build", Variant: "debug", Started: time.Now(), Handle: h,
		OKLabel: "ok-label", FailLabel: "fail-label"}
	m.setPhase(PhaseBuilding)
	_, done := drainAll(t, &m)
	if done.err != nil || done.label != "ok-label" {
		t.Fatalf("done=%+v", done)
	}
	m.onDone(done)
	if m.phase != PhaseOK {
		t.Fatalf("phase=%s", m.phase)
	}
}

func TestKeyMapSafeDigits(t *testing.T) {
	m, _ := testModel(t)
	m = sized(t, m, 120, 30)
	// Bare digits must remain typeable in the bar (alt+N switches windows).
	next, _ := m.Update(keyRune("2"))
	m = next.(Model)
	if m.input.Value() != "2" {
		t.Fatalf("digit stolen by window switch: %q", m.input.Value())
	}
	if m.active != WinStatus {
		t.Fatal("digit switched the active window")
	}
	next, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("4"), Alt: true})
	m = next.(Model)
	if m.active != WinBuild {
		t.Fatal("alt+4 did not switch to Console")
	}
}
