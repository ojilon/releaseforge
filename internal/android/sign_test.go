package android

import (
	"os"
	"path/filepath"
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
