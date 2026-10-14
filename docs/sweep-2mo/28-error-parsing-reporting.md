# 28 — Error Parsing & Reporting (Gradle/CMake/Go/NDK)

## Current state

`internal/build/errors.go` (+`errors_test.go`): regexes for Gradle/CMake/NDK
+ JUnit summaries → `Report{Items,Summary}`. TUI shows summary (fail-red) +
items + dimmed `hint:` lines (`tui.RenderReport`, `update.go:onDone`);
CLI prints summary + items (`cmd/build.go:printReport`, `release.go:runStep`).

## Corpus-driven plan

1. Collect 50+ real failure logs (Gradle OOM, license-not-accepted,
   NDK missing, CMake configure fail, `connectedDebug` no-device,
   `go vet` fails) under `testdata/failures/<tool>/`.
2. Parser returns `[]Finding{Tool, File, Line, Code, Hint, Fix}`; `Format()`
   prints top-5 + `see logs show <name>`.
3. Each finding has a `Fix` suggestion (`doctor` command or doc link).
4. Fuzz line-splitting on `\r\n`, ANSI escapes, long lines (>8KB).

Strip ANSI before parse:

```go
var ansi = regexp.MustCompile("\x1b\\[[0-9;]*m")
clean := ansi.ReplaceAllString(line, "")
```

## Read

- Gradle error output + `--stacktrace`; CMake `Error at …` format; NDK install errors.
- Your `scripts/` failure modes (missing SDK → message text to preserve).

## Tasks

- [ ] Corpus + golden `Report.Format()` snapshots.
- [ ] `Fix` hints wired to `doctor` checks.
- [ ] Bench parser on 100MB log (Odin candidate — see `18` if slow).

## Acceptance

```powershell
go test ./internal/build/ -run TestErrors -v
```
