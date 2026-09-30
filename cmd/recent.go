package cmd

import (
	"fmt"

	"github.com/ojilon/releaseforge/internal/history"
	"github.com/spf13/cobra"
)

var recentCmd = &cobra.Command{
	Use:   "recent",
	Short: "List recently opened/scanned projects",
	RunE: func(cmd *cobra.Command, args []string) error {
		root, _, err := requireDataRoot()
		if err != nil {
			return err
		}
		rf, err := history.LoadRecent(root)
		if err != nil {
			return err
		}
		if len(rf.Items) == 0 {
			fmt.Println("no recent projects — run `releaseforge scan <path>`")
			return nil
		}
		for i, e := range rf.Items {
			fmt.Printf("%d. %s  [%s]  %s\n", i+1, e.Name, e.Type, e.Path)
		}
		return nil
	},
}
