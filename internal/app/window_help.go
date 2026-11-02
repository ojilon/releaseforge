package app

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/ojilon/releaseforge/internal/tui"
)

// helpWindow renders the command reference generated from tui.Verbs plus the
// keybinding cheat sheet. Nothing here is hand-written, so a new verb can
// never leave help stale (docs/sweep-2mo/06).
type helpWindow struct {
	windowBase
	table table.Model
	keys  viewport.Model
	hints string
	binds []key.Binding
}

func newHelpWindow() *helpWindow {
	w := &helpWindow{}
	w.table = themedTable([]table.Column{
		{Title: "command", Width: 34}, {Title: "what it does", Width: 0},
	}, nil, 60, 10)
	w.keys = viewport.New(60, 5)
	w.keys.MouseWheelEnabled = true
	w.binds = []key.Binding{
		key.NewBinding(key.WithKeys("up", "down"), key.WithHelp("↑↓", "commands")),
		key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "this window")),
	}
	return w
}

func (w *helpWindow) ID() WindowID        { return WinHelp }
func (w *helpWindow) Title() string       { return WindowTitles[WinHelp] }
func (w *helpWindow) Keys() []key.Binding { return w.binds }
func (w *helpWindow) Init() tea.Cmd       { return nil }

func (w *helpWindow) Resize(l tui.Layout) {
	w.windowBase.Resize(l)
	w.table.SetWidth(w.mainWidth())
	w.keys.Width = w.mainWidth()
	h := w.rect.Height - 7
	if h < 3 {
		h = 3
	}
	w.table.SetHeight(h)
	w.keys.Height = 4
}

func (w *helpWindow) Refresh(ctx Context) {
	var rows []table.Row
	for _, v := range tui.Verbs {
		rows = append(rows, table.Row{tui.VerbUsage(v), v.Help})
	}
	w.table.SetRows(rows)
	w.table.SetColumns(fitColumns([]table.Column{
		{Title: "command", Width: 34}, {Title: "what it does", Width: 0},
	}, w.mainWidth()))

	var b strings.Builder
	b.WriteString("  windows  1 status · 2 tests · 3 git · 4 console · 5 logs · 6 help\n")
	b.WriteString("  keys     enter run · ↑↓ history · tab complete · esc back/focus\n")
	b.WriteString("           ctrl+p palette · ctrl+r history · ctrl+o folder · ctrl+x cancel\n")
	b.WriteString("           ctrl+l clear · ]/[ next/prev window · r refresh · ctrl+c quit\n")
	b.WriteString("  theme    " + ctx.ThemeName + "  ·  `theme dark|light|mono`  ·  CLI: `releaseforge commands`")
	w.hints = b.String()
	w.keys.SetContent(w.hints)
}

func (w *helpWindow) Update(msg tea.Msg, _ Context) tea.Cmd {
	var c1, c2 tea.Cmd
	w.table, c1 = w.table.Update(msg)
	w.keys, c2 = w.keys.Update(msg)
	return tea.Batch(c1, c2)
}

func (w *helpWindow) View(_ Context) string {
	return w.fit(
		section("Command reference", w.mainWidth()) + "\n" +
			w.table.View() + "\n" +
			section("Keys", w.mainWidth()) + "\n" +
			tui.HelpStyle.Render(w.hints))
}
