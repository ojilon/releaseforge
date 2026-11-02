package app

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/ojilon/releaseforge/internal/history"
	rflog "github.com/ojilon/releaseforge/internal/log"
	"github.com/ojilon/releaseforge/internal/storage"
	"github.com/ojilon/releaseforge/internal/tui"
)

// paletteModel is the ctrl+p / ctrl+r overlay: one query input, a
// bubbles/list of fuzzy-ranked candidates. Sources are commands, recent
// projects, command history, persisted logs and window shortcuts.
type paletteModel struct {
	query textinput.Model
	list  list.Model
	all   []simpleItem
	shown []simpleItem
	mode  string
}

func newPalette() *paletteModel {
	q := textinput.New()
	q.Placeholder = "search commands, projects, history, logs…"
	q.Prompt = "> "
	q.CharLimit = 120
	tui.PaintInput(&q)
	p := &paletteModel{query: q, list: themedList(nil, 50, 10)}
	return p
}

func (p *paletteModel) open(items []simpleItem, mode string) tea.Cmd {
	p.all = items
	p.mode = mode
	p.query.SetValue("")
	p.query.Focus()
	p.rank()
	return textinput.Blink
}

// rank applies the pure fuzzy scorer and preserves the choice -> value map.
func (p *paletteModel) rank() {
	q := strings.TrimSpace(p.query.Value())
	var picked []simpleItem
	if q == "" {
		picked = p.all
	} else {
		titles := make([]string, len(p.all))
		for i, it := range p.all {
			titles[i] = it.title
		}
		for _, mt := range tui.Rank(q, titles, 20) {
			picked = append(picked, p.all[mt.Index])
		}
	}
	if len(picked) > 20 {
		picked = picked[:20]
	}
	p.shown = picked
	items := make([]list.Item, len(picked))
	for i, it := range picked {
		items[i] = it
	}
	_ = p.list.SetItems(items)
	p.list.Select(0)
}

func (p *paletteModel) selected() (simpleItem, bool) {
	it, ok := p.list.SelectedItem().(simpleItem)
	return it, ok
}

func (p *paletteModel) Resize(w, h int) {
	p.list.SetSize(w, h)
	p.query.Width = w - 4
}

func (p *paletteModel) View(w int) string {
	title := tui.TitleStyle.Render(" command palette ")
	if p.mode == "history" {
		title = tui.TitleStyle.Render(" command history ")
	}
	count := fmt.Sprintf("%d matches", len(p.shown))
	return title + " " + tui.HelpStyle.Render(count) + "\n" +
		p.query.View() + "\n" +
		tui.Rule(w) + "\n" + p.list.View()
}

// paletteItems builds all candidates from live data. Never panics on missing
// files: the palette is best-effort by design.
func (m *Model) paletteItems() []simpleItem {
	var out []simpleItem
	for _, v := range tui.Verbs {
		kind := "fill"
		if v.Args == "" && len(v.Subs) == 0 {
			kind = "act"
		}
		out = append(out, simpleItem{
			title: v.Name + "  " + v.Help,
			value: v.Name + " ", kind: kind,
		})
	}
	for i, label := range WindowTitles {
		out = append(out, simpleItem{
			title: fmt.Sprintf("window %d  %s", i+1, label),
			value: fmt.Sprintf("window %d", i+1), kind: "act",
		})
	}
	if rf, err := history.LoadRecent(m.dataRoot); err == nil {
		for _, e := range rf.Items {
			out = append(out, simpleItem{
				title: "open " + e.Name + "  " + e.Path,
				value: "open " + quotePath(e.Path), kind: "act",
			})
		}
	}
	if m.info != nil {
		for _, n := range rflog.List(storage.LogsDir(m.dataRoot, m.info.Name)) {
			out = append(out, simpleItem{
				title: "logs " + n, value: "logs show " + n, kind: "act",
			})
		}
	}
	seen := map[string]bool{}
	for _, line := range m.history {
		line = strings.TrimSpace(line)
		if line == "" || seen[line] {
			continue
		}
		seen[line] = true
		out = append(out, simpleItem{title: "history  " + line, value: line, kind: "history"})
	}
	return out
}

func quotePath(p string) string {
	if strings.ContainsAny(p, " \t") {
		return `"` + strings.ReplaceAll(p, `"`, `\"`) + `"`
	}
	return p
}

// historyItems restricts the palette to previously typed commands.
func (m *Model) historyItems() []simpleItem {
	seen := map[string]bool{}
	var out []simpleItem
	for i := len(m.history) - 1; i >= 0; i-- {
		line := strings.TrimSpace(m.history[i])
		if line == "" || seen[line] {
			continue
		}
		seen[line] = true
		out = append(out, simpleItem{title: line, value: line, kind: "history"})
	}
	return out
}

// openPicker creates the bubbles/filepicker directory browser. This is the
// in-terminal folder picker the vision asked for; typed paths still work.
func (m *Model) openPicker() tea.Cmd {
	fp := filepicker.New()
	fp.DirAllowed = true
	fp.FileAllowed = false
	fp.ShowHidden = false
	fp.ShowSize = false
	fp.ShowPermissions = false
	fp.CurrentDirectory = m.projectDir
	fp.SetHeight(12)
	fp.Cursor = tui.TitleStyle.Render("▸ ")
	fp.Styles.Directory = tui.TitleStyle
	fp.Styles.Cursor = tui.TitleStyle
	fp.Styles.Selected = tui.TitleStyle
	fp.Styles.File = tui.HelpStyle
	fp.Styles.DisabledFile = tui.HelpStyle
	fp.Styles.EmptyDirectory = tui.HelpStyle
	m.picker = &fp
	m.focus = FocusPicker
	return fp.Init()
}
