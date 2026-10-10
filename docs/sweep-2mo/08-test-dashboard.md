# 08 — Test Dashboard (unit + instrumented)

## Current state

`doTest:683` starts `TestArgs(kind)` async; JUnit summaries exist in
`internal/build/errors.go`; `test [kind]` streams to the single viewport.
No pass/fail board, no per-class view, no `--device` precheck surfacing
beyond errors.

## Target

Tests window (`2`) with:

- Header: project + kind + elapsed + `esc/ctrl+x` cancel hint.
- Table: suite/class → pass/fail/skip counts, duration; failed rows
  expandable to first error + log excerpt path.
- Footer: `test unit|instrumented|all [--device ID]`, rerun-failed key `r`.
- Persist: JUnit XML copied under `logs/`; `metrics.Append` already records
  outcome (`app.go:211`) — add per-suite rows.

## Snippet — JUnit row model

```go
type SuiteRow struct {
    Name string; Pass, Fail, Skip, Total int; Dur time.Duration
    LogExcerpt string // path:line into persisted log
}
func RowsFromJUnit(xmlPath string) ([]SuiteRow, error) { /* encoding/xml */ }
```

Gradle kinds (see `13`): `:app:testDebugUnitTest`, `:app:connectedDebugAndroidTest`.
Go: `go test ./... -json` streaming. Instrumented requires `adb doctor`
pass + device picker (huh) — block with clear message otherwise.

## Read

- JUnit XML schema; Gradle test task docs: https://docs.gradle.org/current/userguide/java_testing.html
- `go test -json`: https://pkg.go.dev/cmd/go#hdr-Test_packages
- Your `internal/build/errors_test.go` + `async_test.go` patterns.

## Tasks

- [ ] `internal/build/junit.go` parser + golden XML fixtures.
- [ ] Tests window table (bubbles `table`) + expand-failure pane.
- [ ] `test --device` precheck wired to `adb doctor`; `r` reruns failures.

## Acceptance

```powershell
go test ./internal/build/ -v
# TUI: `test unit` → Tests window shows suites; `r` reruns; `logs last` has JUnit XML
```
