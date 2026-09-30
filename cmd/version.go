package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionSet string

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show or set project version in its source of truth",
	RunE: func(cmd *cobra.Command, args []string) error {
		if versionSet != "" {
			fmt.Printf("version: set to %s\n", versionSet)
			return notImplemented("version set")
		}
		fmt.Println("version: will read current version from project")
		return notImplemented("version get")
	},
}

func init() {
	versionCmd.Flags().StringVar(&versionSet, "set", "", "new version name (also increments code where applicable)")
}
