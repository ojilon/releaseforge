package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	rflog "github.com/ojilon/releaseforge/internal/log"
	"github.com/ojilon/releaseforge/internal/storage"
	"github.com/ojilon/releaseforge/internal/tui"
)

// logsWindow is a persisted-log browser: bubbles/list on the left (or top in
// compact mode), a scrollable preview on the right. Selection is remembered
// across window switches.
type logsWindow struct {
	windowBase
	browser list.Model
	preview viewport.Model
	dir     string
	current string
	keys    []key.Binding
}

func newLogsWindow() *logsWindow {
	w := &logsWindow{}
	w.browser = themedList(nil, 30, 8)
	w.preview = viewport.New(60, 8)
	w.preview.MouseWheelEnabled = true
	w.keys = []key.Binding{
		key.NewBinding(key.WithKeys("up", "down"), key.WithHelp("↑↓", "logs")),
		key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "refresh")),
		key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "preview")),
	}
	return w
}

func (w *logsWindow) ID() WindowID        { return WinLogs }
func (w *logsWindow) Title() string       { return WindowTitles[WinLogs] }
func (w *logsWindow) Keys() []key.Binding { return w.keys }
func (w *logsWindow) Init() tea.Cmd       { return nil }

func (w *logsWindow) Resize(l tui.Layout) {
	w.windowBase.Resize(l)
	if w.horizontal() {
		lw := w.rect.Width / 3
		if lw > 34 {
			lw = 34
		}
		if lw < 20 {
			lw = 20
		}
		w.browser.SetSize(lw, w.rect.Height-2)
		w.preview.Width = w.rect.Width - lw - 2
		w.preview.Height = w.rect.Height - 2
	} else {
		w.browser.SetSize(w.rect.Width, 5)
		w.preview.Width = w.rect.Width
		w.preview.Height = w.rect.Height - 8
		if w.preview.Height < 2 {
			w.preview.Height = 2
		}
	}
}

func (w *logsWindow) horizontal() bool { return w.rect.Width >= 90 }

func (w *logsWindow) Refresh(ctx Context) {
	w.dir = storage.LogsDir(ctx.DataRoot, ctx.ProjectName)
	names := rflog.List(w.dir)
	items := make([]list.Item, 0, len(names))
	for _, n := range names {
		label := n
		if st, err := os.Stat(filepath.Join(w.dir, n)); err == nil {
			label = fmt.Sprintf("%s  %s", n, humanBytes(st.Size()))
		}
		items = append(items, simpleItem{title: label, value: n, kind: "log"})
	}
	_ = w.browser.SetItems(items)
	w.loadPreview()
}

func (w *logsWindow) loadPreview() {
	item, ok := w.browser.SelectedItem().(simpleItem)
	if !ok {
		w.preview.SetContent("no logs yet — run `build` or `test`")
		w.current = ""
		return
	}
	if item.value == w.current {
		return
	}
	w.current = item.value
	lines, err := rflog.Tail(filepath.Join(w.dir, item.value), 300)
	if err != nil {
		w.preview.SetContent("cannot read log: " + err.Error())
	} else {
		w.preview.SetContent(strings.Join(lines, "\n"))
	}
	w.preview.GotoBottom()
}

func (w *logsWindow) Update(msg tea.Msg, ctx Context) tea.Cmd {
	if k, ok := msg.(tea.KeyMsg); ok && k.String() == "r" {
		w.current = ""
		w.Refresh(ctx)
		return nil
	}
	var cmd tea.Cmd
	w.browser, cmd = w.browser.Update(msg)
	w.preview, _ = w.preview.Update(msg)
	w.loadPreview()
	return cmd
}

func (w *logsWindow) View(_ Context) string {
	head := section("Persisted logs — "+w.dir, w.width()) + "\n"
	var body string
	if w.horizontal() {
		body = lipgloss.JoinHorizontal(lipgloss.Top,
			tui.Clamp(w.browser.View(), w.browser.Width(), w.rect.Height-1),
			" ", w.preview.View())
	} else {
		body = w.browser.View() + "\n" +
			tui.MutedStyle.Render(strings.Repeat("─", w.width())) + "\n" + w.preview.View()
	}
	return w.fit(head + body)
}

func humanBytes(n int64) string {
	switch {
	case n < 1024:
		return fmt.Sprintf("%dB", n)
	case n < 1024*1024:
		return fmt.Sprintf("%.0fK", float64(n)/1024)
	default:
		return fmt.Sprintf("%.1fM", float64(n)/(1024*1024))
	}
}
