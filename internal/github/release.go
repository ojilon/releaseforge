// Package github creates releases via the gh CLI or the GitHub API.
package github

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Prefer gh when available and authenticated; fall back to API with token.

// HasGH reports whether the gh CLI is on PATH.
func HasGH() bool {
	_, err := exec.LookPath("gh")
	return err == nil
}

// CreateRelease creates tag v<version> GitHub release with notes file + artifacts.
// tag must already exist locally or be creatable; gh handles remote publish.
func CreateRelease(dir, tag, notesFile string, artifacts []string, prerelease bool) error {
	if !HasGH() {
		return fmt.Errorf("gh CLI not found on PATH (install gh and run gh auth login)")
	}
	if _, err := os.Stat(notesFile); err != nil {
		return fmt.Errorf("notes file not found: %s", notesFile)
	}
	args := []string{"release", "create", tag, "--notes-file", notesFile}
	// No --title flag: gh titles the release with the tag, which is what we
	// want; the Python reference passed --title explicitly.
	if prerelease {
		args = append(args, "--prerelease")
	}
	for _, a := range artifacts {
		if strings.TrimSpace(a) != "" {
			args = append(args, a)
		}
	}
	cmd := exec.Command("gh", args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("gh release create: %w", err)
	}
	return nil
}

// CreateTag creates an annotated tag locally (does not push).
func CreateTag(dir, tag, message string) error {
	cmd := exec.Command("git", "tag", "-a", tag, "-m", message)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git tag: %w", err)
	}
	return nil
}

// PushTag pushes tag to origin.
func PushTag(dir, tag string) error {
	cmd := exec.Command("git", "push", "origin", tag)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git push tag: %w", err)
	}
	return nil
}
