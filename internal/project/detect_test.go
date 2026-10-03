package project

import (
	"os"
	"path/filepath"
	"testing"
)

func write(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(filepath.Join(dir, name)), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestDetectAndroid(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "settings.gradle", "")
	write(t, dir, "app/build.gradle", "android{}")
	write(t, dir, "gradle.properties", "app.versionCode=4\napp.versionName=0.0.3_2\n")
	write(t, dir, "gradlew.bat", "")
	info, err := Detect(dir)
	if err != nil {
		t.Fatal(err)
	}
	if info.Type != "android-gradle" {
		t.Fatalf("got %q", info.Type)
	}
	code, name, err := GradleVersion(info.Root, "gradle.properties")
	if err != nil {
		t.Fatal(err)
	}
	if code != "4" || name != "0.0.3_2" {
		t.Fatalf("got %q %q", code, name)
	}
}

func TestSetGradleVersionIncrements(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "gradle.properties", "app.versionCode=4\napp.versionName=0.0.3_2\nndkVersion=29\n")
	next, err := SetGradleVersion(dir, "gradle.properties", "0.0.4")
	if err != nil {
		t.Fatal(err)
	}
	if next != "5" {
		t.Fatalf("got %q", next)
	}
	code, name, err := GradleVersion(dir, "gradle.properties")
	if err != nil {
		t.Fatal(err)
	}
	if code != "5" || name != "0.0.4" {
		t.Fatalf("got %q %q", code, name)
	}
}

func TestDetectGoAndGeneric(t *testing.T) {
	godir := t.TempDir()
	write(t, godir, "go.mod", "module example.com/x\n\ngo 1.22\n")
	info, err := Detect(godir)
	if err != nil {
		t.Fatal(err)
	}
	if info.Type != "go" {
		t.Fatalf("got %q", info.Type)
	}
	emptydir := t.TempDir()
	info, err = Detect(emptydir)
	if err != nil {
		t.Fatal(err)
	}
	if info.Type != "generic" {
		t.Fatalf("got %q", info.Type)
	}
}

func TestDetectWails(t *testing.T) {
	dir := t.TempDir()
	write(t, dir, "wails.json", `{"name":"x"}`)
	info, err := Detect(dir)
	if err != nil {
		t.Fatal(err)
	}
	if info.Type != "wails" {
		t.Fatalf("got %q", info.Type)
	}
}

func TestBinaryBaseName(t *testing.T) {
	if got := BinaryBaseName(Info{Name: "releaseforge"}); got != "releaseforge" {
		t.Fatalf("got %q", got)
	}
	if got := BinaryBaseName(Info{Name: "My App"}); got != "My App" {
		t.Fatalf("got %q", got)
	}
	if got := BinaryBaseName(Info{}); got != "app" {
		t.Fatalf("fallback got %q", got)
	}
}
