# 27 — Logging & Metrics (ring, files, last-build)

## Current state

`internal/log/` (`Ring` cap 2000, `LogPath/List/Newest/Resolve/Tail/Head`,
index), `internal/metrics/` (`Append`, `LastBuildLine` in status).
`update.go:onDone` records build/test outcome + duration (script runs count
as `kind="script"`); failures render via `tui.RenderReport`.

## Targets

- Ring stays in-memory cap; files are complete. `.exit` sidecar per run (`12`).
- `logs show|last|tail -n --follow` parity CLI↔TUI; `tail -f` in Logs window.
- Metrics rows: `{project,kind,variant,success,duration_ms,tool,git_rev,abis}`.
  `status` shows last build; Tests window aggregates (`08`).
- Redaction: password/token patterns scrubbed before persist (test with fixtures).

## Snippet — scrub

```go
var secrets = regexp.MustCompile(`(?i)(--ks-pass[ =]\S+|password=\S+)`)
func Scrub(s string) string { return secrets.ReplaceAllString(s, "$1=***") }
```

## Read

- Your `internal/log/*_test.go` patterns; Go `regexp`, `bufio.Scanner` limits.

## Tasks

- [ ] `.exit` sidecar + `tail -f` + scrub tests.
- [ ] Metrics CSV/JSONL query (`metrics show --last 20`).

## Acceptance

```powershell
go test ./internal/log/ ./internal/metrics/ -v
releaseforge logs tail --lines 20 last
```
