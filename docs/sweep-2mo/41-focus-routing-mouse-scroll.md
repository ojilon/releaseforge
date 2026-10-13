# 41 — Focus Routing, Mouse, Scroll (no more focus bugs)

## Current state

`Update:281` forwards EVERY msg to input + viewport; `render()` follow-mode
only scroll policy; mouse wheel on (`WithMouseCellMotion`), click-to-focus absent.

## Target

- Router (`03`) owns focus: `Focus{Window, Pane, Bar}`. Keys go to exactly
  ONE consumer: bar (typing) vs window (nav). `tab`/`esc` move focus;
  letters never leak into viewport scroll while typing.
- Scroll policies per window: `Follow` (build/logs tail), `Pin` (diffs keep
  hunk), `Free` (help). Store `viewport.YOffset` per window on switch (`07`).
- Mouse: wheel scrolls active pane; click focuses pane; `alt+wheel` horizontal
  for wide diffs. All behind `MOUSE=1` default-on, flag-off for broken terminals.

## Snippet — single-consumer dispatch

```go
if m.focus == FocusBar {
    m.input, cmd = m.input.Update(msg) // window does NOT see keys
    return m, cmd
}
w := r.windows[r.active]
next, cmd := w.Update(msg) // bar does NOT see keys
```

## Read

- Bubbles viewport mouse + `HighPerformanceRendering`; Lipgloss `Place`.
- Windows coninput: https://github.com/erikgeiser/coninput (already indirect dep).

## Tasks

- [ ] Focus enum + dispatch tests (typed keys never scroll; nav keys never type).
- [ ] Per-window offset preservation + scroll-policy unit tests.

## Acceptance

```powershell
go test ./internal/app/ -run 'TestFocus|TestScroll' -v
# TUI: type while build streams → input intact, viewport follows only if at bottom
```
