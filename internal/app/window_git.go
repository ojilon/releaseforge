package app

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/ojilon/releaseforge/internal/git"
	"github.com/ojilon/releaseforge/internal/tui"
)

type gitView uint8

const (
	gitStatus gitView = iota
	gitDiff
	gitLog
)

// gitWindow is the diff/commit dashboard from docs/sweep-2mo/09. Three
// subviews share the window; each keeps its own cursor/scroll offset.
type gitWindow struct {
	windowBase
	view    gitView
	table   table.Model
	detail  viewport.Model
	summary string
	keys    []key.Binding
}

func newGitWindow() *gitWindow {
	w := &gitWindow{view: gitStatus}
	w.table = themedTable([]table.Column{
		{Title: "sha", Width: 9}, {Title: "date", Width: 12},
		{Title: "subject", Width: 0},
	}, nil, 60, 5)
	w.detail = viewport.New(60, 8)
	w.detail.MouseWheelEnabled = true
	w.keys = []key.Binding{
		key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "status")),
		key.NewBinding(key.WithKeys("d"), key.WithHelp("d", "diff")),
		key.NewBinding(key.WithKeys("l"), key.WithHelp("l", "log")),
		key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "refresh")),
		key.NewBinding(key.WithKeys("up", "down"), key.WithHelp("↑↓", "navigate")),
	}
	return w
}

func (w *gitWindow) ID() WindowID        { return WinGit }
func (w *gitWindow) Title() string       { return WindowTitles[WinGit] }
func (w *gitWindow) Keys() []key.Binding { return w.keys }
func (w *gitWindow) Init() tea.Cmd       { return nil }

func (w *gitWindow) Resize(l tui.Layout) {
	w.windowBase.Resize(l)
	w.table.SetWidth(w.mainWidth())
	w.detail.Width = w.mainWidth()
	h := w.rect.Height - 5
	if h < 3 {
		h = 3
	}
	w.table.SetHeight(h)
	w.detail.Height = h
}

func (w *gitWindow) modeLabel() string {
	switch w.view {
	case gitDiff:
		return "diff"
	case gitLog:
		return "log"
	default:
		return "status"
	}
}

func (w *gitWindow) Refresh(ctx Context) {
	root := ctx.projectRoot()
	if !git.IsRepo(root) {
		w.summary = "not a git repository — " + root
		w.detail.SetContent("scan a folder that contains a .git directory")
		w.table.SetRows(nil)
		return
	}
	branch := git.CurrentBranch(root)
	state := "clean"
	if !git.IsClean(root) {
		state = "dirty"
	}
	w.summary = fmt.Sprintf("%s · %s", branch, state)
	if a, b, ok := git.AheadBehind(root); ok {
		w.summary += fmt.Sprintf(" · ↑%d ↓%d", a, b)
	}
	if tag := git.LatestTag(root); tag != "" {
		w.summary += " · tag " + tag
	}

	// Commit table (log subview).
	var rows []table.Row
	if commits, err := git.Log(root, 40); err == nil {
		for _, c := range commits {
			rows = append(rows, table.Row{c.SHA, c.Date, c.Subject})
		}
	}
	w.table.SetRows(rows)
	w.table.SetColumns(fitColumns([]table.Column{
		{Title: "sha", Width: 9}, {Title: "date", Width: 12},
		{Title: "subject", Width: 0},
	}, w.mainWidth()))

	// Detail text (status / diff subviews).
	var b strings.Builder
	switch w.view {
	case gitDiff:
		stat, err := git.DiffStat(root)
		if err != nil {
			b.WriteString("git diff unavailable: " + err.Error())
		} else if strings.TrimSpace(stat) == "" {
			b.WriteString("no unstaged changes")
		} else {
			b.WriteString(stat)
		}
	default:
		changes, err := git.StatusPorcelain(root)
		if err != nil {
			b.WriteString("git status unavailable: " + err.Error())
		} else if len(changes) == 0 {
			b.WriteString("working tree clean")
		} else {
			for _, l := range changes {
				b.WriteString(l + "\n")
			}
		}
		if tags := git.RecentTags(root, 5); len(tags) > 0 {
			b.WriteString("\nrecent tags: " + strings.Join(tags, ", "))
		}
	}
	w.detail.SetContent(strings.TrimRight(b.String(), "\n"))
	w.detail.GotoTop()
}

func (w *gitWindow) Update(msg tea.Msg, ctx Context) tea.Cmd {
	if k, ok := msg.(tea.KeyMsg); ok {
		switch k.String() {
		case "s":
			w.view = gitStatus
			w.Refresh(ctx)
			return nil
		case "d":
			w.view = gitDiff
			w.Refresh(ctx)
			return nil
		case "l":
			w.view = gitLog
			w.Refresh(ctx)
			return nil
		case "r":
			w.Refresh(ctx)
			return nil
		}
	}
	var c1, c2 tea.Cmd
	w.table, c1 = w.table.Update(msg)
	w.detail, c2 = w.detail.Update(msg)
	return tea.Batch(c1, c2)
}

func (w *gitWindow) View(_ Context) string {
	var b strings.Builder
	b.WriteString(section("Git "+w.modeLabel()+" — "+w.summary, w.mainWidth()) + "\n")
	if w.view == gitLog {
		if len(w.table.Rows()) == 0 {
			b.WriteString(tui.HelpStyle.Render("  no commits"))
		} else {
			b.WriteString(w.table.View())
		}
	} else {
		b.WriteString(w.detail.View())
	}
	b.WriteString("\n\n" + tui.HelpStyle.Render("s status · d diff · l log · r refresh"))
	return w.fit(b.String())
}
