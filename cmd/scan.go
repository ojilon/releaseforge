package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var scanCmd = &cobra.Command{
	Use:   "scan [path]",
	Short: "Detect project type, version, tests, signing config and write local config",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := projectDir
		if len(args) > 0 {
			path = args[0]
		}
		fmt.Printf("scan: will detect type for %s (Gradle/Wails/CMake/...)\n", path)
		return notImplemented("scan")
	},
}
