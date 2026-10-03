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

// Well-known Gradle APK outputs, used as fallback when scan/config supply
// nothing better.
const (
	DebugApkPath   = "app/build/outputs/apk/debug/app-debug.apk"
	ReleaseApkPath = "app/build/outputs/apk/release/app-release-unsigned.apk"
)

// ResolveApk returns the first existing path among candidates, resolving
// relative entries against projectRoot (typically scan/config values first,
// the well-known outputs last).
func ResolveApk(projectRoot string, candidates ...string) (string, error) {
	for _, c := range candidates {
		if strings.TrimSpace(c) == "" {
			continue
		}
		p := c
		if !filepath.IsAbs(p) {
			p = filepath.Join(projectRoot, c)
		}
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			return p, nil
		}
	}
	return "", fmt.Errorf("no APK found among %d candidate(s) under %s", len(candidates), projectRoot)
}

// PackageFiles copies resolved APKs into versionDir with
// <AppName>-<version>-debug.apk / -release.apk names.
func PackageFiles(version, appName, versionDir, debugSrc, releaseSrc string) (debugOut, releaseOut string, err error) {
	if strings.TrimSpace(version) == "" {
		return "", "", fmt.Errorf("package: version required")
	}
	if strings.TrimSpace(appName) == "" {
		appName = "app"
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

// PackageRelease resolves the well-known Gradle outputs and packages them.
// Callers with scan/config-known locations should ResolveApk first (config
// values take precedence) and call PackageFiles directly.
func PackageRelease(projectRoot, version, appName, versionDir string) (string, string, error) {
	debugSrc, err := ResolveApk(projectRoot, DebugApkPath)
	if err != nil {
		return "", "", fmt.Errorf("debug APK not found (run build debug): %w", err)
	}
	releaseSrc, err := ResolveApk(projectRoot, ReleaseApkPath)
	if err != nil {
		return "", "", fmt.Errorf("release APK not found (run build release): %w", err)
	}
	return PackageFiles(version, appName, versionDir, debugSrc, releaseSrc)
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
