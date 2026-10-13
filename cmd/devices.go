package cmd

import (
	"fmt"

	"github.com/ojilon/releaseforge/internal/android"
	"github.com/spf13/cobra"
)

var devicesCmd = &cobra.Command{
	Use:   "devices",
	Short: "List connected adb devices",
	RunE: func(cmd *cobra.Command, args []string) error {
		devs, err := android.Devices()
		if err != nil {
			return fmt.Errorf("%v (is adb on PATH? run `doctor`)", err)
		}
		if len(devs) == 0 {
			fmt.Println("no devices connected")
			return nil
		}
		for _, d := range devs {
			fmt.Println(d)
		}
		return nil
	},
}
