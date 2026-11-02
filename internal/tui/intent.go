package tui

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

// Verb is one command-bar entry. This table is the only place a command is
// described: the Help window, the CLI `commands` output, tab completion,
// the fuzzy palette and the unknown-command hint are all generated from it.
// Adding a verb without touching this table is intentionally impossible to
// document (docs/sweep-2mo/06).
type Verb struct {
	Name  string
	Args  string
	Help  string
	Subs  []string
	Group string
	Keys  string
}

// Verbs is the full command grammar. Order is the order shown in Help.
var Verbs = []Verb{
	{Name: "open", Args: "[path|#n]", Help: "scan a folder (or recent entry #n); no path opens the folder picker", Group: "project", Keys: "ctrl+o"},
	{Name: "scan", Args: "[path|#n]", Help: "alias of open", Group: "project"},
	{Name: "rescan", Args: "", Help: "re-scan the current project", Group: "project"},
	{Name: "recent", Args: "", Help: "list recently opened projects", Group: "project"},
	{Name: "status", Args: "", Help: "tool, project, version, data root, last build", Group: "project", Keys: "1"},
	{Name: "version", Args: "[--set X|bump X|--code-only]", Help: "show or set the project version (never commits)", Group: "project"},
	{Name: "build", Args: "debug|release [version]", Help: "async Gradle/Go build with live output", Group: "work", Keys: "4",
		Subs: []string{"debug", "release"}},
	{Name: "test", Args: "[unit|instrumented|all]", Help: "async tests, streamed and persisted", Group: "work", Keys: "2",
		Subs: []string{"unit", "instrumented", "all"}},
	{Name: "bat", Args: "<script> [args...]", Help: "run .bat/.cmd/.ps1/.sh/.py with live logs", Group: "work"},
	{Name: "sign", Args: "[apk] [--alias X]", Help: "sign an APK with the data-root keystore", Group: "work"},
	{Name: "install", Args: "[debug|release] [--device S]", Help: "install the built APK via adb", Group: "work"},
	{Name: "devices", Args: "", Help: "list connected adb devices", Group: "work"},
	{Name: "notes", Args: "[version]", Help: "draft release notes from git history", Group: "work"},
	{Name: "git", Args: "log|status|diff|tags", Help: "git state while analysing", Group: "inspect", Keys: "3",
		Subs: []string{"log", "status", "diff", "tags"}},
	{Name: "logs", Args: "[show|last|tail|name]", Help: "browse persisted build/test logs", Group: "inspect", Keys: "5",
		Subs: []string{"show", "last", "tail"}},
	{Name: "ls", Args: "[path]", Help: "list a folder (dir is an alias)", Group: "inspect"},
	{Name: "pwd", Args: "", Help: "print the current project directory", Group: "inspect"},
	{Name: "cat", Args: "<file>", Help: "print a small text file (capped)", Group: "inspect"},
	{Name: "theme", Args: "[dark|light|mono]", Help: "switch or list color presets (auto-detected at startup)", Group: "ui",
		Subs: []string{"dark", "light", "mono"}},
	{Name: "window", Args: "1..6|name", Help: "switch the active window", Group: "ui", Keys: "1..6",
		Subs: []string{"1", "2", "3", "4", "5", "6", "status", "tests", "git", "console", "logs", "help"}},
	{Name: "help", Args: "", Help: "open the generated command reference", Group: "ui", Keys: "?"},
	{Name: "clear", Args: "", Help: "clear the console", Group: "ui", Keys: "ctrl+l"},
	{Name: "quit", Args: "", Help: "leave the TUI", Group: "ui", Keys: "ctrl+c"},
}

// Aliases map shorthand verbs to their canonical name.
var Aliases = map[string]string{
	"-h": "help", "?": "help", "q": "quit", "exit": "quit",
	"dir": "ls", "run": "bat", "win": "window",
}

// Groups order the generated help.
var Groups = []struct{ ID, Title string }{
	{"project", "Project"},
	{"work", "Build · test · release"},
	{"inspect", "Inspect"},
	{"ui", "Interface"},
}

// Intent is a parsed command-bar line.
type Intent struct {
	Verb  string
	Args  []string
	Known bool
}

// Canonical resolves aliases so `?` and `-h` behave exactly like `help`.
func (i Intent) Canonical() Intent {
	if c, ok := Aliases[i.Verb]; ok {
		i.Verb = c
	}
	return i
}

// FindVerb returns the table entry for a verb, if any.
func FindVerb(name string) (Verb, bool) {
	for _, v := range Verbs {
		if v.Name == name {
			return v, true
		}
	}
	return Verb{}, false
}

// Tokenize splits a command line on whitespace while honoring single and
// double quotes. Backslashes are literal (Windows paths) except when they
// escape a quote or another backslash. An unterminated quote is an error
// instead of a silently mangled path.
func Tokenize(line string) ([]string, error) {
	var out []string
	var cur strings.Builder
	var quote rune
	started := false
	runes := []rune(line)
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if quote != 0 {
			if r == quote {
				quote = 0
				continue
			}
			// Inside double quotes a backslash can escape a quote or
			// another backslash; all other backslashes stay literal so
			// Windows paths in quotes are untouched.
			if quote == '"' && r == '\\' && i+1 < len(runes) &&
				(runes[i+1] == '"' || runes[i+1] == '\\') {
				cur.WriteRune(runes[i+1])
				i++
				continue
			}
			cur.WriteRune(r)
			continue
		}
		switch {
		case r == '\'' || r == '"':
			quote = r
			started = true
		case r == '\\' && i+1 < len(runes) &&
			(runes[i+1] == '\'' || runes[i+1] == '"' || runes[i+1] == '\\'):
			cur.WriteRune(runes[i+1])
			started = true
			i++
		case r == ' ' || r == '\t' || r == '\n':
			if started {
				out = append(out, cur.String())
				cur.Reset()
				started = false
			}
		default:
			cur.WriteRune(r)
			started = true
		}
	}
	if quote != 0 {
		return nil, errors.New("unterminated quote — close it with ' or \"")
	}
	if started {
		out = append(out, cur.String())
	}
	return out, nil
}

// ParseIntent parses a line into a verb + args. Empty input yields an empty
// verb. Unknown verbs are returned with Known=false so the caller can give a
// helpful hint.
func ParseIntent(line string) (Intent, error) {
	tokens, err := Tokenize(strings.TrimSpace(line))
	if err != nil {
		return Intent{}, err
	}
	if len(tokens) == 0 {
		return Intent{}, nil
	}
	canonical := tokens[0]
	if c, ok := Aliases[canonical]; ok {
		canonical = c
	}
	i := Intent{Verb: strings.ToLower(canonical), Args: tokens[1:]}
	_, i.Known = FindVerb(i.Verb)
	return i, nil
}

// Completions returns full-line completions for line. extras are contextual
// candidates (recent project paths, log names, history lines) offered when a
// verb accepts a free-form argument.
func Completions(line string, extras []string) []string {
	trailingSpace := len(line) > 0 && (line[len(line)-1] == ' ' || line[len(line)-1] == '\t')
	tokens, err := Tokenize(line)
	if err != nil || len(tokens) == 0 {
		return verbCandidates("")
	}
	if len(tokens) == 1 && !trailingSpace {
		return verbCandidates(tokens[0])
	}
	verb := strings.ToLower(tokens[0])
	if c, ok := Aliases[verb]; ok {
		verb = c
	}
	partial := ""
	if !trailingSpace {
		partial = tokens[len(tokens)-1]
	}
	head := tokens[:len(tokens)-1]
	if trailingSpace {
		head = tokens
	}
	var pool []string
	if v, ok := FindVerb(verb); ok {
		pool = append(pool, v.Subs...)
	}
	pool = append(pool, extras...)
	var out []string
	for _, c := range pool {
		if partial == "" || strings.HasPrefix(strings.ToLower(c), strings.ToLower(partial)) {
			out = append(out, strings.Join(append(append([]string{}, head...), c), " ")+" ")
		}
	}
	return dedupeSorted(out)
}

func verbCandidates(prefix string) []string {
	var out []string
	for _, v := range Verbs {
		if strings.HasPrefix(v.Name, strings.ToLower(prefix)) {
			out = append(out, v.Name+" ")
		}
	}
	return dedupeSorted(out)
}

func dedupeSorted(in []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, s := range in {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	sort.Strings(out)
	return out
}

// CommonPrefix is the longest common prefix of a set of completions.
func CommonPrefix(candidates []string) string {
	if len(candidates) == 0 {
		return ""
	}
	p := candidates[0]
	for _, c := range candidates[1:] {
		i := 0
		for i < len(p) && i < len(c) && p[i] == c[i] {
			i++
		}
		p = p[:i]
	}
	return p
}

// VerbUsage renders "name args" padded for the generated help.
func VerbUsage(v Verb) string {
	if v.Args == "" {
		return v.Name
	}
	return v.Name + " " + v.Args
}

// HelpText is the plain-text command reference used by the CLI `commands`
// verb and echoed into the console by `help`. It is generated, never edited
// by hand.
func HelpText() string {
	var b strings.Builder
	b.WriteString("commands (type in the bar, or use the command palette ctrl+p)\n")
	for _, g := range Groups {
		fmt.Fprintf(&b, "\n%s\n", strings.ToUpper(g.Title))
		for _, v := range Verbs {
			if v.Group != g.ID {
				continue
			}
			fmt.Fprintf(&b, "  %-34s %s\n", VerbUsage(v), v.Help)
		}
	}
	b.WriteString("\nkeys\n")
	for _, km := range DefaultKeyMap().Global("", false) {
		fmt.Fprintf(&b, "  %-12s %s\n", km.Help().Key, km.Help().Desc)
	}
	b.WriteString("  1..6        switch window · esc back/focus · r refresh\n")
	return strings.TrimRight(b.String(), "\n")
}
