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

// AheadBehind returns commits ahead/behind the upstream. ok is false when
// there is no upstream (not an error).
func AheadBehind(dir string) (ahead, behind int, ok bool) {
	out, err := run(dir, "rev-list", "--left-right", "--count", "HEAD...@{u}")
	if err != nil {
		return 0, 0, false
	}
	var a, b int
	if _, serr := fmt.Sscanf(out, "%d %d", &a, &b); serr != nil {
		return 0, 0, false
	}
	return a, b, true
}

// MaxChangelogCommits bounds grouped notes; overflow is reported, not dumped.
const MaxChangelogCommits = 100

// ChangeGroup is one conventional-commit section of a changelog.
type ChangeGroup struct {
	Title   string
	Commits []Commit
}

var groupOrder = []string{"Features", "Fixes", "Docs", "Maintenance", "Other"}

func groupFor(prefix string) string {
	switch strings.ToLower(prefix) {
	case "feat":
		return "Features"
	case "fix":
		return "Fixes"
	case "docs":
		return "Docs"
	case "refactor", "perf", "test", "build", "ci", "style", "chore":
		return "Maintenance"
	default:
		return "Other"
	}
}

// GroupCommits buckets commits by conventional-commit prefix
// (`type(scope)!: subject`, case-insensitive). The `!` marker stays in the
// subject; unknown shapes land in Other. Group order is fixed.
func GroupCommits(commits []Commit) []ChangeGroup {
	buckets := map[string][]Commit{}
	for _, c := range commits {
		buckets[groupFor(conventionalPrefix(c.Subject))] = append(
			buckets[groupFor(conventionalPrefix(c.Subject))], c)
	}
	var out []ChangeGroup
	for _, title := range groupOrder {
		if items := buckets[title]; len(items) > 0 {
			out = append(out, ChangeGroup{Title: title, Commits: items})
		}
	}
	return out
}

// conventionalPrefix extracts the type from "type(scope)!: subject".
func conventionalPrefix(subject string) string {
	s := strings.TrimSpace(subject)
	colon := strings.Index(s, ":")
	if colon <= 0 {
		return ""
	}
	head := s[:colon]
	if i := strings.Index(head, "("); i >= 0 {
		head = head[:i]
	}
	return strings.TrimSuffix(strings.TrimSpace(head), "!")
}

// Changelog groups commits since the last tag, capped at max (<=0 means
// MaxChangelogCommits). It returns groups, total commits, and whether output
// was truncated.
func Changelog(dir string, max int) (groups []ChangeGroup, total int, truncated bool, err error) {
	if max <= 0 {
		max = MaxChangelogCommits
	}
	prev := LatestTag(dir)
	commits, err := LogSince(dir, prev)
	if err != nil {
		return nil, 0, false, err
	}
	total = len(commits)
	if len(commits) > max {
		commits = commits[:max]
		truncated = true
	}
	return GroupCommits(commits), total, truncated, nil
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

// DraftGroupedNotes builds release notes with Changes grouped by
// conventional-commit prefix. Headers match DraftNotes; empty groups are
// omitted and truncation is reported with a total count.
func DraftGroupedNotes(appName, version string, groups []ChangeGroup, total int, truncated bool, prerelease bool) string {
	kind := "release"
	if prerelease {
		kind = "pre-release"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "# %s %s\n\n**Type:** %s\n**Date:** %s\n\n", appName, version, kind, time.Now().Format("2006-01-02"))
	b.WriteString("## Highlights\n-\n\n")
	shown := 0
	for _, g := range groups {
		fmt.Fprintf(&b, "## %s\n\n", g.Title)
		for _, c := range g.Commits {
			fmt.Fprintf(&b, "- %s (%s, %s)\n", c.Subject, c.SHA, c.Date)
			shown++
		}
		b.WriteString("\n")
	}
	if shown == 0 {
		b.WriteString("## Changes\n\n-\n\n")
	}
	if truncated {
		fmt.Fprintf(&b, "…and %d more (see git log)\n\n", total-shown)
	}
	b.WriteString("## Bug fixes\n-\n\n## Known issues\n-\n")
	return b.String()
}

// Notes are drafted from commits between tags; AI polish is a later phase.
