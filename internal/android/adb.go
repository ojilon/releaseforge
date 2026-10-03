package android

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/huh"
)

// withSerial prepends -s when a serial is given.
func withSerial(serial string, args []string) []string {
	if strings.TrimSpace(serial) != "" {
		return append([]string{"-s", strings.TrimSpace(serial)}, args...)
	}
	return args
}

// adbStream runs adb with inherited stdio (install, launch).
func adbStream(serial string, args ...string) error {
	cmd := exec.Command("adb", withSerial(serial, args)...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// adbOut runs adb and returns combined output.
func adbOut(serial string, args ...string) (string, error) {
	out, err := exec.Command("adb", withSerial(serial, args)...).CombinedOutput()
	return string(out), err
}

// Devices lists connected adb devices (serials with state device).
func Devices() ([]string, error) {
	out, err := adbOut("", "devices")
	if err != nil {
		return nil, fmt.Errorf("adb devices: %w", err)
	}
	var devs []string
	for _, line := range strings.Split(out, "\n") {
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

// Install pushes apk to a device via adb (serial optional).
func Install(apkPath, serial string) error {
	if _, err := os.Stat(apkPath); err != nil {
		return fmt.Errorf("install: apk not found: %s", apkPath)
	}
	if err := adbStream(serial, "install", "-r", apkPath); err != nil {
		return fmt.Errorf("adb install failed: %w", err)
	}
	return nil
}

// isTTY reports whether stdin is a terminal (picker only offered there).
func isTTY() bool {
	fi, err := os.Stdin.Stat()
	return err == nil && (fi.Mode()&os.ModeCharDevice) != 0
}

// selectDevice is the pure picker core: preferred wins; one device returns
// itself; none errors; many need interactive selection.
func selectDevice(preferred string, devs []string, interactive bool) (string, error) {
	if strings.TrimSpace(preferred) != "" {
		return strings.TrimSpace(preferred), nil
	}
	switch len(devs) {
	case 0:
		return "", fmt.Errorf("no devices connected — connect one and check `adb devices`")
	case 1:
		return devs[0], nil
	default:
		if !interactive {
			return "", fmt.Errorf("%d devices connected — pick one with --device <serial>", len(devs))
		}
		var choice string
		form := huh.NewForm(huh.NewGroup(
			huh.NewSelect[string]().
				Title(fmt.Sprintf("%d devices connected — pick one", len(devs))).
				Options(huh.NewOptions(devs...)...).
				Value(&choice),
		))
		if err := form.Run(); err != nil {
			return "", err
		}
		return choice, nil
	}
}

// PickDevice resolves the target device: --device wins, else auto-pick a lone
// device, else an interactive picker on TTYs (clear --device error otherwise).
func PickDevice(preferred string) (string, error) {
	if strings.TrimSpace(preferred) != "" {
		return strings.TrimSpace(preferred), nil
	}
	devs, err := Devices()
	if err != nil {
		return "", fmt.Errorf("%v (is adb on PATH? run `doctor`)", err)
	}
	return selectDevice("", devs, isTTY())
}

// LaunchApp starts the launcher activity via monkey.
func LaunchApp(serial, pkg string) error {
	if strings.TrimSpace(pkg) == "" {
		return fmt.Errorf("launch: unknown package id — scan the project first")
	}
	if err := adbStream(serial,
		"shell", "monkey", "-p", pkg, "-c", "android.intent.category.LAUNCHER", "1"); err != nil {
		return fmt.Errorf("launch failed: %w", err)
	}
	return nil
}

// LogcatTail returns the last n lines of `adb logcat -d`, optionally filtered
// to one package (pidof when available, substring fallback).
func LogcatTail(serial string, n int, pkg string) ([]string, error) {
	out, err := adbOut(serial, "logcat", "-d")
	if err != nil {
		return nil, fmt.Errorf("adb logcat: %w", err)
	}
	lines := splitLogLines(out)
	if strings.TrimSpace(pkg) != "" {
		lines = filterPackage(serial, lines, strings.TrimSpace(pkg))
	}
	if n > 0 && len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	return lines, nil
}

// filterPackage keeps lines for pkg: pid match via pidof, else substring.
func filterPackage(serial string, lines []string, pkg string) []string {
	if out, err := adbOut(serial, "shell", "pidof", pkg); err == nil {
		if pid := strings.Fields(out); len(pid) > 0 {
			var kept []string
			for _, l := range lines {
				if f := strings.Fields(l); len(f) > 2 && f[2] == pid[0] {
					kept = append(kept, l)
				}
			}
			if len(kept) > 0 {
				return kept
			}
		}
	}
	var kept []string
	for _, l := range lines {
		if strings.Contains(l, pkg) {
			kept = append(kept, l)
		}
	}
	return kept
}

func splitLogLines(s string) []string {
	var out []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			out = append(out, strings.TrimRight(s[start:i], "\r"))
			start = i + 1
		}
	}
	if start < len(s) {
		out = append(out, s[start:])
	}
	return out
}
