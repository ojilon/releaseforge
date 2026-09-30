package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ojilon/releaseforge/internal/git"
	"github.com/ojilon/releaseforge/internal/project"
	"github.com/ojilon/releaseforge/internal/storage"
	"github.com/spf13/cobra"
)

var notesVersion string

var notesCmd = &cobra.Command{
	Use:   "notes [version]",
	Short: "Draft release notes from git history into the data-root releases dir",
	Long: `Merges git history since the previous tag into a notes.md template.

With a version argument (or --version), writes
<data-root>/projects/<name>/releases/<version>/notes.md.
Without one, prints the draft to stdout. Never commits.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ver := strings.TrimSpace(notesVersion)
		if len(args) > 0 {
			ver = strings.TrimSpace(args[0])
		}
		info, err := project.Detect(projectDir)
		if err != nil {
			return err
		}
		var commits []git.Commit
		if git.IsRepo(info.Root) {
			prev := git.LatestTag(info.Root)
			if cl, err := git.LogSince(info.Root, prev); err == nil {
				commits = cl
			}
		}
		body := git.DraftNotes(info.Name, firstNonEmpty(ver, "unreleased"), commits, true)
		if ver == "" {
			fmt.Print(body)
			return nil
		}
		root, _, err := requireDataRoot()
		if err != nil {
			return err
		}
		verDir := storage.ReleaseVersionDir(root, info.Name, ver)
		if err := os.MkdirAll(verDir, 0o755); err != nil {
			return err
		}
		path := filepath.Join(verDir, "notes.md")
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			return err
		}
		fmt.Printf("notes: %s (%d commits since last tag)\n", path, len(commits))
		return nil
	},
}

func firstNonEmpty(ss ...string) string {
	for _, s := range ss {
		if strings.TrimSpace(s) != "" {
			return s
		}
	}
	return ""
}

func init() {
	notesCmd.Flags().StringVar(&notesVersion, "version", "", "release version folder for notes.md")
}
