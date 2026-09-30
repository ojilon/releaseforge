package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "First-run wizard: choose data root (e.g. D:) and write global config",
	Long: `Interactive installation wizard (Huh forms).

Creates the managed data-root layout:
  <data-root>/
    config/global.json
    projects/
    history/
    global-cache/

Prefer a non-system drive (D:) for builds, logs, releases and caches.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("init: will run Huh-based wizard (choose drive/path).")
		return notImplemented("init")
	},
}
