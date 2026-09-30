package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "List or show persisted build/test logs from the data root",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("logs: list recent log files under data-root")
		return notImplemented("logs")
	},
}
