package project

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// stageGradleFixture copies testdata/gradle into a fake android project tree.
func stageGradleFixture(t *testing.T, moduleFile string) string {
	t.Helper()
	src := filepath.Join("testdata", "gradle")
	dir := t.TempDir()
	copyFile := func(srcRel, dstRel string) {
		t.Helper()
		data, err := os.ReadFile(filepath.Join(src, srcRel))
		if err != nil {
			t.Fatal(err)
		}
		dst := filepath.Join(dir, dstRel)
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(dst, data, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	copyFile("gradle.properties", "gradle.properties")
	copyFile("settings.gradle", "settings.gradle")
	copyFile(moduleFile, "app/build.gradle")
	copyFile("gradle-wrapper.properties", filepath.Join("gradle", "wrapper", "gradle-wrapper.properties"))
	copyFile("local.properties", "local.properties")
	copyFile("libs.versions.toml", filepath.Join("gradle", "libs.versions.toml"))
	copyFile("AndroidManifest.xml", filepath.Join("app", "src", "main", "AndroidManifest.xml"))
	if err := os.MkdirAll(filepath.Join(dir, "backend"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "backend", "CMakeLists.txt"), []byte("cmake_minimum_required(VERSION 3.22)\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestParseProperties(t *testing.T) {
	m := parsePropertiesFile(filepath.Join("testdata", "gradle", "gradle.properties"), "gradle.properties")
	if m["app.versionCode"].Value != "12" || m["app.versionCode"].Line != 2 {
		t.Fatalf("got %+v", m["app.versionCode"])
	}
	if m["aurora.abiFilters"].Value != "arm64-v8a,armeabi-v7a" {
		t.Fatalf("got %+v", m["aurora.abiFilters"])
	}
}

func TestScanGradleGroovy(t *testing.T) {
	dir := stageGradleFixture(t, "app-build.gradle")
	g := ScanGradle(dir)
	if len(g.Modules) != 2 || g.Modules[0] != ":app" || g.Modules[1] != ":shared" {
		t.Fatalf("modules: %v", g.Modules)
	}
	a := g.Android
	if a.Namespace.Value != "com.example.app" || a.Namespace.File != "app/build.gradle" {
		t.Fatalf("namespace: %+v", a.Namespace)
	}
	if a.ApplicationID.Value != "com.example.app" {
		t.Fatalf("appId: %+v", a.ApplicationID)
	}
	if a.CompileSdk.Value != "34" || a.MinSdk.Value != "26" || a.TargetSdk.Value != "34" {
		t.Fatalf("sdks: %+v %+v %+v", a.CompileSdk, a.MinSdk, a.TargetSdk)
	}
	if a.NdkVersion.Value != "29.0.14206865" {
		t.Fatalf("ndk: %+v", a.NdkVersion)
	}
	if len(a.AbiFilters) != 2 || a.AbiFilters[0] != "arm64-v8a" {
		t.Fatalf("abis: %v", a.AbiFilters)
	}
	if a.CmakePath.Value != "src/main/cpp/CMakeLists.txt" {
		t.Fatalf("cmake: %+v", a.CmakePath)
	}
	if a.VersionCode.Value != "12" || a.VersionName.Value != "1.2.0" {
		t.Fatalf("version: %+v %+v", a.VersionCode, a.VersionName)
	}
	if g.Wrapper.GradleVersion != "8.10" {
		t.Fatalf("wrapper: %+v", g.Wrapper)
	}
	if g.SdkDir.Value != "<set>" {
		t.Fatalf("sdkdir: %+v", g.SdkDir)
	}
	if g.Versions["agp"] != "8.5.2" || g.VersionsFile != "gradle/libs.versions.toml" {
		t.Fatalf("toml: %+v %s", g.Versions, g.VersionsFile)
	}
	if g.Manifest.Package != "com.example.app" || !g.Manifest.LauncherFound {
		t.Fatalf("manifest: %+v", g.Manifest)
	}
	// sdk.dir value must never appear in serialized output
	raw, _ := json.Marshal(g)
	if strings.Contains(string(raw), "tester") {
		t.Fatal("sdk path leaked into snapshot")
	}
}

func TestScanGradleKts(t *testing.T) {
	dir := t.TempDir()
	data, err := os.ReadFile(filepath.Join("testdata", "gradle", "app-build.gradle.kts"))
	if err != nil {
		t.Fatal(err)
	}
	os.MkdirAll(filepath.Join(dir, "app"), 0o755)
	os.WriteFile(filepath.Join(dir, "app", "build.gradle.kts"), data, 0o644)
	os.WriteFile(filepath.Join(dir, "settings.gradle.kts"), []byte(`include(":app")`+"\n"), 0o644)
	g := ScanGradle(dir)
	if g.Android.Namespace.Value != "com.example.kts" {
		t.Fatalf("namespace: %+v", g.Android.Namespace)
	}
	if g.Android.CompileSdk.Value != "35" || g.Android.VersionName.Value != "0.3.0" {
		t.Fatalf("got %+v %+v", g.Android.CompileSdk, g.Android.VersionName)
	}
	if len(g.Android.AbiFilters) != 2 {
		t.Fatalf("abis: %v", g.Android.AbiFilters)
	}
}

func TestUnresolvedVersionReported(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "app"), 0o755)
	body := "android {\n    defaultConfig {\n        versionName = property(\"appVersion\")\n        versionCode = extra[\"code\"]\n    }\n}\n"
	os.WriteFile(filepath.Join(dir, "app", "build.gradle"), []byte(body), 0o644)
	g := ScanGradle(dir)
	if g.Android.VersionName.Unresolved == "" {
		t.Fatalf("expected unresolved versionName: %+v", g.Android.VersionName)
	}
	if g.Android.VersionCode.Unresolved == "" {
		t.Fatalf("expected unresolved versionCode: %+v", g.Android.VersionCode)
	}
}

func TestCmakeBounds(t *testing.T) {
	dir := t.TempDir()
	deep := filepath.Join(dir, "a", "b", "c", "d", "e")
	os.MkdirAll(deep, 0o755)
	os.WriteFile(filepath.Join(deep, "CMakeLists.txt"), []byte("# deep\n"), 0o644)
	os.MkdirAll(filepath.Join(dir, "native"), 0o755)
	os.WriteFile(filepath.Join(dir, "native", "CMakeLists.txt"), []byte("# top\n"), 0o644)
	got := FindCmakeFiles(dir)
	for _, f := range got {
		if strings.Contains(f, filepath.Join("a", "b")) {
			t.Fatalf("depth bound violated: %v", got)
		}
	}
	if len(got) != 1 || got[0] != filepath.Join("native", "CMakeLists.txt") {
		t.Fatalf("got %v", got)
	}
}

func TestDeepNonAndroidErrors(t *testing.T) {
	if _, err := DeepProperties(Info{Type: "go", Root: t.TempDir()}, t.TempDir()); err == nil {
		t.Fatal("expected type error without spawning anything")
	}
}

func TestScanAttachesGradleAndCmake(t *testing.T) {
	dir := stageGradleFixture(t, "app-build.gradle")
	snap, info, err := Scan(dir)
	if err != nil {
		t.Fatal(err)
	}
	if info.Type != "android-gradle" {
		t.Fatalf("type: %s", info.Type)
	}
	if snap.Gradle == nil || snap.Gradle.Android.ApplicationID.Value != "com.example.app" {
		t.Fatalf("gradle snapshot missing: %+v", snap.Gradle)
	}
	found := false
	for _, f := range snap.CmakeFiles {
		if f == filepath.Join("backend", "CMakeLists.txt") {
			found = true
		}
	}
	if !found {
		t.Fatalf("cmake files: %v", snap.CmakeFiles)
	}
}
