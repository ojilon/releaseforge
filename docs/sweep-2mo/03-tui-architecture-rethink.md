# 03 — TUI Architecture Rethink (kill the god-model)

## Current state

`internal/app/app.go:56 Model` owns project + input + viewport + ring +
history + phase + async handle. `Update:172` + `exec:383` + `View:288`
in one file (~833 lines). Works for single viewport + command bar, but:

## Problems

1. No windows/tabs — everything appends to one ring; test output drowns git info.
2. Focus bugs: input + viewport both `Update()` every key (`:281`); esc/cancel
   races with `doneMsg`; `render()` follow-mode is the only scroll policy.
3. Business logic in view (`doScan/doVersion/doLogs/doNotes`) — untestable
   without a TUI harness; CLI duplicates paths.
4. Styling inline (`tui.*Style` per call) — no layout manager, narrow-terminal
   fallback is a branch in `View`, not a layout pass.

## Target — Elm-style window manager

```
AppModel
 ├─ Router (active window id, focus stack, key dispatch)
 ├─ Windows: Status | Tests | Diffs | Build | Logs | Help (each a tea.Model)
 ├─ Services (no UI): scanner, runner, gitSvc, storeSvc — injected interfaces
 └─ Shell: header + window chrome + command bar + footer (only chrome here)
```

- Each window: `Init/Update/View + Title() + Keys()`. Router forwards
  `WindowSizeMsg`, `linesMsg/doneMsg` only to owner window.
- Command bar parses into `Intent{Verb,Args}` (pure, tested); router routes
  intent → window or service. `exec()` string-switch disappears.
- Services return `(Report, error)`; windows render reports. Same services
  used by `cmd/`.

## Snippet — router skeleton

```go
type Window interface {
    tea.Model
    Title() string
    Keys() []key.Binding
}
type Router struct {
    windows map[string]Window
    active  string
    stack   []string // focus history for Esc-back
}
func (r *Router) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    if k, ok := msg.(tea.KeyMsg); ok {
        if b, ok := globalKeys[k.String()]; ok { return r.activate(b) }
        // else forward to active window only
    }
    w := r.windows[r.active]
    next, cmd := w.Update(msg)
    r.windows[r.active] = next.(Window)
    return r, cmd
}
```

Keys: `1..6` switch windows, `tab` cycles, `esc` pops focus (not cancel —
cancel is `ctrl+x` on Build/Tests only; see `06`).

## Read

- Bubble Tea docs: https://github.com/charmbracelet/bubbletea (Elm arch,
  commands vs messages). Bubbles: https://github.com/charmbracelet/bubbles
- `docs/04-tui-design.md`, `internal/app/app_test.go` (keep parse tests green).
- Lipgloss layout: https://github.com/charmbracelet/lipgloss

## Tasks

- [ ] Extract `parse()` → `internal/tui/intent.go` with table tests.
- [ ] Extract services (`doScan/doVersion/doLogs/doNotes`) → `internal/app/services.go` (no tea imports).
- [ ] Introduce `Router` + 2 windows (Status, Logs) behind flag; single-viewport stays default until parity.
- [ ] Move follow/scroll policy into viewport wrapper with tests.

## Acceptance

```powershell
go vet ./...; go test ./internal/app/ ./internal/tui/ -v
releaseforge tui  # help, status, scan, logs identical to before
```
