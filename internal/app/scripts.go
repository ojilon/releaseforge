package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ojilon/releaseforge/internal/build"
	rflog "github.com/ojilon/releaseforge/internal/log"
	"github.com/ojilon/releaseforge/internal/project"
	"github.com/ojilon/releaseforge/internal/storage"
	"github.com/ojilon/releaseforge/internal/tui"
)

// scriptExts is the allowlist for the bat/run commands.
var scriptExts = map[string]bool{
	".bat": true, ".cmd": true, ".ps1": true, ".sh": true, ".py": true,
}

// doBat runs a .bat/.cmd/.ps1/.sh/.py script with live streaming logs.
// Usage: bat scripts/foo.bat [args...]
// This is how you run scripts from a previous app version on your PC,
// e.g. compiling releaseforge 0.0.4 while running old scripts/*.bat.
func (m *Model) doBat(args []string) (string, tea.Cmd) {
	if len(args) == 0 {
		return tui.ErrorStyle.Render("bat: usage: bat <script> [args...]"), nil
	}
	script := m.resolvePath(args[0])
	st, err := os.Stat(script)
	if err != nil {
		return tui.ErrorStyle.Render("bat: " + err.Error()), nil
	}
	if st.IsDir() {
		return tui.ErrorStyle.Render("bat: not a file: " + script), nil
	}
	ext := strings.ToLower(filepath.Ext(script))
	if !scriptExts[ext] {
		return tui.ErrorStyle.Render(fmt.Sprintf("bat: %s not allowed (use .bat .cmd .ps1 .sh .py)", ext)), nil
	}
	return m.startScript("bat", script, ext, args[1:])
}

// startScript launches a script async with streaming logs (like build/test).
func (m *Model) startScript(kind, script, ext string, extra []string) (string, tea.Cmd) {
	if m.handle != nil {
		return "a build/test/script is already running (esc cancels)", nil
	}
	program, programArgs := scriptProgram(script, ext, extra)
	info, err := project.Detect(m.projectDir)
	name := filepath.Base(m.projectDir)
	root := m.projectDir
	if err == nil {
		name = info.Name
		root = info.Root
	}
	_ = storage.EnsureProjectLayout(m.dataRoot, name)
	logPath := rflog.LogPath(storage.LogsDir(m.dataRoot, name), kind+"-"+filepath.Base(script))
	h := build.Start(program, programArgs, build.Options{Dir: root, LogPath: logPath, Project: name})
	m.handle = h
	m.running = true
	m.status = "running " + filepath.Base(script)
	m.phase = "building"
	m.pendingRun.kind, m.pendingRun.variant, m.pendingRun.started = "script", filepath.Base(script), time.Now()
	m.pendingOK, m.pendingErr = kind+" ok — log "+logPath, kind+" "+filepath.Base(script)
	return fmt.Sprintf("started %s (log %s, esc cancels)", script, logPath), m.waitLine()
}

// scriptProgram maps a script file to its launcher.
func scriptProgram(script, ext string, extra []string) (string, []string) {
	switch ext {
	case ".bat", ".cmd":
		return "cmd.exe", append([]string{"/C", script}, extra...)
	case ".ps1":
		return "powershell.exe", append([]string{"-ExecutionPolicy", "Bypass", "-File", script}, extra...)
	case ".sh":
		return "sh", append([]string{script}, extra...)
	case ".py":
		return "python", append([]string{script}, extra...)
	default:
		return script, extra
	}
}
