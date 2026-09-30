package cmd

import (
	"github.com/spf13/cobra"
)

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch the interactive TUI (same as running releaseforge with no args)",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runTUI()
	},
}
