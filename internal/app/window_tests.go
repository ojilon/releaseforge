package app

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	rflog "github.com/ojilon/releaseforge/internal/log"
	"github.com/ojilon/releaseforge/internal/metrics"
	"github.com/ojilon/releaseforge/internal/storage"
	"github.com/ojilon/releaseforge/internal/tui"
)

// testsWindow is the test dashboard skeleton from docs/sweep-2mo/08: run
// history plus the tail of the last test log. It never fabricates results —
// everything comes from metrics.jsonl and persisted logs.
type testsWindow struct {
	windowBase
	table   table.Model
	preview viewport.Model
	head    string
	keys    []key.Binding
}

func newTestsWindow() *testsWindow {
	w := &testsWindow{}
	w.table = themedTable([]table.Column{
		{Title: "when", Width: 11}, {Title: "kind", Width: 8},
		{Title: "variant", Width: 12}, {Title: "result", Width: 8},
		{Title: "duration", Width: 0},
	}, nil, 60, 5)
	w.preview = viewport.New(60, 6)
	w.preview.MouseWheelEnabled = true
	w.keys = []key.Binding{
		key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "refresh")),
		key.NewBinding(key.WithKeys("up", "down"), key.WithHelp("↑↓", "runs")),
	}
	return w
}

func (w *testsWindow) ID() WindowID        { return WinTests }
func (w *testsWindow) Title() string       { return WindowTitles[WinTests] }
func (w *testsWindow) Keys() []key.Binding { return w.keys }
func (w *testsWindow) Init() tea.Cmd       { return nil }

func (w *testsWindow) Resize(l tui.Layout) {
	w.windowBase.Resize(l)
	w.table.SetWidth(w.mainWidth())
	w.preview.Width = w.mainWidth()
	h := w.rect.Height / 2
	if h < 3 {
		h = 3
	}
	w.table.SetHeight(h - 1)
	w.preview.Height = w.rect.Height - h - 3
	if w.preview.Height < 2 {
		w.preview.Height = 2
	}
}

func (w *testsWindow) Refresh(ctx Context) {
	var tr []table.Row
	for _, r := range metrics.Recent(ctx.DataRoot, ctx.ProjectName, 20) {
		if r.Kind != "test" && r.Kind != "build" {
			continue
		}
		result := "ok"
		if !r.Success {
			result = "FAILED"
		}
		tr = append(tr, table.Row{metrics.Ago(r.At), r.Kind, r.Variant, result,
			metrics.DurationText(r.DurationMs)})
	}
	w.table.SetRows(tr)
	w.table.SetColumns(fitColumns([]table.Column{
		{Title: "when", Width: 11}, {Title: "kind", Width: 8},
		{Title: "variant", Width: 12}, {Title: "result", Width: 8},
		{Title: "duration", Width: 0},
	}, w.mainWidth()))

	dir := storage.LogsDir(ctx.DataRoot, ctx.ProjectName)
	path, label := "", "no test log yet — run `test`"
	if name, err := rflog.Newest(dir); err == nil {
		// Prefer a test log when one exists, else show the newest log.
		for _, n := range rflog.List(dir) {
			if strings.HasPrefix(strings.ToLower(n), "test-") {
				name = n
				break
			}
		}
		p := filepath.Join(dir, name)
		if lines, err := rflog.Tail(p, 40); err == nil {
			label = p
			path = strings.Join(lines, "\n")
		} else if data, err := os.ReadFile(p); err == nil {
			label = p
			path = string(data)
		}
	}
	w.head = label
	w.preview.SetContent(path)
	w.preview.GotoBottom()
}

func (w *testsWindow) Update(msg tea.Msg, _ Context) tea.Cmd {
	var c1, c2 tea.Cmd
	w.table, c1 = w.table.Update(msg)
	w.preview, c2 = w.preview.Update(msg)
	return tea.Batch(c1, c2)
}

func (w *testsWindow) View(_ Context) string {
	var b strings.Builder
	b.WriteString(section("Run history", w.mainWidth()) + "\n")
	if len(w.table.Rows()) == 0 {
		b.WriteString(tui.HelpStyle.Render("  no runs recorded yet — `build debug` or `test`") + "\n")
	} else {
		b.WriteString(w.table.View() + "\n")
	}
	b.WriteString("\n" + section("Last log — "+tui.Truncate(w.head, w.mainWidth()-24), w.mainWidth()) + "\n")
	if strings.TrimSpace(w.preview.View()) == "" {
		b.WriteString(tui.HelpStyle.Render("  (empty)"))
	} else {
		b.WriteString(w.preview.View())
	}
	return w.fit(b.String())
}
