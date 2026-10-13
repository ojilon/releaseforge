package app

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ojilon/releaseforge/internal/config"
	"github.com/ojilon/releaseforge/internal/history"
	"github.com/ojilon/releaseforge/internal/project"
	"github.com/ojilon/releaseforge/internal/storage"
	"github.com/ojilon/releaseforge/internal/tui"
)

// doScan detects the project at path, writes scan cache + config,
// and switches the TUI to that project.
func (m *Model) doScan(path string) string {
	snap, info, err := project.Scan(path)
	if err != nil {
		return tui.ErrorStyle.Render("scan: " + err.Error())
	}
	scanPath, err := project.WriteScan(m.dataRoot, snap)
	if err != nil {
		return tui.ErrorStyle.Render("scan: " + err.Error())
	}
	var existing *config.ProjectConfig
	if loaded, err := config.LoadProject(storage.ProjectConfigPath(m.dataRoot, info.Name)); err == nil {
		existing = &loaded
	}
	merged, notes := project.MergeConfig(project.SeedConfig(info),
		info.Root+"/.releaseforge.json", existing)
	cfgPath := storage.ProjectConfigPath(m.dataRoot, info.Name)
	savedAs := "unchanged"
	if existing == nil {
		savedAs = "written"
	}
	if existing == nil || !configsEqual(merged, *existing) {
		if err := config.SaveProject(cfgPath, merged); err != nil {
			return tui.ErrorStyle.Render("scan: " + err.Error())
		}
		if existing != nil {
			savedAs = "updated"
		}
	}
	if err := history.TouchRecent(m.dataRoot, info.Root, info.Name, info.Type, true); err != nil {
		return tui.ErrorStyle.Render("scan: " + err.Error())
	}
	m.projectDir = info.Root
	m.refreshProject()
	m.status = "scanned " + info.Name
	out := fmt.Sprintf("scanned %s (%s) → %s (config %s)", info.Name, info.Type, scanPath, savedAs)
	for _, n := range notes {
		out += "\nnote: " + n
	}
	return out
}

// doRecent lists recently opened projects.
func (m *Model) doRecent() string {
	return m.recentText()
}

func (m *Model) recentText() string {
	rf, err := history.LoadRecent(m.dataRoot)
	if err != nil {
		return tui.ErrorStyle.Render("recent: " + err.Error())
	}
	if len(rf.Items) == 0 {
		return "no recent projects — run `scan <path>`"
	}
	var b strings.Builder
	for i, e := range rf.Items {
		fmt.Fprintf(&b, "%d. %s  [%s]  %s\n", i+1, e.Name, e.Type, e.Path)
	}
	return strings.TrimRight(b.String(), "\n")
}

// configsEqual compares two project configs by canonical JSON.
func configsEqual(a, b config.ProjectConfig) bool {
	aj, err := json.Marshal(a)
	if err != nil {
		return false
	}
	bj, err := json.Marshal(b)
	if err != nil {
		return false
	}
	return string(aj) == string(bj)
}
