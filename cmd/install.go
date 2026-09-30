package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ojilon/releaseforge/internal/android"
	"github.com/ojilon/releaseforge/internal/project"
	"github.com/ojilon/releaseforge/internal/storage"
	"github.com/spf13/cobra"
)

var installDevice string

var installCmd = &cobra.Command{
	Use:   "install [debug|release]",
	Short: "Install last built APK to a connected device via adb",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		variant := "debug"
		if len(args) > 0 {
			variant = args[0]
		}
		variant = strings.ToLower(strings.TrimSpace(variant))
		if variant != "debug" && variant != "release" {
			return fmt.Errorf("install: unknown variant %q", variant)
		}
		info, err := project.Detect(projectDir)
		if err != nil {
			return err
		}
		if info.Type != "android-gradle" {
			return fmt.Errorf("install: project type %q not supported", info.Type)
		}
		// Prefer packaged artifacts under data-root releases, fall back to Gradle outputs.
		var apk string
		root, _, err := resolveDataRoot()
		if err == nil {
			relDir := storage.ReleasesDir(root, info.Name)
			entries, _ := os.ReadDir(relDir)
			// newest version dir first (lexicographic is good enough for v1)
			best := ""
			for _, e := range entries {
				if e.IsDir() && strings.Contains(e.Name(), ".") || e.IsDir() {
					if e.Name() > best {
						best = e.Name()
					}
				}
			}
			if best != "" {
				cand := filepath.Join(relDir, best, fmt.Sprintf("*-"+variant+".apk"))
				matches, _ := filepath.Glob(cand)
				if len(matches) > 0 {
					apk = matches[len(matches)-1]
				}
			}
		}
		if apk == "" {
			if variant == "debug" {
				apk = filepath.Join(info.Root, "app", "build", "outputs", "apk", "debug", "app-debug.apk")
			} else {
				// signed release may live in data-root; try plain output first
				apk = filepath.Join(info.Root, "app", "build", "outputs", "apk", "release", "app-release-unsigned.apk")
			}
		}
		if _, err := os.Stat(apk); err != nil {
			return fmt.Errorf("install: apk not found: %s (build first)", apk)
		}
		serial := strings.TrimSpace(installDevice)
		if serial == "" {
			if devs, err := android.Devices(); err == nil && len(devs) == 1 {
				serial = devs[0]
			}
		}
		fmt.Printf("installing %s ...\n", apk)
		return android.Install(apk, serial)
	},
}

func init() {
	installCmd.Flags().StringVar(&installDevice, "device", "", "adb serial (default: first device)")
}
