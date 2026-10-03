package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ojilon/releaseforge/internal/project"
)

func TestDryRunNoSideEffects(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}
	dir := t.TempDir()
	git := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v %s", args, err, out)
		}
	}
	git("init")
	git("config", "user.email", "test@example.com")
	git("config", "user.name", "Test")
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/dry\n"), 0o644)
	os.WriteFile(filepath.Join(dir, "VERSION"), []byte("0.5.0\n"), 0o644)
	git("add", ".")
	git("commit", "-m", "x")

	data := t.TempDir()
	info, err := project.Detect(dir)
	if err != nil {
		t.Fatal(err)
	}
	r, err := project.For(info)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := dryRunPlan(info, r, "0.6.0", data)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"test: go test", "build release: go build", "notes:", "zip:", "tag: v0.6.0", "gh release create"} {
		if !strings.Contains(plan, want) {
			t.Fatalf("plan missing %q:\n%s", want, plan)
		}
	}
	// Zero side effects: no tag, VERSION untouched, no releases created.
	out, _ := exec.Command("git", "-C", dir, "tag", "-l").Output()
	if strings.TrimSpace(string(out)) != "" {
		t.Fatalf("tag created: %s", out)
	}
	ver, _ := os.ReadFile(filepath.Join(dir, "VERSION"))
	if strings.TrimSpace(string(ver)) != "0.5.0" {
		t.Fatalf("VERSION changed: %q", ver)
	}
	entries, _ := os.ReadDir(filepath.Join(data, "projects"))
	if len(entries) != 0 {
		t.Fatalf("data root touched: %v", entries)
	}
}
