package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestDraftNotes(t *testing.T) {
	notes := DraftNotes("App", "0.0.4", []Commit{{SHA: "abc1234", Subject: "Fix crash", Date: "2026-09-29"}}, true)
	if notes == "" {
		t.Fatal("empty notes")
	}
}

func TestTagExists(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}
	dir := t.TempDir()
	mustGit(t, dir, "init")
	if TagExists(dir, "v9.9.9") {
		t.Fatal("tag should not exist yet")
	}
	mustGit(t, dir, "config", "user.email", "test@example.com")
	mustGit(t, dir, "config", "user.name", "Test")
	os.WriteFile(filepath.Join(dir, "f.txt"), []byte("hi"), 0o644)
	mustGit(t, dir, "add", ".")
	mustGit(t, dir, "commit", "-m", "x")
	mustGit(t, dir, "tag", "-a", "v9.9.9", "-m", "x")
	if !TagExists(dir, "v9.9.9") {
		t.Fatal("tag should exist")
	}
}

func mustGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v %s", args, err, out)
	}
}

func TestLogInTempRepo(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}
	dir := t.TempDir()
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v %s", args, err, out)
		}
	}
	run("init")
	run("config", "user.email", "test@example.com")
	run("config", "user.name", "Test")
	if err := os.WriteFile(dir+"/f.txt", []byte("hi"), 0o644); err != nil {
		t.Fatal(err)
	}
	run("add", ".")
	run("commit", "-m", "first commit")
	if !IsRepo(dir) {
		t.Fatal("expected repo")
	}
	commits, err := Log(dir, 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(commits) != 1 || commits[0].Subject != "first commit" {
		t.Fatalf("got %+v", commits)
	}
}
