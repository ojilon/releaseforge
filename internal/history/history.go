// Package history stores command history and provides suggestion candidates.
package history

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/ojilon/releaseforge/internal/storage"
)

// MaxCommands caps the persisted command history.
const MaxCommands = 500

// AppendCommand records one command-bar line, keeping the newest MaxCommands.
// Best-effort: errors are swallowed so history never breaks the TUI.
func AppendCommand(dataRoot, line string) {
	line = strings.TrimSpace(line)
	if line == "" {
		return
	}
	path := storage.HistoryFile(dataRoot)
	var lines []string
	if data, err := os.ReadFile(path); err == nil {
		for _, l := range strings.Split(string(data), "\n") {
			if t := strings.TrimSpace(l); t != "" {
				lines = append(lines, t)
			}
		}
	}
	lines = append(lines, line)
	if len(lines) > MaxCommands {
		lines = lines[len(lines)-MaxCommands:]
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, []byte(strings.Join(lines, "\n")+"\n"), 0o644); err != nil {
		return
	}
	_ = os.Rename(tmp, path)
}

// LoadCommands returns persisted command history oldest-first, empty when absent.
func LoadCommands(dataRoot string) []string {
	data, err := os.ReadFile(storage.HistoryFile(dataRoot))
	if err != nil {
		return nil
	}
	var out []string
	for _, l := range strings.Split(string(data), "\n") {
		if t := strings.TrimSpace(l); t != "" {
			out = append(out, t)
		}
	}
	return out
}
