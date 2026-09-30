// Package tui holds Lip Gloss styles, shared Bubbles components and layout helpers.
package tui

import (
	"github.com/charmbracelet/lipgloss"
)

// Styles, colors, and layout constants live here so views stay declarative.
var (
	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("12")).
			Background(lipgloss.Color("0")).
			Padding(0, 1)

	StatusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("10")).
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
