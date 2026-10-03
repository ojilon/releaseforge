package cmd

import (
	"fmt"

	"github.com/ojilon/releaseforge/internal/project"
	"github.com/ojilon/releaseforge/internal/version"
	"github.com/spf13/cobra"
)

var versionSet string

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show or set project version in its source of truth",
	Long: `Shows the tool version plus the open project's version.

Android Gradle: gradle.properties (app.versionCode/app.versionName).
Go: VERSION file at project root (no auto-commit; commit manually).
Use --set to write the project version file (never commits).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("releaseforge: %s\n", version.ToolVersion)
		info, err := project.Detect(projectDir)
		if err != nil {
			return err
		}
		r, err := project.For(info)
		if err != nil {
			return fmt.Errorf("version: %w (root %s)", err, info.Root)
		}
		if versionSet != "" {
			next, err := r.VersionWrite(versionSet)
			if err != nil {
				return err
			}
			if next == "" {
				fmt.Printf("version: %s (wrote VERSION, not committed)\n", versionSet)
				return nil
			}
			printVersion(next, versionSet)
			return nil
		}
		code, name, err := r.VersionRead()
		if err != nil {
			return err
		}
		printVersion(code, name)
		return nil
	},
}

// printVersion renders a version with a code line only when the project type
// has a numeric code (gradle); otherwise it prints the plain version.
func printVersion(code, name string) {
	if code != "" {
		fmt.Printf("versionCode: %s\nversionName: %s\n", code, name)
		return
	}
	fmt.Printf("version: %s\n", name)
}

func init() {
	versionCmd.Flags().StringVar(&versionSet, "set", "", "new version (gradle: also increments code; go: writes VERSION, no commit)")
}
