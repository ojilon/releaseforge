package app

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ojilon/releaseforge/internal/build"
	rflog "github.com/ojilon/releaseforge/internal/log"
	"github.com/ojilon/releaseforge/internal/project"
	"github.com/ojilon/releaseforge/internal/storage"
	"github.com/ojilon/releaseforge/internal/tui"
)

// doBuild starts an async build (debug|release [version]). With a version it
// writes the version file first (file only, never committed) so the stamp
// uses it. Returns a status line plus the streaming command.
func (m *Model) doBuild(variant, setVer string) (string, tea.Cmd) {
	variant = strings.ToLower(strings.TrimSpace(variant))
	setVer = strings.TrimSpace(setVer)
	if m.run != nil {
		return "a build/test/script is already running (ctrl+x cancels)", nil
	}
	info, err := project.Detect(m.projectDir)
	if err != nil {
		return tui.ErrorStyle.Render("build: " + err.Error()), nil
	}
	r, err := project.For(info)
	if err != nil {
		return tui.ErrorStyle.Render("build: " + err.Error()), nil
	}
	if setVer != "" {
		if err := project.ValidateVersion(setVer); err != nil {
			return tui.ErrorStyle.Render("build: " + err.Error()), nil
		}
		if _, err := r.VersionWrite(setVer); err != nil {
			return tui.ErrorStyle.Render("build: version: " + err.Error()), nil
		}
	}
	_ = storage.EnsureProjectLayout(m.dataRoot, info.Name)
	out := r.BuildOutput(m.dataRoot, variant)
	bargs, err := r.BuildArgs(variant, out, "")
	if err != nil {
		return tui.ErrorStyle.Render("build: " + err.Error()), nil
	}
	logPath := rflog.LogPath(storage.LogsDir(m.dataRoot, info.Name), "build-"+variant)
	h := build.Start(r.Program(), bargs, build.Options{Dir: info.Root, LogPath: logPath, Project: info.Name})
	m.installRun("build", variant, PhaseBuilding, h,
		"build "+variant+" ok — log "+logPath, "build "+variant)
	m.status = "building " + variant
	m.refreshProject()

	extra := ""
	if out != "" {
		extra = " → " + out
	}
	verNote := ""
	if setVer != "" {
		verNote = " (version " + setVer + ")"
	}
	return fmt.Sprintf("started build %s%s%s (log %s, ctrl+x cancels)", variant, verNote, extra, logPath),
		tea.Batch(m.console().StartSpinner(), m.waitLine())
}

// doTest starts an async test run (same pattern as doBuild).
func (m *Model) doTest(kind string) (string, tea.Cmd) {
	if m.run != nil {
		return "a build/test/script is already running (ctrl+x cancels)", nil
	}
	info, err := project.Detect(m.projectDir)
	if err != nil {
		return tui.ErrorStyle.Render("test: " + err.Error()), nil
	}
	r, err := project.For(info)
	if err != nil {
		return tui.ErrorStyle.Render("test: " + err.Error()), nil
	}
	tasks, err := r.TestArgs(kind)
	if err != nil {
		return tui.ErrorStyle.Render("test: " + err.Error()), nil
	}
	_ = storage.EnsureProjectLayout(m.dataRoot, info.Name)
	logPath := rflog.LogPath(storage.LogsDir(m.dataRoot, info.Name), "test-"+kind)
	timeout := time.Duration(0)
	if kind == "instrumented" || kind == "all" {
		timeout = 30 * time.Minute
	}
	h := build.Start(r.Program(), tasks, build.Options{Dir: info.Root, LogPath: logPath,
		Project: info.Name, Timeout: timeout})
	m.installRun("test", kind, PhaseTesting, h,
		"test "+kind+" ok — log "+logPath, "test "+kind)
	m.status = "testing " + kind
	m.refreshProject()
	return fmt.Sprintf("started test %s (log %s, ctrl+x cancels)", kind, logPath),
		tea.Batch(m.console().StartSpinner(), m.waitLine())
}

// installRun is the single place an async handle becomes app state.
func (m *Model) installRun(kind, variant string, phase Phase, h *build.Handle, okLabel, failLabel string) {
	m.run = &RunState{
		Kind: kind, Variant: variant, Started: time.Now(), Handle: h,
		OKLabel: okLabel, FailLabel: failLabel,
	}
	m.setPhase(phase)
}
