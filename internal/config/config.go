// Package config loads/saves global and per-project configuration.
//
// Schemas are defined in docs/05-config-and-storage.md.
// Example JSON lives in configs/.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ojilon/releaseforge/internal/storage"
)

// Supported project types (see docs/02-project-types.md).
const (
	TypeAndroidGradle = "android-gradle"
	TypeWails         = "wails"
	TypeCMake         = "cmake"
	TypePython        = "python"
	TypeJavaCLI       = "java-cli"
	TypeGo            = "go"
	TypeGeneric       = "generic"
)

// SupportedTypes lists all valid project type identifiers.
func SupportedTypes() []string {
	return []string{
		TypeAndroidGradle,
		TypeWails,
		TypeCMake,
		TypePython,
		TypeJavaCLI,
		TypeGo,
		TypeGeneric,
	}
}

// AIConfig holds optional AI-assisted release-note settings.
// Disabled by default in v1 (see docs/00-vision.md non-goals).
type AIConfig struct {
	Enabled  bool   `json:"enabled"`
	Provider string `json:"provider"`
	Model    string `json:"model"`
}

// GlobalConfig is stored at <data-root>/config/global.json.
type GlobalConfig struct {
	DataRoot       string  `json:"data_root"`
	GithubTokenEnv string  `json:"github_token_env"`
	PreferredDevice *string `json:"preferred_device"`
	Editor         string  `json:"editor"`
	AI             AIConfig `json:"ai"`
}

// VersionConfig describes where a project's version lives.
type VersionConfig struct {
	File     string `json:"file"`
	CodeKey  string `json:"code_key,omitempty"`
	NameKey  string `json:"name_key,omitempty"`
	JsonPath string `json:"json_path,omitempty"`
}

// BuildConfig describes logical build tasks and platform keys.
type BuildConfig struct {
	GradleWrapper bool              `json:"gradle_wrapper,omitempty"`
	Tasks         map[string]string `json:"tasks,omitempty"`
	AbiFiltersKey string            `json:"abi_filters_key,omitempty"`
	NdkVersionKey string            `json:"ndk_version_key,omitempty"`
	OutputDir     string            `json:"output_dir,omitempty"`
}

// SigningConfig holds Android signing settings.
// Passwords are never stored; Apksigner may be null (auto-detect).
type SigningConfig struct {
	Keystore string  `json:"keystore,omitempty"`
	Alias    string  `json:"alias,omitempty"`
	Apksigner *string `json:"apksigner"`
}

// ArtifactsConfig lists known artifact locations relative to project root.
type ArtifactsConfig struct {
	DebugApk        string `json:"debug_apk,omitempty"`
	ReleaseUnsigned string `json:"release_unsigned,omitempty"`
	AppName         string `json:"app_name,omitempty"`
}

// GithubConfig identifies the upstream repo for releases.
type GithubConfig struct {
	Owner string `json:"owner,omitempty"`
	Repo  string `json:"repo,omitempty"`
}

// ProjectConfig is stored at <data-root>/projects/<name>/config.json
// (and optionally seeded from a repo-local .releaseforge.json).
type ProjectConfig struct {
	Type                   string           `json:"type"`
	Name                   string           `json:"name"`
	Root                   string           `json:"root"`
	Version                VersionConfig    `json:"version"`
	Build                  BuildConfig      `json:"build,omitempty"`
	Signing                *SigningConfig   `json:"signing,omitempty"`
	Artifacts              *ArtifactsConfig `json:"artifacts,omitempty"`
	Github                 *GithubConfig    `json:"github,omitempty"`
	MirrorReleaseDirInRepo bool             `json:"mirror_release_dir_in_repo,omitempty"`
}

// DefaultGlobal returns a GlobalConfig with sane defaults for a new data root.
func DefaultGlobal(dataRoot string) GlobalConfig {
	return GlobalConfig{
		DataRoot:       dataRoot,
		GithubTokenEnv: "GITHUB_TOKEN",
		PreferredDevice: nil,
		Editor:         "",
		AI:             AIConfig{Enabled: false},
	}
}

// Validate checks required global fields.
func (g GlobalConfig) Validate() error {
	if strings.TrimSpace(g.DataRoot) == "" {
		return fmt.Errorf("data_root must not be empty")
	}
	return nil
}

// Validate checks required project fields.
func (p ProjectConfig) Validate() error {
	if strings.TrimSpace(p.Name) == "" {
		return fmt.Errorf("project name must not be empty")
	}
	if strings.TrimSpace(p.Root) == "" {
		return fmt.Errorf("project root must not be empty")
	}
	valid := false
	for _, t := range SupportedTypes() {
		if p.Type == t {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("unknown project type %q (want one of %s)", p.Type, strings.Join(SupportedTypes(), ", "))
	}
	if p.Type != TypeGeneric && strings.TrimSpace(p.Version.File) == "" {
		return fmt.Errorf("version.file must not be empty (type %q)", p.Type)
	}
	return nil
}

// LoadGlobal reads and parses a global.json file.
func LoadGlobal(path string) (GlobalConfig, error) {
	var cfg GlobalConfig
	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, err
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("parse %s: %w", path, err)
	}
	if err := cfg.Validate(); err != nil {
		return cfg, fmt.Errorf("invalid %s: %w", path, err)
	}
	return cfg, nil
}

// SaveGlobal validates, creates parent dirs, and writes global.json atomically.
func SaveGlobal(path string, cfg GlobalConfig) error {
	if err := cfg.Validate(); err != nil {
		return err
	}
	return writeJSON(path, cfg, 0o644)
}

// LoadProject reads and parses a per-project config.json.
func LoadProject(path string) (ProjectConfig, error) {
	var cfg ProjectConfig
	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, err
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("parse %s: %w", path, err)
	}
	if err := cfg.Validate(); err != nil {
		return cfg, fmt.Errorf("invalid %s: %w", path, err)
	}
	return cfg, nil
}

// SaveProject validates, creates parent dirs, and writes config.json atomically.
func SaveProject(path string, cfg ProjectConfig) error {
	if err := cfg.Validate(); err != nil {
		return err
	}
	return writeJSON(path, cfg, 0o644)
}

// DiscoverDataRoot resolves the active data root and its global.json path.
//
// Precedence:
//  1. dataRootOverride (--data-root flag)
//  2. cfgFileOverride (--config flag): if the file exists, its data_root wins;
//     otherwise the root is inferred as the grandparent of the given path
//     (<root>/config/global.json -> <root>).
//  3. $RELEASEFORGE_DATA_ROOT env var
//  4. Installed layout: <install>/bin/<exe> with <install>/data seeded by
//     the installer (zero-config first run after installation).
//  5. storage.DefaultRoot(); if a global.json already exists there, its
//     data_root value is canonical (supports relocated roots).
func DiscoverDataRoot(cfgFileOverride, dataRootOverride string) (root string, configPath string, err error) {
	if strings.TrimSpace(dataRootOverride) != "" {
		root, err = storage.ResolveRoot(dataRootOverride)
		if err != nil {
			return "", "", err
		}
		return root, storage.GlobalConfigPath(root), nil
	}

	if strings.TrimSpace(cfgFileOverride) != "" {
		abs, err := filepath.Abs(strings.TrimSpace(cfgFileOverride))
		if err != nil {
			return "", "", err
		}
		if data, readErr := os.ReadFile(abs); readErr == nil {
			var cfg GlobalConfig
			if jsonErr := json.Unmarshal(data, &cfg); jsonErr == nil && strings.TrimSpace(cfg.DataRoot) != "" {
				root, err = storage.ResolveRoot(cfg.DataRoot)
				if err != nil {
					return "", "", err
				}
				return root, storage.GlobalConfigPath(root), nil
			}
		}
		// Infer root from <root>/config/global.json layout.
		inferred := filepath.Dir(filepath.Dir(abs))
		root, err = storage.ResolveRoot(inferred)
		if err != nil {
			return "", "", err
		}
		return root, abs, nil
	}

	// Env var, installed layout, or default.
	root, err = storage.ResolveRoot("")
	if err != nil {
		return "", "", err
	}
	if strings.TrimSpace(os.Getenv(storage.EnvDataRoot)) == "" {
		if sib, ok := dataRootFromExe(); ok {
			root = sib
		}
	}
	configPath = storage.GlobalConfigPath(root)
	if data, readErr := os.ReadFile(configPath); readErr == nil {
		var cfg GlobalConfig
		if jsonErr := json.Unmarshal(data, &cfg); jsonErr == nil && strings.TrimSpace(cfg.DataRoot) != "" {
			canonical, resErr := storage.ResolveRoot(cfg.DataRoot)
			if resErr != nil {
				return "", "", resErr
			}
			return canonical, storage.GlobalConfigPath(canonical), nil
		}
	}
	return root, configPath, nil
}

// dataRootFromExe discovers the installer-seeded data root: when the running
// binary lives in <install>/bin/, and <install>/data/config/global.json
// exists, that data dir wins over the machine default (flags and env still
// win over it). This makes the first run after installation zero-config.
func dataRootFromExe() (string, bool) {
	return dataRootFromExePath(exePath())
}

func exePath() string {
	if p, err := os.Executable(); err == nil && strings.TrimSpace(p) != "" {
		return p
	}
	return os.Args[0]
}

func dataRootFromExePath(exe string) (string, bool) {
	dir := filepath.Dir(exe)
	if !strings.EqualFold(filepath.Base(dir), "bin") {
		return "", false
	}
	data := filepath.Join(filepath.Dir(dir), "data")
	if st, err := os.Stat(filepath.Join(data, "config", "global.json")); err != nil || st.IsDir() {
		return "", false
	}
	abs, err := filepath.Abs(data)
	if err != nil {
		return "", false
	}
	return filepath.Clean(abs), true
}

func writeJSON(path string, v any, perm os.FileMode) error {	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, perm); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
