// Installer for ReleaseForge (v0.0.1 scope).
//
// A tiny interactive CLI built with the same Charm stack as the main app.
// It does not embed the payload: it installs a releaseforge.exe found next
// to itself (or via --bin) into <chosen-folder>/ReleaseForge/{bin,data},
// creating the storage tree and seed files. Re-runs update binaries and
// top up missing structure without touching existing projects/history.
//
// Build order:
//
//	go build -o releaseforge.exe .
//	go build -o installer.exe ./installer
//	# put both in one folder (dist/) and run installer.exe from anywhere.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/ojilon/releaseforge/internal/config"
	"github.com/ojilon/releaseforge/internal/storage"
	"github.com/ojilon/releaseforge/internal/version"
)

// folderCap bounds drive-root listings (roots can hold hundreds of entries).
const folderCap = 50

// InstallRecord tracks the last install (used later by the 0.0.3 updater).
type InstallRecord struct {
	Version     string `json:"version"`
	InstalledAt string `json:"installed_at"`
	SourceBin   string `json:"source_bin"`
	AppRoot     string `json:"app_root"`
}

func main() {
	binFlag := flag.String("bin", "", "payload releaseforge.exe (default: next to installer, else ./releaseforge.exe)")
	destFlag := flag.String("dest", "", "parent folder to install into (skips wizard folder pick)")
	driveFlag := flag.String("drive", "", "drive/root to use (skips wizard drive pick, e.g. D:\\)")
	appName := flag.String("app-name", "ReleaseForge", "app folder name created under the chosen folder")
	verFlag := flag.String("version", version.ToolVersion, "version recorded in install.json")
	nonInteractive := flag.Bool("non-interactive", false, "no prompts; requires --dest")
	flag.Parse()

	if err := run(*binFlag, *destFlag, *driveFlag, *appName, *verFlag, *nonInteractive); err != nil {
		fmt.Fprintln(os.Stderr, "installer:", err)
		os.Exit(1)
	}
}

func run(binFlag, destFlag, driveFlag, appName, ver string, nonInteractive bool) error {
	payload, err := findPayload(binFlag)
	if err != nil {
		return err
	}

	parent := strings.TrimSpace(destFlag)
	if parent == "" && nonInteractive {
		return fmt.Errorf("--dest is required with --non-interactive")
	}
	if parent == "" {
		drive, err := pickDrive(driveFlag)
		if err != nil {
			return err
		}
		parent, err = pickFolder(drive)
		if err != nil {
			return err
		}
	}
	appRoot := filepath.Join(parent, appName)
	if err := EnsureTree(appRoot, payload, ver); err != nil {
		return err
	}
	fmt.Printf("\ninstalled releaseforge %s\n  app: %s\n  bin: %s\n  data: %s\n",
		ver, appRoot, filepath.Join(appRoot, "bin"), filepath.Join(appRoot, "data"))
	fmt.Printf("\nnext:\n  1. add to PATH for this session:\n     $env:Path += \";%s\"\n  2. run: releaseforge scan <project-path>\n",
		filepath.Join(appRoot, "bin"))
	return nil
}

// findPayload resolves the releaseforge binary: --bin, then next to the
// installer executable, then ./releaseforge.exe.
func findPayload(flagVal string) (string, error) {
	cands := []string{strings.TrimSpace(flagVal)}
	if exe, err := os.Executable(); err == nil {
		cands = append(cands, filepath.Join(filepath.Dir(exe), exeName("releaseforge")))
	}
	cands = append(cands, exeName("releaseforge"))
	for _, c := range cands {
		if c == "" {
			continue
		}
		if st, err := os.Stat(c); err == nil && !st.IsDir() {
			abs, _ := filepath.Abs(c)
			return abs, nil
		}
	}
	return "", fmt.Errorf("releaseforge payload not found (put %s next to installer.exe or pass --bin)", exeName("releaseforge"))
}

func exeName(base string) string {
	if runtime.GOOS == "windows" {
		return base + ".exe"
	}
	return base
}

// listDrives returns existing roots. RF_INSTALL_DRIVES (path-list separated)
// overrides enumeration for tests.
func listDrives() []string {
	if override := strings.TrimSpace(os.Getenv("RF_INSTALL_DRIVES")); override != "" {
		var out []string
		for _, p := range filepath.SplitList(override) {
			if st, err := os.Stat(p); err == nil && st.IsDir() {
				out = append(out, p)
			}
		}
		return out
	}
	if runtime.GOOS == "windows" {
		var out []string
		for c := 'A'; c <= 'Z'; c++ {
			p := string([]rune{c}) + `:\`
			if st, err := os.Stat(p); err == nil && st.IsDir() {
				out = append(out, p)
			}
		}
		return out
	}
	return []string{"/"}
}

// listFolders returns sorted subdirectory names of dir, capped.
func listFolders(dir string, cap int) (names []string, truncated bool, err error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, false, err
	}
	for _, e := range entries {
		if e.IsDir() {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	if len(names) > cap {
		return names[:cap], true, nil
	}
	return names, false, nil
}

func pickDrive(prefill string) (string, error) {
	drives := listDrives()
	if len(drives) == 0 {
		return "", fmt.Errorf("no drives found")
	}
	if d := strings.TrimSpace(prefill); d != "" {
		if st, err := os.Stat(d); err == nil && st.IsDir() {
			return d, nil
		}
		return "", fmt.Errorf("drive %q not accessible", d)
	}
	if len(drives) == 1 {
		return drives[0], nil
	}
	var choice string
	form := huh.NewForm(huh.NewGroup(
		huh.NewSelect[string]().
			Title("Where should ReleaseForge live? Pick a drive").
			Options(huh.NewOptions(drives...)...).
			Value(&choice),
	))
	if err := form.Run(); err != nil {
		return "", err
	}
	return choice, nil
}

func pickFolder(drive string) (string, error) {
	names, truncated, err := listFolders(drive, folderCap)
	if err != nil {
		return "", err
	}
	const createNew = "[Create a new folder…]"
	const useRoot = "[Use the drive root]"
	opts := append([]string{useRoot}, names...)
	opts = append(opts, createNew)
	hint := ""
	if truncated {
		hint = fmt.Sprintf(" (showing first %d; type a name for others)", folderCap)
	}
	var choice string
	form := huh.NewForm(huh.NewGroup(
		huh.NewSelect[string]().
			Title("Pick a folder on "+drive+hint).
			Options(huh.NewOptions(opts...)...).
			Value(&choice),
	))
	if err := form.Run(); err != nil {
		return "", err
	}
	switch choice {
	case useRoot:
		return drive, nil
	case createNew:
		var name string
		f2 := huh.NewForm(huh.NewGroup(
			huh.NewInput().Title("New folder name").Value(&name),
		))
		if err := f2.Run(); err != nil {
			return "", err
		}
		name = strings.TrimSpace(name)
		if name == "" || name == "." || strings.ContainsAny(name, `/\:*?"<>|`) {
			return "", fmt.Errorf("invalid folder name %q", name)
		}
		return filepath.Join(drive, name), nil
	default:
		return filepath.Join(drive, choice), nil
	}
}

// EnsureTree creates/updates the install tree. Idempotent: binaries are
// overwritten, missing dirs/files are topped up, existing data is preserved.
func EnsureTree(appRoot, payloadBin, ver string) error {
	binDir := filepath.Join(appRoot, "bin")
	dataRoot := filepath.Join(appRoot, "data")
	for _, d := range []string{binDir, dataRoot} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return err
		}
	}
	if err := storage.EnsureLayout(dataRoot); err != nil {
		return fmt.Errorf("storage layout: %w", err)
	}
	// Payload binary (overwrite on re-install).
	dst := filepath.Join(binDir, exeName("releaseforge"))
	if err := copyFile(payloadBin, dst); err != nil {
		return fmt.Errorf("copy payload: %w", err)
	 }
	// Seed global.json only when missing; otherwise keep user settings but
	// make sure data_root points at this install's data dir.
	gpath := storage.GlobalConfigPath(dataRoot)
	if !storage.Exists(gpath) {
		if err := config.SaveGlobal(gpath, config.DefaultGlobal(dataRoot)); err != nil {
			return fmt.Errorf("seed global.json: %w", err)
		}
	} else if cur, err := config.LoadGlobal(gpath); err == nil && cur.DataRoot != dataRoot {
		cur.DataRoot = dataRoot
		if err := config.SaveGlobal(gpath, cur); err != nil {
			return fmt.Errorf("fix data_root: %w", err)
		}
	}
	rec := InstallRecord{
		Version:     ver,
		InstalledAt: time.Now().UTC().Format(time.RFC3339),
		SourceBin:   payloadBin,
		AppRoot:     appRoot,
	}
	data, _ := json.MarshalIndent(rec, "", "  ")
	if err := os.WriteFile(filepath.Join(dataRoot, "install.json"), append(data, '\n'), 0o644); err != nil {
		return fmt.Errorf("write install.json: %w", err)
	}
	return nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}
