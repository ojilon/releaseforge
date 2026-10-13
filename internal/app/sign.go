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

// doSign signs an APK with your keystore from the data root.
// Usage: sign [apk-path] [--alias X] [--keystore PATH]
//
// Your files live OUTSIDE the repo:
//
//	<data-root>/signing/<project>/keystore.jks
//	<data-root>/signing/<project>/signing.json  {"alias": "..."}
//
// Password comes only from env RF_KEYSTORE_PASS (never stored).
func (m *Model) doSign(args []string) string {
	info, err := project.Detect(m.projectDir)
	if err != nil {
		return tui.ErrorStyle.Render("sign: " + err.Error())
	}
	apkArg, alias, keystore := "", "", ""
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--alias":
			if i+1 < len(args) {
				alias = args[i+1]
				i++
			}
		case "--keystore":
			if i+1 < len(args) {
				keystore = args[i+1]
				i++
			}
		default:
			if apkArg == "" && !strings.HasPrefix(args[i], "--") {
				apkArg = args[i]
			}
		}
	}
	apk := strings.TrimSpace(apkArg)
	if apk == "" {
		// Newest signed-candidate: packaged releases first, then Gradle outputs.
		if best := newestSignedCandidate(m.dataRoot, info.Name); best != "" {
			apk = best
		} else if r, err := android.ResolveApk(info.Root, android.ReleaseApkPath, android.DebugApkPath); err == nil {
			apk = r
		}
	} else if !filepath.IsAbs(apk) {
		apk = filepath.Join(info.Root, apk)
	}
	if apk == "" {
		return tui.ErrorStyle.Render("sign: no APK found (build release first, or pass a path)")
	}
	cfg, dir, err := android.ResolveSigning(m.dataRoot, info.Name, keystore, alias, "")
	if err != nil {
		return tui.ErrorStyle.Render("sign: " + err.Error())
	}
	if strings.TrimSpace(cfg.Alias) == "" {
		return tui.ErrorStyle.Render(fmt.Sprintf("sign: no alias — add {\"alias\": \"...\"} to %s or pass --alias", android.SigningConfigPath(m.dataRoot, info.Name)))
	}
	pass, err := android.KeystorePassword()
	if err != nil {
		return tui.ErrorStyle.Render("sign: " + err.Error())
	}
	out, err := android.SignApkCapture(apk, cfg.KeystorePath, cfg.Alias, pass, cfg.ApksignerPath)
	var b strings.Builder
	fmt.Fprintf(&b, "signed %s\nkeystore: %s\n", apk, cfg.KeystorePath)
	if strings.TrimSpace(out) != "" {
		b.WriteString(strings.TrimSpace(out) + "\n")
	}
	_ = dir
	if err != nil {
		return tui.ErrorStyle.Render("sign: " + err.Error())
	}
	m.status = "signed " + filepath.Base(apk)
	return strings.TrimRight(b.String(), "\n")
}

// newestSignedCandidate finds the newest *-release.apk under data-root releases.
func newestSignedCandidate(dataRoot, projectName string) string {
	relDir := storage.ReleasesDir(dataRoot, projectName)
	entries, err := os.ReadDir(relDir)
	if err != nil {
		return ""
	}
	best := ""
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		matches, _ := filepath.Glob(filepath.Join(relDir, e.Name(), "*-release.apk"))
		for _, mt := range matches {
			if mt > best {
				best = mt
			}
		}
	}
	return best
}
