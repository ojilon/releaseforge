package tui

import (
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

// Theme is the single place for all app colors.
// Beginner rule: to recolor the app, edit only DarkTheme below
// (or call SetTheme("light")). Nothing else needs to change.
//
// Fields are plain lipgloss colors so you can use names ("red")
// or hex ("#7aa2f7"). Hex looks much better than 0-15 numbers.
type Theme struct {
	Name string

	HeaderFg lipgloss.Color // header text
	HeaderBg lipgloss.Color // header background
	Text     lipgloss.Color // normal text
	Dim      lipgloss.Color // help, hints, separators
	Border   lipgloss.Color // panel borders
	Accent   lipgloss.Color // input box prompt + bar

	OK   lipgloss.Color // success pill
	Busy lipgloss.Color // building/testing pill
	Fail lipgloss.Color // failed pill + errors
}

// DarkTheme is the default. Edit these hex values to taste.
// Accent is a soft light-blue used sparingly (input prompt, left border).
// Keep it small: only chrome uses it, never body text.
var DarkTheme = Theme{
	Name: "dark",

	HeaderFg: "#c0caf5",
	HeaderBg: "#1a1b26",
	Text:     "#c0caf5",
	Dim:      "#565f89",
	Border:   "#3b4261",
	Accent:   "#8aadf4",

	OK:   "#9ece6a",
	Busy: "#e0af68",
	Fail: "#f7768e",
}

// LightTheme is a second preset so you can see how SetTheme works.
var LightTheme = Theme{
	Name: "light",

	HeaderFg: "#1a1b26",
	HeaderBg: "#e9e9ed",
	Text:     "#1a1b26",
	Dim:      "#848cb0",
	Border:   "#b4b7d0",
	Accent:   "#34548a",

	OK:   "#2e7d32",
	Busy: "#9a6a00",
	Fail: "#c62828",
}

// Current is the active theme. Change it only via SetTheme
// so all styles rebuild together.
var Current = DarkTheme

// SetTheme switches theme by name ("dark" or "light").
// Returns false when the name is unknown.
func SetTheme(name string) bool {
	switch name {
	case "dark":
		Current = DarkTheme
	case "light":
		Current = LightTheme
	default:
		return false
	}
	refreshStyles()
	return true
}

// PaintInput colors a bubbles textinput from the current theme.
// Called once in app/model.go New(). Edit here to restyle the input box.
func PaintInput(ti *textinput.Model) {
	ti.PromptStyle = lipgloss.NewStyle().Foreground(Current.Accent).Bold(true)
	ti.TextStyle = lipgloss.NewStyle().Foreground(Current.Text)
	ti.PlaceholderStyle = lipgloss.NewStyle().Foreground(Current.Dim)
	ti.CursorStyle = lipgloss.NewStyle().Foreground(Current.Accent)
}
