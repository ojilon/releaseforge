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
	rflog "github.com/ojilon/releaseforge/internal/log"
	"github.com/ojilon/releaseforge/internal/project"
	"github.com/ojilon/releaseforge/internal/storage"
	"github.com/spf13/cobra"
)

var (
	releasePre       bool
	releaseSkipTests bool
	releaseNotesOnly bool
)

var releaseCmd = &cobra.Command{
	Use:   "release <version>",
	Short: "Full release pipeline: version → test → build → package → notes → tag → GitHub",
	Long: `Android Gradle:
  1. set version (gradle.properties, code+1, no commit)
  2. unit tests (unless --skip-tests)
  3. assembleDebug + assembleRelease
  4. package APKs into data-root releases/<version>/
  5. sign release APK (password prompted once, never stored)
  6. notes from git history
  7. zip artifacts
  8. tag v<version>, push, gh release create

Go (e.g. ReleaseForge itself):
  1. write VERSION file (no commit — commit manually)
  2. go test ./... (unless --skip-tests)
  3. go build -trimpath with version stamp into releases/<version>/
  4. notes from git history
  5. zip binary
  6. tag v<version>, push, gh release create

Without the gh CLI, release stops after the local tag + zip
and reports the artifact paths (exit 0).`,
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
		r, err := project.For(info)
		if err != nil {
			return fmt.Errorf("release: %w", err)
		}
		switch r.(type) {
		case project.GradleRunner:
			return releaseAndroid(info, r, version)
		case project.GoRunner:
			return releaseGo(info, r, version)
		default:
			return fmt.Errorf("release: project type %q not supported yet", info.Type)
		}
	},
}

func releaseAndroid(info project.Info, r project.Runner, version string) error {
	root, _, err := requireDataRoot()
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

	fmt.Printf("==> version %s (gradle.properties, not committed)\n", version)
	if _, err := r.VersionWrite(version); err != nil {
		return fmt.Errorf("set version: %w", err)
	}

	prog := r.Program()
	unitArgs, err := r.TestArgs("unit")
	if err != nil {
		return err
	}
	debugArgs, err := r.BuildArgs("debug", "")
	if err != nil {
		return err
	}
	releaseArgs, err := r.BuildArgs("release", "")
	if err != nil {
		return err
	}
	if !releaseSkipTests {
		if err := runStep(root, info, "test-unit", prog, unitArgs...); err != nil {
			return err
		}
	}
	if err := runStep(root, info, "build-debug", prog, debugArgs...); err != nil {
		return err
	}
	if err := runStep(root, info, "build-release", prog, releaseArgs...); err != nil {
		return err
	}

	fmt.Printf("==> package %s\n", verDir)
	debugApk, releaseApk, err := android.PackageRelease(info.Root, version, appName, verDir)
	if err != nil {
		return err
	}
	fmt.Printf("  debug:   %s\n  release: %s\n", debugApk, releaseApk)

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

	return publishTagAndRelease(info.Root, version, notesPath, []string{debugApk, releaseApk, zipPath}, verDir)
}

func releaseGo(info project.Info, r project.Runner, version string) error {
	root, _, err := requireDataRoot()
	if err != nil {
		return err
	}
	if err := storage.EnsureProjectLayout(root, info.Name); err != nil {
		return err
	}
	verDir := storage.ReleaseVersionDir(root, info.Name, version)

	if releaseNotesOnly {
		notesPath, err := writeNotes(info.Root, verDir, info.Name, version, releasePre)
		if err != nil {
			return err
		}
		fmt.Printf("notes: %s\n", notesPath)
		return nil
	}

	if git.IsRepo(info.Root) && !git.IsClean(info.Root) {
		fmt.Fprintln(os.Stderr, "warning: working tree is dirty — release continues, but consider committing first")
	}

	fmt.Printf("==> version %s (VERSION file, not committed)\n", version)
	if _, err := r.VersionWrite(version); err != nil {
		return fmt.Errorf("set version: %w", err)
	}

	if !releaseSkipTests {
		testArgs, err := r.TestArgs("unit")
		if err != nil {
			return err
		}
		if err := runStep(root, info, "test", r.Program(), testArgs...); err != nil {
			return err
		}
	}

	out := filepath.Join(verDir, build.GoBinaryName(r.BinaryBaseName()+"-"+version))
	fmt.Printf("==> build release → %s\n", out)
	buildArgs, err := r.BuildArgs("release", out)
	if err != nil {
		return err
	}
	logPath := rflog.LogPath(storage.LogsDir(root, info.Name), "build-release")
	res := build.Run(r.Program(), buildArgs, build.Options{
		Dir:     info.Root,
		LogPath: logPath,
		OnLine:  func(t string, _ bool) { fmt.Println(t) },
	})
	fmt.Printf("log: %s\n", res.LogPath)
	if !res.Success {
		for _, e := range res.Errors {
			fmt.Printf("  ! %s\n", e)
		}
		return fmt.Errorf("build release failed (exit %d)", res.ExitCode)
	}
	fmt.Printf("binary: %s\n", out)

	notesPath, err := writeNotes(info.Root, verDir, info.Name, version, releasePre)
	if err != nil {
		return err
	}
	fmt.Printf("notes: %s\n", notesPath)

	zipPath := filepath.Join(verDir, fmt.Sprintf("%s-%s.zip", info.Name, version))
	if err := zipFiles(zipPath, []string{out}); err != nil {
		return fmt.Errorf("zip: %w", err)
	}
	fmt.Printf("zip: %s\n", zipPath)

	return publishTagAndRelease(info.Root, version, notesPath, []string{out, zipPath}, verDir)
}

// runStep executes a build/test step with live output + persisted log.
func runStep(dataRoot string, info project.Info, prefix, prog string, args ...string) error {
	logPath := rflog.LogPath(storage.LogsDir(dataRoot, info.Name), prefix)
	fmt.Printf("==> %s (log %s)\n", prefix, logPath)
	res := build.Run(prog, args, build.Options{
		Dir:     info.Root,
		LogPath: logPath,
		OnLine:  func(t string, _ bool) { fmt.Println(t) },
	})
	if !res.Success {
		printReport(res.Report)
		for _, e := range res.Errors {
			fmt.Printf("  ! %s\n", e)
		}
		return fmt.Errorf("%s failed (exit %d, log %s)", prefix, res.ExitCode, res.LogPath)
	}
	return nil
}

// publishTagAndRelease creates + pushes tag v<version>, then publishes via gh.
// Without gh on PATH it stops after the local tag + zip (exit 0) and prints
// the artifact paths so the user can publish manually later.
func publishTagAndRelease(projectRoot, version, notesPath string, artifacts []string, verDir string) error {
	tag := "v" + strings.TrimPrefix(version, "v")
	if !git.IsRepo(projectRoot) {
		fmt.Println("not a git repo — skipping tag + GitHub publish")
		fmt.Printf("release %s done locally. Artifacts under %s\n", version, verDir)
		return nil
	}
	fmt.Printf("==> tag %s\n", tag)
	if err := github.CreateTag(projectRoot, tag, fmt.Sprintf("Release %s", version)); err != nil {
		return err
	}
	if err := github.PushTag(projectRoot, tag); err != nil {
		return fmt.Errorf("%w (local tag %s kept; push manually with `git push origin %s`)", err, tag, tag)
	}
	if !github.HasGH() {
		fmt.Printf("gh CLI not found — stopping after local tag + artifacts (exit 0).\n")
		fmt.Printf("tag: %s (pushed)\nnotes: %s\n", tag, notesPath)
		for _, a := range artifacts {
			fmt.Printf("artifact: %s\n", a)
		}
		fmt.Printf("publish later: gh release create %s --notes-file %s %s\n", tag, notesPath, strings.Join(artifacts, " "))
		return nil
	}
	fmt.Printf("==> github release %s\n", tag)
	if err := github.CreateRelease(projectRoot, tag, notesPath, artifacts, releasePre); err != nil {
		return err
	}
	fmt.Printf("release %s done. Artifacts under %s\n", version, verDir)
	return nil
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
