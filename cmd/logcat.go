package cmd

import (
	"fmt"

	"github.com/ojilon/releaseforge/internal/android"
	"github.com/ojilon/releaseforge/internal/project"
	"github.com/spf13/cobra"
)

var (
	logcatN      int
	logcatPkg    string
	logcatDevice string
)

var logcatCmd = &cobra.Command{
	Use:   "logcat",
	Short: "Print the tail of adb logcat, optionally filtered to a package",
	RunE: func(cmd *cobra.Command, args []string) error {
		if _, err := project.Detect(projectDir); err != nil {
			return err
		}
		serial, err := android.PickDevice(logcatDevice)
		if err != nil {
			return err
		}
		lines, err := android.LogcatTail(serial, logcatN, logcatPkg)
		if err != nil {
			return err
		}
		for _, l := range lines {
			fmt.Println(l)
		}
		return nil
	},
}

func init() {
	logcatCmd.Flags().IntVarP(&logcatN, "n", "n", 200, "lines to show")
	logcatCmd.Flags().StringVar(&logcatPkg, "pid", "", "filter to package (pidof when available, substring fallback)")
	logcatCmd.Flags().StringVar(&logcatDevice, "device", "", "adb serial (picker when 2+ devices)")
}
