package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var installDevice string

var installCmd = &cobra.Command{
	Use:   "install [debug|release]",
	Short: "Install last built APK to a connected device via adb",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		variant := "debug"
		if len(args) > 0 {
			variant = args[0]
		}
		fmt.Printf("install: variant=%s device=%s\n", variant, installDevice)
		return notImplemented("install")
	},
}

func init() {
	installCmd.Flags().StringVar(&installDevice, "device", "", "adb serial (default: first device)")
}
