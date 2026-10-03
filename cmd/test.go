package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/ojilon/releaseforge/internal/android"
	"github.com/ojilon/releaseforge/internal/build"
	rflog "github.com/ojilon/releaseforge/internal/log"
	"github.com/ojilon/releaseforge/internal/project"
	"github.com/ojilon/releaseforge/internal/storage"
	"github.com/spf13/cobra"
)

var (
	testKind       string // unit | instrumented | all (android); go ignores kind
	testStacktrace bool
	testInfo       bool
	testDebug      bool
)

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
		root, _, err := requireDataRoot()
		if err != nil {
			return err
		}
		if err := storage.EnsureProjectLayout(root, info.Name); err != nil {
			return err
		}
		for _, k := range expandKinds(r, kind) {
			targs, err := r.TestArgs(k)
			if err != nil {
				return err
			}
			extra, err := gradleVerbosity(r, testStacktrace, testInfo, testDebug)
			if err != nil {
				return err
			}
			targs = append(targs, extra...)
			if k == "instrumented" {
				// TODO(doc10): use PickDevice; for now fail fast without devices.
				if devs, derr := android.Devices(); derr != nil || len(devs) == 0 {
					return fmt.Errorf("test instrumented: no devices connected — connect one and check `adb devices`")
				}
			}
			logPath := rflog.LogPath(storage.LogsDir(root, info.Name), "test-"+k)
			res := build.Run(r.Program(), targs, build.Options{
				Dir:     info.Root,
				LogPath: logPath,
				Project: info.Name,
				OnLine:  func(t string, _ bool) { fmt.Println(t) },
			})
			fmt.Printf("log: %s\n", res.LogPath)
			if !res.Success {
				fmt.Printf("test %s FAILED (exit %d)\n", k, res.ExitCode)
				printReport(res.Report)
				for _, e := range res.Errors {
					fmt.Printf("  ! %s\n", e)
				}
				return fmt.Errorf("test %s failed", k)
			}
			fmt.Printf("test %s passed\n", k)
		}
		if summary, path, err := build.SummarizeTests(
			filepath.Join(info.Root, "app", "build", "test-results"),
			storage.ReportsDir(root, info.Name)); err != nil {
			return fmt.Errorf("junit summary: %w", err)
		} else if summary != "" {
			fmt.Printf("tests: %s\nreport: %s\n", summary, path)
		}
		return nil
	},
}

func init() {
	testCmd.Flags().StringVar(&testKind, "kind", "", "unit | instrumented | all (android only)")
	testCmd.Flags().BoolVar(&testStacktrace, "stacktrace", false, "pass --stacktrace to Gradle (Android only)")
	testCmd.Flags().BoolVar(&testInfo, "info", false, "pass --info to Gradle (Android only)")
	testCmd.Flags().BoolVar(&testDebug, "debug", false, "pass --debug to Gradle (Android only)")
}

// expandKinds splits kind "all" into separate logged runs for Gradle; other
// runners and kinds run once as given.
func expandKinds(r project.Runner, kind string) []string {
	if kind == "all" {
		if _, ok := r.(project.GradleRunner); ok {
			return []string{"unit", "instrumented"}
		}
	}
	return []string{kind}
}
