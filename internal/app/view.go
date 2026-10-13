package app

import (
	"strings"

	"github.com/ojilon/releaseforge/internal/tui"
	"github.com/ojilon/releaseforge/internal/version"
)

// View draws the whole screen: header, log pane, input bar, footer.
func (m Model) View() string {
	if !m.ready {
		return tui.RenderHeader(version.ToolVersion, m.projectName(), m.version, m.status) + "\n" + m.welcome()
	}
	bar := tui.BarStyle.Render(m.input.View())
	if m.width > 0 && m.width < 80 {
		header := "ReleaseForge " + version.ToolVersion + " · " + m.projectName()
		return header + "\n" + m.viewport.View() + "\n" + bar + "\n" +
			tui.HelpStyle.Render("help|quit")
	}
	top := tui.PanelBorder.Render(
		tui.RenderHeader(version.ToolVersion, m.projectName(), m.version, m.status) +
			" " + tui.Pill(m.phase) + "\n" + m.contextLine())
	return top + "\n" + m.viewport.View() + "\n" + bar + "\n" + m.footer()
}

// contextLine renders branch, dirtiness, and the commit sparkline.
func (m *Model) contextLine() string {
	line := "no git"
	if m.branch != "" {
		line = m.branch
		if m.dirty {
			line += " · ● dirty"
		} else {
			line += " · clean"
		}
	}
	if m.spark != "" {
		line += " · " + m.spark
	}
	return tui.HelpStyle.Render(line)
}

// footer renders state-dependent key hints.
func (m *Model) footer() string {
	switch m.phase {
	case "building", "testing":
		return tui.HelpStyle.Render("esc cancels · logs tail")
	case "failed":
		return tui.HelpStyle.Render("logs last to inspect · help · quit")
	default:
		return tui.HelpStyle.Render("help · status · scan [path] · recent · version · build · test · notes · logs · quit")
	}
}

// appendLine adds one line to the log ring and repaints.
func (m *Model) appendLine(s string) {
	if !m.ready {
		return
	}
	m.ring.Append(s)
	m.render()
}

// render paints the ring tail with follow-mode: it sticks to the bottom only
// when the user was already there, so reading scrollback survives new output.
func (m *Model) render() {
	if !m.ready {
		return
	}
	follow := m.viewport.AtBottom()
	m.viewport.SetContent(strings.Join(m.ring.Snapshot(), "\n"))
	if follow {
		m.viewport.GotoBottom()
	}
}

func (m *Model) welcome() string {
	return "ReleaseForge TUI — type `help` for commands, `scan <path>` to open a project.\n"
}
