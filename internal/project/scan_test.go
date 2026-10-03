package project

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ojilon/releaseforge/internal/config"
)

func TestWriteScanMetaFresh(t *testing.T) {
	root := t.TempDir()
	snap := Snapshot{ScannedAt: "2026-01-01T00:00:00Z", Root: root, Name: "p", Type: "go"}
	if _, err := WriteScan(root, snap); err != nil {
		t.Fatal(err)
	}
	if !Fresh(root, "p", root) {
		t.Fatal("expected fresh cache")
	}
	if Fresh(root, "p", t.TempDir()) {
		t.Fatal("expected stale cache for other root")
	}
	if Fresh(root, "missing", root) {
		t.Fatal("expected stale for unknown project")
	}
}

func TestLoadScanRoundTripAndMissing(t *testing.T) {
	root := t.TempDir()
	snap := Snapshot{Root: root, Name: "p", Type: "go", Version: VersionSnap{File: "VERSION", Name: "1.0"}}
	if _, err := WriteScan(root, snap); err != nil {
		t.Fatal(err)
	}
	got, err := LoadScan(root, "p")
	if err != nil {
		t.Fatal(err)
	}
	if got.Version.Name != "1.0" {
		t.Fatalf("got %+v", got.Version)
	}
	if _, err := LoadScan(root, "nope"); err == nil {
		t.Fatal("expected missing error")
	}
}

func TestCachedVersion(t *testing.T) {
	root := t.TempDir()
	if _, ok := CachedVersion(root, "p", root); ok {
		t.Fatal("no cache yet")
	}
	snap := Snapshot{Root: root, Name: "p", Type: "go", Version: VersionSnap{File: "VERSION", Name: "2.0"}}
	if _, err := WriteScan(root, snap); err != nil {
		t.Fatal(err)
	}
	v, ok := CachedVersion(root, "p", root)
	if !ok || v.Name != "2.0" {
		t.Fatalf("got %+v %v", v, ok)
	}
}

func TestMergePreservesUserKeys(t *testing.T) {
	alias := "mine"
	existing := config.ProjectConfig{
		Type: "android-gradle", Name: "a", Root: "/r",
		Version:   config.VersionConfig{File: "gradle.properties"},
		Signing:   &config.SigningConfig{Keystore: "k", Alias: alias},
		Github:    &config.GithubConfig{Owner: "o", Repo: "r"},
		Artifacts: &config.ArtifactsConfig{AppName: "Custom"},
		Build:     config.BuildConfig{Tasks: map[string]string{"custom": "myTask"}},
	}
	detected := SeedConfig(Info{Type: "android-gradle", Name: "a", Root: "/r2",
		VersionFile: "gradle.properties", CodeKey: "c", NameKey: "n"})
	merged, _ := MergeConfig(detected, filepath.Join(t.TempDir(), "none.json"), &existing)
	if merged.Signing.Alias != alias || merged.Github.Owner != "o" {
		t.Fatalf("user keys lost: %+v", merged)
	}
	if merged.Artifacts.AppName != "Custom" {
		t.Fatalf("app name lost: %+v", merged.Artifacts)
	}
	if merged.Build.Tasks["custom"] != "myTask" || merged.Build.Tasks["unit_test"] == "" {
		t.Fatalf("tasks wrong: %+v", merged.Build.Tasks)
	}
	if merged.Root != "/r2" || merged.Version.File != "gradle.properties" {
		t.Fatalf("tool keys not refreshed: %+v", merged)
	}
}

func TestMergeSeedApplied(t *testing.T) {
	dir := t.TempDir()
	seed := filepath.Join(dir, ".releaseforge.json")
	raw := `{"type":"go","name":"x","root":"/x","version":{"file":"VERSION"},"github":{"owner":"seed","repo":"r"}}`
	if err := os.WriteFile(seed, []byte(raw), 0o644); err != nil {
		t.Fatal(err)
	}
	detected := SeedConfig(Info{Type: "go", Name: "x", Root: "/x", VersionFile: "VERSION"})
	merged, notes := MergeConfig(detected, seed, nil)
	if merged.Github == nil || merged.Github.Owner != "seed" {
		t.Fatalf("seed not applied: %+v", merged.Github)
	}
	if len(notes) == 0 {
		t.Fatal("expected seed notes")
	}
}

func TestMergeSecretSeedIgnored(t *testing.T) {
	dir := t.TempDir()
	seed := filepath.Join(dir, ".releaseforge.json")
	raw := `{"signing":{"keystore":"k","alias":"a","password":"hunter2"}}`
	if err := os.WriteFile(seed, []byte(raw), 0o644); err != nil {
		t.Fatal(err)
	}
	detected := SeedConfig(Info{Type: "go", Name: "x", Root: "/x", VersionFile: "VERSION"})
	merged, notes := MergeConfig(detected, seed, nil)
	if merged.Signing != nil {
		t.Fatalf("secret seed must be skipped: %+v", merged.Signing)
	}
	found := false
	for _, n := range notes {
		if len(n) >= 7 && n[:7] == "warning" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected warning, got %v", notes)
	}
}
