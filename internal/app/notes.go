package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ojilon/releaseforge/internal/git"
	"github.com/ojilon/releaseforge/internal/project"
	"github.com/ojilon/releaseforge/internal/storage"
	"github.com/ojilon/releaseforge/internal/tui"
)

// doNotes drafts release notes from git history.
// Without a version it prints to screen; with one it saves notes.md.
func (m *Model) doNotes(ver string) string {
	info, err := project.Detect(m.projectDir)
	if err != nil {
		return tui.ErrorStyle.Render("notes: " + err.Error())
	}
	var commits []git.Commit
	if git.IsRepo(info.Root) {
		prev := git.LatestTag(info.Root)
		if cl, err := git.LogSince(info.Root, prev); err == nil {
			commits = cl
		}
	}
	ver = strings.TrimSpace(ver)
	if ver == "" {
		return git.DraftNotes(info.Name, "unreleased", commits, true)
	}
	verDir := storage.ReleaseVersionDir(m.dataRoot, info.Name, ver)
	if err := os.MkdirAll(verDir, 0o755); err != nil {
		return tui.ErrorStyle.Render("notes: " + err.Error())
	}
	path := filepath.Join(verDir, "notes.md")
	if err := os.WriteFile(path, []byte(git.DraftNotes(info.Name, ver, commits, true)), 0o644); err != nil {
		return tui.ErrorStyle.Render("notes: " + err.Error())
	}
	return fmt.Sprintf("notes: %s (%d commits since last tag)", path, len(commits))
}
