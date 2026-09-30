package cmd

import (
	"github.com/ojilon/releaseforge/internal/app"
	"github.com/spf13/cobra"
)

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch the interactive TUI (same as running releaseforge with no args)",
	RunE: func(cmd *cobra.Command, args []string) error {
		root, _, err := resolveDataRoot()
		if err != nil {
			return err
		}
		return app.Run(root, projectDir)
	},
}
