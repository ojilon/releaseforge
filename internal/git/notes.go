// Package git provides log/tag helpers and release-notes generation from history.
package git

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// Commit is a single parsed git log entry.
type Commit struct {
	SHA     string
	Subject string
	Date    string
}

func run(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git %s: %w", strings.Join(args, " "), err)
	}
	return strings.TrimSpace(string(out)), nil
}

// IsRepo reports whether dir is inside a git work tree.
func IsRepo(dir string) bool {
	out, err := run(dir, "rev-parse", "--is-inside-work-tree")
	return err == nil && out == "true"
}

// CurrentBranch returns the current branch (or HEAD short sha on detached).
func CurrentBranch(dir string) string {
	if out, err := run(dir, "rev-parse", "--abbrev-ref", "HEAD"); err == nil && out != "" {
		return out
	}
	return ""
}

// Head returns the short HEAD sha.
func Head(dir string) string {
	if out, err := run(dir, "rev-parse", "--short", "HEAD"); err == nil {
		return out
	}
	return ""
}

// LatestTag returns the most recent tag by creation date, or "" if none.
func LatestTag(dir string) string {
	out, err := run(dir, "tag", "-l", "--sort=-creatordate")
	if err != nil || out == "" {
		return ""
	}
	lines := strings.Split(out, "\n")
	for _, l := range lines {
		if t := strings.TrimSpace(l); t != "" {
			return t
		}
	}
	return ""
}

// Log returns up to n oneline commits (newest first).
func Log(dir string, n int) ([]Commit, error) {
	if n <= 0 {
		n = 30
	}
	out, err := run(dir, "log", fmt.Sprintf("-n%d", n), "--pretty=format:%H%x1f%h%x1f%ad%x1f%s", "--date=short")
	if err != nil {
		return nil, err
	}
	var commits []Commit
	for _, line := range strings.Split(out, "\n") {
		parts := strings.SplitN(strings.TrimSpace(line), "\x1f", 4)
		if len(parts) != 4 {
			continue
		}
		commits = append(commits, Commit{SHA: parts[1], Subject: parts[3], Date: parts[2]})
	}
	return commits, nil
}

// LogSince returns commits in range sinceTag..HEAD (or all if sinceTag empty).
func LogSince(dir, sinceTag string) ([]Commit, error) {
	rev := "HEAD"
	if strings.TrimSpace(sinceTag) != "" {
		rev = sinceTag + "..HEAD"
	}
	out, err := run(dir, "log", rev, "--pretty=format:%H%x1f%h%x1f%ad%x1f%s", "--date=short")
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(out) == "" {
		return nil, nil
	}
	var commits []Commit
	for _, line := range strings.Split(out, "\n") {
		parts := strings.SplitN(strings.TrimSpace(line), "\x1f", 4)
		if len(parts) != 4 {
			continue
		}
		commits = append(commits, Commit{SHA: parts[1], Subject: parts[3], Date: parts[2]})
	}
	return commits, nil
}

// IsClean reports whether the work tree has no staged/unstaged changes.
func IsClean(dir string) bool {
	out, err := run(dir, "status", "--porcelain")
	return err == nil && strings.TrimSpace(out) == ""
}

// Origin returns the remote origin URL, or "" when absent.
func Origin(dir string) string {
	out, err := run(dir, "config", "--get", "remote.origin.url")
	if err != nil {
		return ""
	}
	return out
}

// RecentTags returns up to n recent tags by creation date.
func RecentTags(dir string, n int) []string {
	out, err := run(dir, "tag", "-l", "--sort=-creatordate")
	if err != nil || out == "" {
		return nil
	}
	var tags []string
	for _, l := range strings.Split(out, "\n") {
		if t := strings.TrimSpace(l); t != "" {
			tags = append(tags, t)
			if n > 0 && len(tags) >= n {
				break
			}
		}
	}
	return tags
}

// TagExists reports whether tag exists locally.
func TagExists(dir, tag string) bool {
	out, err := run(dir, "rev-parse", "--verify", "refs/tags/"+tag)
	return err == nil && strings.TrimSpace(out) != ""
}

// DraftNotes builds release notes markdown from commits since the previous tag.
func DraftNotes(appName, version string, commits []Commit, prerelease bool) string {
	kind := "release"
	if prerelease {
		kind = "pre-release"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "# %s %s\n\n**Type:** %s\n**Date:** %s\n\n", appName, version, kind, time.Now().Format("2006-01-02"))
	b.WriteString("## Highlights\n-\n\n## Changes\n\n")
	if len(commits) == 0 {
		b.WriteString("-\n")
	} else {
		for _, c := range commits {
			fmt.Fprintf(&b, "- %s (%s, %s)\n", c.Subject, c.SHA, c.Date)
		}
	}
	b.WriteString("\n## Bug fixes\n-\n\n## Known issues\n-\n")
	return b.String()
}

// Notes are drafted from commits between tags; AI polish is a later phase.
