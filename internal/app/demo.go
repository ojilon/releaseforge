package app

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
	"github.com/ojilon/releaseforge/internal/metrics"
	"github.com/ojilon/releaseforge/internal/project"
	"github.com/ojilon/releaseforge/internal/storage"
	"github.com/ojilon/releaseforge/internal/tui"
)

// DemoFrame is one captured screen. ANSI is the actual View() output with
// colors, so the HTML demo is a faithful recording of the running TUI rather
// than a mock-up.
type DemoFrame struct {
	Title   string
	Caption string
	Width   int
	Height  int
	ANSI    string
}

const demoWidth, demoHeight = 132, 36

// DemoFrames builds a throwaway project (git history, scan cache, metrics,
// logs) and drives the real Model through a scripted tour. No user data is
// touched and the temp tree is removed before returning.
func DemoFrames() []DemoFrame {
	// Deterministic colors for the recording, independent of the caller's
	// terminal profile.
	lipgloss.SetColorProfile(termenv.TrueColor)

	tmp, err := os.MkdirTemp("", "releaseforge-demo-")
	if err != nil {
		return []DemoFrame{{Title: "error", Caption: err.Error(), Width: 40, Height: 3, ANSI: err.Error()}}
	}
	defer os.RemoveAll(tmp)
	dataRoot := filepath.Join(tmp, "data")
	projectDir := filepath.Join(tmp, "atlas-app")
	if err := os.MkdirAll(projectDir, 0o755); err != nil {
		return nil
	}
	_ = storage.EnsureLayout(dataRoot)
	seedDemoProject(projectDir)
	seedDemoGit(projectDir)
	if snap, _, err := project.Scan(projectDir); err == nil {
		_, _ = project.WriteScan(dataRoot, snap)
	}
	seedDemoHistory(dataRoot, projectDir)

	m := New(dataRoot, projectDir)
	m.setTheme("dark")
	m = applySize(m, demoWidth, demoHeight)

	var frames []DemoFrame
	add := func(title, caption string, w, h int) {
		if w != m.width || h != m.height {
			m = applySize(m, w, h)
			m.refreshWindows() // rebuild cached panes at the new width
		}
		frames = append(frames, DemoFrame{
			Title: title, Caption: caption, Width: w, Height: h, ANSI: m.View(),
		})
	}

	add("01 · Status", "Landing window: project facts, cached scan, git state, recent runs and a live side card.", demoWidth, demoHeight)

	m.execCommand("help")
	add("02 · Generated help", "`help` or ? opens the command reference generated from the verb table — it can never go stale.", demoWidth, demoHeight)

	m.input.SetValue("")
	m.switchTo(WinGit)
	m.windows[WinGit].(*gitWindow).view = gitLog
	m.refreshActive()
	add("03 · Git", "Window 3: commits as a bubbles/table. Keys s / d / l switch status, diff and log; r refreshes.", demoWidth, demoHeight)

	m.switchTo(WinLogs)
	add("04 · Logs", "Window 5: a bubbles/list of persisted logs with a scrollable preview. Selection survives window switches.", demoWidth, demoHeight)

	// Scripted live build in the Console.
	m.switchTo(WinBuild)
	m.phase = PhaseBuilding
	m.status = "building release"
	m.run = &RunState{Kind: "build", Variant: "release", Started: time.Now().Add(-12 * time.Second)}
	m.elapsed = 12 * time.Second
	m.console().Append(
		tui.ChipStyle.Render("❯ build release 0.0.4"),
		"started build release → "+filepath.Join(dataRoot, "projects", "atlas-app", "builds"),
		"",
		tui.SubtitleStyle.Render("  GO BUILD"),
		"  compiling github.com/example/atlas-app/internal/scan",
		"  compiling github.com/example/atlas-app/internal/build",
		"  linking ./atlas-app",
		"",
		tui.HelpStyle.Render("  output streams into the console; the full log is persisted under the data root"),
	)
	m.console().StartSpinner()
	m.refreshWindows()
	add("05 · Live build", "Window 4: streamed output, a bubbles/spinner with elapsed time, follow-mode scrolling and ctrl+x cancel.", demoWidth, demoHeight)

	// Palette overlay.
	m.phase = PhaseIdle
	m.run = nil
	m.elapsed = 0
	m.status = "ready"
	m.refreshWindows()
	m.focus = FocusPalette
	m.palette.open(m.paletteItems(), "all")
	m.palette.query.SetValue("bui")
	m.palette.rank()
	add("06 · Command palette", "ctrl+p: fuzzy search over commands, recent projects, logs and history. ctrl+r searches history only.", demoWidth, demoHeight)

	// Folder picker overlay.
	m.focus = FocusBar
	m.palette = newPalette()
	pw, ph := m.paletteSize()
	m.palette.Resize(pw, ph)
	if cmd := m.openPicker(); cmd != nil {
		if msg := cmd(); msg != nil {
			updated, _ := m.picker.Update(msg)
			*m.picker = updated
		}
	}
	add("07 · Folder picker", "ctrl+o or `open` with no path: bubbles/filepicker in directory mode. `s` uses the folder you are in.", demoWidth, demoHeight)

	// Compact responsive layout.
	m.focus = FocusBar
	m.picker = nil
	m.switchTo(WinStatus)
	add("08 · Compact terminal", "The same shell at 70×20: numbered tabs, single-line header, no side card. Nothing wraps.", 70, 20)

	// Tiny terminal.
	add("09 · Tiny terminal", "Below 60 columns the chrome collapses to bare lines so the body still has room.", 52, 14)

	// Light theme.
	m = applySize(m, demoWidth, demoHeight)
	m.setTheme("light")
	m.switchTo(WinHelp)
	add("10 · Light theme", "`theme light` (or auto-detection on a light terminal) repaints chrome, tables and the input together.", demoWidth, demoHeight)
	m.setTheme("dark")

	return frames
}

func applySize(m Model, w, h int) Model {
	next, _ := m.Update(tea.WindowSizeMsg{Width: w, Height: h})
	return next.(Model)
}

func seedDemoProject(dir string) {
	files := map[string]string{
		"go.mod":      "module github.com/example/atlas-app\n\ngo 1.23.0\n",
		"VERSION":     "0.0.4\n",
		"README.md":   "# Atlas App\n\nA small local release pipeline demo.\n",
		"main.go":     "package main\n\nfunc main() {}\n",
		"internal.go": "package main\n\nfunc scan() {}\n",
	}
	for name, body := range files {
		_ = os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644)
	}
}

func seedDemoGit(dir string) {
	if _, err := exec.LookPath("git"); err != nil {
		return // demo still works without git; windows show "not a repository"
	}
	run := func(args ...string) {
		c := exec.Command("git", args...)
		c.Dir = dir
		_ = c.Run()
	}
	run("init", "-q")
	messages := []string{
		"feat: add project scan cache",
		"fix: honour quoted Windows paths",
		"refactor: split the TUI into windows",
		"docs: describe the release pipeline",
	}
	for i, msg := range messages {
		_ = os.WriteFile(filepath.Join(dir, "changelog.txt"),
			[]byte(fmt.Sprintf("step %d: %s\n", i+1, msg)), 0o644)
		run("add", ".")
		run("-c", "user.name=ReleaseForge Demo", "-c", "user.email=demo@example.invalid",
			"commit", "-q", "-m", msg)
	}
	run("tag", "v0.0.3")
}

func seedDemoHistory(dataRoot, projectDir string) {
	name := "atlas-app"
	for i, r := range []metrics.Record{
		{Kind: "test", Variant: "unit", Success: true, DurationMs: 1840},
		{Kind: "build", Variant: "debug", Success: true, DurationMs: 3210},
		{Kind: "build", Variant: "release", Success: false, DurationMs: 5400},
		{Kind: "test", Variant: "unit", Success: true, DurationMs: 1720},
	} {
		r.Project = name
		r.At = time.Now().Add(time.Duration(-4+i) * time.Hour).UTC().Format(time.RFC3339)
		metrics.Append(dataRoot, r)
	}
	logs := storage.LogsDir(dataRoot, name)
	_ = os.MkdirAll(logs, 0o755)
	_ = os.WriteFile(filepath.Join(logs, "build-release.log"), []byte(strings.Join([]string{
		"[build] release 0.0.4",
		"[go] compiling internal/scan",
		"[go] compiling internal/build",
		"[go] linking ./atlas-app",
		"[build] ok 3.2s",
	}, "\n")), 0o644)
	_ = os.WriteFile(filepath.Join(logs, "test-unit.log"), []byte(strings.Join([]string{
		"ok  github.com/example/atlas-app/internal/project",
		"ok  github.com/example/atlas-app/internal/tui",
		"ok  github.com/example/atlas-app/internal/app",
		"PASS",
	}, "\n")), 0o644)
}
