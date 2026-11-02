// Package tui holds the theme, Lip Gloss styles, layout helpers, key
// bindings, the command-intent table and the fuzzy scorer.
//
// It is deliberately free of application state: everything here is a pure
// presentation primitive that the Bubble Tea program (internal/app) and the
// Cobra commands (cmd/) can share.
//
//	theme.go  - color tokens, presets, detection, textinput painting
//	styles.go - styles and render helpers built FROM the theme
//	layout.go - responsive shell geometry + runewidth-safe clamping
//	keys.go   - bubbles/key bindings + bubbles/help footer
//	intent.go - one verb table driving help, completion and dispatch
//	fuzzy.go  - tiny subsequence scorer for the command palette
package tui

import (
	"os"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
)

// Theme is the single place for all app colors.
//
// Every style in styles.go is derived from Current by refreshStyles(). To
// recolor the app, edit a preset below or add a new one and register it in
// SetTheme. Nothing in internal/app hardcodes a color.
type Theme struct {
	Name string

	HeaderFg lipgloss.Color
	HeaderBg lipgloss.Color
	Text     lipgloss.Color
	Dim      lipgloss.Color
	Border   lipgloss.Color
	Accent   lipgloss.Color
	Surface  lipgloss.Color // selected rows, active tab, chips

	OK   lipgloss.Color
	Busy lipgloss.Color
	Fail lipgloss.Color

	// Mono disables all color output (NO_COLOR). Bold/underline still work.
	Mono bool
}

// DarkTheme is the default Tokyo-night-ish palette.
var DarkTheme = Theme{
	Name: "dark",

	HeaderFg: "#c0caf5",
	HeaderBg: "#1a1b26",
	Text:     "#c0caf5",
	Dim:      "#565f89",
	Border:   "#3b4261",
	Accent:   "#8aadf4",
	Surface:  "#24283b",

	OK:   "#9ece6a",
	Busy: "#e0af68",
	Fail: "#f7768e",
}

// LightTheme is tuned for terminals with a light background.
var LightTheme = Theme{
	Name: "light",

	HeaderFg: "#1a1b26",
	HeaderBg: "#e9e9ed",
	Text:     "#1a1b26",
	Dim:      "#72769a",
	Border:   "#b4b7d0",
	Accent:   "#34548a",
	Surface:  "#dfe3f2",

	OK:   "#2e7d32",
	Busy: "#9a6a00",
	Fail: "#c62828",
}

// MonoTheme emits no color at all. Used for NO_COLOR, dumb terminals, pipes
// and the ASCII snapshot tests.
var MonoTheme = Theme{Name: "mono", Mono: true}

// themes is the registry used by SetTheme / ThemeNames.
var themes = map[string]Theme{
	"dark":  DarkTheme,
	"light": LightTheme,
	"mono":  MonoTheme,
}

// Current is the active theme. Change it only through SetTheme so every
// derived style, key hint and input color rebuilds together.
var Current = DarkTheme

// ThemeNames returns the sorted-by-declaration-order list of preset names.
func ThemeNames() []string { return []string{"dark", "light", "mono"} }

// IsTheme reports whether name is a registered preset.
func IsTheme(name string) bool {
	_, ok := themes[strings.ToLower(strings.TrimSpace(name))]
	return ok
}

// SetTheme switches the active theme by name and rebuilds all derived
// styles. It returns false for unknown names (callers must surface the error
// instead of silently ignoring it — see docs/CORRECTIONS.md).
func SetTheme(name string) bool {
	t, ok := themes[strings.ToLower(strings.TrimSpace(name))]
	if !ok {
		return false
	}
	Current = t
	refreshStyles()
	return true
}

// DetectTheme picks a sensible preset from the environment without ever
// blocking on a terminal query. Precedence:
//
//	RELEASEFORGE_THEME (dark|light|mono) > NO_COLOR > COLORFGBG >
//	terminal background (only when stdout is a TTY) > dark
//
// hasDarkBackground is injected so tests can exercise both branches without
// a terminal. Pass nil to use lipgloss.HasDarkBackground.
func DetectTheme(getenv func(string) string, isTTY bool, hasDarkBackground func() bool) string {
	if getenv == nil {
		getenv = os.Getenv
	}
	if v := strings.ToLower(strings.TrimSpace(getenv("RELEASEFORGE_THEME"))); IsTheme(v) {
		return v
	}
	if getenv("NO_COLOR") != "" || getenv("TERM") == "dumb" {
		return "mono"
	}
	// COLORFGBG is "fg;bg"; a light background is a high bg index (7..15)
	// or the literal "default" pair used by some terminals.
	if v := getenv("COLORFGBG"); v != "" {
		parts := strings.Split(v, ";")
		if n, err := strconv.Atoi(strings.TrimSpace(parts[len(parts)-1])); err == nil && n >= 7 {
			return "light"
		}
	}
	if isTTY && hasDarkBackground != nil && !hasDarkBackground() {
		return "light"
	}
	return "dark"
}

// ApplyDetectedTheme is called once at startup. It never fails: an unknown
// override simply falls back to the detected default.
func ApplyDetectedTheme() string {
	isTTY := false
	if fi, err := os.Stdout.Stat(); err == nil {
		isTTY = fi.Mode()&os.ModeCharDevice != 0
	}
	name := DetectTheme(os.Getenv, isTTY, lipgloss.HasDarkBackground)
	if !SetTheme(name) {
		_ = SetTheme("dark")
		return "dark"
	}
	return name
}

// fg returns c unless the current theme is monochrome.
func fg(c lipgloss.Color) lipgloss.TerminalColor {
	if Current.Mono {
		return nil
	}
	return c
}

// bg returns c unless the current theme is monochrome.
func bg(c lipgloss.Color) lipgloss.TerminalColor {
	if Current.Mono {
		return nil
	}
	return c
}

// PaintInput colors a bubbles textinput from the current theme. It is called
// at startup and on every theme change: a theme switch must repaint the
// already-focused input, not just the chrome.
func PaintInput(ti *textinput.Model) {
	ti.PromptStyle = lipgloss.NewStyle().Foreground(fg(Current.Accent)).Bold(true)
	ti.TextStyle = lipgloss.NewStyle().Foreground(fg(Current.Text))
	ti.PlaceholderStyle = lipgloss.NewStyle().Foreground(fg(Current.Dim))
	ti.CursorStyle = lipgloss.NewStyle().Foreground(fg(Current.Accent))
}
