package build

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// ModulePath reads the module path from go.mod (fallback "main").
func ModulePath(dir string) string {
	data, err := os.ReadFile(filepath.Join(dir, "go.mod"))
	if err != nil {
		return "main"
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module "))
		}
	}
	return "main"
}

// GoTestArgs runs the full module test suite.
func GoTestArgs() []string { return []string{"test", "./..."} }

// GoBuildArgs builds the module root into outPath, stamping the tool version.
// variant "release" adds trimpath and stripped symbols.
func GoBuildArgs(dir, outPath, versionStamp, variant string) []string {
	ld := fmt.Sprintf("-X %s/internal/version.ToolVersion=%s", ModulePath(dir), versionStamp)
	if variant == "release" {
		ld = "-s -w " + ld
		return []string{"build", "-trimpath", "-ldflags", ld, "-o", outPath, "."}
	}
	return []string{"build", "-ldflags", ld, "-o", outPath, "."}
}

// GoBinaryName returns the platform-appropriate binary name.
func GoBinaryName(base string) string {
	if runtime.GOOS == "windows" {
		return base + ".exe"
	}
	return base
}
