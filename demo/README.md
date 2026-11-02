# ReleaseForge TUI demo

Open **`index.html`** in a browser (or serve this folder) to step through a
scripted tour of the real interface. No build, no network, no CDN: the file
is self-contained.

## What you are looking at

Every frame is the actual `app.Model.View()` string, rendered with
`lipgloss.SetColorProfile(termenv.TrueColor)` and translated from ANSI SGR
codes to HTML spans by `tools/demo`. It is a recording, not a mock-up — if
the layout changes, regenerate and the demo changes with it.

| # | Frame | Shows |
|---|-------|-------|
| 01 | Status | project facts, cached scan, git state, recent-runs table, wide side card |
| 02 | Generated help | `help` opens the reference generated from the verb table |
| 03 | Git | commit table (log subview), `s`/`d`/`l` switch status/diff/log |
| 04 | Logs | `bubbles/list` browser + scrollable preview |
| 05 | Live build | Console with `bubbles/spinner`, elapsed time, streamed output, follow mode |
| 06 | Command palette | `ctrl+p` fuzzy search over commands, projects, logs, history |
| 07 | Folder picker | `ctrl+o` / bare `open` — `bubbles/filepicker` in directory mode |
| 08 | Compact terminal | 70 × 20: numeric tabs, single-line header, no side card |
| 09 | Tiny terminal | 52 × 14: chrome collapses, nothing wraps |
| 10 | Light theme | `theme light` repaints chrome, tables and the input |

`frames/*.ans` are the raw ANSI captures. In a terminal that understands
UTF-8 and 24-bit color:

```bash
cat frames/01-status.ans
```

## Player controls

- `←` / `→` or the Prev / Next buttons to step
- `space` or Autoplay to advance every 4.5 s
- Number buttons jump to a frame
- The footer of each frame lists the real app's keybindings

## Regenerate

From the repo root:

```bash
go run ./tools/demo -out demo
```

The generator (`internal/app/demo.go` + `tools/demo/main.go`) creates a
throwaway project with git history, a scan cache, metrics records and two
logs under a temp directory, drives the real `Model` through the scripted
tour, and deletes the temp tree. No user data is touched.

Regenerate after any change to the layout, theme, verb table or window
rendering. Frames are checked in so the demo is viewable without a Go
toolchain.
