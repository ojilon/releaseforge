package app

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ojilon/releaseforge/internal/tui"
)

// doLs lists folder contents like `ls` (dir works too).
// Usage: ls [path] — relative paths resolve against the project.
func (m *Model) doLs(args []string) string {
	dir := m.projectDir
	if len(args) > 0 {
		dir = m.resolvePath(args[0])
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return tui.ErrorStyle.Render("ls: " + err.Error())
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		n := e.Name()
		if e.IsDir() {
			n += "/"
		}
		names = append(names, n)
	}
	sort.Strings(names)
	if len(names) == 0 {
		return dir + ": empty"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "%s:\n", dir)
	for _, n := range names {
		b.WriteString("  " + n + "\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

// doPwd prints the current project directory.
func (m *Model) doPwd() string {
	abs, err := filepath.Abs(m.projectDir)
	if err != nil {
		return m.projectDir
	}
	return abs
}

// doCat prints a small text file (capped: 100 lines, 64KB).
// Usage: cat <file>
func (m *Model) doCat(args []string) string {
	if len(args) == 0 {
		return tui.ErrorStyle.Render("cat: usage: cat <file>")
	}
	path := m.resolvePath(args[0])
	data, err := os.ReadFile(path)
	if err != nil {
		return tui.ErrorStyle.Render("cat: " + err.Error())
	}
	const maxBytes = 64 * 1024
	truncated := false
	if len(data) > maxBytes {
		data = data[:maxBytes]
		truncated = true
	}
	lines := strings.Split(string(data), "\n")
	if len(lines) > 100 {
		lines = lines[:100]
		truncated = true
	}
	out := strings.TrimRight(strings.Join(lines, "\n"), "\n")
	if truncated {
		out += "\n…truncated (100 lines / 64KB max)"
	}
	return out
}

// resolvePath resolves p against the project dir when relative.
func (m *Model) resolvePath(p string) string {
	if filepath.IsAbs(p) {
		return filepath.Clean(p)
	}
	return filepath.Join(m.projectDir, filepath.Clean(p))
}
