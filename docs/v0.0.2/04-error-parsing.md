# v0.0.2 doc 04: Structured error parsing

## Goal

Turn red log dumps into actionable errors: file, line, message, and a fix
hint — plus a JUnit summary for tests. This is what makes Android support
"solid" instead of merely present.

## Current state

`ExtractErrors` (`internal/build/runner.go:39-65`) substring-matches 10
patterns, returns ≤15 raw lines, no file/line split, no hints. Nothing reads
`app/build/test-results/*.xml`; `reports/` is created and never used.
Matchers are Gradle-flavored (`e: `, `Execution failed`, `What went wrong`)
with no CMake/NDK/javac/apksigner coverage.

## Design

New `internal/build/errors.go` (stdlib `regexp`/`encoding/xml` only):

    type ErrorItem struct { File string; Line int; Message, Rule, Hint string }
    type Report struct { Items []ErrorItem; Summary string; LogPath string }
    func ParseErrors(lines []string) Report

Rules, first match wins per line, each with a static hint:

- Kotlin/Java compile: `^e: (.+?):(\d+)(?::(\d+))?\s*(.*)` → hint by keyword
  (e.g. `Unresolved reference` → "check import / dependency").
- javac: `^(.+\.java):(\d+): error: (.*)`.
- Gradle failure block: `Execution failed for task ':(.*)'` + following
  `> (.*)` lines attached as message.
- CMake/NDK: `CMake Error at ([^:]+):(\d+)`, `ninja: build stopped`,
  `FAILED: .*\.ninja`, `ld: error:`, `clang.*error:`.
- apksigner/zipalign: `ERROR: (.*)` after an apksigner invocation.
- Cap: 25 items; overflow noted in `Summary` (`+N more, see log`).

JUnit: `SummarizeTests(resultsDir)` parses `TEST-*.xml` (`testsuite`
`tests/failures/errors/skipped` attrs + failed `testcase` names) and writes
`reports/junit-<stamp>.txt` (+ keeps the parsed summary in `Report`). Absent
dir = skip silently (Go projects have no XML).

CLI prints the report after failure (replacing the raw `!` dump, same shape:
`log:` path first, then items as `file:line message` + `hint:`), exit code
unchanged. `Result` gains `Report *Report` (nil when clean) — additive only.

## Files to touch

- New `internal/build/errors.go` (+ thorough test with golden strings).
- `internal/build/runner.go` (attach `Report` on failure; keep `Errors`).
- `cmd/build.go`, `cmd/test.go`, `cmd/release.go:runStep` (print report).
- `internal/storage/storage.go` (helper `ReportPath` next to `LogPath`
  pattern; `reports/` already created).

## Steps

1. Implement matchers + hints with golden tests (real Gradle/CMake snippets
   as fixtures in `testdata/`).
2. Implement JUnit parse + summary writer; test with fixture XML.
3. Wire `Report` into `Result`; update the three CLI print sites.
4. TUI: print the same `file:line` items into the viewport (layout polish
   stays in doc 14).

## Tests to add

- Golden tests per rule (kotlin, javac, gradle task block, cmake, ninja,
  apksigner) + non-matching lines yield no items.
- Cap behavior (30 error lines → 25 items + overflow note).
- JUnit fixture: counts + failed-case names + missing-dir skip.

## Acceptance checks

- `go vet ./...`, `go test ./...` pass.
- Feed a fixture failing log through a test `build` (temp android-like
  project with a fake failing script) and observe `path:line` + hint output.
- `reports/` gains a junit summary after an android unit-test run (or a
  documented skip when no XML exists).

## Out of scope

Fix suggestions beyond static hints (no AI), historical failure stats (15),
TUI panels (14). No Gradle spawned by tests — fixtures only (Celeron rule).
