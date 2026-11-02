package cmd

import (
	"fmt"

	"github.com/ojilon/releaseforge/internal/tui"
	"github.com/spf13/cobra"
)

// commandsCmd prints exactly the same generated reference the TUI Help
// window shows. One verb table, two surfaces (docs/sweep-2mo/06).
var commandsCmd = &cobra.Command{
	Use:   "commands",
	Short: "Print the TUI command reference (generated from the verb table)",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		_, err := fmt.Fprintln(cmd.OutOrStdout(), tui.HelpText())
		return err
	},
}
