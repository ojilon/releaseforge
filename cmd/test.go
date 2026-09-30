package cmd

import (
	"fmt"

	"github.com/ojilon/releaseforge/internal/build"
	"github.com/ojilon/releaseforge/internal/project"
	"github.com/ojilon/releaseforge/internal/storage"
	rflog "github.com/ojilon/releaseforge/internal/log"
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
		info, err := project.Detect(projectDir)
		if err != nil {
			return err
		}
		if info.Type != "android-gradle" {
			return fmt.Errorf("test: project type %q not supported yet (root %s)", info.Type, info.Root)
		}
		var tasks []string
		switch kind {
		case "unit":
			tasks = []string{":app:testDebugUnitTest"}
		case "instrumented":
			tasks = []string{":app:connectedDebugAndroidTest"}
		case "all":
			tasks = []string{":app:testDebugUnitTest", ":app:connectedDebugAndroidTest"}
		default:
			return fmt.Errorf("test: unknown kind %q (want unit|instrumented|all)", kind)
		}
		root, _, err := resolveDataRoot()
		if err != nil {
			return err
		}
		if err := storage.EnsureProjectLayout(root, info.Name); err != nil {
			return err
		}
		logPath := rflog.LogPath(storage.LogsDir(root, info.Name), "test-"+kind)
		wrapper := build.GradleWrapper(info.Root)
		res := build.Run(wrapper, tasks, build.Options{
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
	testCmd.Flags().StringVar(&testKind, "kind", "", "unit | instrumented | all")
}
