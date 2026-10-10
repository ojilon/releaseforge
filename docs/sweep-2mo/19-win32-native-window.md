# 19 — NEAR-COMPLETE: Own Native Window (Win32 API path)

> Horizon: weeks 7–8+. The TUI stays default; a native window becomes an
> OPTION. Provide the Win32 path + Odin alternative. Do not start early.

## Why + non-goals

- TUI is limited: no rich diff colors, no drag-drop APK install, no taskbar
  progress. A native window gives window chrome, menus, file-drop, dialogs.
- Non-goal: replacing TUI/CLI. `tui` remains primary; `window` is additive.
- Two implementations, one interface: **Go+Win32** (default) and
  **Odin+Win32** (optional spike). Pick Go unless Odin proves simpler.

## Design — `rf window` opens native shell hosting same services

```
cmd/window.go → internal/window/
  Window{ router (same Intent router as 03), services }
  backend: win32.go (RegisterClass/CreateWindow/MessageLoop) OR wails/webview fallback
```

- Phase 1 (no new deps): `internal/window/win32.go` via `golang.org/x/sys/windows`
  — `RegisterClassEx`, `CreateWindowEx`, `ShowWindow`, message pump in a
  goroutine; draw with GDI text + listbox (ugly but proves plumbing).
- Phase 2: real UI via **Wails v2** (already a project type — see `20`) OR
  `walk`/`lxn/walk` bindings; reuse windows `07` as web views.
- Phase 3 (optional): Odin window proc (`native/odin/window.odin`) loaded
  as DLL; Go passes intents over a C ABI. Only if Phase 1/2 stalls on perf.

File-drop install: `WM_DROPFILES` → `DragQueryFile` → `adb install`.
Taskbar progress: `ITaskbarList3::SetProgressValue` during builds.

## Snippet — minimal Win32 skeleton (Go)

```go
//go:build windows
package window
import "golang.org/x/sys/windows"
var cls = windows.UTF16PtrFromString("ReleaseForgeWnd")
func Open(title string) error {
    // RegisterClassEx + CreateWindowEx + message loop
    // WM_DROPFILES → DragAcceptFiles(hwnd, TRUE)
    return pump()
}
```

Odin alternative sketch:

```odin
// native/odin/window.odin
package main
import win "core:sys/windows"
wnd_proc :: proc "stdcall" (hwnd: win.HWND, msg: win.UINT, w, l: win.UINT_PTR) -> win.LRESULT { … }
```

## Read

- MS Win32: `RegisterClassEx`, `CreateWindowEx`, `WM_DROPFILES`, `ITaskbarList3`:
  https://learn.microsoft.com/windows/win32/api/
- `golang.org/x/sys/windows` examples; `lxn/walk`: https://github.com/lxn/walk
- Wails: https://wails.io/docs/introduction (you already detect `wails.json`).
- Odin `core:sys/windows`: https://odin-lang.org/docs/overview/ + vendor examples.

## Tasks

- [ ] `internal/window/` iface + `rf window --backend win32-gdi` spike (empty window + drop files logged).
- [ ] Wire Status/Logs windows into native shell (read-only first).
- [ ] Decision: Wails vs walk for Phase 2; Odin spike only after.

## Acceptance

```powershell
rf window --backend win32-gdi  # empty native window opens, drops log paths
```
