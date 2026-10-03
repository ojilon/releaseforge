// Package storage manages the data-root layout chosen at init time.
//
// Layout (see docs/05-config-and-storage.md):
//
//	<data-root>/
//	  config/global.json
//	  projects/<name>/config.json, builds/, logs/, releases/, reports/, cache/
//	  history/commands.jsonl
//	  global-cache/
//
// The git repo of each project stays clean; heavy artifacts live here.
// Prefer non-system drives (D:) on Windows.
package storage

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// EnvDataRoot is honoured by ResolveRoot when no explicit override is given.
const EnvDataRoot = "RELEASEFORGE_DATA_ROOT"

// DefaultDirName is the folder name created under the chosen drive/home.
const DefaultDirName = "ReleaseForgeData"

// DefaultRoot returns the preferred data-root for a fresh install.
// On Windows it prefers D:/ReleaseForgeData when drive D: exists,
// otherwise it falls back to %USERPROFILE%/ReleaseForgeData.
// On other OSes it uses $HOME/ReleaseForgeData.
func DefaultRoot() string {
	if runtime.GOOS == "windows" {
		if st, err := os.Stat(`D:\`); err == nil && st.IsDir() {
			return filepath.Join(`D:\`, DefaultDirName)
		}
		if home, err := os.UserHomeDir(); err == nil && home != "" {
			return filepath.Join(home, DefaultDirName)
		}
		return filepath.Join(`C:\`, DefaultDirName)
	}
	if home, err := os.UserHomeDir(); err == nil && home != "" {
		return filepath.Join(home, DefaultDirName)
	}
	return filepath.Join(".", DefaultDirName)
}

// ResolveRoot normalises an explicit override, environment variable,
// or the default into an absolute path.
func ResolveRoot(override string) (string, error) {
	candidate := strings.TrimSpace(override)
	if candidate == "" {
		candidate = strings.TrimSpace(os.Getenv(EnvDataRoot))
	}
	if candidate == "" {
		candidate = DefaultRoot()
	}
	abs, err := filepath.Abs(candidate)
	if err != nil {
		return "", err
	}
	return filepath.Clean(abs), nil
}

// GlobalConfigPath returns <root>/config/global.json.
func GlobalConfigPath(root string) string {
	return filepath.Join(root, "config", "global.json")
}

// ProjectDir returns <root>/projects/<sanitized-name>.
func ProjectDir(root, project string) string {
	return filepath.Join(root, "projects", SanitizeProjectName(project))
}

// ProjectConfigPath returns <root>/projects/<name>/config.json.
func ProjectConfigPath(root, project string) string {
	return filepath.Join(ProjectDir(root, project), "config.json")
}

// BuildsDir returns <root>/projects/<name>/builds.
func BuildsDir(root, project string) string {
	return filepath.Join(ProjectDir(root, project), "builds")
}

// LogsDir returns <root>/projects/<name>/logs.
func LogsDir(root, project string) string {
	return filepath.Join(ProjectDir(root, project), "logs")
}

// ReleasesDir returns <root>/projects/<name>/releases.
func ReleasesDir(root, project string) string {
	return filepath.Join(ProjectDir(root, project), "releases")
}

// ReleaseVersionDir returns <root>/projects/<name>/releases/<version>.
func ReleaseVersionDir(root, project, version string) string {
	return filepath.Join(ReleasesDir(root, project), version)
}

// ReportsDir returns <root>/projects/<name>/reports.
func ReportsDir(root, project string) string {
	return filepath.Join(ProjectDir(root, project), "reports")
}

// CacheDir returns <root>/projects/<name>/cache.
func CacheDir(root, project string) string {
	return filepath.Join(ProjectDir(root, project), "cache")
}

// HistoryFile returns <root>/history/commands.jsonl.
func HistoryFile(root string) string {
	return filepath.Join(root, "history", "commands.jsonl")
}

// RecentFile returns <root>/history/recent-projects.json.
func RecentFile(root string) string {
	return filepath.Join(root, "history", "recent-projects.json")
}

// ScanFile returns <root>/projects/<name>/cache/scan.json.
func ScanFile(root, project string) string {
	return filepath.Join(CacheDir(root, project), "scan.json")
}

// GlobalCacheDir returns <root>/global-cache.
func GlobalCacheDir(root string) string {
	return filepath.Join(root, "global-cache")
}

// EnsureLayout creates the top-level data-root directories.
// It is idempotent and safe to call on every startup.
func EnsureLayout(root string) error {
	dirs := []string{
		root,
		filepath.Join(root, "config"),
		filepath.Join(root, "projects"),
		filepath.Join(root, "history"),
		GlobalCacheDir(root),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return err
		}
	}
	return nil
}

// EnsureProjectLayout creates per-project subdirectories.
// It is idempotent.
func EnsureProjectLayout(root, project string) error {
	dirs := []string{
		ProjectDir(root, project),
		BuildsDir(root, project),
		LogsDir(root, project),
		ReleasesDir(root, project),
		ReportsDir(root, project),
		CacheDir(root, project),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return err
		}
	}
	return nil
}

// Exists reports whether path exists (file or directory).
func Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// SanitizeProjectName makes a user-supplied project name safe as a single
// path segment. It replaces path separators and Windows-illegal characters
// with '-', trims trailing spaces/dots (Windows limitation), and falls back
// to "unnamed-project" when nothing remains.
func SanitizeProjectName(name string) string {
	s := strings.TrimSpace(name)
	if s == "" {
		return "unnamed-project"
	}
	// filepath base first so "a/b" collapses to "b"? No — we want to keep
	// it a single segment, so replace separators instead of taking base.
	replacer := strings.NewReplacer(
		"/", "-",
		"\\", "-",
		":", "-",
		"<", "-",
		">", "-",
		"\"", "-",
		"|", "-",
		"?", "-",
		"*", "-",
	)
	s = replacer.Replace(s)
	// Remove control characters.
	s = strings.Map(func(r rune) rune {
		if r < 32 || r == 127 {
			return -1
		}
		return r
	}, s)
	s = strings.TrimSpace(s)
	s = strings.Trim(s, ".")
	if s == "" {
		return "unnamed-project"
	}
	if len([]rune(s)) > 64 {
		s = string([]rune(s)[:64])
		s = strings.Trim(s, ".- ")
		if s == "" {
			return "unnamed-project"
		}
	}
	return s
}

// ProjectNameFromPath derives a default project name from a filesystem path
// (base directory name, sanitized).
func ProjectNameFromPath(dir string) string {
	base := filepath.Base(filepath.Clean(dir))
	return SanitizeProjectName(base)
}
