# 38 — Logcat Viewer (filters, buffers, crash hunting)

## Current state

`cmd/logcat.go` exists; adb logcat invoked ad hoc. No filters, no ring,
no crash highlighting.

## Target — Logs window gets a logcat pane

- `logcat [--device ID] [--pid PKG] [--level I|W|E] [--tag TAG] [-c clear]`:
  wraps `adb logcat -v threadtime`, streams into runner (`12`) ring.
- Crash detector: lines matching `FATAL EXCEPTION`, `AndroidRuntime`,
  `Native crash`, `tombstone` get red pill + auto-jump key `J`.
- Filters as flags AND interactive keys (`e` errors-only, `a` all, `/` regex).
- Persist: `logcat-<ts>.log` under project logs + symbolize hint
  (`ndk-stack` path when native crash).

## Snippet

```go
args := []string{"logcat", "-v", "threadtime"}
if pid != "" { args = append(args, "--pid="+pid) }
if level != "" { args = append(args, "*:"+level) }
// stream via build.Start("adb", args, ...) into Logs window pane
```

Crash regexes (seed corpus in `testdata/logcat/`):

```go
var crashRe = regexp.MustCompile(`FATAL EXCEPTION|AndroidRuntime|Fatal signal|Build fingerprint`)
```

## Read

- logcat: https://developer.android.com/tools/logcat ; `ndk-stack` docs.
- Your `cmd/logcat.go` + `internal/log/ring.go` reuse.

## Tasks

- [ ] logcat runner pane + level/tag filters + crash jump.
- [ ] `ndk-stack -sym` hint on native crashes; golden logcat fixtures.

## Acceptance

```powershell
releaseforge logcat --level E  # streams; FATAL lines highlighted, J jumps
```
