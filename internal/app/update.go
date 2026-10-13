package app

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	rflog "github.com/ojilon/releaseforge/internal/log"
	"github.com/ojilon/releaseforge/internal/metrics"
	"github.com/ojilon/releaseforge/internal/tui"
	"time"
)

// Init starts the blinking cursor. Required by Bubble Tea.
func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles one event: window resize, async output, or a key press.
// It returns the updated model plus an optional follow-up command.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.onResize(msg)

	case linesMsg:
		for _, ln := range msg.lines {
			m.ring.Append(ln.Text)
		}
		m.render()
		if m.handle != nil {
			return m, m.waitLine()
		}
		return m, nil

	case doneMsg:
		m.onDone(msg)

	case tea.KeyMsg:
		quit, follow := m.onKey(msg)
		if quit {
			return m, tea.Quit
		}
		if follow != nil {
			return m, follow
		}
		// Plain typing keys fall through to the input/viewport update below.
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	cmds = append(cmds, cmd)
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

// onResize creates the log pane on first size, then keeps it fitted.
func (m *Model) onResize(msg tea.WindowSizeMsg) {
	headerH := 3
	footerH := 3
	if !m.ready {
		m.viewport = viewport.New(msg.Width, msg.Height-headerH-footerH)
		m.viewport.MouseWheelEnabled = true
		m.viewport.SetContent(strings.Join(m.ring.Snapshot(), "\n"))
		m.viewport.GotoBottom()
		m.ready = true
	} else {
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - headerH - footerH
	}
	m.width = msg.Width
	m.input.Width = msg.Width - 4
}

// onDone records metrics and prints the final ok/fail line.
func (m *Model) onDone(msg doneMsg) {
	m.running = false
	m.handle = nil
	if msg.err != nil {
		m.phase = "failed"
	} else {
		m.phase = "ok"
	}
	if m.pendingRun.kind != "" && !m.pendingRun.started.IsZero() {
		metrics.Append(m.dataRoot, metrics.Record{Project: m.projectName(),
			Kind: m.pendingRun.kind, Variant: m.pendingRun.variant,
			Success:    msg.err == nil,
			DurationMs: time.Since(m.pendingRun.started).Milliseconds()})
		m.pendingRun.kind = ""
	}
	m.status = msg.label
	for _, l := range msg.report {
		m.appendLine(l)
	}
	if msg.err != nil {
		m.appendLine(tui.ErrorStyle.Render("✗ " + msg.label + ": " + msg.err.Error()))
	} else {
		m.appendLine("✓ " + msg.label)
	}
	m.appendLine(tui.Rule(m.ruleWidth()))
	m.refreshProject()
}

// onKey handles Enter (run command), Up/Down (history), Esc (cancel),
// Ctrl+C (quit), Ctrl+L (clear). It returns (quit, follow-up command).
// Returning follow == nil means "keep typing".
func (m *Model) onKey(msg tea.KeyMsg) (bool, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC:
		return true, nil
	case tea.KeyEsc:
		if m.handle != nil {
			m.handle.Cancel()
			m.appendLine("(cancelling…)")
		}
		return false, nil
	case tea.KeyUp:
		m.historyUp()
		return false, nil
	case tea.KeyDown:
		m.historyDown()
		return false, nil
	case tea.KeyEnter:
		return m.onEnter()
	case tea.KeyCtrlL:
		m.ring = rflog.NewRing(ringCap)
		m.viewport.SetContent("")
		return false, nil
	}
	return false, nil
}

// historyUp fills the input with the previous command.
func (m *Model) historyUp() {
	if len(m.history) == 0 {
		return
	}
	if m.histIndex > 0 {
		m.histIndex--
	}
	m.input.SetValue(m.history[m.histIndex])
}

// historyDown moves forward through history, or clears the input at the end.
func (m *Model) historyDown() {
	if len(m.history) == 0 {
		return
	}
	if m.histIndex < len(m.history)-1 {
		m.histIndex++
		m.input.SetValue(m.history[m.histIndex])
	} else {
		m.histIndex = len(m.history)
		m.input.SetValue("")
	}
}
