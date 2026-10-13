package app

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/ojilon/releaseforge/internal/git"
	"github.com/ojilon/releaseforge/internal/project"
	"github.com/ojilon/releaseforge/internal/tui"
)

// doGit shows git state while analysing or building.
// Usage: git log [-n 20] | git status | git diff | git tags
func (m *Model) doGit(args []string) string {
	info, err := project.Detect(m.projectDir)
	if err != nil {
		return tui.ErrorStyle.Render("git: " + err.Error())
	}
	if !git.IsRepo(info.Root) {
		return "git: not a repo (" + info.Root + ")"
	}
	sub := "log"
	rest := args
	if len(args) > 0 {
		sub = strings.ToLower(args[0])
		rest = args[1:]
	}
	switch sub {
	case "log":
		n := 20
		if len(rest) > 0 {
			fmt.Sscanf(rest[0], "%d", &n)
		}
		commits, err := git.Log(info.Root, n)
		if err != nil {
			return tui.ErrorStyle.Render("git log: " + err.Error())
		}
		if len(commits) == 0 {
			return "git log: no commits"
		}
		var b strings.Builder
		for _, c := range commits {
			fmt.Fprintf(&b, "%s %s (%s)\n", c.SHA, c.Subject, c.Date)
		}
		return strings.TrimRight(b.String(), "\n")
	case "status":
		branch := git.CurrentBranch(info.Root)
		state := "clean"
		if !git.IsClean(info.Root) {
			state = "dirty"
		}
		out := fmt.Sprintf("branch: %s\nstate:  %s", branch, state)
		if a, bh, ok := git.AheadBehind(info.Root); ok {
			out += fmt.Sprintf("\nahead:  %d\nbehind: %d", a, bh)
		}
		if origin := git.Origin(info.Root); origin != "" {
			out += "\norigin: " + origin
		}
		return out
	case "diff":
		out, err := exec.Command("git", "-C", info.Root, "diff", "--stat").CombinedOutput()
		if err != nil {
			return tui.ErrorStyle.Render("git diff: " + err.Error())
		}
		if strings.TrimSpace(string(out)) == "" {
			return "git diff: clean"
		}
		return strings.TrimRight(string(out), "\n")
	case "tags":
		tags := git.RecentTags(info.Root, 10)
		if len(tags) == 0 {
			return "git tags: none"
		}
		return "tags:\n  " + strings.Join(tags, "\n  ")
	default:
		return tui.ErrorStyle.Render("git: usage: git log|status|diff|tags")
	}
}
