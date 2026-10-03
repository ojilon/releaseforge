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
passes e.g. 30m for instrumented). `Cancel` kills the process
(`cmd.Process.Kill`; no SIGTERM dance on Windows). `Stream` becomes the hub:
one writer goroutine, subscribers get a bounded channel (256, drop-oldest —
never block the writer). `Lines()` is replaced for live use by a small
ring buffer type (`type Ring struct …​`, cap 2000, `Append/Snapshot`) living
in `internal/log`; file persistence unchanged.

TUI (wiring only — full architecture is doc 13): `exec` returns a `tea.Cmd`
that drains `Handle.Lines` into `lineMsg` and posts `doneMsg` at the end;
`esc` during a run calls `Cancel`. CLI path (`runStep`) keeps calling
blocking `Run` — no behavior change there.

## Files to touch

- `internal/build/runner.go` (`Start`, `Handle`, `Timeout`, kill path)
- `internal/log/stream.go` (drop-oldest subscribers, `Ring` type)
- `internal/app/app.go` (minimal: send the existing `logMsg`/`doneMsg` for
  real; `esc` cancels) — full message/UX design stays in doc 13
- `cmd/release.go:runStep` unchanged (still blocking `Run`)

## Steps

1. Add `Ring` + tests.
2. Change `Stream` fan-out to drop-oldest; test with a slow subscriber.
3. Extract pipe-reading from `Run` into `Start`; `Run` = `Start` + drain.
4. Add `Cancel` (kill) + `Timeout` (timer → Cancel); test both with a
   sleeper process (`go run` a temp program or `ping -n`).
5. Wire TUI `doBuild/doTest` to `Start`: lines → `logMsg`, end → `doneMsg`,
   `esc` → `Cancel`. Keep the 200-line replay cap for now.

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
