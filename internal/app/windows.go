package app

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ojilon/releaseforge/internal/project"
	"github.com/ojilon/releaseforge/internal/tui"
)

// WindowID identifies one of the six shell windows.
type WindowID uint8

const (
	WinStatus WindowID = iota
	WinTests
	WinGit
	WinBuild
	WinLogs
	WinHelp
	windowCount
)

// WindowTitles are also the tab labels; 1..6 map to this order.
var WindowTitles = []string{"Status", "Tests", "Git", "Console", "Logs", "Help"}

// Window is a focused pane. Only the active window receives keys and mouse
// messages; each window owns its own viewport/cursor/offset so switching
// never loses scroll position (docs/sweep-2mo/03, 07, 41).
type Window interface {
	ID() WindowID
	Title() string
	Keys() []key.Binding
	Resize(tui.Layout)
	Refresh(Context)
	Update(tea.Msg, Context) tea.Cmd
	View(Context) string
	Init() tea.Cmd
}

// Context is an immutable snapshot of application state handed to windows.
// Windows never reach back into Model, which keeps rendering pure and makes
// snapshot tests trivial.
type Context struct {
	DataRoot    string
	ProjectDir  string
	ProjectName string
	Info        *project.Info
	Version     string
	Branch      string
	Dirty       bool
	Spark       string
	Status      string
	ThemeName   string
	Phase       Phase
	RunKind     string
	RunVariant  string
	Elapsed     time.Duration
	Layout      tui.Layout
}

func (c Context) projectRoot() string {
	if c.Info != nil {
		return c.Info.Root
	}
	return c.ProjectDir
}

// windowBase supplies sizing + final clamping for every window. primary and
// side are the two columns of a wide layout; side is zero-sized when the
// terminal is too narrow.
type windowBase struct {
	layout  tui.Layout
	rect    tui.Rect
	primary tui.Rect
	side    tui.Rect
}

func (b *windowBase) Resize(l tui.Layout) {
	b.layout = l
	b.rect = l.Content
	b.primary, b.side = l.Split()
}

// fit guarantees the window never exceeds its rect, whatever the data does.
func (b *windowBase) fit(s string) string {
	return tui.Clamp(s, b.rect.Width, b.rect.Height)
}

func (b *windowBase) mainWidth() int {
	if b.primary.Valid() {
		return b.primary.Width
	}
	return b.rect.Width
}

func (b *windowBase) width() int  { return b.rect.Width }
func (b *windowBase) height() int { return b.rect.Height }

// section renders a small titled band used inside windows.
func section(title string, width int) string {
	// Keep the caller's casing: titles often contain paths, SHAs and branch
	// names that must not be uppercased.
	label := tui.SubtitleStyle.Render(" " + title + " ")
	fill := width - lipgloss.Width(label) - 1
	if fill < 0 {
		fill = 0
	}
	return label + tui.MutedStyle.Render(strings.Repeat("─", fill))
}

// emptyState is a friendly placeholder for windows with no data yet.
func emptyState(title, hint string, width int) string {
	return tui.TitleStyle.Render(title) + "\n\n" + tui.HelpStyle.Render(hint)
}

// themedTable builds a bubbles/table.Model using the current theme.
func themedTable(cols []table.Column, rows []table.Row, w, h int) table.Model {
	t := table.New(
		table.WithColumns(cols),
		table.WithRows(rows),
		table.WithWidth(w),
		table.WithHeight(h),
		table.WithFocused(true),
	)
	t.SetStyles(table.Styles{
		Header:   tui.TableHeader,
		Cell:     tui.TableCell,
		Selected: tui.TableSel,
	})
	return t
}

// themedList builds a one-line-per-row bubbles/list.Model for overlays and
// browsers. Filtering/pagination chrome is disabled: the palette owns its
// own query input and ranking.
func themedList(items []list.Item, w, h int) list.Model {
	d := list.NewDefaultDelegate()
	d.ShowDescription = false
	d.SetSpacing(0)
	d.SetHeight(1)
	d.Styles.NormalTitle = lipgloss.NewStyle().Foreground(tui.ColorAccent).Padding(0, 1)
	d.Styles.SelectedTitle = lipgloss.NewStyle().
		Foreground(tui.ColorAccent).Bold(true).
		Background(tui.ColorSurface).Padding(0, 1)
	l := list.New(items, d, w, h)
	l.SetFilteringEnabled(false)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetShowPagination(false)
	l.SetShowHelp(false)
	return l
}

// fitColumns distributes a table's remaining width across its columns.
func fitColumns(cols []table.Column, width int) []table.Column {
	fixed := 0
	flex := -1
	for i, c := range cols {
		if c.Width <= 0 {
			flex = i
		} else {
			fixed += c.Width
		}
	}
	if flex >= 0 {
		// table.Model renders each cell with Padding(0,1), so every column
		// costs 2 extra cells on top of its declared width.
		remaining := width - fixed - 2*len(cols) - 2
		if remaining < 8 {
			remaining = 8
		}
		cols[flex].Width = remaining
	}
	return cols
}

// simpleItem is a one-line bubbles/list row used by the Logs browser and the
// command palette. value/kind carry the action without a second type.
type simpleItem struct {
	title string
	desc  string
	value string
	kind  string
}

func (i simpleItem) Title() string       { return i.title }
func (i simpleItem) Description() string { return i.desc }
func (i simpleItem) FilterValue() string { return i.title }

func windowByID(windows []Window, id WindowID) Window {
	if int(id) < len(windows) {
		return windows[id]
	}
	return nil
}
