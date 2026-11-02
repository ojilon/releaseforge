package app

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	rflog "github.com/ojilon/releaseforge/internal/log"
	"github.com/ojilon/releaseforge/internal/tui"
)

// ringCap bounds console memory; the persisted log file stays complete.
const ringCap = 2000

// buildWindow is the console: command echo, command output and streamed
// build/test/script lines. It owns the ring and its viewport, so switching
// windows never loses scrollback (docs/sweep-2mo/41: Follow / Free policy).
type buildWindow struct {
	windowBase
	ring    *rflog.Ring
	view    viewport.Model
	spin    spinner.Model
	running bool
	label   string
	keys    []key.Binding
}

func newBuildWindow() *buildWindow {
	w := &buildWindow{ring: rflog.NewRing(ringCap)}
	w.view = viewport.New(60, 10)
	w.view.MouseWheelEnabled = true
	w.spin = spinner.New(spinner.WithSpinner(spinner.Dot))
	w.spin.Style = tui.SpinnerStyle
	w.keys = []key.Binding{
		key.NewBinding(key.WithKeys("up", "down", "pgup", "pgdown"), key.WithHelp("↑↓/pgup", "scroll")),
		key.NewBinding(key.WithKeys("g"), key.WithHelp("g", "top")),
		key.NewBinding(key.WithKeys("G"), key.WithHelp("G", "bottom")),
		key.NewBinding(key.WithKeys("f"), key.WithHelp("f", "follow")),
	}
	return w
}

func (w *buildWindow) ID() WindowID        { return WinBuild }
func (w *buildWindow) Title() string       { return WindowTitles[WinBuild] }
func (w *buildWindow) Keys() []key.Binding { return w.keys }
func (w *buildWindow) Init() tea.Cmd       { return nil }

func (w *buildWindow) Resize(l tui.Layout) {
	w.windowBase.Resize(l)
	w.view.Width = w.mainWidth()
	w.view.Height = w.rect.Height - 2
	if w.view.Height < 2 {
		w.view.Height = 2
	}
}

func (w *buildWindow) Refresh(ctx Context) {
	w.spin.Style = tui.SpinnerStyle
	w.running = ctx.Phase.Busy()
	if w.running {
		w.label = ctx.RunKind + " " + ctx.RunVariant
	} else {
		w.label = ctx.Status
	}
}

// Append adds lines to the console. Follow-mode sticks to the bottom only
// when the user was already there, so reading scrollback survives output.
func (w *buildWindow) Append(lines ...string) {
	if len(lines) == 0 {
		return
	}
	follow := w.view.AtBottom()
	for _, l := range lines {
		w.ring.Append(l)
	}
	w.view.SetContent(strings.Join(w.ring.Snapshot(), "\n"))
	if follow {
		w.view.GotoBottom()
	}
}

// AppendText appends a possibly multi-line block (command output).
func (w *buildWindow) AppendText(s string) {
	if s == "" {
		return
	}
	w.Append(strings.Split(strings.TrimRight(s, "\n"), "\n")...)
}

func (w *buildWindow) Clear() {
	w.ring = rflog.NewRing(ringCap)
	w.view.SetContent("")
	w.view.GotoTop()
}

// StartSpinner returns the command that kicks off the spinner animation.
func (w *buildWindow) StartSpinner() tea.Cmd {
	w.running = true
	w.spin.Style = tui.SpinnerStyle
	return func() tea.Msg { return w.spin.Tick() }
}

func (w *buildWindow) Update(msg tea.Msg, ctx Context) tea.Cmd {
	if s, ok := msg.(spinner.TickMsg); ok {
		if !w.running {
			return nil
		}
		var cmd tea.Cmd
		w.spin, cmd = w.spin.Update(s)
		return cmd
	}
	if k, ok := msg.(tea.KeyMsg); ok {
		switch k.String() {
		case "f":
			w.view.GotoBottom()
			return nil
		case "g":
			w.view.GotoTop()
			return nil
		case "G":
			w.view.GotoBottom()
			return nil
		}
	}
	var cmd tea.Cmd
	w.view, cmd = w.view.Update(msg)
	return cmd
}

func (w *buildWindow) View(ctx Context) string {
	var head string
	if w.running {
		head = w.spin.View() + " " + tui.WarnStyle.Render(w.label) +
			"  " + tui.HelpStyle.Render(fmt.Sprintf("%.0fs · ctrl+x cancels · logs in window 5", ctx.Elapsed.Seconds()))
	} else {
		pill := tui.Pill(ctx.Phase.String())
		head = pill + " " + tui.ValueStyle.Render(w.label) +
			"  " + tui.HelpStyle.Render("f follow · g top · G bottom")
	}
	return w.fit(section("Console", w.mainWidth()) + "\n" + head + "\n" + w.view.View())
}
