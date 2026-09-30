package storage

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveRootOverride(t *testing.T) {
	root, err := ResolveRoot(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if !filepath.IsAbs(root) {
		t.Fatalf("expected absolute path, got %q", root)
	}
}

func TestResolveRootEnv(t *testing.T) {
	dir := t.TempDir()
	t.Setenv(EnvDataRoot, dir)
	root, err := ResolveRoot("")
	if err != nil {
		t.Fatal(err)
	}
	want, _ := filepath.Abs(dir)
	want = filepath.Clean(want)
	if root != want {
		t.Fatalf("got %q want %q", root, want)
	}
}

func TestDefaultRootNonEmpty(t *testing.T) {
	if DefaultRoot() == "" {
		t.Fatal("DefaultRoot must not be empty")
	}
}

func TestEnsureLayoutCreatesDirs(t *testing.T) {
	root := filepath.Join(t.TempDir(), "data")
	if err := EnsureLayout(root); err != nil {
		t.Fatal(err)
	}
	for _, d := range []string{
		filepath.Join(root, "config"),
		filepath.Join(root, "projects"),
		filepath.Join(root, "history"),
		GlobalCacheDir(root),
	} {
		if st, err := os.Stat(d); err != nil || !st.IsDir() {
			t.Fatalf("expected dir %s", d)
		}
	}
	if got := GlobalConfigPath(root); got != filepath.Join(root, "config", "global.json") {
		t.Fatalf("unexpected global path %q", got)
	}
}

func TestEnsureProjectLayout(t *testing.T) {
	root := t.TempDir()
	if err := EnsureLayout(root); err != nil {
		t.Fatal(err)
	}
	if err := EnsureProjectLayout(root, "Conductino-Android"); err != nil {
		t.Fatal(err)
	}
	for _, d := range []string{
		ProjectDir(root, "Conductino-Android"),
		BuildsDir(root, "Conductino-Android"),
		LogsDir(root, "Conductino-Android"),
		ReleasesDir(root, "Conductino-Android"),
		ReportsDir(root, "Conductino-Android"),
		CacheDir(root, "Conductino-Android"),
	} {
		if st, err := os.Stat(d); err != nil || !st.IsDir() {
			t.Fatalf("expected dir %s", d)
		}
	}
	ver := ReleaseVersionDir(root, "Conductino-Android", "0.1.0")
	want := filepath.Join(root, "projects", "Conductino-Android", "releases", "0.1.0")
	if ver != want {
		t.Fatalf("got %q want %q", ver, want)
	}
}

func TestSanitizeProjectName(t *testing.T) {
	cases := map[string]string{
		"Conductino-Android": "Conductino-Android",
		"a/b\\c:d":           "a-b-c-d",
		"":                   "unnamed-project",
		"...":                "unnamed-project",
		"  name. ":            "name",
	}
	for in, want := range cases {
		if got := SanitizeProjectName(in); got != want {
			t.Errorf("SanitizeProjectName(%q)=%q want %q", in, got, want)
		}
	}
}

func TestProjectNameFromPath(t *testing.T) {
	if got := ProjectNameFromPath(`D:\Dev\Conductino-Android\`); got != "Conductino-Android" {
		t.Fatalf("got %q", got)
	}
}
