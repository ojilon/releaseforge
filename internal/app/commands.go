package app

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ojilon/releaseforge/internal/history"
	rflog "github.com/ojilon/releaseforge/internal/log"
	"github.com/ojilon/releaseforge/internal/metrics"
	"github.com/ojilon/releaseforge/internal/project"
	"github.com/ojilon/releaseforge/internal/tui"
	"github.com/ojilon/releaseforge/internal/version"
)

// parse splits a command-bar line into verb + args (pure, tested).
func parse(line string) (verb string, args []string) {
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return "", nil
	}
	return strings.ToLower(fields[0]), fields[1:]
}

// onEnter runs the line currently typed in the input box.
// Returns (quit, follow-up command for async jobs).
func (m *Model) onEnter() (bool, tea.Cmd) {
	line := strings.TrimSpace(m.input.Value())
	if line == "" {
		return false, nil
	}
	m.history = append(m.history, line)
	m.histIndex = len(m.history)
	history.AppendCommand(m.dataRoot, line)
	m.input.SetValue("")
	m.appendLine("> " + line)
	quit, follow := m.exec(line)
	if quit {
		return true, nil
	}
	return false, follow
}

// exec runs a command bar line. It returns (quit, follow-up command).
func (m *Model) exec(line string) (bool, tea.Cmd) {
	verb, args := parse(line)
	if verb == "" {
		return false, nil
	}
	switch verb {
	case "quit", "exit", "q":
		return true, nil
	case "help", "-h", "?":
		m.appendLine(helpText())
	case "clear":
		m.ring = rflog.NewRing(ringCap)
		m.viewport.SetContent("")
	case "status":
		m.appendLine(m.statusText())
	case "scan", "open", "rescan":
		path := m.projectDir
		if len(args) > 0 {
			path = args[0]
		}
		m.appendLine(m.doScan(path))
	case "recent":
		m.appendLine(m.doRecent())
	case "version":
		m.appendLine(m.doVersion(args))
	case "build":
		variant := "debug"
		if len(args) > 0 {
			variant = args[0]
		}
		s, c := m.doBuild(variant)
		m.appendLine(s)
		return false, c
	case "test":
		kind := "unit"
		if len(args) > 0 {
			kind = args[0]
		}
		s, c := m.doTest(kind)
		m.appendLine(s)
		return false, c
	case "logs":
		m.appendLine(m.doLogs(args))
	case "sign":
		m.appendLine(m.doSign(args))
	case "install":
		m.appendLine(m.doInstall(args))
	case "devices":
		m.appendLine(m.doDevices())
	case "git":
		m.appendLine(m.doGit(args))
	case "ls", "dir":
		m.appendLine(m.doLs(args))
	case "pwd":
		m.appendLine(m.doPwd())
	case "cat":
		m.appendLine(m.doCat(args))
	case "bat", "run":
		s, c := m.doBat(args)
		m.appendLine(s)
		return false, c
	case "notes":
		ver := ""
		if len(args) > 0 {
			ver = args[0]
		}
		m.appendLine(m.doNotes(ver))
	case "theme":
		tui.SetTheme(args[0])
	default:
		m.appendLine(tui.ErrorStyle.Render("unknown command: " + verb + " (try `help`)"))
	}
	return false, nil
}

func helpText() string {
	return `commands:
  open|scan [path]      detect project, write scan cache + config
  recent                list recently opened projects
  status                tool version, project, version, data root
  version [--set X|bump X|--code-only] show or set version
  build debug|release   gradle assemble / go build + persist log
  test [kind]           gradle tasks / go test ./...
  sign [apk] [--alias X]  sign APK (keystore in data-root/signing)
  install [debug|release] install APK via adb, no copying
  devices               list connected adb devices
  git log|status|diff|tags  git state while analysing
  ls|dir [path] · pwd · cat <file>  browse folders
  bat <script> [args]   run .bat/.ps1/.sh/.py with live logs
  notes [version]       draft notes from git history
  logs                  list recent persisted logs
  clear                 clear viewport
  esc cancels a running build/test/script · ctrl+l clears too
  quit                  exit`
}

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
	return fmt.Sprintf("releaseforge: %s\nproject: %s\nroot: %s\ntype: %s\nversion: %s\ndata-root: %s\n%s",
		version.ToolVersion, info.Name, info.Root, info.Type, ver, m.dataRoot,
		metrics.LastBuildLine(m.dataRoot, info.Name))
}
