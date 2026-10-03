package log

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// MaxIndexEntries caps the per-project log index.
const MaxIndexEntries = 200

// Entry describes one persisted run for the log index.
type Entry struct {
	Name      string `json:"name"`
	Command   string `json:"command"`
	StartedAt string `json:"started_at"`
	ExitCode  int    `json:"exit_code"`
	Project   string `json:"project"`
}

// IndexPath returns the index file for a logs directory.
func IndexPath(dir string) string { return filepath.Join(dir, "index.json") }

// AppendIndex records a run. Errors are swallowed: the index must never fail
// a build.
func AppendIndex(dir string, e Entry) {
	entries, _ := ReadIndex(dir)
	entries = append(entries, e)
	if len(entries) > MaxIndexEntries {
		entries = entries[len(entries)-MaxIndexEntries:]
	}
	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return
	}
	tmp := IndexPath(dir) + ".tmp"
	if err := os.WriteFile(tmp, append(data, '\n'), 0o644); err != nil {
		return
	}
	_ = os.Rename(tmp, IndexPath(dir))
}

// ReadIndex returns entries oldest-first; missing file yields nil, nil.
func ReadIndex(dir string) ([]Entry, error) {
	data, err := os.ReadFile(IndexPath(dir))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var entries []Entry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, err
	}
	return entries, nil
}

// LastIndexed returns the newest indexed entry, if any.
func LastIndexed(dir string) (Entry, bool) {
	entries, err := ReadIndex(dir)
	if err != nil || len(entries) == 0 {
		return Entry{}, false
	}
	return entries[len(entries)-1], true
}

// listByMtime returns log filenames newest-first, for index fallback.
func listByMtime(dir string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	type fi struct {
		name string
		mod  int64
	}
	var files []fi
	for _, e := range entries {
		if e.IsDir() || e.Name() == "index.json" {
			continue
		}
		if st, err := e.Info(); err == nil {
			files = append(files, fi{e.Name(), st.ModTime().UnixNano()})
		}
	}
	sort.Slice(files, func(i, j int) bool { return files[i].mod > files[j].mod })
	out := make([]string, 0, len(files))
	for _, f := range files {
		out = append(out, f.name)
	}
	return out
}

// List returns log filenames newest-first, excluding the index itself.
func List(dir string) []string { return listByMtime(dir) }

// Newest returns the newest log filename, preferring the index.
func Newest(dir string) (string, error) {
	if e, ok := LastIndexed(dir); ok && e.Name != "" {
		if st, err := os.Stat(filepath.Join(dir, e.Name)); err == nil && !st.IsDir() {
			return e.Name, nil
		}
	}
	if names := listByMtime(dir); len(names) > 0 {
		return names[0], nil
	}
	return "", os.ErrNotExist
}

// Resolve maps a target (exact name or unambiguous prefix) to a filename.
func Resolve(dir, target string) (string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() && e.Name() != "index.json" {
			names = append(names, e.Name())
		}
	}
	for _, n := range names {
		if n == target {
			return n, nil
		}
	}
	var cands []string
	for _, n := range names {
		if target != "" && strings.HasPrefix(n, target) {
			cands = append(cands, n)
		}
	}
	switch len(cands) {
	case 0:
		return "", os.ErrNotExist
	case 1:
		return cands[0], nil
	default:
		return "", &AmbiguousError{Candidates: cands}
	}
}

// AmbiguousError lists the candidates of an ambiguous prefix.
type AmbiguousError struct {
	Candidates []string
}

func (e *AmbiguousError) Error() string {
	out := "ambiguous log name, candidates:"
	for _, c := range e.Candidates {
		out += "\n  " + c
	}
	return out
}

// Tail returns the last n lines of a file.
func Tail(path string, n int) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if n <= 0 {
		return nil, nil
	}
	lines := splitLines(string(data))
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	return lines, nil
}

// Head returns up to maxBytes plus whether output was truncated.
func Head(path string, maxBytes int) ([]byte, bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, false, err
	}
	if len(data) > maxBytes {
		return data[:maxBytes], true, nil
	}
	return data, false, nil
}

func splitLines(s string) []string {
	var out []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			out = append(out, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		out = append(out, s[start:])
	}
	return out
}
