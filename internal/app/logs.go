package app

import (
	"fmt"
	"strings"

	"github.com/ojilon/releaseforge/internal/project"
	rflog "github.com/ojilon/releaseforge/internal/log"
	"github.com/ojilon/releaseforge/internal/storage"
	"github.com/ojilon/releaseforge/internal/tui"
	"os"
	"path/filepath"
)

// doLogs lists logs or shows one. Usage:
// logs | logs last | logs tail [-n 50] [name] | logs show <name> | logs <name>
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
