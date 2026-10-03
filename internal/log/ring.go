package log

import "sync"

// Ring is a bounded oldest-dropped string buffer for live output views.
// The persisted log file stays complete; the Ring only bounds memory.
type Ring struct {
	mu    sync.Mutex
	cap   int
	buf   []string
	next  int
	count int
}

// NewRing creates a Ring holding at most cap lines (cap < 1 means 1).
func NewRing(cap int) *Ring {
	if cap < 1 {
		cap = 1
	}
	return &Ring{cap: cap, buf: make([]string, cap)}
}

// Append adds a line, overwriting the oldest when full.
func (r *Ring) Append(s string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.buf[r.next] = s
	r.next = (r.next + 1) % r.cap
	if r.count < r.cap {
		r.count++
	}
}

// Snapshot returns lines oldest-first.
func (r *Ring) Snapshot() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]string, 0, r.count)
	for i := 0; i < r.count; i++ {
		out = append(out, r.buf[(r.next-r.count+i+r.cap*2)%r.cap])
	}
	return out
}

// Len returns the number of stored lines.
func (r *Ring) Len() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.count
}
