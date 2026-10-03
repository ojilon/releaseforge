package project

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/ojilon/releaseforge/internal/build"
	"github.com/ojilon/releaseforge/internal/storage"
)

// PropVal is a scanned value with provenance. Unresolved holds the raw
// expression (property()/extra[]/…) when the value could not be determined;
// such values are reported honestly and never guessed.
type PropVal struct {
	Value      string `json:"value,omitempty"`
	File       string `json:"file"`
	Line       int    `json:"line"`
	Unresolved string `json:"unresolved,omitempty"`
}

// AndroidBlock holds parsed android-module fields.
type AndroidBlock struct {
	Namespace      PropVal  `json:"namespace,omitempty"`
	ApplicationID  PropVal  `json:"application_id,omitempty"`
	CompileSdk     PropVal  `json:"compile_sdk,omitempty"`
	MinSdk         PropVal  `json:"min_sdk,omitempty"`
	TargetSdk      PropVal  `json:"target_sdk,omitempty"`
	NdkVersion     PropVal  `json:"ndk_version,omitempty"`
	AbiFilters     []string `json:"abi_filters,omitempty"`
	AbiFiltersFrom PropVal  `json:"abi_filters_from,omitempty"`
	CmakePath      PropVal  `json:"cmake_path,omitempty"`
	VersionCode    PropVal  `json:"version_code,omitempty"`
	VersionName    PropVal  `json:"version_name,omitempty"`
}

// WrapperInfo holds Gradle wrapper provenance.
type WrapperInfo struct {
	DistributionURL string `json:"distribution_url,omitempty"`
	GradleVersion   string `json:"gradle_version,omitempty"`
	File            string `json:"file,omitempty"`
}

// ManifestInfo holds manifest basics (never the full XML).
type ManifestInfo struct {
	Package       string `json:"package,omitempty"`
	LauncherFound bool   `json:"launcher_found"`
	File          string `json:"file,omitempty"`
}

// GradleInfo is the deep Android Gradle snapshot attached to Snapshot.Gradle.
type GradleInfo struct {
	Properties map[string]PropVal `json:"properties,omitempty"`
	Modules    []string           `json:"modules,omitempty"`
	Android    AndroidBlock       `json:"android"`
	Wrapper    WrapperInfo        `json:"wrapper,omitempty"`
	SdkDir     PropVal            `json:"sdk_dir,omitempty"`
	Versions   map[string]string  `json:"versions,omitempty"`
	VersionsFile string           `json:"versions_file,omitempty"`
	Manifest   ManifestInfo       `json:"manifest,omitempty"`
}

var (
	reInclude     = regexp.MustCompile(`include\s*\(?\s*["']([^"']+)["']`)
	reQuotedEq    = regexp.MustCompile(`=\s*["']([^"']+)["']`)
	reGroovyStr   = regexp.MustCompile(`\s+["']([^"']+)["']\s*$`)
	reAbiFilters  = regexp.MustCompile(`abiFilters\s*(?:\+=)?\s*[\(\[]?\s*([^\]\)\n]+)`)
	reCmakePath   = regexp.MustCompile(`path\s*=\s*["']([^"']+)["']|path\s+["']([^"']+)["']`)
	reIntToken    = regexp.MustCompile(`(?:=\s*)?["']?([^\s"'}),]+)["']?`)
	reUnresolved  = regexp.MustCompile(`property\s*\(|extra\s*\[|project\.|providers\.|rootProject|libs\.`)
	reDistVersion = regexp.MustCompile(`gradle-([0-9][0-9.]*)-`)
	reManifestPkg = regexp.MustCompile(`<manifest[^>]*\spackage\s*=\s*"([^"]+)"`)
)

func pv(file string, line int, value string) PropVal {
	return PropVal{Value: value, File: file, Line: line}
}

func pvUnresolved(file string, line int, expr string) PropVal {
	return PropVal{File: file, Line: line, Unresolved: expr}
}

// linesWithNumbers splits text into (lineNo, line) pairs, 1-based.
func linesWithNumbers(text string) []struct {
	No   int
	Text string
} {
	raw := strings.Split(text, "\n")
	out := make([]struct {
		No   int
		Text string
	}, 0, len(raw))
	for i, l := range raw {
		out = append(out, struct {
			No   int
			Text string
		}{No: i + 1, Text: strings.TrimRight(l, "\r")})
	}
	return out
}

// parsePropertiesFile reads a gradle.properties-style file into key → value
// with line numbers. Comments (#/!) and blank lines are skipped.
func parsePropertiesFile(path, rel string) map[string]PropVal {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	out := map[string]PropVal{}
	for _, l := range linesWithNumbers(string(data)) {
		t := strings.TrimSpace(l.Text)
		if t == "" || strings.HasPrefix(t, "#") || strings.HasPrefix(t, "!") {
			continue
		}
		idx := strings.Index(t, "=")
		if idx < 0 {
			idx = strings.Index(t, ":")
		}
		if idx <= 0 {
			continue
		}
		key := strings.TrimSpace(t[:idx])
		val := strings.TrimSpace(t[idx+1:])
		if key == "" {
			continue
		}
		out[key] = pv(rel, l.No, val)
	}
	return out
}

// strField extracts name = "value" or name "value" from one line.
func strField(line, name, file string, no int) (PropVal, bool) {
	prefix := strings.Index(line, name)
	if prefix < 0 {
		return PropVal{}, false
	}
	rest := line[prefix+len(name):]
	if m := reQuotedEq.FindStringSubmatch(rest); m != nil {
		return pv(file, no, m[1]), true
	}
	if m := reGroovyStr.FindStringSubmatch(rest); m != nil && !strings.Contains(rest, "(") {
		return pv(file, no, m[1]), true
	}
	return PropVal{}, false
}

// tokenField extracts a bare token (int, identifier, or expression) for name.
func tokenField(line, name, file string, no int) (string, bool) {
	idx := strings.Index(line, name)
	if idx < 0 {
		return "", false
	}
	rest := strings.TrimSpace(line[idx+len(name):])
	rest = strings.TrimPrefix(rest, "=")
	rest = strings.TrimSpace(rest)
	m := reIntToken.FindStringSubmatch(rest)
	if m == nil || m[1] == "" {
		return "", false
	}
	tok := m[1]
	if reUnresolved.MatchString(tok) || reUnresolved.MatchString(rest) {
		return "unresolved:" + tok, true
	}
	return tok, true
}

func unvalue(file string, no int, tok string) PropVal {
	if strings.HasPrefix(tok, "unresolved:") {
		return pvUnresolved(file, no, strings.TrimPrefix(tok, "unresolved:"))
	}
	return pv(file, no, strings.Trim(tok, `"'`))
}

// parseModuleFile extracts android{} fields from one build.gradle[.kts] file.
func parseModuleFile(path, rel string, blk *AndroidBlock) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	setStr := func(dst *PropVal, name string) {
		for _, l := range linesWithNumbers(string(data)) {
			if v, ok := strField(l.Text, name, rel, l.No); ok {
				*dst = v
				return
			}
		}
	}
	setTok := func(dst *PropVal, name string) {
		for _, l := range linesWithNumbers(string(data)) {
			if tok, ok := tokenField(l.Text, name, rel, l.No); ok {
				*dst = unvalue(rel, l.No, tok)
				return
			}
		}
	}
	setStr(&blk.Namespace, "namespace")
	setStr(&blk.ApplicationID, "applicationId")
	setTok(&blk.CompileSdk, "compileSdk")
	setTok(&blk.MinSdk, "minSdk")
	setTok(&blk.TargetSdk, "targetSdk")
	setStr(&blk.NdkVersion, "ndkVersion")
	for _, l := range linesWithNumbers(string(data)) {
		if m := reAbiFilters.FindStringSubmatch(l.Text); m != nil {
			var abis []string
			for _, p := range strings.Split(m[1], ",") {
				p = strings.Trim(strings.TrimSpace(p), `"'\s()`)
				p = strings.TrimPrefix(p, "listOf(")
				if p != "" && p != "..." {
					abis = append(abis, p)
				}
			}
			if len(abis) > 0 {
				blk.AbiFilters = abis
				blk.AbiFiltersFrom = pv(rel, l.No, strings.Join(abis, ","))
			}
			break
		}
	}
	for _, l := range linesWithNumbers(string(data)) {
		if m := reCmakePath.FindStringSubmatch(l.Text); m != nil {
			for _, g := range m[1:] {
				if g != "" {
					blk.CmakePath = pv(rel, l.No, g)
					break
				}
			}
			break
		}
	}
	setTok(&blk.VersionCode, "versionCode")
	for _, l := range linesWithNumbers(string(data)) {
		idx := strings.Index(l.Text, "versionName")
		if idx < 0 {
			continue
		}
		rest := strings.TrimSpace(l.Text[idx+len("versionName"):])
		rest = strings.TrimPrefix(rest, "=")
		rest = strings.TrimSpace(rest)
		if reUnresolved.MatchString(rest) {
			blk.VersionName = pvUnresolved(rel, l.No, rest)
		} else if m := reQuotedEq.FindStringSubmatch("=" + rest); m != nil {
			blk.VersionName = pv(rel, l.No, m[1])
		} else if f := strings.Fields(rest); len(f) > 0 {
			blk.VersionName = pv(rel, l.No, strings.Trim(f[0], `"',`))
		}
		break
	}
}

// ScanGradle deep-parses Android Gradle metadata under root. It never executes
// anything and never fails the scan: missing files yield empty sections.
func ScanGradle(root string) *GradleInfo {
	g := &GradleInfo{}
	if props := parsePropertiesFile(filepath.Join(root, "gradle.properties"), "gradle.properties"); len(props) > 0 {
		g.Properties = props
	}
	for _, rel := range []string{"settings.gradle", "settings.gradle.kts"} {
		data, err := os.ReadFile(filepath.Join(root, rel))
		if err != nil {
			continue
		}
		seen := map[string]bool{}
		for _, m := range reInclude.FindAllStringSubmatch(string(data), -1) {
			if !seen[m[1]] {
				seen[m[1]] = true
				g.Modules = append(g.Modules, m[1])
			}
		}
	}
	for _, rel := range []string{"app/build.gradle", "app/build.gradle.kts"} {
		if _, err := os.Stat(filepath.Join(root, rel)); err == nil {
			parseModuleFile(filepath.Join(root, rel), rel, &g.Android)
		}
	}
	if data, err := os.ReadFile(filepath.Join(root, "gradle", "wrapper", "gradle-wrapper.properties")); err == nil {
		for _, l := range linesWithNumbers(string(data)) {
			if strings.HasPrefix(strings.TrimSpace(l.Text), "distributionUrl") {
				parts := strings.SplitN(l.Text, "=", 2)
				if len(parts) == 2 {
					url := strings.TrimSpace(parts[1])
					g.Wrapper = WrapperInfo{DistributionURL: url, File: "gradle/wrapper/gradle-wrapper.properties"}
					if m := reDistVersion.FindStringSubmatch(url); m != nil {
						g.Wrapper.GradleVersion = m[1]
					}
				}
			}
		}
	}
	if data, err := os.ReadFile(filepath.Join(root, "local.properties")); err == nil {
		for _, l := range linesWithNumbers(string(data)) {
			t := strings.TrimSpace(l.Text)
			if strings.HasPrefix(t, "sdk.dir") {
				g.SdkDir = PropVal{Value: "<set>", File: "local.properties", Line: l.No}
				break
			}
		}
	}
	if data, err := os.ReadFile(filepath.Join(root, "gradle", "libs.versions.toml")); err == nil {
		inVersions := false
		vers := map[string]string{}
		for _, l := range linesWithNumbers(string(data)) {
			t := strings.TrimSpace(l.Text)
			if strings.HasPrefix(t, "[") {
				inVersions = t == "[versions]"
				continue
			}
			if inVersions && strings.Contains(t, "=") && !strings.HasPrefix(t, "#") {
				kv := strings.SplitN(t, "=", 2)
				vers[strings.TrimSpace(kv[0])] = strings.Trim(strings.TrimSpace(kv[1]), `"`)
			}
		}
		if len(vers) > 0 {
			g.Versions = vers
			g.VersionsFile = "gradle/libs.versions.toml"
		}
	}
	g.Manifest = findManifest(root)
	return g
}

func findManifest(root string) ManifestInfo {
	var cands []string
	seen := map[string]bool{}
	add := func(abs string) {
		rel, err := filepath.Rel(root, abs)
		if err != nil || seen[rel] {
			return
		}
		seen[rel] = true
		cands = append(cands, rel)
	}
	if _, err := os.Stat(filepath.Join(root, "app", "src", "main", "AndroidManifest.xml")); err == nil {
		add(filepath.Join(root, "app", "src", "main", "AndroidManifest.xml"))
	}
	_ = filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
		if err != nil || len(cands) >= 5 {
			return nil
		}
		if d.IsDir() && (d.Name() == ".git" || d.Name() == "build" || d.Name() == ".gradle") {
			return filepath.SkipDir
		}
		if !d.IsDir() && d.Name() == "AndroidManifest.xml" && p != filepath.Join(root, "app", "src", "main", "AndroidManifest.xml") {
			if rel, rerr := filepath.Rel(root, p); rerr == nil && strings.Count(rel, string(filepath.Separator)) <= 3 {
				add(p)
			}
		}
		return nil
	})
	for _, rel := range cands {
		data, err := os.ReadFile(filepath.Join(root, rel))
		if err != nil {
			continue
		}
		mi := ManifestInfo{File: rel}
		if m := reManifestPkg.FindStringSubmatch(string(data)); m != nil {
			mi.Package = m[1]
		}
		if strings.Contains(string(data), "android.intent.action.MAIN") &&
			strings.Contains(string(data), "android.intent.category.LAUNCHER") {
			mi.LauncherFound = true
		}
		return mi
	}
	return ManifestInfo{}
}

var skipDirs = map[string]bool{".git": true, "build": true, "node_modules": true, ".gradle": true, ".idea": true}

// DeepProperties runs `gradlew :app:properties` and stores the raw output
// under cache/properties.txt. Explicit opt-in only: it spawns Gradle, so it
// never runs as part of a normal scan.
func DeepProperties(info Info, dataRoot string) (string, error) {
	if info.Type != "android-gradle" {
		return "", fmt.Errorf("deep scan only supports android-gradle projects")
	}
	if err := storage.EnsureProjectLayout(dataRoot, info.Name); err != nil {
		return "", err
	}
	cmd := exec.Command(build.GradleWrapper(info.Root), ":app:properties")
	cmd.Dir = info.Root
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("gradlew :app:properties: %w", err)
	}
	path := filepath.Join(storage.CacheDir(dataRoot, info.Name), "properties.txt")
	if err := os.WriteFile(path, out, 0o644); err != nil {
		return "", err
	}
	return path, nil
}

// FindCmakeFiles lists CMakeLists.txt files under root, bounded by depth and
// count so large trees stay cheap. It skips VCS/build/output directories.
func FindCmakeFiles(root string) []string {
	var out []string
	_ = filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		rel, _ := filepath.Rel(root, p)
		if d.IsDir() {
			if p != root && skipDirs[d.Name()] {
				return filepath.SkipDir
			}
			if strings.Count(rel, string(filepath.Separator)) >= 4 {
				return filepath.SkipDir
			}
			return nil
		}
		if d.Name() == "CMakeLists.txt" {
			out = append(out, rel)
			if len(out) >= 20 {
				return filepath.SkipDir
			}
		}
		return nil
	})
	sort.Strings(out)
	return out
}
