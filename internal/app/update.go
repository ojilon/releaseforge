package app

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/ojilon/releaseforge/internal/history"
	"github.com/ojilon/releaseforge/internal/metrics"
	"github.com/ojilon/releaseforge/internal/tui"
)

// Update is the single-consumer router: every message goes to exactly one
// component. Typed keys never scroll a viewport; nav keys never type into
// the command bar (docs/sweep-2mo/41).
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.onResize(msg)
		return m, nil

	case linesMsg:
		lines := make([]string, 0, len(msg.lines))
		for _, ln := range msg.lines {
			lines = append(lines, ln.Text)
		}
		m.console().Append(lines...)
		if m.run != nil {
			return m, m.waitLine()
		}
		return m, nil

	case doneMsg:
		m.onDone(msg)
		return m, nil

	case tickMsg:
		if m.run != nil && !m.run.Started.IsZero() {
			m.elapsed = time.Since(m.run.Started)
		} else {
			m.elapsed = 0
		}
		return m, tickCmd()

	case spinner.TickMsg:
		return m, m.console().Update(msg, m.context())

	case tea.KeyMsg:
		quit, cmd := m.onKey(msg)
		if quit {
			return m, tea.Quit
		}
		return m, cmd

	default:
		return m, m.onOther(msg)
	}
}

// onResize gives every window its exact rect, then refreshes cached data on
// the first real size message.
func (m *Model) onResize(msg tea.WindowSizeMsg) {
	m.width, m.height = msg.Width, msg.Height
	m.layout = tui.Compute(msg.Width, msg.Height)
	bar := m.layout.BarRect()
	m.input.Width = bar.Width - 4
	if m.input.Width < 8 {
		m.input.Width = 8
	}
	for _, w := range m.windows {
		w.Resize(m.layout)
	}
	pw, ph := m.paletteSize()
	m.palette.Resize(pw, ph)
	if !m.ready {
		m.ready = true
		m.refreshWindows()
	}
}

func (m *Model) paletteSize() (int, int) {
	w := m.layout.Width * 60 / 100
	if w > 76 {
		w = 76
	}
	if w < 40 {
		w = 40
	}
	h := m.layout.Height * 60 / 100
	if h > 22 {
		h = 22
	}
	if h < 10 {
		h = 10
	}
	return w - 2, h - 6
}

// onOther routes non-key messages (cursor blink, mouse wheel, filepicker
// directory reads) to the focused overlay or the bar/window pair.
func (m *Model) onOther(msg tea.Msg) tea.Cmd {
	switch m.focus {
	case FocusPalette:
		var cmd tea.Cmd
		m.palette.query, cmd = m.palette.query.Update(msg)
		return cmd
	case FocusPicker:
		if m.picker != nil {
			var cmd tea.Cmd
			*m.picker, cmd = m.picker.Update(msg)
			return cmd
		}
		return nil
	}
	var c1, c2 tea.Cmd
	m.input, c1 = m.input.Update(msg)
	if _, isMouse := msg.(tea.MouseMsg); isMouse {
		c2 = m.window().Update(msg, m.context())
	}
	return tea.Batch(c1, c2)
}

// onKey handles global shortcuts first, then dispatches to the focused
// component. Returns (quit, follow-up command).
func (m *Model) onKey(k tea.KeyMsg) (bool, tea.Cmd) {
	if key.Matches(k, m.keys.Quit) {
		return true, nil
	}
	switch m.focus {
	case FocusPalette:
		return m.onPaletteKey(k)
	case FocusPicker:
		return m.onPickerKey(k)
	}

	// Window switching: alt+1..6 always; bare 1..6 only while the window
	// owns focus, so typing "open 2" never hijacks the digit.
	for i := range m.keys.Windows {
		plain := m.focus == FocusWindow && key.Matches(k, m.keys.Windows[i])
		if plain || key.Matches(k, m.keys.WindowsAlt[i]) {
			m.switchTo(WindowID(i))
			return false, nil
		}
	}
	if key.Matches(k, m.keys.NextWindow) {
		m.switchTo((m.active + 1) % windowCount)
		return false, nil
	}
	if key.Matches(k, m.keys.PrevWindow) {
		m.switchTo((m.active + windowCount - 1) % windowCount)
		return false, nil
	}
	if key.Matches(k, m.keys.Help) {
		m.switchTo(WinHelp)
		return false, nil
	}
	if key.Matches(k, m.keys.Palette) {
		m.focus = FocusPalette
		return false, m.palette.open(m.paletteItems(), "all")
	}
	if key.Matches(k, m.keys.History) {
		m.focus = FocusPalette
		return false, m.palette.open(m.historyItems(), "history")
	}
	if key.Matches(k, m.keys.OpenFolder) {
		return false, m.openPicker()
	}
	if key.Matches(k, m.keys.Cancel) {
		m.cancelRun()
		return false, nil
	}
	if key.Matches(k, m.keys.Clear) {
		m.console().Clear()
		m.hint = ""
		return false, nil
	}
	if key.Matches(k, m.keys.WindowFocus) {
		if m.focus == FocusBar {
			m.focus = FocusWindow
		} else {
			m.focus = FocusBar
		}
		return false, nil
	}
	if key.Matches(k, m.keys.Focus) { // esc
		if m.focus == FocusWindow {
			m.focus = FocusBar
		} else if m.phase.Busy() {
			m.cancelRun()
		} else if m.input.Value() != "" {
			m.input.SetValue("")
			m.hint = ""
		}
		return false, nil
	}

	if m.focus == FocusWindow {
		if k.String() == "r" {
			m.refreshWindows()
			m.hint = "refreshed"
		} else {
			return false, m.window().Update(k, m.context())
		}
		return false, nil
	}

	// FocusBar: one consumer, the text input.
	switch k.Type {
	case tea.KeyEnter:
		return m.onEnter()
	case tea.KeyUp:
		m.historyUp()
		return false, nil
	case tea.KeyDown:
		m.historyDown()
		return false, nil
	case tea.KeyTab:
		m.complete()
		return false, nil
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(k)
	if k.Type == tea.KeyRunes || k.Type == tea.KeySpace || k.Type == tea.KeyBackspace {
		m.hint = ""
	}
	return false, cmd
}

// onEnter runs the current command bar line.
func (m *Model) onEnter() (bool, tea.Cmd) {
	line := strings.TrimSpace(m.input.Value())
	if line == "" {
		return false, nil
	}
	m.history = append(m.history, line)
	m.histIndex = len(m.history)
	history.AppendCommand(m.dataRoot, line)
	m.input.SetValue("")
	m.hint = ""
	quit, cmd := m.exec(line)
	if quit {
		return true, nil
	}
	return false, cmd
}

// complete is tab completion generated from the verb table + live data.
func (m *Model) complete() {
	line := m.input.Value()
	candidates := tui.Completions(line, m.completionExtras())
	switch len(candidates) {
	case 0:
		m.hint = "no completion — type `help` for commands"
	case 1:
		m.input.SetValue(candidates[0])
		m.input.CursorEnd()
		m.hint = ""
	default:
		prefix := tui.CommonPrefix(candidates)
		if len(prefix) > len(line) {
			m.input.SetValue(prefix)
			m.input.CursorEnd()
		}
		names := make([]string, 0, len(candidates))
		for _, c := range candidates {
			names = append(names, strings.TrimSpace(c))
		}
		m.hint = "matches: " + strings.Join(names, " ")
	}
}

func (m *Model) completionExtras() []string {
	var out []string
	if rf, err := history.LoadRecent(m.dataRoot); err == nil {
		for _, e := range rf.Items {
			out = append(out, quotePath(e.Path))
		}
	}
	out = append(out, tui.ThemeNames()...)
	for _, w := range WindowTitles {
		out = append(out, strings.ToLower(w))
	}
	return out
}

func (m *Model) historyUp() {
	if len(m.history) == 0 {
		return
	}
	if m.histIndex > 0 {
		m.histIndex--
	}
	m.input.SetValue(m.history[m.histIndex])
	m.input.CursorEnd()
}

// historyDown walks forward and clears the input once past the newest entry.
// (The old implementation got stuck on the last item.)
func (m *Model) historyDown() {
	if len(m.history) == 0 {
		return
	}
	if m.histIndex < len(m.history)-1 {
		m.histIndex++
		m.input.SetValue(m.history[m.histIndex])
		m.input.CursorEnd()
	} else {
		m.histIndex = len(m.history)
		m.input.SetValue("")
	}
}

// cancelRun marks the run cancelled BEFORE killing it, so a late doneMsg is
// classified as cancelled rather than failed (docs/sweep-2mo/40).
func (m *Model) cancelRun() {
	if m.run == nil || m.run.Handle == nil {
		m.hint = "nothing is running"
		return
	}
	m.run.Cancelled = true
	m.run.Handle.Cancel()
	m.appendLine(tui.WarnStyle.Render("(cancelling… esc / ctrl+x)"))
	m.hint = "cancelling"
}

// onDone records metrics and prints the final result into the console.
func (m *Model) onDone(msg doneMsg) {
	if m.run != nil && !m.run.Started.IsZero() {
		metrics.Append(m.dataRoot, metrics.Record{
			Project: m.projectName(), Kind: m.run.Kind, Variant: m.run.Variant,
			Success:    msg.err == nil && !msg.cancelled,
			DurationMs: time.Since(m.run.Started).Milliseconds(),
		})
	}
	m.run = nil
	m.elapsed = 0
	switch {
	case msg.cancelled:
		m.setPhase(PhaseCancelled)
	case msg.err != nil:
		m.setPhase(PhaseFailed)
	default:
		m.setPhase(PhaseOK)
	}
	if len(msg.report) > 0 {
		m.appendLine(tui.RenderReport(msg.summary, msg.report))
	}
	if msg.cancelled {
		m.appendLine(tui.WarnStyle.Render("■ " + msg.label))
	} else if msg.err != nil {
		m.appendLine(tui.ErrorStyle.Render("✗ " + msg.label + ": " + msg.err.Error()))
	} else {
		m.appendLine(tui.SuccessStyle.Render("✓ " + msg.label))
	}
	m.appendLine(tui.Rule(m.ruleWidth()))
	m.status = msg.label
	m.refreshProject()
}

// onPaletteKey owns the palette overlay.
func (m *Model) onPaletteKey(k tea.KeyMsg) (bool, tea.Cmd) {
	switch k.String() {
	case "esc", "ctrl+c":
		m.focus = FocusBar
		m.hint = ""
		return false, nil
	case "enter":
		it, ok := m.palette.selected()
		m.focus = FocusBar
		m.hint = ""
		if !ok {
			return false, nil
		}
		if it.kind == "act" {
			return m.exec(strings.TrimSpace(it.value))
		}
		m.input.SetValue(it.value)
		m.input.CursorEnd()
		return false, nil
	case "up", "down", "pgup", "pgdown", "ctrl+n", "ctrl+j", "ctrl+k":
		var cmd tea.Cmd
		m.palette.list, cmd = m.palette.list.Update(k)
		return false, cmd
	}
	var cmd tea.Cmd
	m.palette.query, cmd = m.palette.query.Update(k)
	m.palette.rank()
	return false, cmd
}

// onPickerKey owns the folder picker overlay.
func (m *Model) onPickerKey(k tea.KeyMsg) (bool, tea.Cmd) {
	if m.picker == nil {
		m.focus = FocusBar
		return false, nil
	}
	switch k.String() {
	case "esc":
		m.focus = FocusBar
		m.picker = nil
		return false, nil
	case "s": // select the directory currently being viewed
		dir := m.picker.CurrentDirectory
		m.focus = FocusBar
		m.picker = nil
		return m.exec("open " + quotePath(dir))
	}
	var cmd tea.Cmd
	*m.picker, cmd = m.picker.Update(k)
	if ok, path := m.picker.DidSelectFile(k); ok {
		m.focus = FocusBar
		m.picker = nil
		return m.exec("open " + quotePath(path))
	}
	return false, cmd
}

// execCommand is used by the demo generator and tests to feed a line through
// the same path a keystroke would take.
func (m *Model) execCommand(line string) (bool, tea.Cmd) {
	m.input.SetValue(line)
	return m.onEnter()
}
