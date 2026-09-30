// Package log provides live streaming writers and persisted log file management.
package log

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Line is a single streamed output line.
type Line struct {
	Text string
	// IsErr is true for stderr lines.
	IsErr bool
}

// Stream fans output lines out to subscribers and persists them to a file.
type Stream struct {
	mu   sync.Mutex
	path string
	file *os.File
	subs []chan Line
	lines []string
}

// New creates a Stream persisting to path (parent dirs are created).
func New(path string) (*Stream, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	f, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	header := fmt.Sprintf("# ReleaseForge log — %s\n", time.Now().Format(time.RFC3339))
	if _, err := f.WriteString(header); err != nil {
		f.Close()
		return nil, err
	}
	return &Stream{path: path, file: f}, nil
}

// Path returns the backing log file.
func (s *Stream) Path() string { return s.path }

// Subscribe returns a channel receiving future lines (buffered).
func (s *Stream) Subscribe() <-chan Line {
	s.mu.Lock()
	defer s.mu.Unlock()
	ch := make(chan Line, 256)
	s.subs = append(s.subs, ch)
	return ch
}

// Write appends a line to the file, memory, and subscribers.
func (s *Stream) Write(text string, isErr bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lines = append(s.lines, text)
	fmt.Fprintln(s.file, text)
	for _, ch := range s.subs {
		select {
		case ch <- Line{Text: text, IsErr: isErr}:
		default:
		}
	}
}

// Lines returns a copy of all lines written so far.
func (s *Stream) Lines() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]string, len(s.lines))
	copy(out, s.lines)
	return out
}

// Close flushes, closes the file, and closes subscriber channels.
func (s *Stream) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	err := s.file.Close()
	for _, ch := range s.subs {
		close(ch)
	}
	s.subs = nil
	return err
}

// LogPath builds a timestamped log filename under dir.
func LogPath(dir, prefix string) string {
	stamp := time.Now().Format("20060102-150405")
	return filepath.Join(dir, fmt.Sprintf("%s-%s.log", prefix, stamp))
}
