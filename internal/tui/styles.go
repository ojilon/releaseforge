// Package tui holds Lip Gloss styles, shared Bubbles components and layout helpers.
package tui

import (
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// Styles, colors, and layout constants live here so views stay declarative.
var (
	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("12")).
			Background(lipgloss.Color("0")).
			Padding(0, 1)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("9")).
			Bold(true)

	BarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("11")).
			Padding(0, 1)

	HelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8"))
)

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

// Color tokens: the single place for panel colors.
const (
	ColorOK     = lipgloss.Color("10")
	ColorRun    = lipgloss.Color("11")
	ColorFail   = lipgloss.Color("9")
	ColorDim    = lipgloss.Color("8")
	ColorBorder = lipgloss.Color("8")
)

var (
	PillOK   = lipgloss.NewStyle().Bold(true).Foreground(ColorOK)
	PillRun  = lipgloss.NewStyle().Bold(true).Foreground(ColorRun)
	PillFail = lipgloss.NewStyle().Bold(true).Foreground(ColorFail)
	PillIdle = lipgloss.NewStyle().Bold(true).Foreground(ColorDim)

	RuleStyle = lipgloss.NewStyle().Foreground(ColorDim)

	PanelBorder = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorder).
			Padding(0, 1)
)

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
