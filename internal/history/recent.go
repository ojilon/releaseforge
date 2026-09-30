// Package history stores command history, recent projects, and suggestions.
package history

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ojilon/releaseforge/internal/storage"
)

// MaxRecent caps the recent-projects list.
const MaxRecent = 30

// RecentEntry is one opened project.
type RecentEntry struct {
	Path       string `json:"path"`
	Name       string `json:"name"`
	Type       string `json:"type"`
	LastOpened string `json:"last_opened"`
	LastScan   string `json:"last_scan"`
}

// RecentFile is the persisted list of opened folders.
type RecentFile struct {
	Items []RecentEntry `json:"items"`
	Max   int           `json:"max"`
}

// LoadRecent reads the recent-projects file; empty when missing.
func LoadRecent(dataRoot string) (RecentFile, error) {
	var rf RecentFile
	data, err := os.ReadFile(storage.RecentFile(dataRoot))
	if err != nil {
		if os.IsNotExist(err) {
			return RecentFile{Max: MaxRecent}, nil
		}
		return rf, err
	}
	if err := json.Unmarshal(data, &rf); err != nil {
		return rf, err
	}
	if rf.Max <= 0 {
		rf.Max = MaxRecent
	}
	return rf, nil
}

// TouchRecent records an opened/scanned project: absolute clean path, dedupe,
// most-recent-first, capped. lastScan is set when isScan is true.
func TouchRecent(dataRoot, path, name, typ string, isScan bool) error {
	abs, err := filepath.Abs(strings.TrimSpace(path))
	if err != nil {
		return err
	}
	abs = filepath.Clean(abs)
	now := time.Now().UTC().Format(time.RFC3339)
	rf, err := LoadRecent(dataRoot)
	if err != nil {
		rf = RecentFile{Max: MaxRecent}
	}
	max := rf.Max
	if max <= 0 {
		max = MaxRecent
	}
	var items []RecentEntry
	for _, e := range rf.Items {
		if filepath.Clean(e.Path) == abs {
			continue
		}
		items = append(items, e)
	}
	entry := RecentEntry{Path: abs, Name: name, Type: typ, LastOpened: now}
	if isScan {
		entry.LastScan = now
	} else {
		for _, e := range rf.Items {
			if filepath.Clean(e.Path) == abs && e.LastScan != "" {
				entry.LastScan = e.LastScan
				break
			}
		}
	}
	items = append([]RecentEntry{entry}, items...)
	if len(items) > max {
		items = items[:max]
	}
	rf = RecentFile{Items: items, Max: max}
	if err := os.MkdirAll(filepath.Dir(storage.RecentFile(dataRoot)), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(rf, "", "  ")
	if err != nil {
		return err
	}
	tmp := storage.RecentFile(dataRoot) + ".tmp"
	if err := os.WriteFile(tmp, append(data, '\n'), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, storage.RecentFile(dataRoot))
}
