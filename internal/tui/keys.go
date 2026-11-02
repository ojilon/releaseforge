package tui

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
)

// KeyMap is the single source of truth for key bindings. The footer, the
// Help window and the docs are all generated from it (docs/sweep-2mo/06).
type KeyMap struct {
	Quit        key.Binding
	Focus       key.Binding
	Cancel      key.Binding
	Clear       key.Binding
	Palette     key.Binding
	History     key.Binding
	OpenFolder  key.Binding
	Help        key.Binding
	Complete    key.Binding
	NextWindow  key.Binding
	PrevWindow  key.Binding
	Windows     [6]key.Binding // plain 1..6 (only when the window owns focus)
	WindowsAlt  [6]key.Binding // alt+1..6 (always, so typing digits is safe)
	ScrollUp    key.Binding
	ScrollDown  key.Binding
	Refresh     key.Binding
	Select      key.Binding
	WindowFocus key.Binding
}

// DefaultKeyMap defines every global shortcut. Window-specific shortcuts are
// supplied by the window itself through Window.Keys().
func DefaultKeyMap() KeyMap {
	b := func(keys, desc string, more ...string) key.Binding {
		all := append([]string{keys}, more...)
		return key.NewBinding(key.WithKeys(all...), key.WithHelp(keys, desc))
	}
	km := KeyMap{
		Quit:        b("ctrl+c", "quit"),
		Focus:       b("esc", "back / focus bar"),
		Cancel:      b("ctrl+x", "cancel run"),
		Clear:       b("ctrl+l", "clear console"),
		Palette:     b("ctrl+p", "command palette"),
		History:     b("ctrl+r", "search history"),
		OpenFolder:  b("ctrl+o", "pick folder"),
		Help:        b("?", "help", "f1"),
		Complete:    b("tab", "complete"),
		NextWindow:  b("]", "next window", "alt+right"),
		PrevWindow:  b("[", "prev window", "alt+left"),
		ScrollUp:    b("pgup", "scroll up", "ctrl+u"),
		ScrollDown:  b("pgdown", "scroll down", "ctrl+d"),
		Refresh:     b("r", "refresh"),
		Select:      b("enter", "run / select"),
		WindowFocus: b("ctrl+w", "focus window"),
	}
	for i := range km.Windows {
		name := string(rune('1' + i))
		km.Windows[i] = b(name, []string{
			"status", "tests", "git", "console", "logs", "help",
		}[i])
		km.WindowsAlt[i] = b("alt+"+name, "window "+name)
	}
	return km
}

// Global returns the bindings shown in the footer while the command bar has
// focus. phase is used to swap in the cancel hint during a run.
func (km KeyMap) Global(phase string, focusedWindow bool) []key.Binding {
	if focusedWindow {
		return []key.Binding{km.ScrollUp, km.ScrollDown, km.Refresh, km.Focus, km.Help, km.Quit}
	}
	base := []key.Binding{km.Select, km.Complete, km.Palette, km.OpenFolder, km.WindowFocus, km.NextWindow, km.Help}
	if phase == "building" || phase == "testing" {
		base = append([]key.Binding{km.Cancel}, base...)
	}
	return append(base, km.Quit)
}

// HelpModel builds a themed bubbles/help view. Using help.Model (rather than a
// hand-written string) gives us automatic width truncation with "…".
func HelpModel(width int) help.Model {
	h := help.New()
	h.Width = width
	h.ShortSeparator = MutedStyle.Render(" · ")
	h.Styles.ShortKey = lipgloss.NewStyle().Foreground(fg(Current.Accent)).Bold(true)
	h.Styles.ShortDesc = lipgloss.NewStyle().Foreground(fg(Current.Dim))
	h.Styles.ShortSeparator = MutedStyle
	h.Styles.Ellipsis = MutedStyle
	return h
}

// ShortHelp renders one line of key hints, clipped to width.
func ShortHelp(width int, bindings []key.Binding) string {
	if width < 1 {
		width = 1
	}
	s := HelpModel(width).ShortHelpView(bindings)
	return Truncate(lipgloss.NewStyle().Render(s), width)
}
