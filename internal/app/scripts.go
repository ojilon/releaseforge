package app

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

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
// Usage: bat <script> [args...] — the bridge to older scripts/*.bat flows.
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
	program, programArgs, err := scriptProgram(script, ext, args[1:])
	if err != nil {
		return tui.ErrorStyle.Render("bat: " + err.Error()), nil
	}
	return m.startScript(program, programArgs, script)
}

// startScript launches a script async with streaming logs (like build/test).
func (m *Model) startScript(program string, programArgs []string, script string) (string, tea.Cmd) {
	if m.run != nil {
		return "a build/test/script is already running (ctrl+x cancels)", nil
	}
	info, err := project.Detect(m.projectDir)
	name := filepath.Base(m.projectDir)
	root := m.projectDir
	if err == nil {
		name, root = info.Name, info.Root
	}
	_ = storage.EnsureProjectLayout(m.dataRoot, name)
	logPath := rflog.LogPath(storage.LogsDir(m.dataRoot, name), "bat-"+filepath.Base(script))
	h := build.Start(program, programArgs, build.Options{Dir: root, LogPath: logPath, Project: name})
	m.installRun("script", filepath.Base(script), PhaseBuilding, h,
		"script ok — log "+logPath, "script "+filepath.Base(script))
	m.status = "running " + filepath.Base(script)
	m.refreshProject()
	return fmt.Sprintf("started %s (log %s, ctrl+x cancels)", script, logPath),
		tea.Batch(m.console().StartSpinner(), m.waitLine())
}

// scriptProgram maps a script file to a launcher available on THIS platform.
// Correction: the old code always chose cmd.exe / powershell.exe / "python",
// which fails on Linux/macOS with a confusing exec error. Windows-only
// extensions now fail with an actionable message instead.
func scriptProgram(script, ext string, extra []string) (string, []string, error) {
	windows := runtime.GOOS == "windows"
	switch ext {
	case ".bat", ".cmd":
		if !windows {
			return "", nil, fmt.Errorf("%s scripts need Windows (cmd.exe)", ext)
		}
		return "cmd.exe", append([]string{"/C", script}, extra...), nil
	case ".ps1":
		if windows {
			return "powershell.exe", append([]string{"-ExecutionPolicy", "Bypass", "-File", script}, extra...), nil
		}
		if p, err := exec.LookPath("pwsh"); err == nil {
			return p, append([]string{"-File", script}, extra...), nil
		}
		return "", nil, fmt.Errorf(".ps1 needs PowerShell (pwsh) on PATH")
	case ".sh":
		if windows {
			if p, err := exec.LookPath("bash"); err == nil {
				return p, append([]string{script}, extra...), nil
			}
			return "", nil, fmt.Errorf(".sh needs bash on PATH (e.g. Git Bash)")
		}
		return "sh", append([]string{script}, extra...), nil
	case ".py":
		prog := "python3"
		if windows {
			prog = "python"
		}
		if p, err := exec.LookPath(prog); err == nil {
			return p, append([]string{script}, extra...), nil
		}
		// Some setups only have "python" even on Unix.
		if p, err := exec.LookPath("python"); err == nil {
			return p, append([]string{script}, extra...), nil
		}
		return "", nil, fmt.Errorf(".py needs python on PATH")
	default:
		return script, extra, nil
	}
}
