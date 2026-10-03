package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ojilon/releaseforge/internal/android"
	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check tool, SDK, and device health for Android work",
	RunE: func(cmd *cobra.Command, args []string) error {
		checks := []struct {
			name string
			ok   string
			bad  string
			run  func() (string, bool)
		}{
			{"adb", "", "missing — install platform-tools and add adb to PATH", toolVersion("adb", "version")},
			{"gh", "", "missing — install gh and run gh auth login", toolPresent("gh")},
			{"java", "", "missing — install JDK 17+", toolVersion("java", "-version")},
			{"gradle wrapper", "", "no gradlew[.bat] in project — builds need one", wrapperPresent},
			{"ANDROID_HOME", "", "unset (warn only — apksigner search falls back to PATH)", envPresent("ANDROID_HOME", "ANDROID_SDK_ROOT")},
		}
		for _, c := range checks {
			detail, ok := c.run()
			if ok {
				fmt.Printf("%-14s ok%s\n", c.name+":", suffix(detail))
				continue
			}
			if c.bad != "" {
				fmt.Printf("%-14s %s\n", c.name+":", c.bad)
			}
		}
		devs, err := android.Devices()
		switch {
		case err != nil:
			fmt.Printf("%-14s adb error (%v)\n", "devices:", err)
		case len(devs) == 0:
			fmt.Printf("%-14s none — connect a device and check `adb devices`\n", "devices:")
		default:
			fmt.Printf("%-14s %d connected: %s\n", "devices:", len(devs), strings.Join(devs, ", "))
		}
		return nil
	},
}

func suffix(detail string) string {
	if detail == "" {
		return ""
	}
	return " (" + detail + ")"
}

// toolPresent reports whether a binary is on PATH.
func toolPresent(name string) func() (string, bool) {
	return func() (string, bool) {
		_, err := exec.LookPath(name)
		return "", err == nil
	}
}

// toolVersion runs `bin arg` and returns its first output line.
func toolVersion(bin, arg string) func() (string, bool) {
	return func() (string, bool) {
		out, err := exec.Command(bin, arg).CombinedOutput()
		if err != nil {
			// java -version exits 0 but writes to stderr; a missing binary
			// errors here. Fall back to presence so java still reports ok.
			if _, lerr := exec.LookPath(bin); lerr == nil && len(out) > 0 {
				return firstLine(string(out)), true
			}
			return "", false
		}
		return firstLine(string(out)), true
	}
}

func firstLine(s string) string {
	if i := strings.Index(s, "\n"); i >= 0 {
		return strings.TrimSpace(s[:i])
	}
	return strings.TrimSpace(s)
}

// wrapperPresent checks for gradlew in the current project dir.
func wrapperPresent() (string, bool) {
	for _, n := range []string{"gradlew.bat", "gradlew"} {
		if st, err := os.Stat(filepath.Join(projectDir, n)); err == nil && !st.IsDir() {
			return n, true
		}
	}
	return "", false
}

// envPresent checks any of the named env vars.
func envPresent(names ...string) func() (string, bool) {
	return func() (string, bool) {
		for _, n := range names {
			if v := strings.TrimSpace(os.Getenv(n)); v != "" {
				return n + " set", true
			}
		}
		return "", false
	}
}
