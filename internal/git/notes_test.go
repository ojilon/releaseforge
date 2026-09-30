package git

import (
	"os"
	"os/exec"
	"testing"
)

func TestDraftNotes(t *testing.T) {
	notes := DraftNotes("App", "0.0.4", []Commit{{SHA: "abc1234", Subject: "Fix crash", Date: "2026-09-29"}}, true)
	if notes == "" {
		t.Fatal("empty notes")
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
