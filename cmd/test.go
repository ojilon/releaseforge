package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var testKind string // unit | instrumented | all

var testCmd = &cobra.Command{
	Use:   "test [kind]",
	Short: "Run unit and/or instrumented tests with live output and report capture",
	Long: `Android:
  unit         → ./gradlew :app:testDebugUnitTest
  instrumented → ./gradlew :app:connectedDebugAndroidTest
  all          → both

Reports and logs land under <data-root>/projects/<name>/reports and /logs.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		kind := "unit"
		if len(args) > 0 {
			kind = args[0]
		}
		if testKind != "" {
			kind = testKind
		}
		fmt.Printf("test: kind=%s project=%s\n", kind, projectDir)
		return notImplemented("test")
	},
}

func init() {
	testCmd.Flags().StringVar(&testKind, "kind", "", "unit | instrumented | all")
}
