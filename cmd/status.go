package cmd

import (
	"fmt"

	"github.com/ojilon/releaseforge/internal/android"
	"github.com/ojilon/releaseforge/internal/git"
	"github.com/ojilon/releaseforge/internal/project"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show project scan summary, last build, version, device, config health",
	RunE: func(cmd *cobra.Command, args []string) error {
		info, err := project.Detect(projectDir)
		if err != nil {
			return err
		}
		root, cfgPath, err := resolveDataRoot()
		if err != nil {
			return err
		}
		ver := info.Type
		if code, name, err := project.CurrentVersion(info); err == nil && name != "" {
			if code != "" {
				ver = fmt.Sprintf("%s (code %s) from %s", name, code, info.VersionFile)
			} else if info.VersionFile != "" {
				ver = fmt.Sprintf("%s from %s", name, info.VersionFile)
			} else {
				ver = name
			}
		}
		fmt.Printf("project:   %s\nroot:      %s\ntype:      %s\nversion:   %s\ndata-root: %s\nconfig:    %s\n",
			info.Name, info.Root, info.Type, ver, root, cfgPath)
		if git.IsRepo(info.Root) {
			fmt.Printf("git:       %s @ %s (clean=%v, tag=%s)\n",
				git.CurrentBranch(info.Root), git.Head(info.Root), git.IsClean(info.Root), git.LatestTag(info.Root))
		} else {
			fmt.Printf("git:       not a repo\n")
		}
		if devs, err := android.Devices(); err == nil && len(devs) > 0 {
			fmt.Printf("devices:   %v\n", devs)
		} else {
			fmt.Printf("devices:   none (adb empty or missing)\n")
		}
		return nil
	},
}
