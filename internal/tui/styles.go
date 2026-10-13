// Package tui holds the theme, Lip Gloss styles and layout helpers.
//
// Where things live:
//   theme.go  - Theme struct + Current colors (EDIT COLORS HERE)
//   styles.go - styles built FROM the theme (edit rarely)
package tui

import (
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// Styles below are built from Current theme in refreshStyles.
// Do not hardcode colors here — change theme.go instead.
var (
	HeaderStyle lipgloss.Style
	ErrorStyle  lipgloss.Style
	BarStyle    lipgloss.Style
	HelpStyle   lipgloss.Style

	// LogBoxStyle is the thin border around the log pane.
	LogBoxStyle lipgloss.Style
	// DashBoxStyle is the thin border around the right-side dashboard.
	DashBoxStyle lipgloss.Style

	PillOK   lipgloss.Style
	PillRun  lipgloss.Style
	PillFail lipgloss.Style
	PillIdle lipgloss.Style

	RuleStyle   lipgloss.Style
	PanelBorder lipgloss.Style
)

// Color tokens kept for callers that need a raw color.
// They mirror the current theme after refreshStyles runs.
var (
	ColorOK     lipgloss.Color
	ColorRun    lipgloss.Color
	ColorFail   lipgloss.Color
	ColorDim    lipgloss.Color
	ColorBorder lipgloss.Color
)

func init() {
	refreshStyles()
}

// refreshStyles rebuilds every style from Current.
// Called once at startup and on every SetTheme.
func refreshStyles() {
	HeaderStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(Current.HeaderFg).
		Background(Current.HeaderBg).
		Padding(0, 1)

	ErrorStyle = lipgloss.NewStyle().
		Foreground(Current.Fail).
		Bold(true)

	// BarStyle paints the input box line (see view.go: bar).
	// Only a thick LEFT border in accent color — the rest stays borderless
	// so the color stays small and never overwhelms the text.
	BarStyle = lipgloss.NewStyle().
		Foreground(Current.Text).
		Border(lipgloss.ThickBorder(), false, false, false, true).
		BorderForeground(Current.Accent).
		Padding(0, 1)

	// LogBoxStyle is the thin border around the logs pane.
	LogBoxStyle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(Current.Border)

	// DashBoxStyle is the thin border around the scan dashboard.
	DashBoxStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Current.Border).
		Padding(0, 1)

	HelpStyle = lipgloss.NewStyle().
		Foreground(Current.Dim)

	ColorOK = Current.OK
	ColorRun = Current.Busy
	ColorFail = Current.Fail
	ColorDim = Current.Dim
	ColorBorder = Current.Border

	PillOK = lipgloss.NewStyle().Bold(true).Foreground(Current.OK)
	PillRun = lipgloss.NewStyle().Bold(true).Foreground(Current.Busy)
	PillFail = lipgloss.NewStyle().Bold(true).Foreground(Current.Fail)
	PillIdle = lipgloss.NewStyle().Bold(true).Foreground(Current.Dim)

	RuleStyle = lipgloss.NewStyle().Foreground(Current.Dim)

	// PanelBorder is the rounded box around the header.
	// Overall app background comes from your terminal;
	// this border + header background is what you actually see.
	PanelBorder = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Current.Border).
		Padding(0, 1)
}

// RenderHeader builds the status header line.
func RenderHeader(toolVersion, project, version, status string) string {
	line := "ReleaseForge"
	if toolVersion != "" {
		line += " " + toolVersion
	}
	if project != "" {
		line += " · " + project
	}
	if version != "" {
		line += " · " + version
	}
	if status != "" {
		line += " · " + status
	}
	return HeaderStyle.Render(line)
}

// Pill renders a state pill for a phase (idle, scanning, building, testing,
// failed, ok).
func Pill(phase string) string {
	switch phase {
	case "ok":
		return PillOK.Render("● ok")
	case "failed":
		return PillFail.Render("● failed")
	case "building", "testing", "scanning":
		return PillRun.Render("● " + phase)
	default:
		return PillIdle.Render("● idle")
	}
}

// Rule renders a faint full-width separator.
func Rule(width int) string {
	if width < 4 {
		width = 4
	}
	out := ""
	for i := 0; i < width-2; i++ {
		out += "─"
	}
	return RuleStyle.Render(out)
}

var sparkBlocks = []rune(" ▁▂▃▄▅▆▇█")

// RenderReport formats structured build error lines cleanly:
// summary in fail color, error lines normal, hint: lines dimmed.
func RenderReport(summary string, lines []string) string {
	var b strings.Builder
	if strings.TrimSpace(summary) != "" {
		b.WriteString(ErrorStyle.Render("✗ " + strings.TrimSpace(summary)))
		b.WriteString("\n")
	}
	for _, l := range lines {
		t := strings.TrimRight(l, "\n")
		if strings.Contains(t, "hint:") {
			idx := strings.Index(t, "hint:")
			b.WriteString(t[:idx])
			b.WriteString(HelpStyle.Render(t[idx:]))
		} else {
			b.WriteString(t)
		}
		b.WriteString("\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

// Sparkline renders 14 daily commit counts as bars. All-zero yields "" so
// callers can hide the line. Longer inputs use the last 14.
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
