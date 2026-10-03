package cmd

import (
	"fmt"

	"github.com/ojilon/releaseforge/internal/build"
	"github.com/ojilon/releaseforge/internal/project"
	"github.com/ojilon/releaseforge/internal/storage"
	rflog "github.com/ojilon/releaseforge/internal/log"
	"github.com/spf13/cobra"
)

var testKind string // unit | instrumented | all (android); go ignores kind

var testCmd = &cobra.Command{
	Use:   "test [kind]",
	Short: "Run project tests with live output and persisted logs",
	Long: `Android Gradle:
  unit         → ./gradlew :app:testDebugUnitTest
  instrumented → ./gradlew :app:connectedDebugAndroidTest
  all          → both

Go:
  any kind     → go test ./...

Logs land under <data-root>/projects/<name>/logs.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		kind := "unit"
		if len(args) > 0 {
			kind = args[0]
		}
		if testKind != "" {
			kind = testKind
		}
		info, err := project.Detect(projectDir)
		if err != nil {
			return err
		}
		r, err := project.For(info)
		if err != nil {
			return fmt.Errorf("test: %w", err)
		}
		targs, err := r.TestArgs(kind)
		if err != nil {
			return err
		}
		root, _, err := requireDataRoot()
		if err != nil {
			return err
		}
		if err := storage.EnsureProjectLayout(root, info.Name); err != nil {
			return err
		}
		logPath := rflog.LogPath(storage.LogsDir(root, info.Name), "test-"+kind)
		res := build.Run(r.Program(), targs, build.Options{
			Dir:     info.Root,
			LogPath: logPath,
			OnLine:  func(t string, _ bool) { fmt.Println(t) },
		})
		fmt.Printf("log: %s\n", res.LogPath)
		if !res.Success {
			fmt.Printf("test %s FAILED (exit %d)\n", kind, res.ExitCode)
			for _, e := range res.Errors {
				fmt.Printf("  ! %s\n", e)
			}
			return fmt.Errorf("test %s failed", kind)
		}
		fmt.Printf("test %s passed\n", kind)
		return nil
	},
}

func init() {
	testCmd.Flags().StringVar(&testKind, "kind", "", "unit | instrumented | all (android only)")
}
