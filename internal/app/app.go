// Package app is the Bubble Tea root model: tabs, command bar, live panes.
package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/ojilon/releaseforge/internal/build"
	"github.com/ojilon/releaseforge/internal/config"
	"github.com/ojilon/releaseforge/internal/project"
	"github.com/ojilon/releaseforge/internal/storage"
	rflog "github.com/ojilon/releaseforge/internal/log"
	"github.com/ojilon/releaseforge/internal/tui"
)

// Root model owns:
//   - current project context
//   - persistent command input + history
//   - live log viewport
//   - status header
//
// See docs/04-tui-design.md.

// NotImplemented is the shared stub error for unbuilt surfaces.
func NotImplemented(name string) error {
	return fmt.Errorf("%s: not implemented yet — see docs/08-implementation-plan.md", name)
}

// logMsg carries one streamed output line into the model.
type logMsg struct {
	text  string
	isErr bool
}

// doneMsg marks completion of an async command.
type doneMsg struct {
	label string
	err   error
}

type asyncResult struct {
	lines []string
	err   error
	label string
}

// Model is the TUI root.
type Model struct {
	dataRoot   string
	projectDir string
	info       *project.Info
	version    string
	status     string

	input    textinput.Model
	viewport viewport.Model
	ready    bool

	history   []string
	histIndex int

	// pending async run
	running bool
}

// New creates the root model.
func New(dataRoot, projectDir string) Model {
	ti := textinput.New()
	ti.Placeholder = "type a command (help, status, scan, version, build, test, logs, quit)"
	ti.Focus()
	ti.CharLimit = 512
	ti.Prompt = "> "

	m := Model{dataRoot: dataRoot, projectDir: projectDir, input: ti, status: "ready"}
	m.refreshProject()
	return m
}

// Run starts the interactive TUI.
func Run(dataRoot, projectDir string) error {
	if strings.TrimSpace(dataRoot) == "" {
		var err error
		dataRoot, err = storage.ResolveRoot("")
		if err != nil {
			return err
		}
	}
	p := tea.NewProgram(New(dataRoot, projectDir), tea.WithAltScreen())
	_, err := p.Run()
	return err
}

func (m *Model) refreshProject() {
	info, err := project.Detect(m.projectDir)
	if err != nil {
		m.info = nil
		m.version = ""
		return
	}
	m.info = &info
	if code, name, err := project.CurrentVersion(info); err == nil && name != "" {
		if code != "" {
			m.version = fmt.Sprintf("v%s (%s)", name, code)
		} else {
			m.version = "v" + name
		}
	} else {
		m.version = info.Type
	}
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		headerH := 3
		footerH := 3
		if !m.ready {
			m.viewport = viewport.New(msg.Width, msg.Height-headerH-footerH)
			m.viewport.SetContent(m.welcome())
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - headerH - footerH
		}
		m.input.Width = msg.Width - 4

	case logMsg:
		m.viewport.SetContent(m.viewport.View() + "\n" + msg.text)
		m.viewport.GotoBottom()

	case doneMsg:
		m.running = false
		m.status = msg.label
		if msg.err != nil {
			m.appendLine(tui.ErrorStyle.Render("✗ " + msg.label + ": " + msg.err.Error()))
		} else {
			m.appendLine("✓ " + msg.label)
		}
		m.appendLine("")
		m.refreshProject()

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyUp:
			if len(m.history) > 0 {
				if m.histIndex > 0 {
					m.histIndex--
				}
				m.input.SetValue(m.history[m.histIndex])
			}
			return m, nil
		case tea.KeyDown:
			if len(m.history) > 0 {
				if m.histIndex < len(m.history)-1 {
					m.histIndex++
					m.input.SetValue(m.history[m.histIndex])
				} else {
					m.histIndex = len(m.history)
					m.input.SetValue("")
				}
			}
			return m, nil
		case tea.KeyEnter:
			line := strings.TrimSpace(m.input.Value())
			if line == "" {
				return m, nil
			}
			m.history = append(m.history, line)
			m.histIndex = len(m.history)
			m.input.SetValue("")
			m.appendLine("> " + line)
			if quit := m.exec(line); quit {
				return m, tea.Quit
			}
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	cmds = append(cmds, cmd)
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	header := tui.RenderHeader(m.projectName(), m.version, m.dataRoot, m.status)
	body := ""
	if m.ready {
		body = m.viewport.View()
	} else {
		body = m.welcome()
	}
	bar := tui.BarStyle.Render(m.input.View())
	help := tui.HelpStyle.Render("help · status · scan [path] · version · build debug|release · test unit · logs · quit")
	return fmt.Sprintf("%s\n%s\n%s\n%s", header, body, bar, help)
}

func (m *Model) projectName() string {
	if m.info != nil {
		return m.info.Name
	}
	return filepath.Base(m.projectDir)
}

func (m *Model) appendLine(s string) {
	if !m.ready {
		return
	}
	m.viewport.SetContent(m.viewport.View() + "\n" + s)
	m.viewport.GotoBottom()
}

func (m *Model) welcome() string {
	return "ReleaseForge TUI — type `help` for commands, `scan <path>` to open a project.\n"
}

// exec runs a command bar line. Returns true when the TUI should quit.
func (m *Model) exec(line string) bool {
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return false
	}
	verb := strings.ToLower(fields[0])
	args := fields[1:]

	switch verb {
	case "quit", "exit", "q":
		return true
	case "help", "-h", "?":
		m.appendLine(helpText())
	case "clear":
		m.viewport.SetContent("")
	case "status":
		m.appendLine(m.statusText())
	case "scan", "open", "rescan":
		path := m.projectDir
		if len(args) > 0 {
			path = args[0]
		}
		m.appendLine(m.doScan(path))
	case "version":
		m.appendLine(m.doVersion(args))
	case "build":
		variant := "debug"
		if len(args) > 0 {
			variant = args[0]
		}
		m.appendLine(m.doBuild(variant))
	case "test":
		kind := "unit"
		if len(args) > 0 {
			kind = args[0]
		}
		m.appendLine(m.doTest(kind))
	case "logs":
		m.appendLine(m.doLogs())
	default:
		m.appendLine(tui.ErrorStyle.Render("unknown command: " + verb + " (try `help`)"))
	}
	return false
}

func helpText() string {
	return `commands:
  open|scan [path]      detect project, write config under data root
  status                project, version, data root
  version [--set X]     show or set version (android-gradle)
  build debug|release   run gradle assemble + persist log
  test unit|instrumented|all
  logs                  list recent persisted logs
  clear                 clear viewport
  quit                  exit`
}

func (m *Model) statusText() string {
	info, err := project.Detect(m.projectDir)
	if err != nil {
		return "status: " + err.Error()
	}
	ver := info.Type
	if code, name, err := project.CurrentVersion(info); err == nil && name != "" {
		ver = fmt.Sprintf("%s (code %s)", name, code)
	}
	return fmt.Sprintf("project: %s\nroot: %s\ntype: %s\nversion: %s\ndata-root: %s",
		info.Name, info.Root, info.Type, ver, m.dataRoot)
}

func (m *Model) doScan(path string) string {
	info, err := project.Detect(path)
	if err != nil {
		return tui.ErrorStyle.Render("scan: " + err.Error())
	}
	if err := storage.EnsureProjectLayout(m.dataRoot, info.Name); err != nil {
		return tui.ErrorStyle.Render("scan: " + err.Error())
	}
	pcfg := config.ProjectConfig{Type: info.Type, Name: info.Name, Root: info.Root}
	if info.VersionFile != "" {
		pcfg.Version = config.VersionConfig{File: info.VersionFile, CodeKey: info.CodeKey, NameKey: info.NameKey}
	}
	cfgPath := storage.ProjectConfigPath(m.dataRoot, info.Name)
	if !storage.Exists(cfgPath) {
		if err := config.SaveProject(cfgPath, pcfg); err != nil {
			return tui.ErrorStyle.Render("scan: " + err.Error())
		}
	}
	m.projectDir = info.Root
	m.refreshProject()
	m.status = "scanned " + info.Name
	return fmt.Sprintf("scanned %s (%s) → %s", info.Name, info.Type, cfgPath)
}

func (m *Model) doVersion(args []string) string {
	info, err := project.Detect(m.projectDir)
	if err != nil {
		return tui.ErrorStyle.Render("version: " + err.Error())
	}
	if info.Type != "android-gradle" {
		return fmt.Sprintf("version: type %q has no managed version file", info.Type)
	}
	if len(args) >= 2 && args[0] == "--set" {
		next, err := project.SetGradleVersion(info.Root, info.VersionFile, args[1])
		if err != nil {
			return tui.ErrorStyle.Render("version: " + err.Error())
		}
		m.refreshProject()
		return fmt.Sprintf("versionCode: %s\nversionName: %s", next, args[1])
	}
	code, name, err := project.GradleVersion(info.Root, info.VersionFile)
	if err != nil {
		return tui.ErrorStyle.Render("version: " + err.Error())
	}
	return fmt.Sprintf("versionCode: %s\nversionName: %s", code, name)
}

func (m *Model) doBuild(variant string) string {
	variant = strings.ToLower(strings.TrimSpace(variant))
	if variant != "debug" && variant != "release" {
		return tui.ErrorStyle.Render("build: want debug|release")
	}
	info, err := project.Detect(m.projectDir)
	if err != nil {
		return tui.ErrorStyle.Render("build: " + err.Error())
	}
	if info.Type != "android-gradle" {
		return tui.ErrorStyle.Render("build: unsupported type " + info.Type)
	}
	task := "assembleDebug"
	if variant == "release" {
		task = "assembleRelease"
	}
	_ = storage.EnsureProjectLayout(m.dataRoot, info.Name)
	logPath := rflog.LogPath(storage.LogsDir(m.dataRoot, info.Name), "build-"+variant)
	wrapper := build.GradleWrapper(info.Root)
	var lines []string
	res := build.Run(wrapper, []string{task}, build.Options{
		Dir: info.Root, LogPath: logPath,
		OnLine: func(t string, _ bool) { lines = append(lines, t) },
	})
	// replay captured lines into viewport (bounded)
	start := 0
	if len(lines) > 200 {
		start = len(lines) - 200
		lines = append(lines[:0:0], lines[start:]...)
		m.appendLine(fmt.Sprintf("(... %d earlier lines in %s)", start, logPath))
	}
	for _, l := range lines {
		m.appendLine(l)
	}
	if !res.Success {
		m.status = "build failed"
		return tui.ErrorStyle.Render(fmt.Sprintf("build %s FAILED — log %s", variant, res.LogPath))
	}
	m.status = "build " + variant + " ok"
	m.refreshProject()
	return fmt.Sprintf("build %s ok — log %s", variant, res.LogPath)
}

func (m *Model) doTest(kind string) string {
	info, err := project.Detect(m.projectDir)
	if err != nil {
		return tui.ErrorStyle.Render("test: " + err.Error())
	}
	if info.Type != "android-gradle" {
		return tui.ErrorStyle.Render("test: unsupported type " + info.Type)
	}
	var tasks []string
	switch kind {
	case "unit":
		tasks = []string{":app:testDebugUnitTest"}
	case "instrumented":
		tasks = []string{":app:connectedDebugAndroidTest"}
	case "all":
		tasks = []string{":app:testDebugUnitTest", ":app:connectedDebugAndroidTest"}
	default:
		return tui.ErrorStyle.Render("test: want unit|instrumented|all")
	}
	_ = storage.EnsureProjectLayout(m.dataRoot, info.Name)
	logPath := rflog.LogPath(storage.LogsDir(m.dataRoot, info.Name), "test-"+kind)
	wrapper := build.GradleWrapper(info.Root)
	var lines []string
	res := build.Run(wrapper, tasks, build.Options{
		Dir: info.Root, LogPath: logPath,
		OnLine: func(t string, _ bool) { lines = append(lines, t) },
	})
	start := 0
	if len(lines) > 200 {
		start = len(lines) - 200
		lines = append(lines[:0:0], lines[start:]...)
		m.appendLine(fmt.Sprintf("(... %d earlier lines in %s)", start, logPath))
	}
	for _, l := range lines {
		m.appendLine(l)
	}
	if !res.Success {
		m.status = "test failed"
		return tui.ErrorStyle.Render(fmt.Sprintf("test %s FAILED — log %s", kind, res.LogPath))
	}
	m.status = "test " + kind + " ok"
	return fmt.Sprintf("test %s ok — log %s", kind, res.LogPath)
}

func (m *Model) doLogs() string {
	info, err := project.Detect(m.projectDir)
	if err != nil {
		return tui.ErrorStyle.Render("logs: " + err.Error())
	}
	dir := storage.LogsDir(m.dataRoot, info.Name)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "logs: none yet (" + dir + ")"
	}
	if len(entries) == 0 {
		return "logs: none yet (" + dir + ")"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "logs under %s:\n", dir)
	n := 0
	for i := len(entries) - 1; i >= 0 && n < 10; i-- {
		fmt.Fprintf(&b, "  %s\n", entries[i].Name())
		n++
	}
	return b.String()
}
