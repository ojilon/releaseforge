package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestGroupCommits(t *testing.T) {
	in := []Commit{
		{Subject: "feat: new thing"},
		{Subject: "Fix: broken thing"},
		{Subject: "docs(readme): words"},
		{Subject: "chore: tidy"},
		{Subject: "refactor!: breaking change"},
		{Subject: "random words"},
	}
	groups := GroupCommits(in)
	if len(groups) != 5 {
		t.Fatalf("got %d groups", len(groups))
	}
	want := map[string]int{"Features": 1, "Fixes": 1, "Docs": 1, "Maintenance": 2, "Other": 1}
	for _, g := range groups {
		if want[g.Title] != len(g.Commits) {
			t.Fatalf("%s: got %d", g.Title, len(g.Commits))
		}
	}
	if groups[0].Title != "Features" || groups[4].Title != "Other" {
		t.Fatal("group order wrong")
	}
	// breaking marker stays in the subject
	for _, g := range groups {
		for _, c := range g.Commits {
			if c.Subject == "refactor!: breaking change" && g.Title != "Maintenance" {
				t.Fatal("breaking commit misplaced")
			}
		}
	}
}

func TestChangelogCap(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}
	dir := t.TempDir()
	mustGit(t, dir, "init")
	mustGit(t, dir, "config", "user.email", "test@example.com")
	mustGit(t, dir, "config", "user.name", "Test")
	os.WriteFile(filepath.Join(dir, "f.txt"), []byte("x"), 0o644)
	mustGit(t, dir, "add", ".")
	for i := 0; i < 8; i++ {
		mustGit(t, dir, "commit", "--allow-empty", "-m", "feat: change")
	}
	groups, total, truncated, err := Changelog(dir, 5)
	if err != nil {
		t.Fatal(err)
	}
	if total != 8 || !truncated {
		t.Fatalf("total=%d truncated=%v", total, truncated)
	}
	n := 0
	for _, g := range groups {
		n += len(g.Commits)
	}
	if n != 5 {
		t.Fatalf("shown=%d", n)
	}
	body := DraftGroupedNotes("App", "1.0", groups, total, truncated, true)
	if len(body) == 0 {
		t.Fatal("empty notes")
	}
}

func TestAheadBehind(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}
	tmp := t.TempDir()
	origin := filepath.Join(tmp, "origin.git")
	mustGit(t, tmp, "init", "--bare", origin)
	clone := func(name string) string {
		d := filepath.Join(tmp, name)
		mustGit(t, tmp, "clone", origin, d)
		mustGit(t, d, "config", "user.email", "test@example.com")
		mustGit(t, d, "config", "user.name", "Test")
		return d
	}
	w1 := clone("w1")
	os.WriteFile(filepath.Join(w1, "a.txt"), []byte("a"), 0o644)
	mustGit(t, w1, "add", ".")
	mustGit(t, w1, "commit", "-m", "first")
	mustGit(t, w1, "push", "-u", "origin", "HEAD")
	w2 := clone("w2")
	os.WriteFile(filepath.Join(w2, "b.txt"), []byte("b"), 0o644)
	mustGit(t, w2, "add", ".")
	mustGit(t, w2, "commit", "-m", "second")
	mustGit(t, w2, "push", "origin", "HEAD")
	os.WriteFile(filepath.Join(w1, "c.txt"), []byte("c"), 0o644)
	mustGit(t, w1, "add", ".")
	mustGit(t, w1, "commit", "-m", "local")
	mustGit(t, w1, "fetch", "origin")
	a, b, ok := AheadBehind(w1)
	if !ok || a != 1 || b != 1 {
		t.Fatalf("ahead=%d behind=%d ok=%v", a, b, ok)
	}
	fresh := t.TempDir()
	mustGit(t, fresh, "init")
	if _, _, ok := AheadBehind(fresh); ok {
		t.Fatal("expected no upstream")
	}
}
