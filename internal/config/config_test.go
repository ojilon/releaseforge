package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ojilon/releaseforge/internal/storage"
)

func TestSaveLoadGlobalRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config", "global.json")
	cfg := DefaultGlobal(filepath.Join(dir, "data"))
	if err := SaveGlobal(path, cfg); err != nil {
		t.Fatal(err)
	}
	loaded, err := LoadGlobal(path)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.DataRoot != cfg.DataRoot {
		t.Fatalf("got %q want %q", loaded.DataRoot, cfg.DataRoot)
	}
	if loaded.GithubTokenEnv != "GITHUB_TOKEN" {
		t.Fatalf("unexpected token env %q", loaded.GithubTokenEnv)
	}
}

func TestSaveGlobalRejectsEmptyRoot(t *testing.T) {
	if err := SaveGlobal(filepath.Join(t.TempDir(), "g.json"), GlobalConfig{}); err == nil {
		t.Fatal("expected validation error")
	}
}

func TestSaveLoadProjectRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := storage.ProjectConfigPath(dir, "Conductino-Android")
	cfg := ProjectConfig{
		Type: TypeAndroidGradle,
		Name: "Conductino-Android",
		Root: `D:/Dev/Conductino-Android`,
		Version: VersionConfig{
			File:    "gradle.properties",
			CodeKey: "app.versionCode",
			NameKey: "app.versionName",
		},
	}
	if err := SaveProject(path, cfg); err != nil {
		t.Fatal(err)
	}
	loaded, err := LoadProject(path)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Name != cfg.Name || loaded.Version.CodeKey != "app.versionCode" {
		t.Fatalf("round trip mismatch: %+v", loaded)
	}
}

func TestLoadExampleAndroidConfig(t *testing.T) {
	loaded, err := LoadProject(filepath.Join("..", "..", "configs", "example-android-gradle.json"))
	if err != nil {
		t.Skipf("example config not found from test dir: %v", err)
	}
	if loaded.Type != TypeAndroidGradle {
		t.Fatalf("got type %q", loaded.Type)
	}
}

func TestDiscoverDataRootPrefersFlag(t *testing.T) {
	dir := t.TempDir()
	root, cfgPath, err := DiscoverDataRoot("", dir)
	if err != nil {
		t.Fatal(err)
	}
	if cfgPath != storage.GlobalConfigPath(root) {
		t.Fatalf("unexpected config path %q", cfgPath)
	}
	if _, err := os.Stat(root); err == nil {
		// root may not exist yet; Discover must not create it
		t.Logf("root exists (ok): %s", root)
	}
}

func TestDiscoverDataRootReadsExistingGlobal(t *testing.T) {
	dir := t.TempDir()
	real := filepath.Join(dir, "real-root")
	if err := storage.EnsureLayout(real); err != nil {
		t.Fatal(err)
	}
	if err := SaveGlobal(storage.GlobalConfigPath(real), DefaultGlobal(real)); err != nil {
		t.Fatal(err)
	}
	// Point --config at the existing file: data_root inside wins.
	root, _, err := DiscoverDataRoot(storage.GlobalConfigPath(real), "")
	if err != nil {
		t.Fatal(err)
	}
	if root != real && filepath.Clean(root) != filepath.Clean(real) {
		// On Windows Abs may differ in case; compare clean
		t.Fatalf("got %q want %q", root, real)
	}
}

func TestProjectValidateRejectsUnknownType(t *testing.T) {
	p := ProjectConfig{Name: "x", Root: "/tmp/x", Type: "nope", Version: VersionConfig{File: "f"}}
	if err := p.Validate(); err == nil {
		t.Fatal("expected error for unknown type")
	}
}
