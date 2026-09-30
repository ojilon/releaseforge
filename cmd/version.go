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
		switch info.Type {
		case "android-gradle":
			if versionSet != "" {
				next, err := project.SetGradleVersion(info.Root, info.VersionFile, versionSet)
				if err != nil {
					return err
				}
				fmt.Printf("versionCode: %s\nversionName: %s\n", next, versionSet)
				return nil
			}
			code, name, err := project.GradleVersion(info.Root, info.VersionFile)
			if err != nil {
				return err
			}
			fmt.Printf("versionCode: %s\nversionName: %s\n", code, name)
			return nil
		case "go":
			if versionSet != "" {
				if err := project.SetVersionFile(info.Root, "VERSION", versionSet); err != nil {
					return err
				}
				fmt.Printf("version: %s (wrote VERSION, not committed)\n", versionSet)
				return nil
			}
			v, err := project.ReadVersionFile(info.Root, "VERSION")
			if err != nil {
				return fmt.Errorf("version: %w (project %s)", err, info.Root)
			}
			fmt.Printf("version: %s\n", v)
			return nil
		default:
			return fmt.Errorf("version: project type %q has no managed version file yet (root %s)", info.Type, info.Root)
		}
	},
}

func init() {
	versionCmd.Flags().StringVar(&versionSet, "set", "", "new version (gradle: also increments code; go: writes VERSION, no commit)")
}
