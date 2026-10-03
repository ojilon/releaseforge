# v0.0.2 doc 03: Async process and log stream

## Goal

Run builds/tests without freezing the TUI, stream lines live, allow cancel,
and bound memory — the foundation docs 04/05/13/14 build on.

## Current state

- `build.Run` (`internal/build/runner.go:68`) blocks until process exit; the
  TUI calls it synchronously in the Enter handler (`app.go:440,491`), so the
  whole UI freezes during Gradle.
- `log.Stream` (`internal/log/stream.go`) supports `Subscribe` channels, but
  nothing subscribes; `build.Run` only file-writes + `OnLine`.
- TUI replays at most 200 lines post-hoc (`app.go:445-453,495-503`); dead
  `logMsg/doneMsg` types show the intended design was never finished.
- No cancellation (Ctrl+C quits the app), no timeouts. Celeron + Gradle =
  a stuck UI is a real risk, so cancel must be first-class.

## Design

Extend `internal/build` (stdlib only, no new deps):

    type Handle struct { Lines <-chan Line; Done <-chan Result; Cancel context.CancelFunc }
    func Start(name string, args []string, opts Options) *Handle  // goroutines, existing pipe code moves here
    func Run(...) Result  // kept as blocking wrapper over Start (CLI keeps working)

`Options` gains `Timeout time.Duration` (0 = none; CLI default none, TUI
passes 30m for instrumented/all). `Cancel` kills the process via
`exec.CommandContext` (no SIGTERM dance on Windows). Lines travel on a
dedicated buffered channel on the `Handle` — `Stream` stays a file+memory
writer only (no subscription machinery; doc 01 removed it and it stays
removed). Timeout/cancel are reported as headline errors (`timed out after
…​` / `cancelled`) prepended to `Result.Errors`.

TUI (wiring only — full architecture is doc 13): `exec` returns a `tea.Cmd`
that drains `Handle.Lines` into `lineMsg` and posts `doneMsg` at the end;
`esc` during a run calls `Cancel`. CLI path (`runStep`) keeps calling
blocking `Run` — no behavior change there.

## Files to touch

- `internal/build/runner.go` (`Start`, `Handle`, `Timeout`, kill path)
- `internal/log/stream.go` (add `Ring` type; `Stream` untouched otherwise)
- `internal/app/app.go` (`exec` returns `(bool, tea.Cmd)`; `doBuild/doTest`
  spawn + drain chain; `esc` cancels; second-run guard)
- `cmd/release.go:runStep` unchanged (still blocking `Run`)

## Steps

1. Add `Ring` (cap 2000, `Append/Snapshot/Len`) to `internal/log` + tests.
   (TUI adoption of `Ring` stays in doc 13; here it is specified + tested.)
2. Split `Run` into `Start` (goroutine, `Handle{Lines,Done,Cancel}`) +
   blocking `Run` wrapper that drains `Lines` into `OnLine` (CLI behavior
   unchanged, byte for byte).
3. Add `Cancel` (via `CommandContext`) + `Timeout` (timer → cancel, reported
   as `timed out after …​`) + tests with a sleeper process.
4. Wire TUI `doBuild/doTest` to `Start`: immediate `started …​ (esc cancels)`
   line, then a chained one-line-at-a-time drain (`lineMsg` → existing
   append path, re-armed each message; `doneMsg` at close, reusing the
   existing handler). No post-hoc replay — lines are already on screen.
   `esc` cancels the active handle; starting a second build/test while one
   runs is refused with a hint.

## Tests to add

- Ring overwrite order + snapshot bounds.
- `Start` streams lines before exit; `Cancel` kills within ~1s; `Timeout`
  fires; `Run` behavior unchanged (existing tests keep passing).
- Race run: `go test -race ./internal/build/ ./internal/log/`.

## Acceptance checks

- `go vet ./...`, `go test ./...`, plus `-race` on the two packages.
- TUI `test` streams lines while running (visible growth, not a frozen
  screen + dump); `esc` aborts and reports `cancelled`.
- CLI `test`/`build` output byte-identical shape to before.

## Out of scope

Error parsing (04), log listing UI (05), tabs/panels (14), progress bars.
No Gradle spawned by this doc itself.
