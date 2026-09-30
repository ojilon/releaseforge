package cmd

import (
	"fmt"
	"strings"

	"github.com/ojilon/releaseforge/internal/build"
	"github.com/ojilon/releaseforge/internal/project"
	"github.com/ojilon/releaseforge/internal/storage"
	rflog "github.com/ojilon/releaseforge/internal/log"
	"github.com/spf13/cobra"
)

var (
	buildVariant string // debug | release
	buildABIs    string
)

var buildCmd = &cobra.Command{
	Use:   "build [variant]",
	Short: "Build project (debug/release) with correct targets and live logs",
	Long: `For Android Gradle:
  assembleDebug / assembleRelease, honouring aurora.abiFilters / NDK.
For Wails:
  wails build
For CMake:
  cmake --build

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
		if variant != "debug" && variant != "release" {
			return fmt.Errorf("build: unknown variant %q (want debug|release)", variant)
		}
		info, err := project.Detect(projectDir)
		if err != nil {
			return err
		}
		if info.Type != "android-gradle" {
			return fmt.Errorf("build: project type %q not supported yet (root %s)", info.Type, info.Root)
		}
		task := "assembleDebug"
		if variant == "release" {
			task = "assembleRelease"
		}
		if strings.TrimSpace(buildABIs) != "" && verbose {
			fmt.Printf("note: --abis %s honoured via gradle.properties aurora.abiFilters (edit file before build)\n", buildABIs)
		}
		root, _, err := resolveDataRoot()
		if err != nil {
			return err
		}
		if err := storage.EnsureProjectLayout(root, info.Name); err != nil {
			return err
		}
		logPath := rflog.LogPath(storage.LogsDir(root, info.Name), "build-"+variant)
		wrapper := build.GradleWrapper(info.Root)
		res := build.Run(wrapper, []string{task}, build.Options{
			Dir:     info.Root,
			LogPath: logPath,
			OnLine:  func(t string, _ bool) { fmt.Println(t) },
		})
		fmt.Printf("log: %s\n", res.LogPath)
		if !res.Success {
			fmt.Printf("build %s FAILED (exit %d)\n", variant, res.ExitCode)
			for _, e := range res.Errors {
				fmt.Printf("  ! %s\n", e)
			}
			return fmt.Errorf("build %s failed", variant)
		}
		fmt.Printf("build %s succeeded\n", variant)
		return nil
	},
}

func init() {
	buildCmd.Flags().StringVar(&buildVariant, "variant", "", "debug or release")
	buildCmd.Flags().StringVar(&buildABIs, "abis", "", "comma-separated ABI filters (Android)")
}
