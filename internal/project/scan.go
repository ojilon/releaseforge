package project

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ojilon/releaseforge/internal/git"
	"github.com/ojilon/releaseforge/internal/storage"
)

// Snapshot is the persisted scan result (docs/09-scan-foundation.md schema v1).
// Unknown fields are ignored by older readers.
type Snapshot struct {
	ScannedAt  string         `json:"scanned_at"`
	Root       string         `json:"root"`
	Name       string         `json:"name"`
	Type       string         `json:"type"`
	Git        GitSnapshot    `json:"git"`
	Tools      []ToolHit      `json:"tools"`
	Frameworks []ToolHit      `json:"frameworks"`
	Configs    []ConfigHit    `json:"configs"`
	Version    VersionSnap    `json:"version"`
	Hints      map[string]bool `json:"hints"`
}

// GitSnapshot holds local git state.
type GitSnapshot struct {
	Present       bool     `json:"present"`
	Branch        string   `json:"branch,omitempty"`
	Head          string   `json:"head,omitempty"`
	OriginURL     string   `json:"origin_url,omitempty"`
	RecentCommits []Commit `json:"recent_commits,omitempty"`
	RecentTags    []string `json:"recent_tags,omitempty"`
}

// Commit is one recent commit.
type Commit struct {
	SHA     string `json:"sha"`
	Subject string `json:"subject"`
	Date    string `json:"date"`
}

// ToolHit records a detected tool/framework and its evidence.
type ToolHit struct {
	ID       string   `json:"id"`
	Evidence []string `json:"evidence"`
}

// ConfigHit records an interesting config file.
type ConfigHit struct {
	Path string `json:"path"`
	Role string `json:"role"`
}

// VersionSnap records the detected version.
type VersionSnap struct {
	File string `json:"file,omitempty"`
	Name string `json:"name,omitempty"`
	Code string `json:"code,omitempty"`
}

// Scan runs the ordered scan pipeline: resolve → git → markers → hints.
func Scan(dir string) (Snapshot, Info, error) {
	info, err := Detect(dir)
	if err != nil {
		return Snapshot{}, Info{}, err
	}
	snap := Snapshot{
		ScannedAt: time.Now().UTC().Format(time.RFC3339),
		Root:      info.Root,
		Name:      info.Name,
		Type:      info.Type,
		Hints:     map[string]bool{},
	}

	// Git (read-only, best-effort).
	if git.IsRepo(info.Root) {
		snap.Git.Present = true
		snap.Git.Branch = git.CurrentBranch(info.Root)
		snap.Git.Head = git.Head(info.Root)
		snap.Git.OriginURL = git.Origin(info.Root)
		if commits, err := git.Log(info.Root, 30); err == nil {
			for _, c := range commits {
				snap.Git.RecentCommits = append(snap.Git.RecentCommits, Commit{SHA: c.SHA, Subject: c.Subject, Date: c.Date})
			}
		}
		snap.Git.RecentTags = git.RecentTags(info.Root, 10)
	}

	// Markers → tools/frameworks/configs.
	addTool := func(id string, evidence ...string) {
		snap.Tools = append(snap.Tools, ToolHit{ID: id, Evidence: evidence})
	}
	has := func(rel string) bool {
		_, err := os.Stat(filepath.Join(info.Root, rel))
		return err == nil
	}
	markConfig := func(rel, role string) {
		if has(rel) {
			snap.Configs = append(snap.Configs, ConfigHit{Path: rel, Role: role})
		}
	}

	if has("gradlew") || has("gradlew.bat") {
		snap.Hints["has_wrapper"] = true
	}
	if has("settings.gradle") || has("settings.gradle.kts") {
		var ev []string
		for _, f := range []string{"settings.gradle", "settings.gradle.kts", "gradlew", "gradlew.bat"} {
			if has(f) {
				ev = append(ev, f)
			}
		}
		addTool("gradle", ev...)
	}
	for _, c := range []string{"CMakeLists.txt", "backend/CMakeLists.txt", "native/CMakeLists.txt"} {
		if has(c) {
			addTool("cmake", c)
			break
		}
	}
	if has("wails.json") {
		addTool("wails", "wails.json")
	}
	if has("go.mod") {
		addTool("go", "go.mod")
	}
	for _, f := range []string{"package.json", "pnpm-lock.yaml", "yarn.lock", "package-lock.json"} {
		if has(f) {
			addTool("node", f)
			break
		}
	}
	if has("pom.xml") || has("mvnw") {
		addTool("maven", "pom.xml")
	}
	for _, f := range []string{"pyproject.toml", "requirements.txt", "setup.py"} {
		if has(f) {
			addTool("python", f)
			break
		}
	}
	if has("Cargo.toml") {
		addTool("rust", "Cargo.toml")
	}
	if has("scripts/release.py") {
		snap.Hints["legacy_release_scripts"] = true
	}

	// Framework hints.
	if data, err := os.ReadFile(filepath.Join(info.Root, "app", "build.gradle")); err == nil {
		if strings.Contains(string(data), "com.android.application") {
			snap.Frameworks = append(snap.Frameworks, ToolHit{ID: "android", Evidence: []string{"com.android.application"}})
		}
	}

	// Config inventory.
	markConfig("gradle.properties", "version+android")
	markConfig("app/build.gradle", "android-module")
	markConfig("app/build.gradle.kts", "android-module")
	markConfig("settings.gradle", "gradle-settings")
	markConfig("wails.json", "wails")
	markConfig("go.mod", "go-module")
	markConfig("VERSION", "version")
	markConfig("CMakeLists.txt", "cmake")
	markConfig("pyproject.toml", "python")
	markConfig("pom.xml", "maven")
	markConfig(".releaseforge.json", "tool-seed")
	if entries, err := os.ReadDir(filepath.Join(info.Root, "scripts")); err == nil {
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".py") {
				snap.Configs = append(snap.Configs, ConfigHit{Path: "scripts/" + e.Name(), Role: "script"})
			}
		}
	}

	// Version hints.
	snap.Version.File = info.VersionFile
	if code, name, err := CurrentVersion(info); err == nil {
		snap.Version.Code = code
		snap.Version.Name = name
	}
	if info.Type == "wails" {
		if v := wailsProductVersion(filepath.Join(info.Root, "wails.json")); v != "" && snap.Version.Name == "" {
			snap.Version.Name = v
		}
	}
	if data, err := os.ReadFile(filepath.Join(info.Root, "gradle.properties")); err == nil {
		if strings.Contains(string(data), "ndkVersion") {
			addTool("ndk", "ndkVersion in gradle.properties")
		}
	}

	return snap, info, nil
}

// WriteScan persists the snapshot to cache/scan.json (creating dirs).
func WriteScan(dataRoot string, snap Snapshot) (string, error) {
	if err := storage.EnsureProjectLayout(dataRoot, snap.Name); err != nil {
		return "", err
	}
	path := storage.ScanFile(dataRoot, snap.Name)
	data, err := json.MarshalIndent(snap, "", "  ")
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		return "", err
	}
	return path, nil
}

func wailsProductVersion(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	var v struct {
		Info struct {
			ProductVersion string `json:"productVersion"`
		} `json:"info"`
	}
	if err := json.Unmarshal(data, &v); err != nil {
		return ""
	}
	return strings.TrimSpace(v.Info.ProductVersion)
}
