package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show project scan summary, last build, version, device, config health",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("status: project=%s\n", projectDir)
		return notImplemented("status")
	},
}
