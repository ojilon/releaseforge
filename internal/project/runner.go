package project

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/ojilon/releaseforge/internal/build"
	"github.com/ojilon/releaseforge/internal/config"
	"github.com/ojilon/releaseforge/internal/storage"
)

// TestSpec is one runnable test kind with its argv.
type TestSpec struct {
	Kind string
	Args []string
}

// Runner executes per-type test/build/version operations. Adding a project
// type means adding a Runner here, not new switches in cmd and the TUI.
type Runner interface {
	// Program is the binary to exec (wrapper path or "go").
	Program() string
	// TestSpecs lists valid test kinds.
	TestSpecs() []TestSpec
	// TestArgs validates kind and returns argv.
	TestArgs(kind string) ([]string, error)
	// BuildArgs validates variant and returns argv (out is the go output
	// path; gradle ignores it). abis is gradle-only: when non-empty it
	// appends -Paurora.abiFilters=<abis> without touching any files.
	BuildArgs(variant, out, abis string) ([]string, error)
	// BuildOutput is the data-root output path a build produces (go), or ""
	// when outputs stay in the project tree (gradle).
	BuildOutput(dataRoot, variant string) string
	// BinaryBaseName is the file-safe base for built binaries.
	BinaryBaseName() string
	// ArtifactGlobs lists repo-relative output locations (gradle APKs).
	ArtifactGlobs(variant string) []string
	// VersionRead returns (code, name); code is "" when the type has none.
	VersionRead() (code, name string, err error)
	// VersionWrite sets the version name (gradle-properties also bumps code).
	VersionWrite(name string) (newCode string, err error)
	// BumpCode increments a numeric version code without renaming.
	// Only gradle.properties supports it; others return an error.
	BumpCode() (newCode string, err error)
}

// GradleRunner implements Runner for Android Gradle projects.
type GradleRunner struct{ info Info }

// GoRunner implements Runner for Go module projects.
type GoRunner struct{ info Info }

// For returns the Runner for info, or an unsupported-type error.
func For(info Info) (Runner, error) {
	switch info.Type {
	case "android-gradle":
		return GradleRunner{info: info}, nil
	case "go":
		return GoRunner{info: info}, nil
	default:
		return nil, fmt.Errorf("project type %q not supported yet (root %s)", info.Type, info.Root)
	}
}

func (r GradleRunner) Program() string { return build.GradleWrapper(r.info.Root) }

func (r GradleRunner) TestSpecs() []TestSpec {
	return []TestSpec{
		{Kind: "unit", Args: []string{":app:testDebugUnitTest"}},
		{Kind: "instrumented", Args: []string{":app:connectedDebugAndroidTest"}},
		{Kind: "all", Args: []string{":app:testDebugUnitTest", ":app:connectedDebugAndroidTest"}},
	}
}

func (r GradleRunner) TestArgs(kind string) ([]string, error) {
	for _, s := range r.TestSpecs() {
		if s.Kind == kind {
			return s.Args, nil
		}
	}
	return nil, fmt.Errorf("test: unknown kind %q (want unit|instrumented|all)", kind)
}

func (r GradleRunner) BuildArgs(variant, _, abis string) ([]string, error) {
	var task string
	switch strings.ToLower(strings.TrimSpace(variant)) {
	case "debug":
		task = "assembleDebug"
	case "release":
		task = "assembleRelease"
	default:
		return nil, fmt.Errorf("build: unknown variant %q (want debug|release)", variant)
	}
	args := []string{task}
	if abis = strings.TrimSpace(abis); abis != "" {
		args = append(args, "-Paurora.abiFilters="+abis)
	}
	return args, nil
}

func (r GradleRunner) BuildOutput(_, _ string) string { return "" }

func (r GradleRunner) BinaryBaseName() string { return BinaryBaseName(r.info) }

func (r GradleRunner) ArtifactGlobs(variant string) []string {
	_ = variant
	return []string{
		"app/build/outputs/apk/debug/app-debug.apk",
		"app/build/outputs/apk/release/app-release-unsigned.apk",
	}
}

func (r GradleRunner) VersionRead() (string, string, error) {
	if r.info.VersionSource == SourceKts {
		v, err := KtsVersion(r.info.Root, r.info.VersionFile)
		if err != nil {
			return "", "", err
		}
		return "", v, nil
	}
	return GradleVersion(r.info.Root, r.info.VersionFile)
}

func (r GradleRunner) VersionWrite(name string) (string, error) {
	if r.info.VersionSource == SourceKts {
		if err := SetKtsVersion(r.info.Root, r.info.VersionFile, name); err != nil {
			return "", err
		}
		return "", nil
	}
	return SetGradleVersion(r.info.Root, r.info.VersionFile, name)
}

func (r GradleRunner) BumpCode() (string, error) {
	if r.info.VersionSource == SourceKts {
		return "", fmt.Errorf("kts versions have no numeric code; use --set")
	}
	return IncrementGradleCode(r.info.Root, r.info.VersionFile)
}

func (r GoRunner) Program() string { return "go" }

func (r GoRunner) TestSpecs() []TestSpec {
	return []TestSpec{{Kind: "unit", Args: build.GoTestArgs()}}
}

func (r GoRunner) TestArgs(_ string) ([]string, error) { return build.GoTestArgs(), nil }

func (r GoRunner) BuildArgs(variant, out, _ string) ([]string, error) {
	v := strings.ToLower(strings.TrimSpace(variant))
	if v != "debug" && v != "release" {
		return nil, fmt.Errorf("build: unknown variant %q (want debug|release)", variant)
	}
	stamp := "dev"
	if _, name, err := CurrentVersion(r.info); err == nil && name != "" {
		stamp = name
	}
	return build.GoBuildArgs(r.info.Root, out, stamp, v), nil
}

func (r GoRunner) BuildOutput(dataRoot, variant string) string {
	base := r.BinaryBaseName()
	if strings.ToLower(strings.TrimSpace(variant)) == "release" {
		base += "-release"
	}
	return filepath.Join(storage.BuildsDir(dataRoot, r.info.Name), build.GoBinaryName(base))
}

func (r GoRunner) BinaryBaseName() string { return BinaryBaseName(r.info) }

func (r GoRunner) ArtifactGlobs(_ string) []string { return nil }

func (r GoRunner) VersionRead() (string, string, error) {
	v, err := ReadVersionFile(r.info.Root, "VERSION")
	if err != nil {
		return "", "", err
	}
	return "", v, nil
}

func (r GoRunner) VersionWrite(name string) (string, error) {
	if err := SetVersionFile(r.info.Root, "VERSION", name); err != nil {
		return "", err
	}
	return "", nil
}

func (r GoRunner) BumpCode() (string, error) {
	return "", fmt.Errorf("go versions have no numeric code; use --set")
}

// SeedConfig builds the first-scan project config for info: tool-owned
// defaults for known types. User-owned sections (signing, github) are filled
// by the merge in doc 07; this only seeds what detection knows.
func SeedConfig(info Info) config.ProjectConfig {
	pcfg := config.ProjectConfig{Type: info.Type, Name: info.Name, Root: info.Root}
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
	return pcfg
}
