module github.com/ojilon/releaseforge

go 1.22

require (
	github.com/charmbracelet/bubbles v0.20.0
	github.com/charmbracelet/bubbletea v1.2.4
	github.com/charmbracelet/huh v0.6.0
	github.com/charmbracelet/lipgloss v1.0.0
	github.com/spf13/cobra v1.8.1
)

// Additional indirects will appear after `go mod tidy`.
// Planned later: github.com/charmbracelet/glamour, go-git, etc.
