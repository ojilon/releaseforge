package android

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPackageReleaseCopies(t *testing.T) {
	proj := t.TempDir()
	debug := filepath.Join(proj, "app", "build", "outputs", "apk", "debug")
	rel := filepath.Join(proj, "app", "build", "outputs", "apk", "release")
	if err := os.MkdirAll(debug, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(rel, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(debug, "app-debug.apk"), []byte("debug"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(rel, "app-release-unsigned.apk"), []byte("release"), 0o644); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(t.TempDir(), "0.0.4")
	d, r, err := PackageRelease(proj, "0.0.4", "Conductino-Study", out)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(d); err != nil {
		t.Fatalf("debug out missing: %v", err)
	}
	if _, err := os.Stat(r); err != nil {
		t.Fatalf("release out missing: %v", err)
	}
}

func TestFindApksignerFallback(t *testing.T) {
	if got := FindApksigner(""); got == "" {
		t.Fatal("expected non-empty fallback")
	}
}

func TestResolveApkPrefersConfig(t *testing.T) {
	proj := t.TempDir()
	std := filepath.Join(proj, "app", "build", "outputs", "apk", "debug")
	if err := os.MkdirAll(std, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(std, "app-debug.apk"), []byte("std"), 0o644); err != nil {
		t.Fatal(err)
	}
	custom := filepath.Join(proj, "custom", "out.apk")
	if err := os.MkdirAll(filepath.Dir(custom), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(custom, []byte("custom"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := ResolveApk(proj, "custom/out.apk", DebugApkPath)
	if err != nil || got != custom {
		t.Fatalf("got %q %v", got, err)
	}
	got, err = ResolveApk(proj, "missing.apk", DebugApkPath)
	if err != nil || !strings.HasSuffix(got, "app-debug.apk") {
		t.Fatalf("fallback: got %q %v", got, err)
	}
	if _, err := ResolveApk(proj, "missing.apk"); err == nil {
		t.Fatal("expected error")
	}
}

func TestPackageFilesNaming(t *testing.T) {
	proj := t.TempDir()
	debug := filepath.Join(proj, "d.apk")
	rel := filepath.Join(proj, "r.apk")
	os.WriteFile(debug, []byte("d"), 0o644)
	os.WriteFile(rel, []byte("r"), 0o644)
	out := t.TempDir()
	d, r, err := PackageFiles("1.0", "MyApp", out, debug, rel)
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(d) != "MyApp-1.0-debug.apk" || filepath.Base(r) != "MyApp-1.0-release.apk" {
		t.Fatalf("got %q %q", d, r)
	}
}
