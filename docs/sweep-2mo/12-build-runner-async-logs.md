# 12 — Build Runner: Async, Cancel, Log Streaming

## Current state

`internal/build/runner.go` `Start()` → `Handle{Lines,Done}`; `app.go:441
waitLine` batches ≤100 lines; `esc` cancels; logs via `rflog.LogPath`.
Timeouts only for instrumented tests (`app.go:701`).

## Hardening checklist

1. Context everywhere: `context.WithCancel` per run; cancel kills process
   group (Windows: `taskkill /T` fallback; document).
2. Timeouts + heartbeats: no-output watchdog (default 10 min, flag).
3. Backpressure: `Lines` buffered (e.g. 1024); slow TUI never blocks build;
   overflow counter surfaced ("+12k lines, tailing").
4. Log files: `<logs>/<kind>-<variant>-<timestamp>.log` + `.exit` sidecar
   (exit code, duration, command line). `Newest/Resolve/Tail/Head` already exist.
5. Env scrubbing: PATH/SDK vars logged (values redacted for secrets).

## Snippet — handle contract

```go
type Options struct {
    Dir string; Env []string; LogPath string
    Project string; Timeout time.Duration; NoOutputTimeout time.Duration
}
type Handle struct {
    Lines <-chan rflog.Line; Done <-chan Result; Cancel func()
}
func Start(program string, args []string, o Options) *Handle
```

Keep `waitLine` batching; add `doneMsg{report}` from `Report.Format()`.

## Read

- Go `os/exec` + `Context`: https://pkg.go.dev/os/exec
- Windows process groups: `CREATE_NEW_PROCESS_GROUP`, `taskkill` docs.

## Tasks

- [ ] No-output watchdog + `.exit` sidecar + overflow counter.
- [ ] Windows kill-tree helper + test (spawns `ping -n` child, cancels).
- [ ] `logs tail -f` follow mode in Logs window.

## Acceptance

```powershell
go test ./internal/build/ -v
# start `build release`, press ctrl+x → process tree dead, `.exit` written
```
