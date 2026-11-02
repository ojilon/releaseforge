package app

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ojilon/releaseforge/internal/git"
	"github.com/ojilon/releaseforge/internal/metrics"
	"github.com/ojilon/releaseforge/internal/project"
	"github.com/ojilon/releaseforge/internal/tui"
)

// statusWindow is the "know where you are" pane: project facts, git state,
// cached scan details and recent run history. It replaces the old fixed
// right-hand dashboard with a real full window (docs/sweep-2mo/07).
type statusWindow struct {
	windowBase
	table table.Model
	facts string
	card  string
	keys  []key.Binding
}

func newStatusWindow() *statusWindow {
	w := &statusWindow{}
	w.table = themedTable([]table.Column{
		{Title: "when", Width: 11},
		{Title: "kind", Width: 9},
		{Title: "variant", Width: 10},
		{Title: "result", Width: 8},
		{Title: "time", Width: 0},
	}, nil, 60, 5)
	w.keys = []key.Binding{
		key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "refresh")),
		key.NewBinding(key.WithKeys("up", "down"), key.WithHelp("↑↓", "runs")),
	}
	return w
}

func (w *statusWindow) ID() WindowID        { return WinStatus }
func (w *statusWindow) Title() string       { return WindowTitles[WinStatus] }
func (w *statusWindow) Keys() []key.Binding { return w.keys }
func (w *statusWindow) Init() tea.Cmd       { return nil }

func (w *statusWindow) Resize(l tui.Layout) {
	w.windowBase.Resize(l)
	w.table.SetWidth(w.mainWidth())
	h := w.rect.Height - 14
	if h < 2 {
		h = 2
	}
	if h > 7 {
		h = 7
	}
	w.table.SetHeight(h)
}

func (w *statusWindow) Refresh(ctx Context) {
	rows := [][2]string{
		{"project", ctx.ProjectName},
		{"root", ctx.projectRoot()},
		{"type", projectType(ctx)},
		{"version", orDash(ctx.Version)},
		{"data-root", ctx.DataRoot},
	}
	if ctx.Info != nil && git.IsRepo(ctx.Info.Root) {
		state := "clean"
		if ctx.Dirty {
			state = "● dirty"
		}
		rows = append(rows, [2]string{"git", ctx.Branch + " · " + state})
		if a, b, ok := git.AheadBehind(ctx.Info.Root); ok {
			rows = append(rows, [2]string{"upstream", fmt.Sprintf("ahead %d · behind %d", a, b)})
		}
		if o := git.Origin(ctx.Info.Root); o != "" {
			rows = append(rows, [2]string{"origin", o})
		}
	} else {
		rows = append(rows, [2]string{"git", "not a repository"})
	}
	rows = append(rows, [2]string{"last build",
		strings.TrimPrefix(metrics.LastBuildLine(ctx.DataRoot, ctx.ProjectName), "last build: ")})

	var b strings.Builder
	b.WriteString(section("Project", w.mainWidth()) + "\n")
	b.WriteString(tui.KeyValue(rows, w.mainWidth()) + "\n\n")

	if snap, err := project.LoadScan(ctx.DataRoot, ctx.ProjectName); err == nil {
		var scanRows [][2]string
		if snap.Git.Head != "" {
			scanRows = append(scanRows, [2]string{"head", shortSHA(snap.Git.Head)})
		}
		if len(snap.Git.RecentTags) > 0 {
			scanRows = append(scanRows, [2]string{"latest tag", snap.Git.RecentTags[0]})
		}
		if snap.Version.File != "" {
			scanRows = append(scanRows, [2]string{"version file", snap.Version.File})
		}
		var ids []string
		for i, t := range snap.Tools {
			if i >= 5 {
				break
			}
			ids = append(ids, t.ID)
		}
		if len(ids) > 0 {
			scanRows = append(scanRows, [2]string{"tools", strings.Join(ids, ", ")})
		}
		if len(scanRows) > 0 {
			b.WriteString(section("Cached scan", w.mainWidth()) + "\n")
			b.WriteString(tui.KeyValue(scanRows, w.mainWidth()) + "\n")
		}
	} else {
		b.WriteString(tui.HelpStyle.Render("no scan cache — `scan <path>` or ctrl+o") + "\n")
	}
	w.facts = b.String()

	// Side card (wide layouts only): live state at a glance.
	cardRows := [][2]string{
		{"phase", ctx.Phase.String()},
		{"theme", ctx.ThemeName},
		{"branch", orDash(ctx.Branch)},
		{"activity", orDash(ctx.Spark)},
		{"status", ctx.Status},
	}
	if ctx.Phase.Busy() {
		cardRows = append(cardRows, [2]string{"running", ctx.RunKind + " " + ctx.RunVariant})
		cardRows = append(cardRows, [2]string{"elapsed", fmt.Sprintf("%.1fs", ctx.Elapsed.Seconds())})
	}
	cardW := 32
	if w.side.Valid() {
		cardW = w.side.Width - 2 // frame borders
		if cardW < 12 {
			cardW = 12
		}
	}
	w.card = section("At a glance", cardW) + "\n" + tui.KeyValue(cardRows, cardW)

	var tr []table.Row
	for _, r := range metrics.Recent(ctx.DataRoot, ctx.ProjectName, 8) {
		result := "ok"
		if !r.Success {
			result = "FAILED"
		}
		tr = append(tr, table.Row{
			metrics.Ago(r.At), r.Kind, r.Variant, result, metrics.DurationText(r.DurationMs),
		})
	}
	w.table.SetRows(tr)
	w.table.SetColumns(fitColumns([]table.Column{
		{Title: "when", Width: 11},
		{Title: "kind", Width: 9},
		{Title: "variant", Width: 10},
		{Title: "result", Width: 8},
		{Title: "time", Width: 0},
	}, w.mainWidth()))
}

func (w *statusWindow) Update(msg tea.Msg, _ Context) tea.Cmd {
	var cmd tea.Cmd
	w.table, cmd = w.table.Update(msg)
	return cmd
}

func (w *statusWindow) View(_ Context) string {
	body := w.facts
	if len(w.table.Rows()) > 0 {
		body += "\n" + section("Recent runs", w.mainWidth()) + "\n" + w.table.View()
	} else {
		body += "\n" + tui.HelpStyle.Render("no runs yet — `build debug` or `test`")
	}
	if w.side.Valid() {
		card := tui.Frame(w.card, w.side.Width, w.side.Height, tui.FrameBorder)
		left := tui.Clamp(body, w.primary.Width, w.rect.Height)
		return w.fit(lipgloss.JoinHorizontal(lipgloss.Top, left, "  ", card))
	}
	return w.fit(body)
}

func projectType(ctx Context) string {
	if ctx.Info != nil {
		return ctx.Info.Type
	}
	return "unknown — scan a folder"
}

func orDash(s string) string {
	if strings.TrimSpace(s) == "" {
		return "—"
	}
	return s
}
