// Package app is the Bubble Tea TUI frontend.
//
// Think of it as: frontend (this package) calls backend
// (internal/project, internal/build, internal/git, ...) and
// prints the returned strings. No backend logic lives here.
//
// Files:
//   model.go    - Model struct + New (what the screen remembers)
//   run.go      - starting the TUI + async job plumbing
//   update.go   - Update: keyboard and window events
//   view.go     - View: drawing header, log pane, input bar
//   commands.go - parse + exec: typed command dispatch
//   scan.go     - scan / recent commands
//   version.go  - version command
//   jobs.go     - build / test commands (start async jobs)
//   logs.go     - logs command
//   notes.go    - notes command
package app

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/ojilon/releaseforge/internal/build"
	"github.com/ojilon/releaseforge/internal/git"
	"github.com/ojilon/releaseforge/internal/history"
	rflog "github.com/ojilon/releaseforge/internal/log"
	"github.com/ojilon/releaseforge/internal/project"
	"github.com/ojilon/releaseforge/internal/tui"
)

// NotImplemented is the shared stub error for unbuilt surfaces.
func NotImplemented(name string) error {
	return fmt.Errorf("%s: not implemented yet — see docs/08-implementation-plan.md", name)
}

// ringCap bounds viewport memory; the log file stays complete.
const ringCap = 2000

// Model is the TUI root. It remembers:
//   - where we are (dataRoot, projectDir, info)
//   - what to show (status, phase, branch, viewport ring)
//   - what is running (handle, pendingRun)
type Model struct {
	dataRoot   string
	projectDir string
	info       *project.Info
	version    string
	status     string

	input    textinput.Model
	viewport viewport.Model
	ring     *rflog.Ring
	ready    bool

	history   []string
	histIndex int

	// Presentation state.
	phase  string // idle|building|testing|failed|ok
	width  int
	branch string
	dirty  bool
	spark  string

	// Active async run (nil when idle).
	handle     *build.Handle
	pendingOK  string
	pendingErr string
	pendingRun struct {
		kind    string // "build" | "test"
		variant string
		started time.Time
	}

	running bool
}

// New creates the root model with empty panes and loaded history.
func New(dataRoot, projectDir string) Model {
	ti := textinput.New()
	ti.Placeholder = "type a command (help, status, scan, recent, version, build, test, notes, logs, quit)"
	ti.Focus()
	ti.CharLimit = 512
	ti.Prompt = "> "

	ring := rflog.NewRing(ringCap)
	hist := history.LoadCommands(dataRoot)
	m := Model{dataRoot: dataRoot, projectDir: projectDir, input: ti,
		ring: ring, status: "ready", phase: "idle",
		history: hist, histIndex: len(hist)}
	m.ring.Append(strings.TrimRight(m.welcome(), "\n"))
	m.refreshProject()
	return m
}

// refreshProject re-detects the project and updates name/version/branch.
func (m *Model) refreshProject() {
	info, err := project.Detect(m.projectDir)
	if err != nil {
		m.info = nil
		m.version = ""
		return
	}
	m.info = &info
	if vsnap, ok := project.CachedVersion(m.dataRoot, info.Name, info.Root); ok {
		if vsnap.Code != "" {
			m.version = fmt.Sprintf("v%s (%s)", vsnap.Name, vsnap.Code)
		} else {
			m.version = "v" + vsnap.Name
		}
		return
	}
	if code, name, err := project.CurrentVersion(info); err == nil && name != "" {
		if code != "" {
			m.version = fmt.Sprintf("v%s (%s)", name, code)
		} else {
			m.version = "v" + name
		}
	} else {
		m.version = info.Type
	}
	m.branch, m.dirty, m.spark = "", false, ""
	if git.IsRepo(info.Root) {
		m.branch = git.CurrentBranch(info.Root)
		m.dirty = !git.IsClean(info.Root)
		if snap, err := project.LoadScan(m.dataRoot, info.Name); err == nil &&
			project.Fresh(m.dataRoot, info.Name, info.Root) {
			var dates []string
			for _, c := range snap.Git.RecentCommits {
				dates = append(dates, c.Date)
			}
			m.spark = tui.Sparkline(tui.BucketCommits(dates, time.Now()))
		}
	}
}

// projectName is the display name for header and metrics.
func (m *Model) projectName() string {
	if m.info != nil {
		return m.info.Name
	}
	return filepath.Base(m.projectDir)
}

// ruleWidth bounds separator rendering to sane widths.
func (m *Model) ruleWidth() int {
	if m.width >= 80 {
		return m.width
	}
	return 80
}
