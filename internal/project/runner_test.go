package project

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestForTypes(t *testing.T) {
	if _, err := For(Info{Type: "android-gradle"}); err != nil {
		t.Fatalf("gradle: %v", err)
	}
	if _, err := For(Info{Type: "go"}); err != nil {
		t.Fatalf("go: %v", err)
	}
	for _, typ := range []string{"generic", "wails", "cmake", "nope"} {
		if _, err := For(Info{Type: typ}); err == nil {
			t.Fatalf("%s: expected unsupported error", typ)
		}
	}
}

func TestGradleMatrix(t *testing.T) {
	r, _ := For(Info{Type: "android-gradle", Root: t.TempDir()})
	if _, err := r.TestArgs("bogus"); err == nil {
		t.Fatal("expected kind error")
	}
	for _, k := range []string{"unit", "instrumented", "all"} {
		args, err := r.TestArgs(k)
		if err != nil || len(args) == 0 {
			t.Fatalf("%s: %v %v", k, args, err)
		}
		if !strings.HasPrefix(args[0], ":app:") {
			t.Fatalf("%s: missing :app: prefix: %v", k, args)
		}
	}
	if _, err := r.BuildArgs("bogus", ""); err == nil {
		t.Fatal("expected variant error")
	}
	if args, _ := r.BuildArgs("debug", ""); args[0] != "assembleDebug" {
		t.Fatalf("got %v", args)
	}
	if r.BuildOutput("d", "debug") != "" {
		t.Fatal("gradle has no data-root output")
	}
}

func TestGoRunnerIgnoresKind(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	r, _ := For(Info{Type: "go", Root: dir, Name: "x"})
	for _, k := range []string{"unit", "whatever"} {
		if args, err := r.TestArgs(k); err != nil || len(args) == 0 {
			t.Fatalf("%s: %v", k, err)
		}
	}
	if r.Program() != "go" {
		t.Fatal("program should be go")
	}
	out := r.BuildOutput(t.TempDir(), "release")
	base := filepath.Base(out)
	if base != "x-release" && base != "x-release.exe" {
		t.Fatalf("got %q", out)
	}
}

func TestRunnerVersionRoundTrip(t *testing.T) {
	gdir := t.TempDir()
	write(t, gdir, "gradle.properties", "app.versionCode=7\napp.versionName=1.0\n")
	gr, _ := For(Info{Type: "android-gradle", Root: gdir, VersionFile: "gradle.properties"})
	code, name, err := gr.VersionRead()
	if err != nil || code != "7" || name != "1.0" {
		t.Fatalf("got %q %q %v", code, name, err)
	}
	if next, err := gr.VersionWrite("1.1"); err != nil || next != "8" {
		t.Fatalf("got %q %v", next, err)
	}

	vdir := t.TempDir()
	write(t, vdir, "VERSION", "0.1.0\n")
	vr, _ := For(Info{Type: "go", Root: vdir, VersionFile: "VERSION", Name: "v"})
	if code, name, err := vr.VersionRead(); err != nil || code != "" || name != "0.1.0" {
		t.Fatalf("got %q %q %v", code, name, err)
	}
	if _, err := vr.VersionWrite("0.2.0"); err != nil {
		t.Fatal(err)
	}
}

func TestSeedConfig(t *testing.T) {
	g := SeedConfig(Info{Type: "android-gradle", Name: "a", Root: "/r", VersionFile: "gradle.properties", CodeKey: "c", NameKey: "n"})
	if !g.Build.GradleWrapper || g.Build.Tasks["unit_test"] != ":app:testDebugUnitTest" {
		t.Fatalf("got %+v", g.Build)
	}
	if g.Artifacts == nil || g.Artifacts.AppName != "a" {
		t.Fatalf("got %+v", g.Artifacts)
	}
	gg := SeedConfig(Info{Type: "go", Name: "b", Root: "/r", VersionFile: "VERSION"})
	if gg.Version.File != "VERSION" || gg.Build.Tasks != nil {
		t.Fatalf("got %+v", gg)
	}
}
