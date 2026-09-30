package cmd

import (
	"fmt"

	"github.com/ojilon/releaseforge/internal/project"
	"github.com/spf13/cobra"
)

var versionSet string

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show or set project version in its source of truth",
	RunE: func(cmd *cobra.Command, args []string) error {
		info, err := project.Detect(projectDir)
		if err != nil {
			return err
		}
		if info.Type != "android-gradle" {
			return fmt.Errorf("version: project type %q has no managed version file yet (root %s)", info.Type, info.Root)
		}
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
	},
}

func init() {
	versionCmd.Flags().StringVar(&versionSet, "set", "", "new version name (also increments code where applicable)")
}
