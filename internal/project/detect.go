// Package project detects type, reads/writes version, and exposes a Project interface.
package project

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ojilon/releaseforge/internal/storage"
)

// Info is the result of Detect for a local directory.
type Info struct {
	// Root is the absolute project directory.
	Root string
	// Name is the sanitized base name.
	Name string
	// Type is one of the config.SupportedTypes values.
	Type string
	// VersionFile is the relative version source of truth, if known.
	VersionFile string
	// VersionSource is properties|kts|file|none ("" when unknown/legacy).
	VersionSource string
	// CodeKey / NameKey for gradle-style projects.
	CodeKey string
	NameKey string
}

// Detection order and markers are documented in docs/02-project-types.md.
func fileExists(root, name string) bool {
	_, err := os.Stat(filepath.Join(root, name))
	return err == nil
}

func anyExists(root string, names ...string) string {
	for _, n := range names {
		if fileExists(root, n) {
			return n
		}
	}
	return ""
}

// Detect inspects a local directory and returns its project type.
// Any local directory is valid; unknown trees return type "generic".
func Detect(dir string) (Info, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return Info{}, err
	}
	abs = filepath.Clean(abs)
	st, err := os.Stat(abs)
	if err != nil {
		return Info{}, fmt.Errorf("scan %s: %w", abs, err)
	}
	if !st.IsDir() {
		return Info{}, fmt.Errorf("scan %s: not a directory", abs)
	}
	name := storage.ProjectNameFromPath(abs)

	// 1. Android Gradle: settings.gradle(.kts) + app module + gradle props/wrapper.
	if anyExists(abs, "settings.gradle", "settings.gradle.kts") != "" {
		if anyExists(abs, "app/build.gradle", "app/build.gradle.kts") != "" ||
			fileExists(abs, "gradle.properties") ||
			anyExists(abs, "gradlew", "gradlew.bat") != "" {
			source, vfile := DetectVersionSource(abs)
			if source == "" {
				vfile = "gradle.properties" // legacy default; reads fail honestly
			}
			return Info{Root: abs, Name: name, Type: "android-gradle",
				VersionFile: vfile, VersionSource: source,
				CodeKey: "app.versionCode", NameKey: "app.versionName"}, nil
		}
	}
	// 2. Wails.
	if fileExists(abs, "wails.json") {
		return Info{Root: abs, Name: name, Type: "wails", VersionFile: "wails.json"}, nil
	}
	// 3. Pure CMake.
	if fileExists(abs, "CMakeLists.txt") {
		return Info{Root: abs, Name: name, Type: "cmake", VersionFile: "CMakeLists.txt"}, nil
	}
	// 4. Python.
	if anyExists(abs, "pyproject.toml", "requirements.txt", "setup.py") != "" {
		return Info{Root: abs, Name: name, Type: "python", VersionFile: "pyproject.toml"}, nil
	}
	// 5. Java CLI: pom.xml or non-android build.gradle, or src/main/java layout.
	if fileExists(abs, "pom.xml") || fileExists(abs, "src/main/java") {
		return Info{Root: abs, Name: name, Type: "java-cli", VersionFile: "pom.xml"}, nil
	}
	if fileExists(abs, "build.gradle") || fileExists(abs, "build.gradle.kts") {
		return Info{Root: abs, Name: name, Type: "java-cli", VersionFile: "build.gradle"}, nil
	}
	// 6. Go module. VERSION file at root is the version source when present.
	if fileExists(abs, "go.mod") {
		vf, vs := "go.mod", SourceNone
		if fileExists(abs, "VERSION") {
			vf, vs = "VERSION", SourceFile
		}
		return Info{Root: abs, Name: name, Type: "go", VersionFile: vf, VersionSource: vs}, nil
	}
	return Info{Root: abs, Name: name, Type: "generic", VersionFile: ""}, nil
}

// BinaryBaseName returns the file-safe base name for built binaries.
// It derives from the sanitized project name (set by Detect).
func BinaryBaseName(info Info) string {
	if strings.TrimSpace(info.Name) == "" {
		return "app"
	}
	return info.Name
}
