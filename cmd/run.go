package cmd

import (
	"fmt"
	"strings"

	"github.com/ojilon/releaseforge/internal/android"
	"github.com/ojilon/releaseforge/internal/build"
	rflog "github.com/ojilon/releaseforge/internal/log"
	"github.com/ojilon/releaseforge/internal/project"
	"github.com/ojilon/releaseforge/internal/storage"
	"github.com/spf13/cobra"
)

var runDevice string

var runCmd = &cobra.Command{
	Use:   "run [debug|release]",
	Short: "Build, install, and launch the app on a connected device",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		variant := "debug"
		if len(args) > 0 {
			variant = args[0]
		}
		variant = strings.ToLower(strings.TrimSpace(variant))
		if variant != "debug" && variant != "release" {
			return fmt.Errorf("run: unknown variant %q", variant)
		}
		info, err := project.Detect(projectDir)
		if err != nil {
			return err
		}
		r, err := project.For(info)
		if err != nil {
			return fmt.Errorf("run: %w", err)
		}
		if _, ok := r.(project.GradleRunner); !ok {
			return fmt.Errorf("run: project type %q not supported", info.Type)
		}
		root, _, err := requireDataRoot()
		if err != nil {
			return err
		}
		if err := storage.EnsureProjectLayout(root, info.Name); err != nil {
			return err
		}
		bargs, err := r.BuildArgs(variant, "", "")
		if err != nil {
			return err
		}
		logPath := rflog.LogPath(storage.LogsDir(root, info.Name), "run-build-"+variant)
		res := build.Run(r.Program(), bargs, build.Options{
			Dir: info.Root, LogPath: logPath, Project: info.Name,
			OnLine: func(t string, _ bool) { fmt.Println(t) },
		})
		fmt.Printf("log: %s\n", res.LogPath)
		if !res.Success {
			printReport(res.Report)
			return fmt.Errorf("run: build failed")
		}
		apk, err := android.ResolveApk(info.Root, variantApk(variant))
		if err != nil {
			return fmt.Errorf("run: %w", err)
		}
		serial, err := android.PickDevice(runDevice)
		if err != nil {
			return err
		}
		fmt.Printf("installing %s ...\n", apk)
		if err := android.Install(apk, serial); err != nil {
			return err
		}
		pkg, err := launchPackage(root, info)
		if err != nil {
			return err
		}
		fmt.Printf("launching %s ...\n", pkg)
		return android.LaunchApp(serial, pkg)
	},
}

// variantApk returns the well-known output for a variant.
func variantApk(variant string) string {
	if variant == "release" {
		return android.ReleaseApkPath
	}
	return android.DebugApkPath
}

// launchPackage resolves the app id: scan cache first (applicationId, then
// manifest package), else a live scan; never guessed.
func launchPackage(dataRoot string, info project.Info) (string, error) {
	if snap, err := project.LoadScan(dataRoot, info.Name); err == nil &&
		project.Fresh(dataRoot, info.Name, info.Root) && snap.Gradle != nil {
		if v := snap.Gradle.Android.ApplicationID.Value; v != "" {
			return v, nil
		}
		if snap.Gradle.Manifest.Package != "" {
			return snap.Gradle.Manifest.Package, nil
		}
	}
	snap, _, err := project.Scan(info.Root)
	if err != nil {
		return "", err
	}
	if snap.Gradle != nil {
		if v := snap.Gradle.Android.ApplicationID.Value; v != "" {
			return v, nil
		}
		if snap.Gradle.Manifest.Package != "" {
			return snap.Gradle.Manifest.Package, nil
		}
	}
	return "", fmt.Errorf("unknown package id — scan the project first")
}

func init() {
	runCmd.Flags().StringVar(&runDevice, "device", "", "adb serial (picker when 2+ devices)")
}
