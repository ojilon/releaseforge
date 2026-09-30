package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/ojilon/releaseforge/internal/storage"
)

func TestEnsureTreeFreshAndRerun(t *testing.T) {
	tmp := t.TempDir()
	payload := filepath.Join(tmp, "releaseforge.exe")
	if err := os.WriteFile(payload, []byte("fake-binary-v1"), 0o644); err != nil {
		t.Fatal(err)
	}
	appRoot := filepath.Join(tmp, "parent", "ReleaseForge")

	if err := EnsureTree(appRoot, payload, "0.0.1"); err != nil {
		t.Fatal(err)
	}
	for _, d := range []string{
		filepath.Join(appRoot, "bin"),
		filepath.Join(appRoot, "data", "config"),
		filepath.Join(appRoot, "data", "projects"),
		filepath.Join(appRoot, "data", "history"),
		filepath.Join(appRoot, "data", "global-cache"),
	} {
		if st, err := os.Stat(d); err != nil || !st.IsDir() {
			t.Fatalf("expected dir %s", d)
		}
	}
	if _, err := os.Stat(storage.GlobalConfigPath(filepath.Join(appRoot, "data"))); err != nil {
		t.Fatalf("global.json missing: %v", err)
	}
	// User data must survive a re-install; binary must update.
	keep := filepath.Join(appRoot, "data", "history", "recent-projects.json")
	if err := os.WriteFile(keep, []byte(`{"items":[],"max":30}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(payload, []byte("fake-binary-v2"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := EnsureTree(appRoot, payload, "0.0.2"); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(filepath.Join(appRoot, "bin", "releaseforge.exe"))
	if err != nil || string(got) != "fake-binary-v2" {
		t.Fatalf("binary not updated: %q %v", got, err)
	}
	if _, err := os.Stat(keep); err != nil {
		t.Fatalf("user data wiped: %v", err)
	}
	raw, err := os.ReadFile(filepath.Join(appRoot, "data", "install.json"))
	if err != nil {
		t.Fatal(err)
	}
	var rec InstallRecord
	if err := json.Unmarshal(raw, &rec); err != nil {
		t.Fatal(err)
	}
	if rec.Version != "0.0.2" {
		t.Fatalf("got version %q", rec.Version)
	}
}

func TestListDrivesOverride(t *testing.T) {
	a := t.TempDir()
	b := t.TempDir()
	t.Setenv("RF_INSTALL_DRIVES", a+string(os.PathListSeparator)+b)
	got := listDrives()
	if len(got) != 2 {
		t.Fatalf("got %v", got)
	}
}

func TestNonInteractiveRequiresDest(t *testing.T) {
	if err := run("", "", "", "ReleaseForge", "0.0.1", true); err == nil {
		t.Fatal("expected error")
	}
}
