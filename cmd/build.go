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
		if variant != "debug" && variant != "release" {
			return fmt.Errorf("build: unknown variant %q (want debug|release)", variant)
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
		var bargs []string
		var extra string
		switch info.Type {
		case "android-gradle":
			prog = build.GradleWrapper(info.Root)
			bargs = []string{"assembleDebug"}
			if variant == "release" {
				bargs = []string{"assembleRelease"}
			}
			if strings.TrimSpace(buildABIs) != "" && verbose {
				fmt.Printf("note: --abis %s honoured via gradle.properties aurora.abiFilters (edit file before build)\n", buildABIs)
			}
		case "go":
			stamp := projectVersionName(info)
			out := goBinaryOut(root, info, variant)
			prog, bargs = "go", build.GoBuildArgs(info.Root, out, stamp, variant)
			extra = out
		default:
			return fmt.Errorf("build: project type %q not supported yet (root %s)", info.Type, info.Root)
		}
		logPath := rflog.LogPath(storage.LogsDir(root, info.Name), "build-"+variant)
		res := build.Run(prog, bargs, build.Options{
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
		if extra != "" {
			fmt.Printf("binary: %s\n", extra)
		}
		fmt.Printf("build %s succeeded\n", variant)
		return nil
	},
}

// projectVersionName returns the display version for stamping (VERSION content
// for go, versionName for gradle, "dev" fallback).
func projectVersionName(info project.Info) string {
	if _, name, err := project.CurrentVersion(info); err == nil && name != "" {
		return name
	}
	return "dev"
}

func init() {
	buildCmd.Flags().StringVar(&buildVariant, "variant", "", "debug or release")
	buildCmd.Flags().StringVar(&buildABIs, "abis", "", "comma-separated ABI filters (Android)")
}
