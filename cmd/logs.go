package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	rflog "github.com/ojilon/releaseforge/internal/log"
	"github.com/ojilon/releaseforge/internal/project"
	"github.com/ojilon/releaseforge/internal/storage"
	"github.com/spf13/cobra"
)

var logsTailN int

var logsCmd = &cobra.Command{
	Use:   "logs [show <name>|last|tail [-n N] [name|last]]",
	Short: "List, show, or tail persisted build/test logs from the data root",
	Long: `No arguments lists recent logs. Subcommands read them:

  logs show <name>   print a log (exact name or unambiguous prefix)
  logs last          print the newest log
  logs tail [-n 50] [name|last]   print the last N lines (default 50, default target last)`,
	Args: cobra.MaximumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		info, err := project.Detect(projectDir)
		if err != nil {
			return err
		}
		root, _, err := resolveDataRoot()
		if err != nil {
			return err
		}
		dir := storage.LogsDir(root, info.Name)
		if len(args) == 0 {
			return listLogs(dir)
		}
		switch args[0] {
		case "show":
			if len(args) < 2 {
				return fmt.Errorf("logs show needs a log name (see `logs`)")
			}
			return showLog(dir, args[1], 0)
		case "last":
			return showLog(dir, "last", 0)
		case "tail":
			target := "last"
			if len(args) > 1 {
				target = args[1]
			}
			return showLog(dir, target, logsTailN)
		default:
			return fmt.Errorf("logs: unknown subcommand %q (want show|last|tail)", args[0])
		}
	},
}

// listLogs prints up to 20 log filenames, newest first.
func listLogs(dir string) error {
	files := rflog.List(dir)
	if len(files) == 0 {
		fmt.Printf("no logs yet under %s\n", dir)
		return nil
	}
	fmt.Printf("logs under %s:\n", dir)
	for i, f := range files {
		if i >= 20 {
			break
		}
		fmt.Printf("  %s\n", filepath.Join(dir, f))
	}
	return nil
}

// showLog prints a log (or its tail). target is a name, prefix, or "last".
func showLog(dir, target string, tailN int) error {
	name := target
	if target == "last" {
		n, err := rflog.Newest(dir)
		if err != nil {
			return fmt.Errorf("no logs yet under %s", dir)
		}
		name = n
	} else {
		n, err := rflog.Resolve(dir, target)
		if err != nil {
			return fmt.Errorf("logs: %v (see `logs`)", err)
		}
		name = n
	}
	path := filepath.Join(dir, name)
	fmt.Printf("# %s\n", path)
	if tailN > 0 {
		lines, err := rflog.Tail(path, tailN)
		if err != nil {
			return err
		}
		for _, l := range lines {
			fmt.Println(l)
		}
		return nil
	}
	const maxOut = 1 << 20
	data, truncated, err := rflog.Head(path, maxOut)
	if err != nil {
		return err
	}
	fmt.Printf("%s", data)
	if len(data) > 0 && data[len(data)-1] != '\n' {
		fmt.Println()
	}
	if truncated {
		st, _ := os.Stat(path)
		fmt.Printf("…truncated (%d bytes total)\n", st.Size())
	}
	return nil
}

func init() {
	logsCmd.Flags().IntVarP(&logsTailN, "n", "n", 50, "lines for logs tail")
}
