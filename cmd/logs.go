package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/ojilon/releaseforge/internal/project"
	"github.com/ojilon/releaseforge/internal/storage"
	"github.com/spf13/cobra"
)

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "List or show persisted build/test logs from the data root",
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
		entries, err := os.ReadDir(dir)
		if err != nil || len(entries) == 0 {
			fmt.Printf("no logs yet under %s\n", dir)
			return nil
		}
		type fi struct {
			name string
			mod  int64
		}
		var files []fi
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			st, err := e.Info()
			if err != nil {
				continue
			}
			files = append(files, fi{e.Name(), st.ModTime().Unix()})
		}
		sort.Slice(files, func(i, j int) bool { return files[i].mod > files[j].mod })
		fmt.Printf("logs under %s:\n", dir)
		for i, f := range files {
			if i >= 20 {
				break
			}
			fmt.Printf("  %s\n", filepath.Join(dir, f.name))
		}
		return nil
	},
}
