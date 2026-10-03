package cmd

import (
	"fmt"
	"strings"

	"github.com/ojilon/releaseforge/internal/build"
	rflog "github.com/ojilon/releaseforge/internal/log"
	"github.com/ojilon/releaseforge/internal/project"
	"github.com/ojilon/releaseforge/internal/storage"
	"github.com/spf13/cobra"
)

var (
	buildVariant string // debug | release
	buildABIs    string
)

var buildCmd = &cobra.Command{
	Use:   "build [variant]",
	Short: "Build project with live logs persisted under the data-root",
	Long: `Android Gradle:
  debug   → assembleDebug
  release → assembleRelease (aurora.abiFilters honoured via gradle.properties)

Go:
  debug   → go build with version stamp into builds/
  release → go build -trimpath with stripped symbols into builds/

Logs are streamed live and persisted under the data-root.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		variant := "debug"
		if len(args) > 0 {
			variant = args[0]
		}
		if buildVariant != "" {
			variant = buildVariant
		}
		variant = strings.ToLower(strings.TrimSpace(variant))
		info, err := project.Detect(projectDir)
		if err != nil {
			return err
		}
		r, err := project.For(info)
		if err != nil {
			return fmt.Errorf("build: %w", err)
		}
		root, _, err := requireDataRoot()
		if err != nil {
			return err
		}
		if err := storage.EnsureProjectLayout(root, info.Name); err != nil {
			return err
		}
		out := r.BuildOutput(root, variant)
		bargs, err := r.BuildArgs(variant, out)
		if err != nil {
			return err
		}
		if _, ok := r.(project.GradleRunner); ok && strings.TrimSpace(buildABIs) != "" && verbose {
			fmt.Printf("note: --abis %s honoured via gradle.properties aurora.abiFilters (edit file before build)\n", buildABIs)
		}
		logPath := rflog.LogPath(storage.LogsDir(root, info.Name), "build-"+variant)
		res := build.Run(r.Program(), bargs, build.Options{
			Dir:     info.Root,
			LogPath: logPath,
			OnLine:  func(t string, _ bool) { fmt.Println(t) },
		})
		fmt.Printf("log: %s\n", res.LogPath)
		if !res.Success {
			fmt.Printf("build %s FAILED (exit %d)\n", variant, res.ExitCode)
			printReport(res.Report)
			for _, e := range res.Errors {
				fmt.Printf("  ! %s\n", e)
			}
			return fmt.Errorf("build %s failed", variant)
		}
		if out != "" {
			fmt.Printf("binary: %s\n", out)
		}
		fmt.Printf("build %s succeeded\n", variant)
		return nil
	},
}

func init() {
	buildCmd.Flags().StringVar(&buildVariant, "variant", "", "debug or release")
	buildCmd.Flags().StringVar(&buildABIs, "abis", "", "comma-separated ABI filters (Android)")
}

// printReport renders structured error items after a failure.
func printReport(rep *build.Report) {
	if rep == nil {
		return
	}
	for _, l := range rep.Format() {
		fmt.Println(l)
	}
}
