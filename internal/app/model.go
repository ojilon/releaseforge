// Package app is the Bubble Tea TUI frontend.
//
// Architecture (v0.0.3 redesign, docs/04-tui-design.md):
//
//	Model (shell)        - header, tab strip, command bar, footer, routing
//	Window (x6)          - Status | Tests | Git | Console | Logs | Help
//	Overlays             - fuzzy command palette (ctrl+p), folder picker (ctrl+o)
//	Backend              - internal/project, internal/build, internal/git, …
//
// The shell routes every message to exactly one consumer (single-consumer
// dispatch, docs/sweep-2mo/41). Windows are independent tea-ish components:
// they own their own viewport/cursor state and never reach back into Model.
package app

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/ojilon/releaseforge/internal/build"
	"github.com/ojilon/releaseforge/internal/git"
	"github.com/ojilon/releaseforge/internal/history"
	"github.com/ojilon/releaseforge/internal/metrics"
	"github.com/ojilon/releaseforge/internal/project"
	"github.com/ojilon/releaseforge/internal/tui"
	rfversion "github.com/ojilon/releaseforge/internal/version"
)

// Run is the single owner of an async build/test/script. Replaces the old
// handle + running + pendingRun + pendingOK/pendingErr scatter.
type RunState struct {
	Kind      string // build | test | script
	Variant   string
	Started   time.Time
	Handle    *build.Handle
	OKLabel   string
	FailLabel string
	Cancelled bool
}

// Model is the shell. It remembers where we are, what is showing, what is
// running and which component owns the keyboard.
type Model struct {
	dataRoot   string
	projectDir string
	info       *project.Info
	version    string
	status     string
	themeName  string

	input     textinput.Model
	history   []string
	histIndex int
	keys      tui.KeyMap

	width, height int
	layout        tui.Layout
	ready         bool

	phase   Phase
	run     *RunState
	elapsed time.Duration

	windows []Window
	active  WindowID
	focus   Focus

	palette   *paletteModel
	picker    *filepicker.Model
	pickerDir string

	branch string
	dirty  bool
	spark  string

	hint string // one-line completion / status hint shown above the footer
}

// New creates the root model: detected theme, six windows, loaded history,
// detected project and a welcome console. No I/O beyond the data root.
func New(dataRoot, projectDir string) Model {
	m := Model{
		dataRoot:   dataRoot,
		projectDir: projectDir,
		status:     "ready",
		phase:      PhaseIdle,
		keys:       tui.DefaultKeyMap(),
		layout:     tui.Compute(100, 30), // safe default until WindowSizeMsg
	}
	m.themeName = tui.ApplyDetectedTheme()

	ti := textinput.New()
	ti.Placeholder = "type a command · tab completes · ctrl+p palette · ctrl+o folder"
	ti.Focus()
	ti.CharLimit = 512
	ti.Prompt = "❯ "
	tui.PaintInput(&ti)
	m.input = ti

	m.history = history.LoadCommands(dataRoot)
	m.histIndex = len(m.history)

	m.windows = []Window{
		newStatusWindow(), newTestsWindow(), newGitWindow(),
		newBuildWindow(), newLogsWindow(), newHelpWindow(),
	}
	m.active = WinStatus
	m.focus = FocusBar
	m.palette = newPalette()
	for _, w := range m.windows {
		w.Resize(m.layout)
	}
	pw, ph := m.paletteSize()
	m.palette.Resize(pw, ph)

	m.refreshProject()
	m.console().AppendText(m.welcome())
	return m
}

// console is the build window, which owns the transcript.
func (m *Model) console() *buildWindow { return m.windows[WinBuild].(*buildWindow) }

// window returns the active window (never nil).
func (m *Model) window() Window { return windowByID(m.windows, m.active) }

// Init starts the blinking cursor and the 1s clock used for elapsed run time.
func (m Model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, tickCmd())
}

// context builds the immutable snapshot handed to windows.
func (m *Model) context() Context {
	c := Context{
		DataRoot:    m.dataRoot,
		ProjectDir:  m.projectDir,
		ProjectName: m.projectName(),
		Info:        m.info,
		Version:     m.version,
		Branch:      m.branch,
		Dirty:       m.dirty,
		Spark:       m.spark,
		Status:      m.status,
		ThemeName:   m.themeName,
		Phase:       m.phase,
		Elapsed:     m.elapsed,
		Layout:      m.layout,
	}
	if m.run != nil {
		c.RunKind, c.RunVariant = m.run.Kind, m.run.Variant
	}
	return c
}

// refreshProject re-detects the project, updates header facts and refreshes
// every window's cached data (only on real events, never on each keystroke).
func (m *Model) refreshProject() {
	info, err := project.Detect(m.projectDir)
	if err != nil {
		m.info = nil
		m.version = ""
		m.branch, m.dirty, m.spark = "", false, ""
	} else {
		m.info = &info
		if vsnap, ok := project.CachedVersion(m.dataRoot, info.Name, info.Root); ok {
			m.version = versionLabel(vsnap.Name, vsnap.Code)
		} else if code, name, err := project.CurrentVersion(info); err == nil && name != "" {
			m.version = versionLabel(name, code)
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
	if m.ready {
		m.refreshWindows()
	}
}

func versionLabel(name, code string) string {
	if code != "" {
		return fmt.Sprintf("v%s (%s)", name, code)
	}
	return "v" + name
}

// refreshWindows rebuilds all cached window data. Called on project change,
// window switch and explicit refresh.
func (m *Model) refreshWindows() {
	ctx := m.context()
	for _, w := range m.windows {
		w.Refresh(ctx)
	}
}

func (m *Model) refreshActive() {
	if w := m.window(); w != nil {
		w.Refresh(m.context())
	}
}

// switchTo activates a window and refreshes its data. Other windows keep
// their scroll/cursor state.
func (m *Model) switchTo(id WindowID) {
	if int(id) >= len(m.windows) {
		return
	}
	m.active = id
	m.refreshActive()
}

func (m *Model) projectName() string {
	if m.info != nil {
		return m.info.Name
	}
	return filepath.Base(m.projectDir)
}

// ruleWidth bounds separator rendering to sane widths.
func (m *Model) ruleWidth() int {
	w := m.layout.Content.Width
	if w < 20 {
		w = 20
	}
	return w
}

// appendLine writes one (possibly multi-line) block to the console.
func (m *Model) appendLine(s string) {
	m.console().AppendText(s)
}

// welcome is the console's first block: an in-app hint card.
func (m *Model) welcome() string {
	return strings.Join([]string{
		tui.TitleStyle.Render("  ReleaseForge — local release, build & test control plane"),
		"",
		tui.HelpStyle.Render("  scan a folder with  open <path>  ·  ctrl+o picks one interactively"),
		tui.HelpStyle.Render("  1..6 switch windows  ·  ctrl+p command palette  ·  ? generated help"),
		tui.HelpStyle.Render("  type a command below; output streams into window 4 (Console)"),
		"",
		tui.Rule(m.ruleWidth()),
	}, "\n")
}

// statusText is the `status` command output (also used by the CLI twin).
func (m *Model) statusText() string {
	info, err := project.Detect(m.projectDir)
	if err != nil {
		return "status: " + err.Error()
	}
	ver := info.Type
	if vsnap, ok := project.CachedVersion(m.dataRoot, info.Name, info.Root); ok {
		if vsnap.Code != "" {
			ver = fmt.Sprintf("%s (code %s)", vsnap.Name, vsnap.Code)
		} else {
			ver = vsnap.Name
		}
	} else if code, name, err := project.CurrentVersion(info); err == nil && name != "" {
		if code != "" {
			ver = fmt.Sprintf("%s (code %s)", name, code)
		} else {
			ver = name
		}
	}
	gitLine := "not a repository"
	if git.IsRepo(info.Root) {
		state := "clean"
		if !git.IsClean(info.Root) {
			state = "dirty"
		}
		gitLine = fmt.Sprintf("%s · %s", git.CurrentBranch(info.Root), state)
	}
	return fmt.Sprintf("releaseforge: %s\nproject: %s\nroot: %s\ntype: %s\nversion: %s\ngit: %s\ndata-root: %s\n%s",
		rfversion.ToolVersion, info.Name, info.Root, info.Type, ver, gitLine, m.dataRoot,
		metrics.LastBuildLine(m.dataRoot, info.Name))
}

// setTheme applies a preset and repaints everything that caches colors.
func (m *Model) setTheme(name string) string {
	if !tui.SetTheme(name) {
		return tui.ErrorStyle.Render(fmt.Sprintf("theme: unknown %q (use %s)", name, strings.Join(tui.ThemeNames(), "|")))
	}
	m.themeName = tui.Current.Name
	tui.PaintInput(&m.input)
	if m.palette != nil {
		tui.PaintInput(&m.palette.query)
	}
	m.refreshWindows()
	return "theme: " + m.themeName
}
