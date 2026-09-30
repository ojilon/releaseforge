package cmd

import (
	"fmt"
	"path/filepath"

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
		root, _, err := requireDataRoot()
		if err != nil {
			return err
		}
		if err := storage.EnsureProjectLayout(root, info.Name); err != nil {
			return err
		}
		var prog string
		var targs []string
		switch info.Type {
		case "android-gradle":
			switch kind {
			case "unit":
				targs = []string{":app:testDebugUnitTest"}
			case "instrumented":
				targs = []string{":app:connectedDebugAndroidTest"}
			case "all":
				targs = []string{":app:testDebugUnitTest", ":app:connectedDebugAndroidTest"}
			default:
				return fmt.Errorf("test: unknown kind %q (want unit|instrumented|all)", kind)
			}
			prog = build.GradleWrapper(info.Root)
		case "go":
			prog, targs = "go", build.GoTestArgs()
		default:
			return fmt.Errorf("test: project type %q not supported yet (root %s)", info.Type, info.Root)
		}
		logPath := rflog.LogPath(storage.LogsDir(root, info.Name), "test-"+kind)
		res := build.Run(prog, targs, build.Options{
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

// goBinaryOut returns the builds-dir output path for a Go build.
func goBinaryOut(dataRoot, projectName, variant string) string {
	base := "releaseforge"
	if variant == "release" {
		base = "releaseforge-release"
	}
	return filepath.Join(storage.BuildsDir(dataRoot, projectName), build.GoBinaryName(base))
}
