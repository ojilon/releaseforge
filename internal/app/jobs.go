package app

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ojilon/releaseforge/internal/build"
	rflog "github.com/ojilon/releaseforge/internal/log"
	"github.com/ojilon/releaseforge/internal/project"
	"github.com/ojilon/releaseforge/internal/storage"
	"github.com/ojilon/releaseforge/internal/tui"
	"time"
)

// doBuild starts an async build (debug|release). Returns a status line
// plus a command that streams output back into Update.
func (m *Model) doBuild(variant string) (string, tea.Cmd) {
	variant = strings.ToLower(strings.TrimSpace(variant))
	info, err := project.Detect(m.projectDir)
	if err != nil {
		return tui.ErrorStyle.Render("build: " + err.Error()), nil
	}
	r, err := project.For(info)
	if err != nil {
		return tui.ErrorStyle.Render("build: " + err.Error()), nil
	}
	if m.handle != nil {
		return "a build/test is already running (esc cancels)", nil
	}
	_ = storage.EnsureProjectLayout(m.dataRoot, info.Name)
	out := r.BuildOutput(m.dataRoot, variant)
	bargs, err := r.BuildArgs(variant, out, "")
	if err != nil {
		return tui.ErrorStyle.Render("build: " + err.Error()), nil
	}
	logPath := rflog.LogPath(storage.LogsDir(m.dataRoot, info.Name), "build-"+variant)
	h := build.Start(r.Program(), bargs, build.Options{Dir: info.Root, LogPath: logPath, Project: info.Name})
	m.handle = h
	m.running = true
	m.status = "building " + variant
	m.phase = "building"
	m.pendingRun.kind, m.pendingRun.variant, m.pendingRun.started = "build", variant, time.Now()
	m.pendingOK, m.pendingErr = "build "+variant+" ok — log "+logPath, "build "+variant
	m.refreshProject()
	extra := ""
	if out != "" {
		extra = " → " + out
	}
	return fmt.Sprintf("started build %s%s (log %s, esc cancels)", variant, extra, logPath), m.waitLine()
}

// doTest starts an async test run. Same pattern as doBuild.
func (m *Model) doTest(kind string) (string, tea.Cmd) {
	info, err := project.Detect(m.projectDir)
	if err != nil {
		return tui.ErrorStyle.Render("test: " + err.Error()), nil
	}
	r, err := project.For(info)
	if err != nil {
		return tui.ErrorStyle.Render("test: " + err.Error()), nil
	}
	if m.handle != nil {
		return "a build/test is already running (esc cancels)", nil
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
	h := build.Start(r.Program(), tasks, build.Options{Dir: info.Root, LogPath: logPath, Project: info.Name, Timeout: timeout})
	m.handle = h
	m.running = true
	m.status = "testing " + kind
	m.phase = "testing"
	m.pendingRun.kind, m.pendingRun.variant, m.pendingRun.started = "test", kind, time.Now()
	m.pendingOK, m.pendingErr = "test "+kind+" ok — log "+logPath, "test "+kind
	return fmt.Sprintf("started test %s (log %s, esc cancels)", kind, logPath), m.waitLine()
}
