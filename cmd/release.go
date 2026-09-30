package cmd

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/ojilon/releaseforge/internal/android"
	"github.com/ojilon/releaseforge/internal/build"
	"github.com/ojilon/releaseforge/internal/config"
	"github.com/ojilon/releaseforge/internal/git"
	"github.com/ojilon/releaseforge/internal/github"
	"github.com/ojilon/releaseforge/internal/project"
	"github.com/ojilon/releaseforge/internal/storage"
	rflog "github.com/ojilon/releaseforge/internal/log"
	"github.com/spf13/cobra"
)

var (
	releasePre       bool
	releaseSkipTests bool
	releaseNotesOnly bool
)

var releaseCmd = &cobra.Command{
	Use:   "release <version>",
	Short: "Full release pipeline: version → test → build → package → sign → notes → tag → GitHub",
	Long: `Orchestrates the same flow currently done by Conductino-Android scripts/release.py,
generalised for all supported project types.

Steps (Android):
  1. set version (gradle.properties)
  2. unit tests
  3. assembleDebug + assembleRelease
  4. package into release/<version>/
  5. sign release APK (apksigner + keystore)
  6. generate notes from git history (template + commits)
  7. zip artifacts
  8. create annotated tag and GitHub release (gh)

Flags control pre-release marking and optional skips.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		version := strings.TrimSpace(args[0])
		if version == "" {
			return fmt.Errorf("release: version required")
		}
		info, err := project.Detect(projectDir)
		if err != nil {
			return err
		}
		if info.Type != "android-gradle" {
			return fmt.Errorf("release: project type %q not supported yet", info.Type)
		}
		root, _, err := resolveDataRoot()
		if err != nil {
			return err
		}
		if err := storage.EnsureProjectLayout(root, info.Name); err != nil {
			return err
		}
		verDir := storage.ReleaseVersionDir(root, info.Name, version)
		appName := info.Name
		var pcfg config.ProjectConfig
		if loaded, err := config.LoadProject(storage.ProjectConfigPath(root, info.Name)); err == nil {
			pcfg = loaded
			if pcfg.Artifacts != nil && pcfg.Artifacts.AppName != "" {
				appName = pcfg.Artifacts.AppName
			}
		}

		// Notes-only mode: draft notes from git history without building.
		if releaseNotesOnly {
			notesPath, err := writeNotes(info.Root, verDir, appName, version, releasePre)
			if err != nil {
				return err
			}
			fmt.Printf("notes: %s\n", notesPath)
			return nil
		}

		if git.IsRepo(info.Root) && !git.IsClean(info.Root) {
			fmt.Fprintln(os.Stderr, "warning: working tree is dirty — release continues, but consider committing first")
		}

		fmt.Printf("==> version %s\n", version)
		if _, err := project.SetGradleVersion(info.Root, info.VersionFile, version); err != nil {
			return fmt.Errorf("set version: %w", err)
		}

		wrapper := build.GradleWrapper(info.Root)
		run := func(prefix string, tasks ...string) error {
			logPath := rflog.LogPath(storage.LogsDir(root, info.Name), prefix)
			fmt.Printf("==> %s (log %s)\n", prefix, logPath)
			res := build.Run(wrapper, tasks, build.Options{
				Dir:     info.Root,
				LogPath: logPath,
				OnLine:  func(t string, _ bool) { fmt.Println(t) },
			})
			if !res.Success {
				for _, e := range res.Errors {
					fmt.Printf("  ! %s\n", e)
				}
				return fmt.Errorf("%s failed (exit %d, log %s)", prefix, res.ExitCode, res.LogPath)
			}
			return nil
		}

		if !releaseSkipTests {
			if err := run("test-unit", ":app:testDebugUnitTest"); err != nil {
				return err
			}
		}
		if err := run("build-debug", "assembleDebug"); err != nil {
			return err
		}
		if err := run("build-release", "assembleRelease"); err != nil {
			return err
		}

		fmt.Printf("==> package %s\n", verDir)
		debugApk, releaseApk, err := android.PackageRelease(info.Root, version, appName, verDir)
		if err != nil {
			return err
		}
		fmt.Printf("  debug:   %s\n  release: %s\n", debugApk, releaseApk)

		// Sign the release APK (password prompted once, never stored).
		keystore, alias, apksignerBin := "conductino-release.jks", "conductino", ""
		if pcfg.Signing != nil {
			if pcfg.Signing.Keystore != "" {
				keystore = pcfg.Signing.Keystore
			}
			if pcfg.Signing.Alias != "" {
				alias = pcfg.Signing.Alias
			}
			if pcfg.Signing.Apksigner != nil {
				apksignerBin = *pcfg.Signing.Apksigner
			}
		}
		if !filepath.IsAbs(keystore) {
			keystore = filepath.Join(info.Root, keystore)
		}
		var password string
		form := huh.NewForm(huh.NewGroup(
			huh.NewInput().Title("Keystore password").EchoMode(huh.EchoModePassword).Value(&password),
		))
		if err := form.Run(); err != nil {
			return fmt.Errorf("password prompt: %w", err)
		}
		fmt.Printf("==> sign %s\n", releaseApk)
		if err := android.SignApk(releaseApk, keystore, alias, password, apksignerBin); err != nil {
			return err
		}

		notesPath, err := writeNotes(info.Root, verDir, appName, version, releasePre)
		if err != nil {
			return err
		}
		fmt.Printf("notes: %s\n", notesPath)

		zipPath := filepath.Join(verDir, fmt.Sprintf("%s-%s.zip", appName, version))
		if err := zipFiles(zipPath, []string{debugApk, releaseApk}); err != nil {
			return fmt.Errorf("zip: %w", err)
		}
		fmt.Printf("zip: %s\n", zipPath)

		tag := "v" + strings.TrimPrefix(version, "v")
		if git.IsRepo(info.Root) {
			fmt.Printf("==> tag %s\n", tag)
			if err := github.CreateTag(info.Root, tag, fmt.Sprintf("Release %s", version)); err != nil {
				return err
			}
			if err := github.PushTag(info.Root, tag); err != nil {
				return err
			}
			fmt.Printf("==> github release %s\n", tag)
			if err := github.CreateRelease(info.Root, tag, notesPath, []string{debugApk, releaseApk, zipPath}, releasePre); err != nil {
				return err
			}
		} else {
			fmt.Println("not a git repo — skipping tag + GitHub publish")
		}
		fmt.Printf("release %s done. Artifacts under %s\n", version, verDir)
		return nil
	},
}

func writeNotes(projectRoot, verDir, appName, version string, prerelease bool) (string, error) {
	var commits []git.Commit
	if git.IsRepo(projectRoot) {
		prev := git.LatestTag(projectRoot)
		if cl, err := git.LogSince(projectRoot, prev); err == nil {
			commits = cl
		}
	}
	body := git.DraftNotes(appName, version, commits, prerelease)
	if err := os.MkdirAll(verDir, 0o755); err != nil {
		return "", err
	}
	path := filepath.Join(verDir, "notes.md")
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		return "", err
	}
	return path, nil
}

func zipFiles(dst string, files []string) error {
	f, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer f.Close()
	w := zip.NewWriter(f)
	defer w.Close()
	for _, src := range files {
		in, err := os.Open(src)
		if err != nil {
			return err
		}
		fw, err := w.Create(filepath.Base(src))
		if err != nil {
			in.Close()
			return err
		}
		if _, err := io.Copy(fw, in); err != nil {
			in.Close()
			return err
		}
		in.Close()
	}
	return nil
}

func init() {
	releaseCmd.Flags().BoolVar(&releasePre, "pre", false, "mark GitHub release as pre-release")
	releaseCmd.Flags().BoolVar(&releaseSkipTests, "skip-tests", false, "skip test step (not recommended)")
	releaseCmd.Flags().BoolVar(&releaseNotesOnly, "notes-only", false, "only generate/update release notes")
}
