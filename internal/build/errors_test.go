package build

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseKotlin(t *testing.T) {
	rep := ParseErrors([]string{"e: app/src/Foo.kt:10:5 Unresolved reference: bar"})
	if len(rep.Items) != 1 {
		t.Fatalf("got %+v", rep.Items)
	}
	it := rep.Items[0]
	if it.File != "app/src/Foo.kt" || it.Line != 10 || it.Rule != "kotlin" {
		t.Fatalf("got %+v", it)
	}
	if !strings.Contains(it.Hint, "import") {
		t.Fatalf("hint: %q", it.Hint)
	}
}

func TestParseJavac(t *testing.T) {
	rep := ParseErrors([]string{"src/Foo.java:42: error: incompatible types"})
	if len(rep.Items) != 1 || rep.Items[0].File != "src/Foo.java" || rep.Items[0].Line != 42 {
		t.Fatalf("got %+v", rep.Items)
	}
}

func TestParseGradleTaskBlock(t *testing.T) {
	rep := ParseErrors([]string{
		"Execution failed for task ':app:compileReleaseKotlin'.",
		"> Compilation error. See log for more details",
		"> Something else",
	})
	if len(rep.Items) != 1 || rep.Items[0].Rule != "gradle" {
		t.Fatalf("got %+v", rep.Items)
	}
	if !strings.Contains(rep.Items[0].Message, ":app:compileReleaseKotlin") ||
		!strings.Contains(rep.Items[0].Message, "Compilation error") {
		t.Fatalf("got %q", rep.Items[0].Message)
	}
}

func TestParseCmakeNinjaLinkerApk(t *testing.T) {
	lines := []string{
		"CMake Error at backend/CMakeLists.txt:17 Could not find curl",
		"FAILED: aurora_core.so",
		"ninja: build stopped: subcommand failed.",
		"ld: error: undefined reference to `aurora_init'",
		"ERROR: Failed to sign the APK",
	}
	rep := ParseErrors(lines)
	if len(rep.Items) != 5 {
		t.Fatalf("got %d: %+v", len(rep.Items), rep.Items)
	}
	if rep.Items[0].File != "backend/CMakeLists.txt" || rep.Items[0].Line != 17 {
		t.Fatalf("cmake: %+v", rep.Items[0])
	}
	if rep.Items[4].Rule != "apksigner" {
		t.Fatalf("apk: %+v", rep.Items[4])
	}
}

func TestParseIgnoresNoise(t *testing.T) {
	rep := ParseErrors([]string{"ok", "BUILD SUCCESSFUL", "> Task :app:assembleDebug", ""})
	if len(rep.Items) != 0 {
		t.Fatalf("got %+v", rep.Items)
	}
}

func TestParseCap(t *testing.T) {
	var lines []string
	for i := 0; i < 40; i++ {
		lines = append(lines, "CMake Error at f.txt:1 boom")
	}
	rep := ParseErrors(lines)
	if len(rep.Items) != MaxErrorItems {
		t.Fatalf("got %d", len(rep.Items))
	}
	if !strings.Contains(rep.Summary, "25") {
		t.Fatalf("summary: %q", rep.Summary)
	}
}

func TestSummarizeTests(t *testing.T) {
	results := filepath.Join("testdata", "junit")
	summary, path, err := SummarizeTests(results, t.TempDir())
	if err != nil || path == "" {
		t.Fatalf("summary=%q path=%q err=%v", summary, path, err)
	}
	if !strings.Contains(summary, "tests: 3") || !strings.Contains(summary, "FooTest.fails") {
		t.Fatalf("summary: %q", summary)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("report not written: %v", err)
	}
}

func TestSummarizeTestsMissing(t *testing.T) {
	s, p, err := SummarizeTests(filepath.Join(t.TempDir(), "none"), t.TempDir())
	if err != nil || s != "" || p != "" {
		t.Fatalf("got %q %q %v", s, p, err)
	}
}
