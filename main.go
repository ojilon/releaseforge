// ReleaseForge — local release, build, test and project-lifecycle tool.
// Entry point delegates to the Cobra command tree (cmd package).
package main

import (
	"os"

	"github.com/ojilon/releaseforge/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
