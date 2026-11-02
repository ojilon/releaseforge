package app

import (
	"fmt"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ojilon/releaseforge/internal/history"
	"github.com/ojilon/releaseforge/internal/tui"
)

// exec runs one command-bar line. Dispatch is generated from the verb table
// in internal/tui/intent.go: unknown verbs get a "did you mean" hint scored
// by the same fuzzy scorer used by the palette.
func (m *Model) exec(line string) (bool, tea.Cmd) {
	intent, err := tui.ParseIntent(line)
	if err != nil {
		m.appendLine(tui.ErrorStyle.Render("command: " + err.Error()))
		return false, nil
	}
	if intent.Verb == "" {
		return false, nil
	}
	verb, args := intent.Verb, intent.Args
	m.appendLine(tui.ChipStyle.Render("❯ " + strings.TrimSpace(line)))

	switch verb {
	case "quit", "exit", "q":
		return true, nil

	case "help", "-h", "?":
		m.switchTo(WinHelp)
		m.appendLine(tui.HelpText())
		return false, nil

	case "clear":
		m.console().Clear()
		m.hint = ""
		return false, nil

	case "status":
		m.switchTo(WinStatus)
		m.appendLine(m.statusText())
		return false, nil

	case "scan", "open", "rescan":
		target, pick := m.scanTarget(verb, args)
		if pick {
			return false, m.openPicker()
		}
		if strings.TrimSpace(target) == "" {
			m.appendLine(tui.ErrorStyle.Render("open: no recent project with that number — run `recent`"))
			return false, nil
		}
		m.appendLine(m.doScan(target))
		m.switchTo(WinStatus)
		return false, nil

	case "recent":
		m.appendLine(m.doRecent())
		m.switchTo(WinStatus)
		return false, nil

	case "version":
		m.appendLine(m.doVersion(args))
		return false, nil

	case "build":
		variant, setVer := "debug", ""
		if len(args) > 0 {
			variant = args[0]
		}
		if len(args) > 1 {
			setVer = args[1]
		}
		s, c := m.doBuild(variant, setVer)
		m.appendLine(s)
		m.switchTo(WinBuild)
		return false, c

	case "test":
		kind := "unit"
		if len(args) > 0 {
			kind = args[0]
		}
		s, c := m.doTest(kind)
		m.appendLine(s)
		m.switchTo(WinBuild)
		return false, c

	case "bat", "run":
		s, c := m.doBat(args)
		m.appendLine(s)
		m.switchTo(WinBuild)
		return false, c

	case "sign":
		m.appendLine(m.doSign(args))
	case "install":
		m.appendLine(m.doInstall(args))
	case "devices":
		m.appendLine(m.doDevices())
	case "git":
		m.appendLine(m.doGit(args))
		m.switchTo(WinGit)
	case "ls", "dir":
		m.appendLine(m.doLs(args))
	case "pwd":
		m.appendLine(m.doPwd())
	case "cat":
		m.appendLine(m.doCat(args))
	case "notes":
		ver := ""
		if len(args) > 0 {
			ver = args[0]
		}
		m.appendLine(m.doNotes(ver))
	case "logs":
		m.appendLine(m.doLogs(args))
		m.switchTo(WinLogs)
		m.refreshActive()
	case "theme":
		if len(args) == 0 {
			m.appendLine(fmt.Sprintf("theme: %s (available: %s)", m.themeName,
				strings.Join(tui.ThemeNames(), ", ")))
		} else {
			m.appendLine(m.setTheme(strings.ToLower(args[0])))
		}
	case "window":
		m.appendLine(m.doWindow(args))
	default:
		m.appendLine(tui.ErrorStyle.Render("unknown command: " + verb + m.didYouMean(verb)))
	}
	return false, nil
}

// scanTarget resolves an open/scan/rescan argument. A bare number or #n opens
// that recent project (correction: `recent` used to be a dead-end list).
func (m *Model) scanTarget(verb string, args []string) (path string, pick bool) {
	if len(args) > 0 {
		a := strings.TrimSpace(args[0])
		idx := -1
		if strings.HasPrefix(a, "#") {
			if n, err := strconv.Atoi(a[1:]); err == nil {
				idx = n
			}
		} else if n, err := strconv.Atoi(a); err == nil {
			idx = n
		}
		if idx > 0 {
			if rf, err := history.LoadRecent(m.dataRoot); err == nil && idx <= len(rf.Items) {
				return rf.Items[idx-1].Path, false
			}
			return "", false // error is reported by doScan
		}
		return a, false
	}
	if verb == "rescan" {
		return m.projectDir, false
	}
	return "", true // typed path omitted → open the folder picker
}

// doWindow switches the active window by number or name.
func (m *Model) doWindow(args []string) string {
	if len(args) == 0 {
		return "usage: window 1..6|" + strings.Join(lowerTitles(), "|")
	}
	a := strings.ToLower(strings.TrimSpace(args[0]))
	if n, err := strconv.Atoi(a); err == nil && n >= 1 && n <= len(WindowTitles) {
		m.switchTo(WindowID(n - 1))
		return "window: " + WindowTitles[n-1]
	}
	for i, label := range WindowTitles {
		if strings.ToLower(label) == a {
			m.switchTo(WindowID(i))
			return "window: " + label
		}
	}
	return tui.ErrorStyle.Render("window: unknown " + a)
}

func lowerTitles() []string {
	out := make([]string, len(WindowTitles))
	for i, t := range WindowTitles {
		out[i] = strings.ToLower(t)
	}
	return out
}

// didYouMean suggests the closest verb using the shared fuzzy scorer.
func (m *Model) didYouMean(verb string) string {
	names := make([]string, len(tui.Verbs))
	for i, v := range tui.Verbs {
		names[i] = v.Name
	}
	matches := tui.Rank(verb, names, 1)
	if len(matches) == 0 {
		return " (type `help`)"
	}
	return fmt.Sprintf(" (did you mean `%s`? · type `help`)", names[matches[0].Index])
}

// helpText is kept as a thin wrapper so CLI/tests can ask the app for the
// same generated reference the Help window shows.
func helpText() string { return tui.HelpText() }
