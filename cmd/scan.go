package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/ojilon/releaseforge/internal/config"
	"github.com/ojilon/releaseforge/internal/history"
	"github.com/ojilon/releaseforge/internal/project"
	"github.com/ojilon/releaseforge/internal/storage"
	"github.com/spf13/cobra"
)

var scanCmd = &cobra.Command{
	Use:   "scan [path]",
	Short: "Detect project type, version, git history and write scan cache",
	Long: `Scan pipeline: resolve path → git state → tool markers → version hints.

Persists cache/scan.json + minimal config.json under the data root and
updates the recent-projects list. Re-run scan/rescan to refresh.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := projectDir
		if len(args) > 0 {
			path = args[0]
		}
		snap, info, err := project.Scan(path)
		if err != nil {
			return err
		}
		root, _, err := requireDataRoot()
		if err != nil {
			return err
		}
		scanPath, err := project.WriteScan(root, snap)
		if err != nil {
			return err
		}
		// Merge fresh detection with seed + preserved user edits; the config is
		// rewritten when anything changed.
		var existing *config.ProjectConfig
		cfgPath := storage.ProjectConfigPath(root, info.Name)
		if loaded, err := config.LoadProject(cfgPath); err == nil {
			existing = &loaded
		}
		merged, notes := project.MergeConfig(project.SeedConfig(info),
			info.Root+"/.releaseforge.json", existing)
		savedAs := "unchanged"
		if existing == nil {
			if err := config.SaveProject(cfgPath, merged); err != nil {
				return fmt.Errorf("write project config: %w", err)
			}
			savedAs = "written"
		} else if !configsEqual(merged, *existing) {
			if err := config.SaveProject(cfgPath, merged); err != nil {
				return fmt.Errorf("write project config: %w", err)
			}
			savedAs = "updated"
		}
		if err := history.TouchRecent(root, info.Root, info.Name, info.Type, true); err != nil {
			return fmt.Errorf("update recent: %w", err)
		}
		versionStr := "(no version source)"
		if snap.Version.Name != "" {
			if snap.Version.Code != "" {
				versionStr = fmt.Sprintf("%s (code %s) from %s", snap.Version.Name, snap.Version.Code, snap.Version.File)
			} else {
				versionStr = fmt.Sprintf("%s from %s", snap.Version.Name, snap.Version.File)
			}
		} else if snap.Version.File != "" {
			versionStr = fmt.Sprintf("(unreadable %s)", snap.Version.File)
		}
		tools := "(none)"
		if len(snap.Tools) > 0 {
			tools = ""
			for i, t := range snap.Tools {
				if i > 0 {
					tools += ", "
				}
				tools += t.ID
			}
		}
		gitStr := "not a repo"
		if snap.Git.Present {
			gitStr = fmt.Sprintf("%s @ %s (%d commits, %d tags)",
				snap.Git.Branch, snap.Git.Head, len(snap.Git.RecentCommits), len(snap.Git.RecentTags))
		}
		fmt.Printf("Project: %s\nRoot:    %s\nType:    %s\nVersion: %s\nGit:     %s\nTools:   %s\nConfig:  %s (%s)\nCache:   %s\n",
			info.Name, info.Root, info.Type, versionStr, gitStr, tools, cfgPath, savedAs, scanPath)
		for _, n := range notes {
			fmt.Printf("note: %s\n", n)
		}
		return nil
	},
}

// configsEqual compares two project configs by their canonical JSON.
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
