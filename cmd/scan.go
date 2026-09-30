package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/ojilon/releaseforge/internal/config"
	"github.com/ojilon/releaseforge/internal/project"
	"github.com/ojilon/releaseforge/internal/storage"
	"github.com/spf13/cobra"
)

var scanCmd = &cobra.Command{
	Use:   "scan [path]",
	Short: "Detect project type, version, tests, signing config and write local config",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := projectDir
		if len(args) > 0 {
			path = args[0]
		}
		info, err := project.Detect(path)
		if err != nil {
			return err
		}
		root, _, err := resolveDataRoot()
		if err != nil {
			return err
		}
		if err := storage.EnsureProjectLayout(root, info.Name); err != nil {
			return err
		}
		// Build minimal project config from detection.
		pcfg := config.ProjectConfig{
			Type: info.Type,
			Name: info.Name,
			Root: info.Root,
		}
		if info.VersionFile != "" {
			pcfg.Version = config.VersionConfig{File: info.VersionFile, CodeKey: info.CodeKey, NameKey: info.NameKey}
		}
		if info.Type == "android-gradle" {
			pcfg.Build = config.BuildConfig{
				GradleWrapper: true,
				Tasks: map[string]string{
					"unit_test":         ":app:testDebugUnitTest",
					"instrumented_test": ":app:connectedDebugAndroidTest",
					"assemble_debug":    "assembleDebug",
					"assemble_release":  "assembleRelease",
				},
				AbiFiltersKey: "aurora.abiFilters",
				NdkVersionKey: "ndkVersion",
			}
			pcfg.Artifacts = &config.ArtifactsConfig{
				DebugApk:        "app/build/outputs/apk/debug/app-debug.apk",
				ReleaseUnsigned: "app/build/outputs/apk/release/app-release-unsigned.apk",
				AppName:         info.Name,
			}
		}
		cfgPath := storage.ProjectConfigPath(root, info.Name)
		if !storage.Exists(cfgPath) {
			if err := config.SaveProject(cfgPath, pcfg); err != nil {
				return fmt.Errorf("write project config: %w", err)
			}
		}
		code, name, verr := project.CurrentVersion(info)
		versionStr := "(no version source)"
		if verr == nil && name != "" {
			versionStr = fmt.Sprintf("%s (code %s) from %s", name, code, info.VersionFile)
		} else if info.VersionFile != "" && verr != nil {
			versionStr = fmt.Sprintf("(unreadable: %v)", verr)
		}
		abs, _ := filepath.Abs(info.Root)
		fmt.Printf("Project: %s\nRoot:    %s\nType:    %s\nVersion: %s\nConfig:  %s\n",
			info.Name, abs, info.Type, versionStr, cfgPath)
		return nil
	},
}
