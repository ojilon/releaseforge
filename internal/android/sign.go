// Package android handles APK location, apksigner, keystore prompts, ABI helpers, adb.
package android

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

// Logic ported from Conductino-Android scripts/sign.py, package.py, etc.
// See docs/06-android-deep-dive.md.

// FindApksigner resolves the apksigner binary:
// custom path -> PATH -> $ANDROID_HOME / $ANDROID_SDK_ROOT build-tools/<newest>.
func FindApksigner(custom string) string {
	if strings.TrimSpace(custom) != "" {
		if st, err := os.Stat(custom); err == nil && !st.IsDir() {
			return custom
		}
	}
	exe := "apksigner"
	if isWindows() {
		exe = "apksigner.bat"
	}
	if p, err := exec.LookPath("apksigner"); err == nil {
		return p
	}
	sdk := os.Getenv("ANDROID_HOME")
	if sdk == "" {
		sdk = os.Getenv("ANDROID_SDK_ROOT")
	}
	if sdk != "" {
		bt := filepath.Join(sdk, "build-tools")
		entries, err := os.ReadDir(bt)
		if err == nil {
			var dirs []string
			for _, e := range entries {
				if e.IsDir() {
					dirs = append(dirs, e.Name())
				}
			}
			sort.Sort(sort.Reverse(sort.StringSlice(dirs)))
			for _, d := range dirs {
				cand := filepath.Join(bt, d, exe)
				if st, err := os.Stat(cand); err == nil && !st.IsDir() {
					return cand
				}
			}
		}
	}
	return exe
}

func isWindows() bool {
	return os.PathSeparator == '\\'
}

// SignApk signs apk with keystore/alias/password and verifies.
// Password is passed via pass: form and never stored.
func SignApk(apkPath, keystorePath, alias, password, apksignerBin string) error {
	if apkPath == "" || keystorePath == "" || alias == "" {
		return fmt.Errorf("sign: apk, keystore and alias are required")
	}
	if _, err := os.Stat(apkPath); err != nil {
		return fmt.Errorf("sign: apk not found: %s", apkPath)
	}
	if _, err := os.Stat(keystorePath); err != nil {
		return fmt.Errorf("sign: keystore not found: %s", keystorePath)
	}
	bin := FindApksigner(apksignerBin)
	sign := exec.Command(bin, "sign",
		"--ks", keystorePath,
		"--ks-key-alias", alias,
		"--ks-pass", "pass:"+password,
		apkPath)
	sign.Stdout = os.Stdout
	sign.Stderr = os.Stderr
	if err := sign.Run(); err != nil {
		return fmt.Errorf("apksigner sign failed: %w", err)
	}
	verify := exec.Command(bin, "verify", "--verbose", apkPath)
	verify.Stdout = os.Stdout
	verify.Stderr = os.Stderr
	if err := verify.Run(); err != nil {
		return fmt.Errorf("apksigner verify failed: %w", err)
	}
	return nil
}

// PackageRelease copies debug + unsigned release APKs into versionDir with
// <AppName>-<version>-debug.apk / -release.apk names (mirrors scripts/package.py
// but with configurable app name and data-root-friendly output dir).
func PackageRelease(projectRoot, version, appName, versionDir string) (debugOut, releaseOut string, err error) {
	if strings.TrimSpace(version) == "" {
		return "", "", fmt.Errorf("package: version required")
	}
	if strings.TrimSpace(appName) == "" {
		appName = "app"
	}
	debugSrc := filepath.Join(projectRoot, "app", "build", "outputs", "apk", "debug", "app-debug.apk")
	releaseSrc := filepath.Join(projectRoot, "app", "build", "outputs", "apk", "release", "app-release-unsigned.apk")
	if _, err := os.Stat(debugSrc); err != nil {
		return "", "", fmt.Errorf("debug APK not found: %s (run build debug)", debugSrc)
	}
	if _, err := os.Stat(releaseSrc); err != nil {
		return "", "", fmt.Errorf("release APK not found: %s (run build release)", releaseSrc)
	}
	if err := os.MkdirAll(versionDir, 0o755); err != nil {
		return "", "", err
	}
	debugOut = filepath.Join(versionDir, fmt.Sprintf("%s-%s-debug.apk", appName, version))
	releaseOut = filepath.Join(versionDir, fmt.Sprintf("%s-%s-release.apk", appName, version))
	if err := copyFile(debugSrc, debugOut); err != nil {
		return "", "", err
	}
	if err := copyFile(releaseSrc, releaseOut); err != nil {
		return "", "", err
	}
	return debugOut, releaseOut, nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
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

// Devices lists connected adb devices (serials with state device).
func Devices() ([]string, error) {
	out, err := exec.Command("adb", "devices").Output()
	if err != nil {
		return nil, fmt.Errorf("adb devices: %w", err)
	}
	var devs []string
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "List of") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) == 2 && fields[1] == "device" {
			devs = append(devs, fields[0])
		}
	}
	return devs, nil
}

// Install pushes apk to a device via adb (serial optional; first device if empty).
func Install(apkPath, serial string) error {
	if _, err := os.Stat(apkPath); err != nil {
		return fmt.Errorf("install: apk not found: %s", apkPath)
	}
	args := []string{"install", "-r"}
	if strings.TrimSpace(serial) != "" {
		args = []string{"-s", serial, "install", "-r"}
	}
	args = append(args, apkPath)
	cmd := exec.Command("adb", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("adb install failed: %w", err)
	}
	return nil
}
