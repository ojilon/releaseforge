package app

import (
	"fmt"
	"strings"

	"github.com/ojilon/releaseforge/internal/project"
	"github.com/ojilon/releaseforge/internal/tui"
)

// dashWidth is the fixed width of the right-side dashboard panel.
const dashWidth = 34

// dashVisible reports whether the dashboard should show beside the logs.
func (m *Model) dashVisible() bool {
	return m.width >= 100 && strings.TrimSpace(m.dash) != ""
}

// buildDash renders scan results as a small dashboard:
// versions, git state, tools. Plain text — view.go adds the border.
func (m *Model) buildDash() string {
	if m.info == nil {
		return ""
	}
	var b strings.Builder
	b.WriteString("scan\n")
	b.WriteString(fmt.Sprintf("name  %s\n", m.info.Name))
	b.WriteString(fmt.Sprintf("type  %s\n", m.info.Type))
	ver := strings.TrimSpace(m.version)
	if ver == "" {
		ver = "—"
	}
	b.WriteString(fmt.Sprintf("ver   %s\n", ver))
	if m.branch != "" {
		state := "clean"
		if m.dirty {
			state = "dirty"
		}
		b.WriteString(fmt.Sprintf("git   %s · %s\n", m.branch, state))
	}
	if m.spark != "" {
		b.WriteString(fmt.Sprintf("act   %s\n", m.spark))
	}
	// Enrich from the cached scan snapshot (best-effort, never fails).
	if snap, err := project.LoadScan(m.dataRoot, m.info.Name); err == nil {
		if snap.Git.Head != "" {
			b.WriteString(fmt.Sprintf("head  %s\n", shortSHA(snap.Git.Head)))
		}
		if len(snap.Git.RecentTags) > 0 {
			b.WriteString(fmt.Sprintf("tag   %s\n", snap.Git.RecentTags[0]))
		}
		if snap.Version.File != "" {
			b.WriteString(fmt.Sprintf("file  %s\n", snap.Version.File))
		}
		if len(snap.Tools) > 0 {
			var ids []string
			for i, t := range snap.Tools {
				if i >= 4 {
					break
				}
				ids = append(ids, t.ID)
			}
			b.WriteString(fmt.Sprintf("tools %s\n", strings.Join(ids, ",")))
		}
		if snap.Hints["legacy_release_scripts"] {
			b.WriteString("oldpy scripts/\n")
		}
	}
	return strings.TrimRight(b.String(), "\n")
}

// shortSHA keeps SHAs readable in the narrow panel.
func shortSHA(s string) string {
	s = strings.TrimSpace(s)
	if len(s) > 12 {
		return s[:12]
	}
	return s
}

// dashBox renders the dashboard with title styling + thin border.
func (m *Model) dashBox() string {
	title := tui.HelpStyle.Render("── dashboard ──")
	return tui.DashBoxStyle.Render(title + "\n" + m.dash)
}
