package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ojilon/releaseforge/internal/android"
	"github.com/ojilon/releaseforge/internal/project"
	"github.com/ojilon/releaseforge/internal/storage"
	"github.com/ojilon/releaseforge/internal/tui"
)

// doInstall installs the last built APK via adb, no copying needed.
// Usage: install [debug|release] [--device SERIAL]
func (m *Model) doInstall(args []string) string {
	variant := "debug"
	serial := ""
	for i := 0; i < len(args); i++ {
		switch {
		case args[i] == "--device" && i+1 < len(args):
			serial = args[i+1]
			i++
		case args[i] == "debug" || args[i] == "release":
			variant = args[i]
		}
	}
	info, err := project.Detect(m.projectDir)
	if err != nil {
		return tui.ErrorStyle.Render("install: " + err.Error())
	}
	apk := pickInstallApk(m.dataRoot, info.Root, info.Name, variant)
	if apk == "" {
		return tui.ErrorStyle.Render(fmt.Sprintf("install: no %s APK found (build %s first)", variant, variant))
	}
	if strings.TrimSpace(serial) == "" {
		devs, err := android.Devices()
		if err != nil {
			return tui.ErrorStyle.Render(fmt.Sprintf("install: %v (is adb on PATH?)", err))
		}
		if len(devs) == 0 {
			return tui.ErrorStyle.Render("install: no devices — connect one, check `devices`")
		}
		if len(devs) > 1 {
			return tui.ErrorStyle.Render(fmt.Sprintf("install: %d devices — rerun with --device <serial> (see `devices`)", len(devs)))
		}
		serial = devs[0]
	}
	out, err := android.InstallCapture(apk, serial)
	var b strings.Builder
	fmt.Fprintf(&b, "install %s → %s\n", apk, serial)
	if strings.TrimSpace(out) != "" {
		b.WriteString(strings.TrimSpace(out))
	}
	if err != nil {
		return tui.ErrorStyle.Render("install: " + err.Error())
	}
	m.status = "installed " + filepath.Base(apk)
	return b.String()
}

// doDevices lists connected adb devices.
func (m *Model) doDevices() string {
	devs, err := android.Devices()
	if err != nil {
		return tui.ErrorStyle.Render(fmt.Sprintf("devices: %v (is adb on PATH? run `doctor`)", err))
	}
	if len(devs) == 0 {
		return "devices: none — connect one and enable USB debugging"
	}
	var b strings.Builder
	b.WriteString("devices:\n")
	for _, d := range devs {
		b.WriteString("  " + d + "\n")
	}
	return strings.TrimRight(b.String(), "\n")
}

// pickInstallApk prefers packaged data-root releases, then Gradle outputs.
func pickInstallApk(dataRoot, projectRoot, projectName, variant string) string {
	relDir := storage.ReleasesDir(dataRoot, projectName)
	if entries, err := os.ReadDir(relDir); err == nil {
		best := ""
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			matches, _ := filepath.Glob(filepath.Join(relDir, e.Name(), "*-"+variant+".apk"))
			for _, mt := range matches {
				if mt > best {
					best = mt
				}
			}
		}
		if best != "" {
			return best
		}
	}
	cands := []string{android.DebugApkPath}
	if variant == "release" {
		cands = []string{android.ReleaseApkPath}
	}
	if resolved, err := android.ResolveApk(projectRoot, cands...); err == nil {
		return resolved
	}
	return ""
}
