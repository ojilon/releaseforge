# 40 — TUI State Machine (idle/building/testing/failed/ok done right)

## Current state

`app.go:72 phase string` (`idle|building|testing|failed|ok`), set inline in
`doBuild/doTest/doneMsg`. No transition table; `footer()` branches on it.
Races: `esc` cancel vs `doneMsg` arrival; double-run guarded by `handle != nil`.

## Target — explicit FSM

```
idle → building|testing → (cancelled|failed|ok) → idle
       ^_______________| (rerun from failed/ok)
```

- `phase.go`: `type Phase int` + `Allow(from, to) bool` table + tests.
  Illegal transitions are no-ops + log (never silent state corruption).
- `running` + `handle` merge into `Run{Kind, Variant, Started, Handle}`;
  single nil-check replaces two fields.
- `doneMsg` after cancel → `cancelled` (not `failed`); footer shows
  `rerun (r) · logs last` on failed, `esc` only pops focus (see `06`).

## Snippet

```go
type Phase int
const (Idle Phase = iota; Building; Testing; Failed; OK; Cancelled)
var trans = map[Phase][]Phase{
    Idle: {Building, Testing},
    Building: {OK, Failed, Cancelled},
    Testing: {OK, Failed, Cancelled},
    Failed: {Building, Testing, Idle},
    OK: {Building, Testing, Idle},
}
```

## Read

- Bubble Tea async patterns (don't mutate model in Cmds); your `app_test.go`.

## Tasks

- [ ] `phase.go` + transition tests incl. cancel-vs-done race.
- [ ] Merge `running/handle/pendingRun` into `Run` struct; update footer/help.

## Acceptance

```powershell
go test ./internal/app/ -run TestPhase -v
# start build, cancel, doneMsg arrives late → phase cancelled, not failed
```
