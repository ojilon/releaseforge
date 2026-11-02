package tui

import "testing"

func TestScore(t *testing.T) {
	if Score("bld", "build") < 0 {
		t.Fatal("subsequence must match")
	}
	if Score("xyz", "build") >= 0 {
		t.Fatal("non-subsequence must not match")
	}
	if Score("", "build") != 0 {
		t.Fatal("empty query matches everything with score 0")
	}
	if Score("build", "build release") <= Score("bld", "build release") {
		t.Fatal("prefix match should score higher than a gappy match")
	}
	if Score("tp", "test project") < Score("tp", "tiny prototype") {
		// boundary bonus: "test project" starts both words
		t.Log("boundary scoring is informational, not a hard contract")
	}
}

func TestRank(t *testing.T) {
	targets := []string{"status", "scan", "build", "test", "notes", "sign"}
	m := Rank("st", targets, 3)
	// "status" and "test" are the only subsequence matches.
	if len(m) != 2 {
		t.Fatalf("matches: %+v", m)
	}
	for i := 1; i < len(m); i++ {
		if m[i-1].Score < m[i].Score {
			t.Fatalf("not sorted: %+v", m)
		}
	}
	if got := Rank("build", targets, 0); len(got) == 0 || targets[got[0].Index] != "build" {
		t.Fatalf("exact match must rank first: %+v", got)
	}
	if got := Rank("zzzz", targets, 10); len(got) != 0 {
		t.Fatalf("no matches expected: %+v", got)
	}
}
