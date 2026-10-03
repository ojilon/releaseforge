package build

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ojilon/releaseforge/internal/storage"
)

// MaxErrorItems caps structured extraction; overflow is noted in Summary.
const MaxErrorItems = 25

// ErrorItem is one parsed problem with location and a fix hint.
type ErrorItem struct {
	File    string
	Line    int
	Message string
	Rule    string
	Hint    string
}

// Report is the structured outcome attached to a failed Result.
type Report struct {
	Items   []ErrorItem
	Summary string
	LogPath string
}

var (
	reKotlin    = regexp.MustCompile(`^e:\s*(.+?):(\d+)(?::(\d+))?\s*(.*)$`)
	reJavac     = regexp.MustCompile(`^(.+\.java):(\d+):\s*error:\s*(.*)$`)
	reTaskFail  = regexp.MustCompile(`Execution failed for task '([^']+)'`)
	reCmakeErr  = regexp.MustCompile(`CMake Error at ([^:\s]+):(\d+)\s*(.*)$`)
	reNinja     = regexp.MustCompile(`^FAILED:\s*(.*)$`)
	reNinjaStop = regexp.MustCompile(`^ninja:\s*build stopped:(.*)$`)
	reClang     = regexp.MustCompile(`^(.*\.(c|cpp|cc|cxx|h|hpp|kt|java))?:?\s*clang.*error:\s*(.*)$`)
	reApkErr    = regexp.MustCompile(`^\s*ERROR:\s*(.*)$`)
)

func atoi(s string) int {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0
		}
		n = n*10 + int(c-'0')
	}
	return n
}

// ParseErrors extracts structured items from output lines. First rule wins
// per line; Gradle `>` continuations attach to the preceding task failure.
func ParseErrors(lines []string) Report {
	var rep Report
	pending := -1
	taskCont := 0
	flush := func() { pending = -1; taskCont = 0 }
	add := func(it ErrorItem) {
		if len(rep.Items) >= MaxErrorItems {
			return
		}
		rep.Items = append(rep.Items, it)
	}
	for _, raw := range lines {
		l := strings.TrimSpace(raw)
		if l == "" {
			continue
		}
		if m := reTaskFail.FindStringSubmatch(l); m != nil {
			add(ErrorItem{Rule: "gradle", Message: "task '" + m[1] + "' failed",
				Hint: "rerun with --stacktrace for the full trace (build --stacktrace)"})
			pending = len(rep.Items) - 1
			taskCont = 0
			continue
		}
		if pending >= 0 && pending < len(rep.Items) && strings.HasPrefix(l, "> ") && taskCont < 3 {
			rep.Items[pending].Message += "; " + strings.TrimPrefix(l, "> ")
			taskCont++
			continue
		}
		flush()
		switch {
		case strings.HasPrefix(l, "e: "):
			if m := reKotlin.FindStringSubmatch(l); m != nil {
				msg := strings.TrimPrefix(strings.TrimSpace(m[4]), ":")
				hint := "open file:line in the IDE"
				if strings.Contains(msg, "Unresolved reference") {
					hint = "check the import or the Gradle dependency providing it"
				} else if strings.Contains(msg, "Type mismatch") {
					hint = "check the expected vs actual types at the location"
				}
				add(ErrorItem{File: m[1], Line: atoi(m[2]), Message: msg, Rule: "kotlin", Hint: hint})
			}
		case strings.Contains(l, ".java:") && strings.Contains(l, "error:"):
			if m := reJavac.FindStringSubmatch(l); m != nil {
				add(ErrorItem{File: m[1], Line: atoi(m[2]), Message: m[3], Rule: "javac",
					Hint: "open file:line; check types and imports"})
			}
		case strings.HasPrefix(l, "CMake Error at"):
			if m := reCmakeErr.FindStringSubmatch(l); m != nil {
				add(ErrorItem{File: m[1], Line: atoi(m[2]), Message: strings.TrimSpace(m[3]),
					Rule: "cmake", Hint: "open the CMakeLists at file:line"})
			}
		case strings.HasPrefix(l, "FAILED:"):
			if m := reNinja.FindStringSubmatch(l); m != nil {
				add(ErrorItem{Message: m[1], Rule: "ninja",
					Hint: "see the failed compile command above in the full log"})
			}
		case strings.HasPrefix(l, "ninja:"):
			if m := reNinjaStop.FindStringSubmatch(l); m != nil {
				add(ErrorItem{Message: "build stopped:" + m[1], Rule: "ninja",
					Hint: "fix the first error above; later ones are fallout"})
			}
		case strings.Contains(l, "undefined reference"):
			add(ErrorItem{Message: l, Rule: "linker",
				Hint: "a native symbol is missing: check CMake sources and libraries"})
		case strings.Contains(l, "clang") && strings.Contains(l, "error:"):
			file, msg := "", l
			if m := reClang.FindStringSubmatch(l); m != nil && m[1] != "" {
				file, msg = m[1], m[3]
			}
			add(ErrorItem{File: file, Message: msg, Rule: "clang",
				Hint: "read the full error in the log"})
		case strings.HasPrefix(l, "ERROR:"):
			if m := reApkErr.FindStringSubmatch(l); m != nil {
				add(ErrorItem{Message: m[1], Rule: "apksigner",
					Hint: "verify keystore path, alias, and password"})
			}
		}
	}
	// Count total matches for the overflow note (cheap second pass avoided:
	// Summary notes the cap only when hit).
	if len(rep.Items) == MaxErrorItems {
		rep.Summary = fmt.Sprintf("%d+ problems (showing first %d)", MaxErrorItems, MaxErrorItems)
	} else if len(rep.Items) > 0 {
		rep.Summary = fmt.Sprintf("%d problem(s) extracted", len(rep.Items))
	}
	return rep
}

// Format renders report items as `file:line message` + `hint:` lines.
func (r *Report) Format() []string {
	var out []string
	for _, it := range r.Items {
		loc := it.File
		if loc == "" {
			loc = it.Rule
		} else if it.Line > 0 {
			loc = fmt.Sprintf("%s:%d", loc, it.Line)
		}
		out = append(out, "  "+loc+"  "+it.Message)
		if it.Hint != "" {
			out = append(out, "    hint: "+it.Hint)
		}
	}
	return out
}

type xmlTestsuite struct {
	Tests    int           `xml:"tests,attr"`
	Failures int           `xml:"failures,attr"`
	Errors   int           `xml:"errors,attr"`
	Skipped  int           `xml:"skipped,attr"`
	Cases    []xmlTestcase `xml:"testcase"`
}

type xmlTestcase struct {
	Classname string      `xml:"classname,attr"`
	Name      string      `xml:"name,attr"`
	Failure   *xmlOutcome `xml:"failure"`
	Error     *xmlOutcome `xml:"error"`
}

type xmlOutcome struct {
	Message string `xml:"message,attr"`
}

// SummarizeTests parses TEST-*.xml under resultsDir, writes a text summary to
// reports/junit-<stamp>.txt under reportsDir, and returns (summary, path).
// A missing results dir returns ("", "", nil): silently skipped.
func SummarizeTests(resultsDir, reportsDir string) (string, string, error) {
	entries, err := os.ReadDir(resultsDir)
	if err != nil {
		return "", "", nil
	}
	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasPrefix(e.Name(), "TEST-") && strings.HasSuffix(e.Name(), ".xml") {
			files = append(files, filepath.Join(resultsDir, e.Name()))
		}
	}
	if len(files) == 0 {
		return "", "", nil
	}
	var b strings.Builder
	total, fail, skip := 0, 0, 0
	var failed []string
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			continue
		}
		var ts xmlTestsuite
		if err := xml.Unmarshal(data, &ts); err != nil {
			continue
		}
		total += ts.Tests
		fail += ts.Failures + ts.Errors
		skip += ts.Skipped
		for _, c := range ts.Cases {
			if c.Failure != nil || c.Error != nil {
				failed = append(failed, c.Classname+"."+c.Name)
			}
		}
	}
	fmt.Fprintf(&b, "tests: %d, failures: %d, skipped: %d\n", total, fail, skip)
	for _, f := range failed {
		fmt.Fprintf(&b, "  FAILED %s\n", f)
	}
	if err := os.MkdirAll(reportsDir, 0o755); err != nil {
		return "", "", err
	}
	path := storage.ReportPath(reportsDir, "junit")
	if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
		return "", "", err
	}
	return strings.TrimSpace(b.String()), path, nil
}
