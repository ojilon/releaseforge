// Package metrics records local build/test history for the status line.
// Data capture only; there is no dashboard UI in v0.0.2.
package metrics

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ojilon/releaseforge/internal/storage"
)

// MaxRecords caps the per-project metrics file.
const MaxRecords = 200

// Record is one finished build or test run.
type Record struct {
	At         string `json:"at"`
	Project    string `json:"project"`
	Kind       string `json:"kind"`    // "build" | "test"
	Variant    string `json:"variant"` // debug|release, unit|instrumented|all
	Success    bool   `json:"success"`
	DurationMs int64  `json:"duration_ms"`
}

func path(dataRoot, project string) string {
	return filepath.Join(storage.ProjectDir(dataRoot, project), "metrics.jsonl")
}

// Append records a run. Best-effort: it never fails the caller.
func Append(dataRoot string, r Record) {
	if strings.TrimSpace(r.At) == "" {
		r.At = time.Now().UTC().Format(time.RFC3339)
	}
	p := path(dataRoot, r.Project)
	var lines []string
	if data, err := os.ReadFile(p); err == nil {
		for _, l := range strings.Split(string(data), "\n") {
			if strings.TrimSpace(l) == "" {
				continue
			}
			var probe Record
			if json.Unmarshal([]byte(l), &probe) != nil {
				continue // tolerate corrupt lines
			}
			lines = append(lines, l)
		}
	}
	data, err := json.Marshal(r)
	if err != nil {
		return
	}
	lines = append(lines, string(data))
	if len(lines) > MaxRecords {
		lines = lines[len(lines)-MaxRecords:]
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return
	}
	tmp := p + ".tmp"
	if err := os.WriteFile(tmp, []byte(strings.Join(lines, "\n")+"\n"), 0o644); err != nil {
		return
	}
	_ = os.Rename(tmp, p)
}

// Last returns the newest record of a kind, or false when none exists.
func Last(dataRoot, project, kind string) (Record, bool) {
	data, err := os.ReadFile(path(dataRoot, project))
	if err != nil {
		return Record{}, false
	}
	var last Record
	found := false
	for _, l := range strings.Split(string(data), "\n") {
		if strings.TrimSpace(l) == "" {
			continue
		}
		var r Record
		if json.Unmarshal([]byte(l), &r) != nil || r.Kind != kind {
			continue
		}
		last, found = r, true
	}
	return last, found
}

// Ago renders an RFC3339 timestamp as "3m ago" style text.
func Ago(at string) string {
	t, err := time.Parse(time.RFC3339, at)
	if err != nil {
		return "unknown time"
	}
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	case d < time.Hour*24:
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	default:
		return fmt.Sprintf("%dd ago", int(d.Hours()/24))
	}
}

// LastBuildLine renders the last-build status line shared by CLI and TUI.
func LastBuildLine(dataRoot, project string) string {
	rec, ok := Last(dataRoot, project, "build")
	if !ok {
		return "last build: none yet"
	}
	result := "ok"
	if !rec.Success {
		result = "FAILED"
	}
	return fmt.Sprintf("last build: %s %s (%s, %s)",
		rec.Variant, result, DurationText(rec.DurationMs), Ago(rec.At))
}

// DurationText renders milliseconds as "850ms" or "12.3s".
func DurationText(ms int64) string {
	if ms < 1000 {
		return fmt.Sprintf("%dms", ms)
	}
	return fmt.Sprintf("%.1fs", float64(ms)/1000)
}
