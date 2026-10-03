package project

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestKtsRoundTrip(t *testing.T) {
	for _, tc := range []struct{ name, body, want string }{
		{"double", "plugins {}\nandroid {\n    version = \"1.0.0\"\n}\n", "1.0.0"},
		{"single", "version = '2.1.0'\n", "2.1.0"},
	} {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "build.gradle.kts"), []byte(tc.body), 0o644); err != nil {
			t.Fatal(err)
		}
		got, err := KtsVersion(dir, "build.gradle.kts")
		if err != nil || got != tc.want {
			t.Fatalf("%s: got %q %v", tc.name, got, err)
		}
		if err := SetKtsVersion(dir, "build.gradle.kts", "9.9.9"); err != nil {
			t.Fatal(err)
		}
		got, err = KtsVersion(dir, "build.gradle.kts")
		if err != nil || got != "9.9.9" {
			t.Fatalf("%s rewrite: got %q %v", tc.name, got, err)
		}
		raw, _ := os.ReadFile(filepath.Join(dir, "build.gradle.kts"))
		if tc.name == "single" && !strings.Contains(string(raw), "'9.9.9'") {
			t.Fatalf("quote style lost: %s", raw)
		}
		if tc.name == "double" && !strings.Contains(string(raw), "\"9.9.9\"") {
			t.Fatalf("quote style lost: %s", raw)
		}
	}
}

func TestKtsNonLiteralRefused(t *testing.T) {
	dir := t.TempDir()
	body := "android {\n    version = property(\"appVersion\")\n}\n"
	if err := os.WriteFile(filepath.Join(dir, "build.gradle.kts"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := KtsVersion(dir, "build.gradle.kts"); err == nil || !strings.Contains(err.Error(), ":2:") {
		t.Fatalf("expected file:line error, got %v", err)
	}
	if err := SetKtsVersion(dir, "build.gradle.kts", "1.0"); err == nil {
		t.Fatal("expected refusal")
	}
	// comments must not match
	dir2 := t.TempDir()
	os.WriteFile(filepath.Join(dir2, "build.gradle.kts"), []byte("// version = \"0.0.1\"\n"), 0o644)
	if _, err := KtsVersion(dir2, "build.gradle.kts"); err == nil {
		t.Fatal("comment must not match")
	}
}

func TestValidateVersion(t *testing.T) {
	for _, bad := range []string{"", "  ", "1.0 beta", "a\nb"} {
		if err := ValidateVersion(bad); err == nil {
			t.Fatalf("%q: expected error", bad)
		}
	}
	for _, good := range []string{"0.0.1", "1.2", "0.0.3_2", "1.0.0-pre"} {
		if err := ValidateVersion(good); err != nil {
			t.Fatalf("%q: %v", good, err)
		}
	}
	if !LooksSemver("1.2.3") || !LooksSemver("0.0.3_2") || LooksSemver("nightly") {
		t.Fatal("semver heuristic wrong")
	}
}

func TestDetectVersionSource(t *testing.T) {
	pdir := t.TempDir()
	os.WriteFile(filepath.Join(pdir, "gradle.properties"), []byte("app.versionCode=1\napp.versionName=1.0\n"), 0o644)
	if s, f := DetectVersionSource(pdir); s != SourceProperties || f != "gradle.properties" {
		t.Fatalf("got %q %q", s, f)
	}
	kdir := t.TempDir()
	os.MkdirAll(filepath.Join(kdir, "app"), 0o755)
	os.WriteFile(filepath.Join(kdir, "app", "build.gradle.kts"), []byte("version = \"3.0\"\n"), 0o644)
	if s, f := DetectVersionSource(kdir); s != SourceKts || f != "app/build.gradle.kts" {
		t.Fatalf("got %q %q", s, f)
	}
	if s, _ := DetectVersionSource(t.TempDir()); s != "" {
		t.Fatalf("got %q", s)
	}
}

func TestBumpCode(t *testing.T) {
	pdir := t.TempDir()
	os.WriteFile(filepath.Join(pdir, "gradle.properties"), []byte("app.versionCode=4\napp.versionName=1.0\n"), 0o644)
	gr, _ := For(Info{Type: "android-gradle", Root: pdir, VersionFile: "gradle.properties", VersionSource: SourceProperties})
	next, err := gr.BumpCode()
	if err != nil || next != "5" {
		t.Fatalf("got %q %v", next, err)
	}
	code, name, _ := gr.VersionRead()
	if code != "5" || name != "1.0" {
		t.Fatalf("name must be untouched: %q %q", code, name)
	}
	kr, _ := For(Info{Type: "android-gradle", Root: pdir, VersionFile: "x.kts", VersionSource: SourceKts})
	if _, err := kr.BumpCode(); err == nil {
		t.Fatal("kts must refuse --code-only")
	}
	gor, _ := For(Info{Type: "go", Root: t.TempDir()})
	if _, err := gor.BumpCode(); err == nil {
		t.Fatal("go must refuse --code-only")
	}
}

func TestKtsRunnerReadWrite(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "app"), 0o755)
	os.WriteFile(filepath.Join(dir, "app", "build.gradle.kts"), []byte("version = \"0.1.0\"\n"), 0o644)
	r, _ := For(Info{Type: "android-gradle", Root: dir, VersionFile: "app/build.gradle.kts", VersionSource: SourceKts})
	if code, name, err := r.VersionRead(); err != nil || code != "" || name != "0.1.0" {
		t.Fatalf("got %q %q %v", code, name, err)
	}
	if _, err := r.VersionWrite("0.2.0"); err != nil {
		t.Fatal(err)
	}
}
