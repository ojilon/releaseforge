package tui

import (
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// Styles below are rebuilt from Current by refreshStyles(). Do not hardcode
// colors anywhere else; edit theme.go instead.
var (
	HeaderStyle lipgloss.Style
	ErrorStyle  lipgloss.Style
	BarStyle    lipgloss.Style
	HelpStyle   lipgloss.Style

	// Legacy aliases, kept for the small helpers that still use them.
	LogBoxStyle  lipgloss.Style
	DashBoxStyle lipgloss.Style

	TitleStyle    lipgloss.Style
	SubtitleStyle lipgloss.Style
	KeyStyle      lipgloss.Style
	ValueStyle    lipgloss.Style
	SuccessStyle  lipgloss.Style
	WarnStyle     lipgloss.Style
	MutedStyle    lipgloss.Style
	ChipStyle     lipgloss.Style

	TabActive   lipgloss.Style
	TabInactive lipgloss.Style
	TabKey      lipgloss.Style
	TabStrip    lipgloss.Style

	FrameBorder       lipgloss.Style
	FrameBorderActive lipgloss.Style
	BarBorder         lipgloss.Style
	OverlayBorder     lipgloss.Style

	PillOK   lipgloss.Style
	PillRun  lipgloss.Style
	PillFail lipgloss.Style
	PillIdle lipgloss.Style

	RuleStyle   lipgloss.Style
	PanelBorder lipgloss.Style

	SpinnerStyle lipgloss.Style
	TableHeader  lipgloss.Style
	TableCell    lipgloss.Style
	TableSel     lipgloss.Style
	ListSel      lipgloss.Style
)

// Color tokens mirror Current after refreshStyles for callers that need a
// raw color (e.g. bubbles progress gradients).
var (
	ColorOK      lipgloss.Color
	ColorRun     lipgloss.Color
	ColorFail    lipgloss.Color
	ColorDim     lipgloss.Color
	ColorBorder  lipgloss.Color
	ColorAccent  lipgloss.Color
	ColorSurface lipgloss.TerminalColor // nil in mono theme
)

func init() { refreshStyles() }

// refreshStyles rebuilds every style from Current. Called at package init and
// on every SetTheme.
func refreshStyles() {
	HeaderStyle = lipgloss.NewStyle().Bold(true).
		Foreground(fg(Current.HeaderFg)).Background(bg(Current.HeaderBg)).Padding(0, 1)

	ErrorStyle = lipgloss.NewStyle().Foreground(fg(Current.Fail)).Bold(true)
	SuccessStyle = lipgloss.NewStyle().Foreground(fg(Current.OK)).Bold(true)
	WarnStyle = lipgloss.NewStyle().Foreground(fg(Current.Busy)).Bold(true)
	MutedStyle = lipgloss.NewStyle().Foreground(fg(Current.Dim))
	HelpStyle = lipgloss.NewStyle().Foreground(fg(Current.Dim))
	TitleStyle = lipgloss.NewStyle().Foreground(fg(Current.Accent)).Bold(true)
	SubtitleStyle = lipgloss.NewStyle().Foreground(fg(Current.Dim)).Bold(true)
	KeyStyle = lipgloss.NewStyle().Foreground(fg(Current.Dim))
	ValueStyle = lipgloss.NewStyle().Foreground(fg(Current.Text))
	ChipStyle = lipgloss.NewStyle().Foreground(fg(Current.Accent)).
		Background(bg(Current.Surface)).Padding(0, 1)

	// BarStyle is the inner command-bar line; BarBorder wraps it in a frame.
	BarStyle = lipgloss.NewStyle().Foreground(fg(Current.Text))
	BarBorder = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).
		BorderForeground(fg(Current.Accent))

	LogBoxStyle = lipgloss.NewStyle().Border(lipgloss.NormalBorder()).
		BorderForeground(fg(Current.Border))
	DashBoxStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).
		BorderForeground(fg(Current.Border)).Padding(0, 1)

	FrameBorder = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).
		BorderForeground(fg(Current.Border))
	FrameBorderActive = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).
		BorderForeground(fg(Current.Accent))
	OverlayBorder = lipgloss.NewStyle().Border(lipgloss.DoubleBorder()).
		BorderForeground(fg(Current.Accent))
	PanelBorder = FrameBorder

	TabActive = lipgloss.NewStyle().Foreground(fg(Current.Text)).
		Background(bg(Current.Surface)).Bold(true).Padding(0, 1)
	TabInactive = lipgloss.NewStyle().Foreground(fg(Current.Dim)).Padding(0, 1)
	TabKey = lipgloss.NewStyle().Foreground(fg(Current.Accent)).Bold(true)
	TabStrip = lipgloss.NewStyle().Foreground(fg(Current.Dim))

	ColorOK, ColorRun, ColorFail = Current.OK, Current.Busy, Current.Fail
	ColorDim, ColorBorder, ColorAccent = Current.Dim, Current.Border, Current.Accent
	ColorSurface = bg(Current.Surface)

	PillOK = lipgloss.NewStyle().Bold(true).Foreground(fg(Current.OK))
	PillRun = lipgloss.NewStyle().Bold(true).Foreground(fg(Current.Busy))
	PillFail = lipgloss.NewStyle().Bold(true).Foreground(fg(Current.Fail))
	PillIdle = lipgloss.NewStyle().Bold(true).Foreground(fg(Current.Dim))
	RuleStyle = lipgloss.NewStyle().Foreground(fg(Current.Dim))
	SpinnerStyle = lipgloss.NewStyle().Foreground(fg(Current.Busy))

	TableHeader = lipgloss.NewStyle().Bold(true).Foreground(fg(Current.Accent)).Padding(0, 1)
	TableCell = lipgloss.NewStyle().Foreground(fg(Current.Text)).Padding(0, 1)
	TableSel = lipgloss.NewStyle().Bold(true).Foreground(fg(Current.Text)).
		Background(bg(Current.Surface)).Padding(0, 1)
	ListSel = lipgloss.NewStyle().Foreground(fg(Current.Accent)).Bold(true)
}

// RenderHeader builds the status header line (tool · project · version ·
// status). Kept as a standalone helper so the compact layout can reuse it.
func RenderHeader(toolVersion, project, version, status string) string {
	parts := []string{"ReleaseForge"}
	if toolVersion != "" {
		parts[0] += " " + toolVersion
	}
	if project != "" {
		parts = append(parts, project)
	}
	if version != "" {
		parts = append(parts, version)
	}
	if status != "" {
		parts = append(parts, status)
	}
	return HeaderStyle.Render(strings.Join(parts, " · "))
}

// Pill renders a state pill for a phase name (idle, building, testing,
// failed, ok, cancelled).
func Pill(phase string) string {
	switch phase {
	case "ok":
		return PillOK.Render("● ok")
	case "failed":
		return PillFail.Render("● failed")
	case "cancelled":
		return WarnStyle.Render("● cancelled")
	case "building", "testing", "scanning":
		return PillRun.Render("● " + phase)
	default:
		return PillIdle.Render("● idle")
	}
}

// Rule renders a faint separator of exactly width cells.
func Rule(width int) string {
	if width < 4 {
		width = 4
	}
	return RuleStyle.Render(strings.Repeat("─", width-2))
}

var sparkBlocks = []rune(" ▁▂▃▄▅▆▇█")

// RenderReport formats structured build error lines: summary in fail color,
// error lines normal, "hint:" tails dimmed.
func RenderReport(summary string, lines []string) string {
	var b strings.Builder
	if strings.TrimSpace(summary) != "" {
		b.WriteString(ErrorStyle.Render("✗ " + strings.TrimSpace(summary)))
		b.WriteString("\n")
	}
	for _, l := range lines {
		t := strings.TrimRight(l, "\n")
		if idx := strings.Index(t, "hint:"); idx >= 0 {
			b.WriteString(t[:idx])
			b.WriteString(HelpStyle.Render(t[idx:]))
		} else {
			b.WriteString(t)
		}
		b.WriteString("\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

// Sparkline renders up to 14 daily commit counts as bars. All-zero yields ""
// so callers can hide the line.
func Sparkline(counts []int) string {
	if len(counts) > 14 {
		counts = counts[len(counts)-14:]
	}
	max := 0
	for _, c := range counts {
		if c > max {
			max = c
		}
	}
	if max == 0 {
		return ""
	}
	var b strings.Builder
	for _, c := range counts {
		idx := (c * 8) / max
		if idx > 8 {
			idx = 8
		}
		b.WriteRune(sparkBlocks[idx])
	}
	return b.String()
}

// BucketCommits buckets YYYY-MM-DD dates into the last 14 days ending today.
func BucketCommits(dates []string, today time.Time) []int {
	counts := make([]int, 14)
	base := today.Truncate(24 * time.Hour)
	for _, d := range dates {
		t, err := time.Parse("2006-01-02", strings.TrimSpace(d))
		if err != nil {
			continue
		}
		delta := int(base.Sub(t.Truncate(24*time.Hour)).Hours() / 24)
		if delta >= 0 && delta < 14 {
			counts[13-delta]++
		}
	}
	return counts
}

// Truncate clips s to width display cells, ANSI-aware and wide-rune safe.
// Unlike byte slicing it never splits a CJK glyph or an emoji.
func Truncate(s string, width int) string {
	if width <= 0 {
		return ""
	}
	if ansi.StringWidth(s) <= width {
		return s
	}
	return ansi.Truncate(s, width, "…")
}

// Clamp forces s to exactly width x height cells: long lines are truncated
// (never wrapped), short lines padded, missing lines filled with blanks.
// This is the last line of defense against layout overflow: every window
// renders through it.
func Clamp(s string, width, height int) string {
	if width < 1 {
		width = 1
	}
	if height < 1 {
		height = 1
	}
	lines := strings.Split(strings.ReplaceAll(s, "\r\n", "\n"), "\n")
	out := make([]string, 0, height)
	for i := 0; i < height; i++ {
		line := ""
		if i < len(lines) {
			line = lines[i]
		}
		w := ansi.StringWidth(line)
		if w > width {
			line = Truncate(line, width)
			w = ansi.StringWidth(line)
		}
		if w < width {
			line += strings.Repeat(" ", width-w)
		}
		out = append(out, line)
	}
	return strings.Join(out, "\n")
}

// Frame draws a bordered box of exactly w x h cells around already-clamped
// content. Which style is used decides the border color.
func Frame(content string, w, h int, style lipgloss.Style) string {
	iw, ih := w-2, h-2
	if iw < 1 {
		iw = 1
	}
	if ih < 1 {
		ih = 1
	}
	return style.Copy().Width(iw).Height(ih).Render(Clamp(content, iw, ih))
}

// Tabs renders the window tab strip. labels are numbered from 1; active is
// zero-based. compact collapses labels to numbers plus the active title.
func Tabs(labels []string, active int, width int, compact bool) string {
	var b strings.Builder
	for i, label := range labels {
		num := strconv.Itoa(i + 1)
		if compact {
			if i == active {
				b.WriteString(TabActive.Render(num + " " + label))
			} else {
				b.WriteString(TabInactive.Render(num))
			}
		} else {
			cell := TabKey.Render(num) + " " + label
			if i == active {
				b.WriteString(TabActive.Render(num + " " + label))
			} else {
				b.WriteString(TabInactive.Render(cell))
			}
		}
		if i < len(labels)-1 {
			b.WriteString(" ")
		}
	}
	return Truncate(b.String(), width)
}

// KeyValue renders aligned "key  value" rows into width cells.
func KeyValue(rows [][2]string, width int) string {
	keyW := 0
	for _, r := range rows {
		if w := ansi.StringWidth(r[0]); w > keyW {
			keyW = w
		}
	}
	var b strings.Builder
	for i, r := range rows {
		key := KeyStyle.Render(r[0])
		pad := keyW - ansi.StringWidth(r[0]) + 2
		line := key + strings.Repeat(" ", pad) + ValueStyle.Render(Truncate(r[1], maxInt(1, width-keyW-pad)))
		b.WriteString(Truncate(line, width))
		if i < len(rows)-1 {
			b.WriteString("\n")
		}
	}
	return b.String()
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
