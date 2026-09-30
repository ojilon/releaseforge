package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	buildVariant string // debug | release
	buildABIs    string
)

var buildCmd = &cobra.Command{
	Use:   "build [variant]",
	Short: "Build project (debug/release) with correct targets and live logs",
	Long: `For Android Gradle:
  assembleDebug / assembleRelease, honouring aurora.abiFilters / NDK.
For Wails:
  wails build
For CMake:
  cmake --build

Logs are streamed live and persisted under the data-root.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		variant := "debug"
		if len(args) > 0 {
			variant = args[0]
		}
		if buildVariant != "" {
			variant = buildVariant
		}
		fmt.Printf("build: variant=%s abis=%s project=%s\n", variant, buildABIs, projectDir)
		return notImplemented("build")
	},
}

func init() {
	buildCmd.Flags().StringVar(&buildVariant, "variant", "", "debug or release")
	buildCmd.Flags().StringVar(&buildABIs, "abis", "", "comma-separated ABI filters (Android)")
}
