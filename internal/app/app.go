// Package app is the Bubble Tea root model: tabs, command bar, live panes.
package app

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/ojilon/releaseforge/internal/build"
	"github.com/ojilon/releaseforge/internal/config"
	"github.com/ojilon/releaseforge/internal/git"
	"github.com/ojilon/releaseforge/internal/history"
	rflog "github.com/ojilon/releaseforge/internal/log"
	"github.com/ojilon/releaseforge/internal/project"
	"github.com/ojilon/releaseforge/internal/storage"
	"github.com/ojilon/releaseforge/internal/tui"
	"github.com/ojilon/releaseforge/internal/version"
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

// ringCap bounds viewport memory; the log file stays complete.
const ringCap = 2000

// linesMsg carries a batch of streamed output lines into the model.
type linesMsg struct {
	lines []rflog.Line
}

// doneMsg marks completion of an async command.
type doneMsg struct {
	label  string
	err    error
	report []string
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
	ring     *rflog.Ring
	ready    bool

	history   []string
	histIndex int

	// TUI presentation state (doc 14).
	phase  string // idle|building|testing|failed|ok
	width  int
	branch string
	dirty  bool
	spark  string

	// Active async run (nil when idle).
	handle     *build.Handle
	pendingOK  string
	pendingErr string

	// pending async run
	running bool
}

// New creates the root model.
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

// Run starts the interactive TUI. The data root must be initialised
// (run `releaseforge init`); otherwise it returns a clear error.
func Run(dataRoot, projectDir string) error {
	if strings.TrimSpace(dataRoot) == "" {
		var err error
		dataRoot, err = storage.ResolveRoot("")
		if err != nil {
			return err
		}
	}
	if !storage.Exists(storage.GlobalConfigPath(dataRoot)) {
		return fmt.Errorf("data root not initialised (%s missing) — run `releaseforge init` first", storage.GlobalConfigPath(dataRoot))
	}
	p := tea.NewProgram(New(dataRoot, projectDir), tea.WithAltScreen(), tea.WithMouseCellMotion())
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
			m.viewport.MouseWheelEnabled = true
			m.viewport.SetContent(strings.Join(m.ring.Snapshot(), "\n"))
			m.viewport.GotoBottom()
			m.ready = true
		} else {
			m.viewport.Width = msg.Width
			m.viewport.Height = msg.Height - headerH - footerH
		}
		m.width = msg.Width
		m.input.Width = msg.Width - 4

	case linesMsg:
		for _, ln := range msg.lines {
			m.ring.Append(ln.Text)
		}
		m.render()
		if m.handle != nil {
			return m, m.waitLine()
		}
		return m, nil

	case doneMsg:
		m.running = false
		m.handle = nil
		if msg.err != nil {
			m.phase = "failed"
		} else {
			m.phase = "ok"
		}
		m.status = msg.label
		for _, l := range msg.report {
			m.appendLine(l)
		}
		if msg.err != nil {
			m.appendLine(tui.ErrorStyle.Render("✗ " + msg.label + ": " + msg.err.Error()))
		} else {
			m.appendLine("✓ " + msg.label)
		}
		m.appendLine(tui.Rule(m.ruleWidth()))
		m.refreshProject()

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyEsc:
			if m.handle != nil {
				m.handle.Cancel()
				m.appendLine("(cancelling…)")
			}
			return m, nil
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
			history.AppendCommand(m.dataRoot, line)
			m.input.SetValue("")
			m.appendLine("> " + line)
			quit, follow := m.exec(line)
			if quit {
				return m, tea.Quit
			}
			return m, follow
		case tea.KeyCtrlL:
			m.ring = rflog.NewRing(ringCap)
			m.viewport.SetContent("")
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
	if !m.ready {
		return tui.RenderHeader(version.ToolVersion, m.projectName(), m.version, m.status) + "\n" + m.welcome()
	}
	bar := tui.BarStyle.Render(m.input.View())
	if m.width > 0 && m.width < 80 {
		header := "ReleaseForge " + version.ToolVersion + " · " + m.projectName()
		return header + "\n" + m.viewport.View() + "\n" + bar + "\n" +
			tui.HelpStyle.Render("help|quit")
	}
	top := tui.PanelBorder.Render(
		tui.RenderHeader(version.ToolVersion, m.projectName(), m.version, m.status) +
			" " + tui.Pill(m.phase) + "\n" + m.contextLine())
	return top + "\n" + m.viewport.View() + "\n" + bar + "\n" + m.footer()
}

// contextLine renders branch, dirtiness, and the commit sparkline.
func (m *Model) contextLine() string {
	line := "no git"
	if m.branch != "" {
		line = m.branch
		if m.dirty {
			line += " · ● dirty"
		} else {
			line += " · clean"
		}
	}
	if m.spark != "" {
		line += " · " + m.spark
	}
	return tui.HelpStyle.Render(line)
}

// footer renders state-dependent key hints.
func (m *Model) footer() string {
	switch m.phase {
	case "building", "testing":
		return tui.HelpStyle.Render("esc cancels · logs tail")
	case "failed":
		return tui.HelpStyle.Render("logs last to inspect · help · quit")
	default:
		return tui.HelpStyle.Render("help · status · scan [path] · recent · version · build · test · notes · logs · quit")
	}
}

// ruleWidth bounds separator rendering to sane widths.
func (m *Model) ruleWidth() int {
	if m.width >= 80 {
		return m.width
	}
	return 80
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
	m.ring.Append(s)
	m.render()
}

// render paints the ring tail with follow-mode: it sticks to the bottom only
// when the user was already there, so reading scrollback survives new output.
func (m *Model) render() {
	if !m.ready {
		return
	}
	follow := m.viewport.AtBottom()
	m.viewport.SetContent(strings.Join(m.ring.Snapshot(), "\n"))
	if follow {
		m.viewport.GotoBottom()
	}
}

func (m *Model) welcome() string {
	return "ReleaseForge TUI — type `help` for commands, `scan <path>` to open a project.\n"
}

// parse splits a command-bar line into verb + args (pure, tested).
func parse(line string) (verb string, args []string) {
	fields := strings.Fields(line)
	if len(fields) == 0 {
		return "", nil
	}
	return strings.ToLower(fields[0]), fields[1:]
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
	case "notes":
		ver := ""
		if len(args) > 0 {
			ver = args[0]
		}
		m.appendLine(m.doNotes(ver))
	default:
		m.appendLine(tui.ErrorStyle.Render("unknown command: " + verb + " (try `help`)"))
	}
	return false, nil
}

// waitLine returns a Cmd delivering the next output batch, or the final
// doneMsg once the process ends and its output is drained. Batching (first
// line blocking, rest opportunistic up to 100) keeps fast output cheap.
func (m *Model) waitLine() tea.Cmd {
	h, okL, failL := m.handle, m.pendingOK, m.pendingErr
	return func() tea.Msg {
		first, ok := <-h.Lines
		if !ok {
			return finishRun(h, okL, failL)
		}
		batch := []rflog.Line{first}
		for len(batch) < 100 {
			select {
			case ln, ok := <-h.Lines:
				if !ok {
					return linesMsg{lines: batch}
				}
				batch = append(batch, ln)
			default:
				return linesMsg{lines: batch}
			}
		}
		return linesMsg{lines: batch}
	}
}

// finishRun drains the result after output is exhausted.
func finishRun(h *build.Handle, okL, failL string) tea.Msg {
	res := <-h.Done
	if res.Success {
		return doneMsg{label: okL}
	}
	var rep []string
	if res.Report != nil {
		rep = res.Report.Format()
	}
	return doneMsg{label: failL, err: firstError(res), report: rep}
}

// firstError summarizes a failed Result for doneMsg.
func firstError(res build.Result) error {
	if len(res.Errors) > 0 {
		return fmt.Errorf("%s", res.Errors[0])
	}
	return fmt.Errorf("exit %d", res.ExitCode)
}

func helpText() string {
	return `commands:
  open|scan [path]      detect project, write scan cache + config
  recent                list recently opened projects
  status                tool version, project, version, data root
  version [--set X|bump X|--code-only] show or set version
  build debug|release   gradle assemble / go build + persist log
  test [kind]           gradle tasks / go test ./...
  notes [version]       draft notes from git history
  logs                  list recent persisted logs
  clear                 clear viewport
  esc cancels a running build/test · ctrl+l clears too
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
	return fmt.Sprintf("releaseforge: %s\nproject: %s\nroot: %s\ntype: %s\nversion: %s\ndata-root: %s",
		version.ToolVersion, info.Name, info.Root, info.Type, ver, m.dataRoot)
}

func (m *Model) doScan(path string) string {
	snap, info, err := project.Scan(path)
	if err != nil {
		return tui.ErrorStyle.Render("scan: " + err.Error())
	}
	scanPath, err := project.WriteScan(m.dataRoot, snap)
	if err != nil {
		return tui.ErrorStyle.Render("scan: " + err.Error())
	}
	var existing *config.ProjectConfig
	if loaded, err := config.LoadProject(storage.ProjectConfigPath(m.dataRoot, info.Name)); err == nil {
		existing = &loaded
	}
	merged, notes := project.MergeConfig(project.SeedConfig(info),
		info.Root+"/.releaseforge.json", existing)
	cfgPath := storage.ProjectConfigPath(m.dataRoot, info.Name)
	savedAs := "unchanged"
	if existing == nil {
		savedAs = "written"
	}
	if existing == nil || !configsEqual(merged, *existing) {
		if err := config.SaveProject(cfgPath, merged); err != nil {
			return tui.ErrorStyle.Render("scan: " + err.Error())
		}
		if existing != nil {
			savedAs = "updated"
		}
	}
	if err := history.TouchRecent(m.dataRoot, info.Root, info.Name, info.Type, true); err != nil {
		return tui.ErrorStyle.Render("scan: " + err.Error())
	}
	m.projectDir = info.Root
	m.refreshProject()
	m.status = "scanned " + info.Name
	out := fmt.Sprintf("scanned %s (%s) → %s (config %s)", info.Name, info.Type, scanPath, savedAs)
	for _, n := range notes {
		out += "\nnote: " + n
	}
	return out
}

func (m *Model) doRecent() string {
	return m.recentText()
}

// configsEqual compares two project configs by canonical JSON.
func configsEqual(a, b config.ProjectConfig) bool {
	aj, err := json.Marshal(a)
	if err != nil {
		return false
	}
	bj, err := json.Marshal(b)
	if err != nil {
		return false
	}
	return string(aj) == string(bj)
}

func (m *Model) recentText() string {
	rf, err := history.LoadRecent(m.dataRoot)
	if err != nil {
		return tui.ErrorStyle.Render("recent: " + err.Error())
	}
	if len(rf.Items) == 0 {
		return "no recent projects — run `scan <path>`"
	}
	var b strings.Builder
	for i, e := range rf.Items {
		fmt.Fprintf(&b, "%d. %s  [%s]  %s\n", i+1, e.Name, e.Type, e.Path)
	}
	return strings.TrimRight(b.String(), "\n")
}

func (m *Model) doVersion(args []string) string {
	info, err := project.Detect(m.projectDir)
	if err != nil {
		return tui.ErrorStyle.Render("version: " + err.Error())
	}
	r, err := project.For(info)
	if err != nil {
		return tui.ErrorStyle.Render("version: " + err.Error())
	}
	header := "releaseforge: " + version.ToolVersion + "\n"
	setName, codeOnly := "", false
	switch {
	case len(args) >= 2 && (args[0] == "--set" || args[0] == "bump"):
		setName = strings.TrimSpace(args[1])
	case len(args) >= 1 && args[0] == "--code-only":
		codeOnly = true
	case len(args) > 0:
		return tui.ErrorStyle.Render("version: usage: version [--set X | bump X | --code-only]")
	}
	if codeOnly {
		next, err := r.BumpCode()
		if err != nil {
			return tui.ErrorStyle.Render("version: " + err.Error())
		}
		m.refreshProject()
		return header + fmt.Sprintf("versionCode: %s", next)
	}
	if setName != "" {
		if err := project.ValidateVersion(setName); err != nil {
			return tui.ErrorStyle.Render("version: " + err.Error())
		}
		next, err := r.VersionWrite(setName)
		if err != nil {
			return tui.ErrorStyle.Render("version: " + err.Error())
		}
		m.refreshProject()
		if next == "" {
			return header + fmt.Sprintf("version: %s (wrote file, not committed)", setName)
		}
		return header + fmt.Sprintf("versionCode: %s\nversionName: %s", next, setName)
	}
	code, name, err := r.VersionRead()
	if err != nil {
		return tui.ErrorStyle.Render("version: " + err.Error())
	}
	if code != "" {
		return header + fmt.Sprintf("versionCode: %s\nversionName: %s", code, name)
	}
	return header + fmt.Sprintf("version: %s", name)
}

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
	m.pendingOK, m.pendingErr = "build "+variant+" ok — log "+logPath, "build "+variant
	m.refreshProject()
	extra := ""
	if out != "" {
		extra = " → " + out
	}
	return fmt.Sprintf("started build %s%s (log %s, esc cancels)", variant, extra, logPath), m.waitLine()
}

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
	m.pendingOK, m.pendingErr = "test "+kind+" ok — log "+logPath, "test "+kind
	return fmt.Sprintf("started test %s (log %s, esc cancels)", kind, logPath), m.waitLine()
}

func (m *Model) doLogs(args []string) string {
	info, err := project.Detect(m.projectDir)
	if err != nil {
		return tui.ErrorStyle.Render("logs: " + err.Error())
	}
	dir := storage.LogsDir(m.dataRoot, info.Name)
	if len(args) == 0 {
		return m.logsList(dir)
	}
	switch args[0] {
	case "show":
		if len(args) < 2 {
			return tui.ErrorStyle.Render("logs: show needs a name")
		}
		return m.logsShow(dir, args[1], 0)
	case "last":
		return m.logsShow(dir, "last", 0)
	case "tail":
		n, target := 50, "last"
		rest := args[1:]
		if len(rest) >= 2 && (rest[0] == "-n" || rest[0] == "--lines") {
			fmt.Sscanf(rest[1], "%d", &n)
			rest = rest[2:]
		}
		if len(rest) > 0 {
			target = rest[0]
		}
		return m.logsShow(dir, target, n)
	default:
		return m.logsShow(dir, args[0], 0)
	}
}

func (m *Model) logsList(dir string) string {
	files := rflog.List(dir)
	if len(files) == 0 {
		return "logs: none yet (" + dir + ")"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "logs under %s:\n", dir)
	for i, f := range files {
		if i >= 10 {
			break
		}
		fmt.Fprintf(&b, "  %s\n", f)
	}
	return b.String()
}

// logsShow prints a log (or its tail). target is a name, prefix, or "last".
func (m *Model) logsShow(dir, target string, tailN int) string {
	name := target
	if target == "last" {
		n, err := rflog.Newest(dir)
		if err != nil {
			return "logs: none yet (" + dir + ")"
		}
		name = n
	} else {
		n, err := rflog.Resolve(dir, target)
		if err != nil {
			return tui.ErrorStyle.Render("logs: " + err.Error() + " (see `logs`)")
		}
		name = n
	}
	path := filepath.Join(dir, name)
	var b strings.Builder
	fmt.Fprintf(&b, "# %s\n", path)
	if tailN > 0 {
		lines, err := rflog.Tail(path, tailN)
		if err != nil {
			return tui.ErrorStyle.Render("logs: " + err.Error())
		}
		for _, l := range lines {
			b.WriteString(l + "\n")
		}
		return strings.TrimRight(b.String(), "\n")
	}
	const maxOut = 1 << 20
	data, truncated, err := rflog.Head(path, maxOut)
	if err != nil {
		return tui.ErrorStyle.Render("logs: " + err.Error())
	}
	b.Write(data)
	out := b.String()
	if truncated {
		if st, err := os.Stat(path); err == nil {
			out += fmt.Sprintf("\n…truncated (%d bytes total)", st.Size())
		}
	}
	return out
}

func (m *Model) doNotes(ver string) string {
	info, err := project.Detect(m.projectDir)
	if err != nil {
		return tui.ErrorStyle.Render("notes: " + err.Error())
	}
	var commits []git.Commit
	if git.IsRepo(info.Root) {
		prev := git.LatestTag(info.Root)
		if cl, err := git.LogSince(info.Root, prev); err == nil {
			commits = cl
		}
	}
	ver = strings.TrimSpace(ver)
	if ver == "" {
		return git.DraftNotes(info.Name, "unreleased", commits, true)
	}
	verDir := storage.ReleaseVersionDir(m.dataRoot, info.Name, ver)
	if err := os.MkdirAll(verDir, 0o755); err != nil {
		return tui.ErrorStyle.Render("notes: " + err.Error())
	}
	path := filepath.Join(verDir, "notes.md")
	if err := os.WriteFile(path, []byte(git.DraftNotes(info.Name, ver, commits, true)), 0o644); err != nil {
		return tui.ErrorStyle.Render("notes: " + err.Error())
	}
	return fmt.Sprintf("notes: %s (%d commits since last tag)", path, len(commits))
}
