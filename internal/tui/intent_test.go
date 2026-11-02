package tui

import (
	"strings"
	"testing"
)

func TestTokenize(t *testing.T) {
	cases := []struct {
		in   string
		want []string
	}{
		{"", nil},
		{"  status  ", []string{"status"}},
		{"build release 0.0.4", []string{"build", "release", "0.0.4"}},
		{`open "C:/Dev/My App/repo"`, []string{"open", "C:/Dev/My App/repo"}},
		{`open 'D:\Dev\My Repo'`, []string{"open", `D:\Dev\My Repo`}},
		{`cat "a\"b.txt"`, []string{"cat", `a"b.txt`}},
		{`scan D:\Dev\repo`, []string{"scan", `D:\Dev\repo`}},
	}
	for _, c := range cases {
		got, err := Tokenize(c.in)
		if err != nil {
			t.Fatalf("Tokenize(%q): %v", c.in, err)
		}
		if len(got) != len(c.want) {
			t.Fatalf("Tokenize(%q)=%v want %v", c.in, got, c.want)
		}
		for i := range got {
			if got[i] != c.want[i] {
				t.Fatalf("Tokenize(%q)[%d]=%q want %q", c.in, i, got[i], c.want[i])
			}
		}
	}
	if _, err := Tokenize(`open "unterminated`); err == nil {
		t.Fatal("unterminated quote must be an error")
	}
}

func TestParseIntent(t *testing.T) {
	cases := map[string][2]string{
		"":                {"", ""},
		"quit":            {"quit", ""},
		"Q":               {"q", ""},
		"?":               {"help", ""},
		"-h":              {"help", ""},
		"OPEN":            {"open", ""},
		"build release":   {"build", "release"},
		"version --set X": {"version", "--set"},
		"frobnicate":      {"frobnicate", ""},
	}
	for in, want := range cases {
		i, err := ParseIntent(in)
		if err != nil || i.Verb != want[0] {
			t.Errorf("ParseIntent(%q)=%q,%v", in, i.Verb, err)
		}
		if want[1] != "" && (len(i.Args) == 0 || i.Args[0] != want[1]) {
			t.Errorf("ParseIntent(%q) args=%v", in, i.Args)
		}
	}
	if i, _ := ParseIntent("status"); !i.Known {
		t.Fatal("status must be a known verb")
	}
	if i, _ := ParseIntent("nonsense"); i.Known {
		t.Fatal("nonsense must be unknown")
	}
}

func TestCompletionsGenerated(t *testing.T) {
	if got := Completions("bui", nil); len(got) != 1 || got[0] != "build " {
		t.Fatalf("verb completion: %v", got)
	}
	if got := Completions("build r", nil); len(got) == 0 || !strings.Contains(got[0], "release") {
		t.Fatalf("sub completion: %v", got)
	}
	if got := Completions("theme ", []string{"dark", "light", "mono"}); len(got) != 3 {
		t.Fatalf("extras: %v", got)
	}
	if p := CommonPrefix([]string{"build ", "bat "}); p != "b" {
		t.Fatalf("common prefix %q", p)
	}
}

func TestHelpTextGeneratedFromVerbs(t *testing.T) {
	h := HelpText()
	for _, v := range Verbs {
		if !strings.Contains(h, v.Name) {
			t.Fatalf("help missing verb %q", v.Name)
		}
		if !strings.Contains(h, v.Help) {
			t.Fatalf("help missing description of %q", v.Name)
		}
	}
}
