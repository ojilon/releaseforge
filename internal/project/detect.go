// Package project detects type, reads/writes version, and exposes a Project interface.
package project

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
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
			return Info{Root: abs, Name: name, Type: "android-gradle",
				VersionFile: "gradle.properties", CodeKey: "app.versionCode", NameKey: "app.versionName"}, nil
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
		vf := "go.mod"
		if fileExists(abs, "VERSION") {
			vf = "VERSION"
		}
		return Info{Root: abs, Name: name, Type: "go", VersionFile: vf}, nil
	}
	return Info{Root: abs, Name: name, Type: "generic", VersionFile: ""}, nil
}

var (
	codeRe = regexp.MustCompile(`(?m)^app\.versionCode\s*=\s*(.+?)\s*$`)
	nameRe = regexp.MustCompile(`(?m)^app\.versionName\s*=\s*(.+?)\s*$`)
)

// GradleVersion reads code + name from gradle.properties.
func GradleVersion(projectRoot, file string) (code, name string, err error) {
	if file == "" {
		file = "gradle.properties"
	}
	data, err := os.ReadFile(filepath.Join(projectRoot, file))
	if err != nil {
		return "", "", err
	}
	text := string(data)
	cm := codeRe.FindStringSubmatch(text)
	nm := nameRe.FindStringSubmatch(text)
	if cm == nil || nm == nil {
		return "", "", fmt.Errorf("version info not found in %s (want app.versionCode/app.versionName)", file)
	}
	return strings.TrimSpace(cm[1]), strings.TrimSpace(nm[1]), nil
}

// SetGradleVersion sets versionName and increments versionCode by 1.
// Semantics match scripts/version.py.
func SetGradleVersion(projectRoot, file, versionName string) (newCode string, err error) {
	if strings.TrimSpace(versionName) == "" {
		return "", fmt.Errorf("version name must not be empty")
	}
	if file == "" {
		file = "gradle.properties"
	}
	path := filepath.Join(projectRoot, file)
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	text := string(data)
	cm := codeRe.FindStringSubmatch(text)
	if cm == nil {
		return "", fmt.Errorf("app.versionCode not found in %s", file)
	}
	if nameRe.FindStringSubmatch(text) == nil {
		return "", fmt.Errorf("app.versionName not found in %s", file)
	}
	var current int
	if _, err := fmt.Sscanf(strings.TrimSpace(cm[1]), "%d", &current); err != nil {
		return "", fmt.Errorf("invalid app.versionCode %q: %w", cm[1], err)
	}
	next := current + 1
	text = codeRe.ReplaceAllString(text, fmt.Sprintf("app.versionCode=%d", next))
	text = nameRe.ReplaceAllString(text, "app.versionName="+versionName)
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		return "", err
	}
	return fmt.Sprintf("%d", next), nil
}

// ReadVersionFile reads a plain version file (e.g. VERSION). No auto-commit;
// callers decide when to commit.
func ReadVersionFile(projectRoot, file string) (string, error) {
	if strings.TrimSpace(file) == "" {
		file = "VERSION"
	}
	data, err := os.ReadFile(filepath.Join(projectRoot, file))
	if err != nil {
		return "", err
	}
	v := strings.TrimSpace(string(data))
	if v == "" {
		return "", fmt.Errorf("%s is empty", file)
	}
	return v, nil
}

// SetVersionFile writes a plain version file (e.g. VERSION). No commit.
func SetVersionFile(projectRoot, file, version string) error {
	version = strings.TrimSpace(version)
	if version == "" {
		return fmt.Errorf("version must not be empty")
	}
	if strings.TrimSpace(file) == "" {
		file = "VERSION"
	}
	return os.WriteFile(filepath.Join(projectRoot, file), []byte(version+"\n"), 0o644)
}

// CurrentVersion returns (code, name) for known types; generic returns ("", "", nil).
func CurrentVersion(info Info) (code, name string, err error) {
	switch info.Type {
	case "android-gradle":
		return GradleVersion(info.Root, info.VersionFile)
	case "go":
		if info.VersionFile == "VERSION" || fileExists(info.Root, "VERSION") {
			v, err := ReadVersionFile(info.Root, "VERSION")
			if err != nil {
				return "", "", err
			}
			return "", v, nil
		}
		return "", "", nil
	default:
		return "", "", nil
	}
}
