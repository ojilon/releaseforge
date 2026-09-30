package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	releasePre     bool
	releaseSkipTests bool
	releaseNotesOnly bool
)

var releaseCmd = &cobra.Command{
	Use:   "release <version>",
	Short: "Full release pipeline: version → test → build → package → sign → notes → tag → GitHub",
	Long: `Orchestrates the same flow currently done by Conductino-Android scripts/release.py,
generalised for all supported project types.

Steps (Android):
  1. set version (gradle.properties)
  2. unit tests
  3. assembleDebug + assembleRelease
  4. package into release/<version>/
  5. sign release APK (apksigner + keystore)
  6. generate notes from git history (template + commits)
  7. zip artifacts
  8. create annotated tag and GitHub release (gh)

Flags control pre-release marking and optional skips.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		version := args[0]
		fmt.Printf("release: version=%s pre=%v skip-tests=%v notes-only=%v\n",
			version, releasePre, releaseSkipTests, releaseNotesOnly)
		return notImplemented("release")
	},
}

func init() {
	releaseCmd.Flags().BoolVar(&releasePre, "pre", false, "mark GitHub release as pre-release")
	releaseCmd.Flags().BoolVar(&releaseSkipTests, "skip-tests", false, "skip test step (not recommended)")
	releaseCmd.Flags().BoolVar(&releaseNotesOnly, "notes-only", false, "only generate/update release notes")
}
