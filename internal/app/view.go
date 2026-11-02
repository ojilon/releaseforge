package app

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/ojilon/releaseforge/internal/tui"
	rfversion "github.com/ojilon/releaseforge/internal/version"
)

// View draws the shell: header frame, tab strip, framed active window,
// framed command bar and a bubbles/help footer. Every block has an exact
// width/height so the screen never wraps at 70, 100 or 160 columns.
func (m Model) View() string {
	if !m.ready {
		return tui.Clamp("ReleaseForge — initialising…", 40, 1)
	}
	l := m.layout
	ctx := m.context()

	header := m.headerLines(ctx)
	headerBox := frame(header, l.Width, l.HeaderH, tui.FrameBorder)

	var blocks []string
	blocks = append(blocks, headerBox)
	if l.TabsH > 0 {
		strip := tui.Tabs(WindowTitles, int(m.active), l.Width, l.Compact)
		blocks = append(blocks, tui.Clamp(strip, l.Width, 1))
	}

	body := m.window().View(ctx)
	if m.focus == FocusPalette || m.focus == FocusPicker {
		body = m.overlayBody()
	}
	blocks = append(blocks, frame(body, l.Width, l.BodyH, tui.FrameBorderActive))
	blocks = append(blocks, frame(m.input.View(), l.Width, l.BarH, tui.BarBorder))
	blocks = append(blocks, tui.Clamp(m.footer(), l.Width, 1))
	return strings.Join(blocks, "\n")
}

// frame falls back to a plain clamped block when there is no room for a
// border (tiny terminals).
func frame(content string, w, h int, style lipgloss.Style) string {
	if h < 3 {
		return tui.Clamp(content, w, h)
	}
	return tui.Frame(content, w, h, style)
}

func (m Model) headerLines(ctx Context) string {
	l := m.layout
	status := m.status
	if l.Compact {
		status = "" // compact header has one line; the phase pill carries state
	}
	line1 := tui.RenderHeader(rfversion.ToolVersion, m.projectName(), m.version, status) +
		"  " + tui.Pill(m.phase.String()) +
		"  " + tui.ChipStyle.Render(m.window().Title())
	if m.focus == FocusWindow {
		line1 += " " + tui.WarnStyle.Render("[window focus]")
	}
	if l.HeaderRect().Height <= 1 {
		return tui.Clamp(line1, l.HeaderRect().Width, 1)
	}
	context := "no git"
	if m.branch != "" {
		context = m.branch
		if m.dirty {
			context += " · ● dirty"
		} else {
			context += " · clean"
		}
	}
	if spark := strings.TrimSpace(m.spark); spark != "" {
		context += " · " + spark
	}
	context += "  ·  " + m.focus.String() + " focus  ·  theme " + m.themeName
	line2 := tui.HelpStyle.Render(" " + context)
	return tui.Clamp(line1+"\n"+line2, l.HeaderRect().Width, l.HeaderRect().Height)
}

// overlayBody centers the palette / folder picker inside the body frame.
func (m Model) overlayBody() string {
	l := m.layout
	pw, ph := m.paletteSize()
	ow, oh := pw+4, ph+7
	if ow > l.Content.Width {
		ow = l.Content.Width
	}
	if oh > l.Content.Height {
		oh = l.Content.Height
	}
	var content string
	switch m.focus {
	case FocusPalette:
		content = m.palette.View(ow-2) + "\n" +
			tui.HelpStyle.Render("↑↓ choose · enter act · esc close")
	case FocusPicker:
		if m.picker != nil {
			content = tui.TitleStyle.Render(" open folder ") + "\n" +
				tui.HelpStyle.Render("enter select · l/→ descend · h/← up · s use this folder · esc close") +
				"\n" + tui.Rule(ow-2) + "\n" + m.picker.View()
		}
	}
	box := frame(tui.Clamp(content, ow-2, oh-2), ow, oh, tui.OverlayBorder)
	return lipgloss.Place(l.Content.Width, l.Content.Height,
		lipgloss.Center, lipgloss.Center, box)
}

// footer is generated from bubbles/key bindings via bubbles/help.
func (m Model) footer() string {
	if m.hint != "" {
		return tui.HelpStyle.Render(tui.Truncate(m.hint, m.layout.Width))
	}
	bindings := m.keys.Global(m.phase.String(), m.focus == FocusWindow)
	if m.focus == FocusWindow {
		bindings = append(bindings, m.window().Keys()...)
	}
	return tui.ShortHelp(m.layout.Width, bindings)
}
