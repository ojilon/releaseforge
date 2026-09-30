package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/ojilon/releaseforge/internal/config"
	"github.com/ojilon/releaseforge/internal/storage"
	"github.com/spf13/cobra"
)

var (
	initForce          bool
	initNonInteractive bool
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "First-run wizard: choose data root (e.g. D:) and write global config",
	Long: `Interactive installation wizard (Huh form).

Creates the managed data-root layout:
  <data-root>/
    config/global.json
    projects/
    history/
    global-cache/

Prefer a non-system drive (D:) for builds, logs, releases and caches.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Explicit flag wins and also enables non-interactive use.
		chosen := strings.TrimSpace(dataRoot)
		if chosen == "" && !initNonInteractive {
			def := storage.DefaultRoot()
			// If a global config already exists, offer its data_root as default.
			if root, _, err := resolveDataRoot(); err == nil && root != "" {
				def = root
			}
			value := def
			// When stdout is not a TTY (CI/scripts), skip the form and use default.
			if fi, err := os.Stdout.Stat(); err == nil && (fi.Mode()&os.ModeCharDevice) == 0 {
				chosen = def
			} else {
				form := huh.NewForm(
					huh.NewGroup(
						huh.NewInput().
							Title("Data root for builds, logs, releases and cache").
							Description("Prefer a non-system drive, e.g. D:/ReleaseForgeData").
							Value(&value),
					),
				)
				if err := form.Run(); err != nil {
					return fmt.Errorf("init wizard: %w", err)
				}
				chosen = strings.TrimSpace(value)
			}
		}
		if chosen == "" {
			chosen = storage.DefaultRoot()
		}

		root, err := storage.ResolveRoot(chosen)
		if err != nil {
			return fmt.Errorf("resolve data root: %w", err)
		}

		if err := storage.EnsureLayout(root); err != nil {
			return fmt.Errorf("create data-root layout: %w", err)
		}

		cfgPath := storage.GlobalConfigPath(root)
		if storage.Exists(cfgPath) && !initForce {
			existing, err := config.LoadGlobal(cfgPath)
			if err != nil {
				return fmt.Errorf("data root already initialised but %s is unreadable: %w", cfgPath, err)
			}
			// Keep existing settings; just confirm the root matches.
			if existing.DataRoot != root {
				existing.DataRoot = root
				if err := config.SaveGlobal(cfgPath, existing); err != nil {
					return err
				}
			}
			fmt.Printf("data root already initialised: %s\nconfig: %s\n", root, cfgPath)
			return nil
		}

		// Preserve non-root settings when re-initialising with --force.
		cfg := config.DefaultGlobal(root)
		if storage.Exists(cfgPath) {
			if existing, err := config.LoadGlobal(cfgPath); err == nil {
				existing.DataRoot = root
				cfg = existing
			}
		}
		if err := config.SaveGlobal(cfgPath, cfg); err != nil {
			return fmt.Errorf("write global config: %w", err)
		}

		fmt.Printf("initialised data root: %s\nconfig: %s\n", root, cfgPath)
		fmt.Printf("projects: %s\nlogs/history/cache managed under the data root.\n",
			filepath.Join(root, "projects", "<name>"))
		return nil
	},
}

func init() {
	initCmd.Flags().BoolVar(&initForce, "force", false, "overwrite existing global.json (keeps data_root updated)")
	initCmd.Flags().BoolVar(&initNonInteractive, "non-interactive", false, "skip wizard; use --data-root, env or default")
}
