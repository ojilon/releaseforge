package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	cfgFile   string
	dataRoot  string
	projectDir string
	verbose   bool
)

// rootCmd is the base command when called without subcommands.
// Running with no args launches the interactive TUI.
var rootCmd = &cobra.Command{
	Use:   "releaseforge",
	Short: "Local release, build, test and lifecycle tool for Android & PC projects",
	Long: `ReleaseForge owns the full local lifecycle of your projects:

  scan → test → build (correct targets/ABIs) → sign → package →
  install-to-device → notes (from git) → tag → GitHub release.

Interactive TUI (default) plus scriptable Cobra subcommands for complex tasks.`,
	// Default action: start TUI when no subcommand given.
	RunE: func(cmd *cobra.Command, args []string) error {
		return runTUI()
	},
}

// Execute adds all child commands and runs the root.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "global config file (default: <data-root>/config/global.json)")
	rootCmd.PersistentFlags().StringVar(&dataRoot, "data-root", "", "override managed data root (e.g. D:/ReleaseForgeData)")
	rootCmd.PersistentFlags().StringVarP(&projectDir, "project", "p", ".", "project directory to operate on")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")

	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(scanCmd)
	rootCmd.AddCommand(tuiCmd)
	rootCmd.AddCommand(buildCmd)
	rootCmd.AddCommand(testCmd)
	rootCmd.AddCommand(releaseCmd)
	rootCmd.AddCommand(installCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(logsCmd)
	rootCmd.AddCommand(statusCmd)
}

// placeholder helpers so the tree compiles before full implementation
func notImplemented(name string) error {
	return fmt.Errorf("%s: not implemented yet — see docs/08-implementation-plan.md", name)
}

func runTUI() error {
	fmt.Fprintln(os.Stderr, "ReleaseForge TUI — foundation scaffold. Implement internal/app + internal/tui.")
	fmt.Fprintln(os.Stderr, "See docs/04-tui-design.md and docs/08-implementation-plan.md.")
	return notImplemented("tui")
}
