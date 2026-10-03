package project

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Version sources for Info.VersionSource.
const (
	SourceProperties = "properties"
	SourceKts        = "kts"
	SourceFile       = "file"
	SourceNone       = "none"
)

var (
	codeRe = regexp.MustCompile(`(?m)^app\.versionCode\s*=\s*(.+?)\s*$`)
	nameRe = regexp.MustCompile(`(?m)^app\.versionName\s*=\s*(.+?)\s*$`)
	semverRe = regexp.MustCompile(`^[0-9]+\.[0-9]+(\.[0-9]+)?([._+\-].*)?$`)
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
	next, err := IncrementCode(strings.TrimSpace(cm[1]))
	if err != nil {
		return "", fmt.Errorf("invalid app.versionCode %q: %w", cm[1], err)
	}
	text = codeRe.ReplaceAllString(text, fmt.Sprintf("app.versionCode=%s", next))
	text = nameRe.ReplaceAllString(text, "app.versionName="+versionName)
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		return "", err
	}
	return next, nil
}

// IncrementCode parses a numeric code and returns code+1.
func IncrementCode(current string) (string, error) {
	var n int
	if _, err := fmt.Sscanf(strings.TrimSpace(current), "%d", &n); err != nil {
		return "", err
	}
	return fmt.Sprintf("%d", n+1), nil
}

// IncrementGradleCode bumps app.versionCode by 1 without touching the name.
func IncrementGradleCode(projectRoot, file string) (string, error) {
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
	next, err := IncrementCode(strings.TrimSpace(cm[1]))
	if err != nil {
		return "", fmt.Errorf("invalid app.versionCode %q: %w", cm[1], err)
	}
	text = codeRe.ReplaceAllString(text, fmt.Sprintf("app.versionCode=%s", next))
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		return "", err
	}
	return next, nil
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

// ktsVersionLine finds a plain `version = "literal"` assignment, returning
// its line number (1-based), quote char, and literal. A non-literal right
// side yields found=false with the raw expression for error reporting.
func ktsVersionLine(path string) (no int, quote byte, literal, raw string, found bool, err error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, 0, "", "", false, err
	}
	for _, l := range linesWithNumbers(string(data)) {
		trimmed := strings.TrimSpace(l.Text)
		if strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "#") || strings.HasPrefix(trimmed, "/*") || strings.HasPrefix(trimmed, "*") {
			continue
		}
		idx := strings.Index(l.Text, "version")
		if idx < 0 {
			continue
		}
		after := l.Text[idx+len("version"):]
		// Skip versionCode/versionName and dotted accesses (e.g. android.version).
		if idx > 0 {
			prev := l.Text[idx-1]
			if prev == '.' || prev == '_' || (prev >= 'a' && prev <= 'z') || (prev >= 'A' && prev <= 'Z') {
				continue
			}
		}
		rest := strings.TrimSpace(after)
		if !strings.HasPrefix(rest, "=") {
			continue
		}
		rest = strings.TrimSpace(strings.TrimPrefix(rest, "="))
		if rest == "" {
			continue
		}
		q := rest[0]
		if q != '"' && q != '\'' {
			return l.No, 0, "", rest, false, nil
		}
		end := strings.IndexByte(rest[1:], q)
		if end < 0 {
			return l.No, 0, "", rest, false, nil
		}
		return l.No, q, rest[1 : 1+end], rest, true, nil
	}
	return 0, 0, "", "", false, nil
}

// KtsVersion reads a plain `version = "…"` literal from a .kts file.
func KtsVersion(projectRoot, file string) (string, error) {
	no, _, literal, raw, found, err := ktsVersionLine(filepath.Join(projectRoot, file))
	if err != nil {
		return "", err
	}
	if !found {
		if raw != "" {
			return "", fmt.Errorf("%s:%d: version is not a plain literal (%s) — set it by hand", file, no, raw)
		}
		return "", fmt.Errorf("no version assignment in %s", file)
	}
	return literal, nil
}

// SetKtsVersion replaces the `version = "…"` literal, preserving quote style
// and indentation. Non-literal assignments are refused with file:line.
func SetKtsVersion(projectRoot, file, version string) error {
	path := filepath.Join(projectRoot, file)
	no, quote, _, raw, found, err := ktsVersionLine(path)
	if err != nil {
		return err
	}
	if !found {
		if raw != "" {
			return fmt.Errorf("%s:%d: version is not a plain literal (%s) — set it by hand", file, no, raw)
		}
		return fmt.Errorf("no version assignment in %s", file)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	lines := strings.Split(string(data), "\n")
	indent := lines[no-1][:strings.Index(lines[no-1], "version")]
	lines[no-1] = fmt.Sprintf("%sversion = %c%s%c", indent, quote, version, quote)
	return os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0o644)
}

// DetectVersionSource reports where an Android version lives: properties
// (gradle.properties with both keys), kts (plain literal), or "".
func DetectVersionSource(root string) (source, file string) {
	if data, err := os.ReadFile(filepath.Join(root, "gradle.properties")); err == nil {
		if codeRe.MatchString(string(data)) && nameRe.MatchString(string(data)) {
			return SourceProperties, "gradle.properties"
		}
	}
	for _, f := range []string{"app/build.gradle.kts", "build.gradle.kts"} {
		if _, _, _, _, found, err := ktsVersionLine(filepath.Join(root, f)); err == nil && found {
			return SourceKts, f
		}
	}
	return "", ""
}

// ValidateVersion rejects unusable version names. Non-semver-like names are
// allowed (Gradle names are free-form); use LooksSemver for the warning.
func ValidateVersion(v string) error {
	if strings.TrimSpace(v) == "" {
		return fmt.Errorf("version must not be empty")
	}
	if strings.ContainsAny(v, " \t\r\n") {
		return fmt.Errorf("version %q must not contain whitespace", v)
	}
	return nil
}

// LooksSemver reports whether v starts like X.Y[.Z].
func LooksSemver(v string) bool { return semverRe.MatchString(v) }

// CurrentVersion returns (code, name) for known types; generic returns ("", "", nil).
func CurrentVersion(info Info) (code, name string, err error) {
	switch info.Type {
	case "android-gradle":
		if info.VersionSource == SourceKts {
			v, err := KtsVersion(info.Root, info.VersionFile)
			if err != nil {
				return "", "", err
			}
			return "", v, nil
		}
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
