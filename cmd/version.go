package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/ojilon/releaseforge/internal/project"
	"github.com/ojilon/releaseforge/internal/version"
	"github.com/spf13/cobra"
)

var (
	versionSet      string
	versionCodeOnly bool
)

var versionCmd = &cobra.Command{
	Use:   "version [bump <name>]",
	Short: "Show or set project version in its source of truth",
	Long: `Shows the tool version plus the open project's version.

Android Gradle: gradle.properties (app.versionCode/app.versionName) or a
  plain version = "…" literal in build.gradle.kts.
Go: VERSION file at project root (no auto-commit; commit manually).
Use --set / bump to write (never commits); --code-only bumps the numeric
code without renaming (gradle.properties only).`,
	Args: cobra.MaximumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("releaseforge: %s\n", version.ToolVersion)
		info, err := project.Detect(projectDir)
		if err != nil {
			return err
		}
		if info.Type == "android-gradle" && info.VersionSource == project.SourceKts {
			fmt.Printf("source: %s\n", info.VersionFile)
		}
		r, err := project.For(info)
		if err != nil {
			return fmt.Errorf("version: %w (root %s)", err, info.Root)
		}
		setName := strings.TrimSpace(versionSet)
		if setName == "" && len(args) == 2 && args[0] == "bump" {
			setName = strings.TrimSpace(args[1])
		}
		if versionCodeOnly && setName != "" {
			return fmt.Errorf("version: --code-only and --set/bump are exclusive")
		}
		switch {
		case versionCodeOnly:
			next, err := r.BumpCode()
			if err != nil {
				return err
			}
			fmt.Printf("versionCode: %s\n", next)
			return nil
		case setName != "":
			if err := project.ValidateVersion(setName); err != nil {
				return err
			}
			warnSemver(setName)
			next, err := r.VersionWrite(setName)
			if err != nil {
				return err
			}
			if next == "" {
				fmt.Printf("version: %s (wrote file, not committed)\n", setName)
				return nil
			}
			printVersion(next, setName)
			return nil
		default:
			if len(args) > 0 {
				return fmt.Errorf("version: usage: version [--set X | bump X | --code-only]")
			}
			code, name, err := r.VersionRead()
			if err != nil {
				return err
			}
			printVersion(code, name)
			return nil
		}
	},
}

// warnSemver notes non-semver-like names on stderr without failing.
func warnSemver(v string) {
	if !project.LooksSemver(v) {
		fmt.Fprintf(os.Stderr, "warning: %q is not semver-like (X.Y[.Z]) — continuing\n", v)
	}
}

// printVersion renders a version with a code line only when the project type
// has a numeric code (gradle); otherwise it prints the plain version.
func printVersion(code, name string) {
	if code != "" {
		fmt.Printf("versionCode: %s\nversionName: %s\n", code, name)
		return
	}
	fmt.Printf("version: %s\n", name)
}

func init() {
	versionCmd.Flags().StringVar(&versionSet, "set", "", "new version (gradle: also increments code; go/kts: writes name, no commit)")
	versionCmd.Flags().BoolVar(&versionCodeOnly, "code-only", false, "bump numeric code only (gradle.properties)")
}
