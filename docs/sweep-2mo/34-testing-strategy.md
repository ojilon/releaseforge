# 34 — Testing Strategy (unit → hardware acceptance)

## Layers

1. **Unit** (`go test ./...`): parse/commit history (`11`), intent/router
   (`03`,`06`), junit (`08`), porcelain (`09`), errors corpus (`28`),
   version R/W (`26`), merge/freshness (`02`,`25`). Table + golden files.
2. **Fixture trees** (`testdata/`): android-mini, wails-mini, cmake-hello,
   zig-hello, maven-mini, generic-node. Scan/build --dry-run against them in CI.
3. **TUI harness**: headless `tea.Program` with fake `WindowSizeMsg` +
   scripted keys; snapshot `View()` at 70/100/160 cols.
4. **Live hardware** (manual gate before any release claim): real SDK build,
   instrumented test on device, signed APK end-to-end, `run`/`logcat`
   against a phone (see `08-implementation-plan.md` "Live-hardware acceptance").

## Commands

```powershell
go vet ./...; go test ./...; gofmt -l .
go test -race ./internal/build/ ./internal/app/
```

## Read

- Go testing + testdata conventions; Bubble Tea testing examples.
- Your `*_test.go` files as style guide (keep that bar).

## Tasks

- [ ] Corpus + fixtures + TUI snapshots wired into `go test`.
- [ ] `LIVE_HARDWARE.md` checklist filled per release.

## Acceptance

```powershell
go vet ./...; go test ./...  # green, every commit, no exceptions
```
