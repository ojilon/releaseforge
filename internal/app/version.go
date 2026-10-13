package app

import (
	"fmt"
	"strings"

	"github.com/ojilon/releaseforge/internal/project"
	"github.com/ojilon/releaseforge/internal/tui"
	"github.com/ojilon/releaseforge/internal/version"
)

// doVersion shows or sets the project version.
// Usage: version | version --set X | version bump X | version --code-only
func (m *Model) doVersion(args []string) string {
	info, err := project.Detect(m.projectDir)
	if err != nil {
		return tui.ErrorStyle.Render("version: " + err.Error())
	}
	r, err := project.For(info)
	if err != nil {
		return tui.ErrorStyle.Render("version: " + err.Error())
	}
	header := "releaseforge: " + version.ToolVersion + "\n"
	setName, codeOnly := "", false
	switch {
	case len(args) >= 2 && (args[0] == "--set" || args[0] == "bump"):
		setName = strings.TrimSpace(args[1])
	case len(args) >= 1 && args[0] == "--code-only":
		codeOnly = true
	case len(args) > 0:
		return tui.ErrorStyle.Render("version: usage: version [--set X | bump X | --code-only]")
	}
	if codeOnly {
		next, err := r.BumpCode()
		if err != nil {
			return tui.ErrorStyle.Render("version: " + err.Error())
		}
		m.refreshProject()
		return header + fmt.Sprintf("versionCode: %s", next)
	}
	if setName != "" {
		if err := project.ValidateVersion(setName); err != nil {
			return tui.ErrorStyle.Render("version: " + err.Error())
		}
		next, err := r.VersionWrite(setName)
		if err != nil {
			return tui.ErrorStyle.Render("version: " + err.Error())
		}
		m.refreshProject()
		if next == "" {
			return header + fmt.Sprintf("version: %s (wrote file, not committed)", setName)
		}
		return header + fmt.Sprintf("versionCode: %s\nversionName: %s", next, setName)
	}
	code, name, err := r.VersionRead()
	if err != nil {
		return tui.ErrorStyle.Render("version: " + err.Error())
	}
	if code != "" {
		return header + fmt.Sprintf("versionCode: %s\nversionName: %s", code, name)
	}
	return header + fmt.Sprintf("version: %s", name)
}
