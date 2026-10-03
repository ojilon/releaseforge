package android

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// fakeAdb writes an adb.bat fixture. argsLog receives "%*" of install/launch
// calls; logcat/pidof/devices get canned responses.
func fakeAdb(t *testing.T, argsLog string) string {
	t.Helper()
	dir := t.TempDir()
	script := "@echo off\r\n" +
		"if \"%1\"==\"devices\" (\r\n" +
		"  echo List of devices attached\r\n" +
		"  echo emulator-5554 device\r\n" +
		"  echo deadbeef unauthorized\r\n" +
		"  exit /b 0\r\n" +
		")\r\n" +
		"if \"%1\"==\"logcat\" (\r\n" +
		"  echo 10-03 20:00:00.000  1234  5678 I FooTag: hello world\r\n" +
		"  echo 10-03 20:00:01.000  9999  9999 I BarTag: other line\r\n" +
		"  echo 10-03 20:00:02.000  1234  5678 I FooTag: second line\r\n" +
		"  exit /b 0\r\n" +
		")\r\n" +
		"if \"%1\"==\"shell\" (\r\n" +
		"  if \"%2\"==\"pidof\" (\r\n" +
		"    echo 1234\r\n" +
		"    exit /b 0\r\n" +
		"  )\r\n" +
		"  echo %* >> \"" + argsLog + "\"\r\n" +
		"  exit /b 0\r\n" +
		")\r\n" +
		"echo %* >> \"" + argsLog + "\"\r\n" +
		"exit /b 0\r\n"
	if err := os.WriteFile(filepath.Join(dir, "adb.bat"), []byte(script), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	return dir
}

func TestDevicesParsesFixture(t *testing.T) {
	fakeAdb(t, filepath.Join(t.TempDir(), "args.log"))
	devs, err := Devices()
	if err != nil {
		t.Fatal(err)
	}
	if len(devs) != 1 || devs[0] != "emulator-5554" {
		t.Fatalf("got %v (unauthorized must be excluded)", devs)
	}
}

func TestSelectDeviceMatrix(t *testing.T) {
	if got, err := selectDevice("S1", []string{"a", "b"}, false); err != nil || got != "S1" {
		t.Fatalf("preferred: %q %v", got, err)
	}
	if got, err := selectDevice("", []string{"only"}, false); err != nil || got != "only" {
		t.Fatalf("lone: %q %v", got, err)
	}
	if _, err := selectDevice("", nil, false); err == nil {
		t.Fatal("expected no-device error")
	}
	if _, err := selectDevice("", []string{"a", "b"}, false); err == nil ||
		!strings.Contains(err.Error(), "--device") {
		t.Fatalf("expected --device error, got %v", err)
	}
}

func TestInstallArgv(t *testing.T) {
	log := filepath.Join(t.TempDir(), "args.log")
	fakeAdb(t, log)
	apk := filepath.Join(t.TempDir(), "a.apk")
	os.WriteFile(apk, []byte("x"), 0o644)
	if err := Install(apk, ""); err != nil {
		t.Fatal(err)
	}
	if err := Install(apk, "S9"); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(log)
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 2 {
		t.Fatalf("got %q", data)
	}
	if !strings.Contains(lines[0], "install -r "+apk) {
		t.Fatalf("line0: %q", lines[0])
	}
	if !strings.Contains(lines[1], "-s S9 install -r "+apk) {
		t.Fatalf("line1: %q", lines[1])
	}
}

func TestLaunchArgv(t *testing.T) {
	log := filepath.Join(t.TempDir(), "args.log")
	fakeAdb(t, log)
	if err := LaunchApp("", "com.example.app"); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(log)
	if !strings.Contains(string(data), "shell monkey -p com.example.app -c android.intent.category.LAUNCHER 1") {
		t.Fatalf("got %q", data)
	}
	if err := LaunchApp("", ""); err == nil {
		t.Fatal("expected empty-package error")
	}
}

func TestLogcatPidFilter(t *testing.T) {
	fakeAdb(t, filepath.Join(t.TempDir(), "args.log"))
	lines, err := LogcatTail("", 0, "com.example.app")
	if err != nil {
		t.Fatal(err)
	}
	// pidof returns 1234, so only pid lines survive (substring would also match).
	if len(lines) != 2 {
		t.Fatalf("got %v", lines)
	}
	for _, l := range lines {
		if !strings.Contains(l, "1234") {
			t.Fatalf("unfiltered line: %q", l)
		}
	}
	all, err := LogcatTail("", 1, "")
	if err != nil || len(all) != 1 {
		t.Fatalf("tail bound: %v %v", all, err)
	}
}

func TestWithSerialPrefix(t *testing.T) {
	if got := withSerial("S", []string{"install"}); len(got) != 3 || got[0] != "-s" {
		t.Fatalf("got %v", got)
	}
	if got := withSerial("", []string{"install"}); len(got) != 1 {
		t.Fatalf("got %v", got)
	}
}
