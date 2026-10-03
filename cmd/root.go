package cmd

import (
	"fmt"

	"github.com/ojilon/releaseforge/internal/app"
	"github.com/ojilon/releaseforge/internal/config"
	"github.com/ojilon/releaseforge/internal/storage"
	"github.com/ojilon/releaseforge/internal/version"
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
	Version: version.ToolVersion,
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
	rootCmd.AddCommand(recentCmd)
	rootCmd.AddCommand(tuiCmd)
	rootCmd.AddCommand(buildCmd)
	rootCmd.AddCommand(testCmd)
	rootCmd.AddCommand(releaseCmd)
	rootCmd.AddCommand(installCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(logsCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(notesCmd)
}

// resolveDataRoot applies --data-root / --config / env / default precedence.
// It returns the data root and the global.json path without creating anything.
func resolveDataRoot() (root string, configPath string, err error) {
	return config.DiscoverDataRoot(cfgFile, dataRoot)
}

// requireDataRoot resolves the data root and fails with a clear init hint
// when storage is not initialised or not accessible.
func requireDataRoot() (root string, configPath string, err error) {
	root, configPath, err = resolveDataRoot()
	if err != nil {
		return "", "", err
	}
	if !storage.Exists(configPath) {
		return "", "", fmt.Errorf("data root not initialised (%s missing) — run `releaseforge init` first", configPath)
	}
	return root, configPath, nil
}

func runTUI() error {
	root, _, err := resolveDataRoot()
	if err != nil {
		return err
	}
	return app.Run(root, projectDir)
}
