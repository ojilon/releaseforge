# ReleaseForge — project state

> Snapshot: **2026-11-02**, branch `arena/bd3af02e-releaseforge`.
> Tool version reported by the binary: `0.0.2` (see "Version truth" below).
> This document is the honest current state, not a plan. For the design
> rationale of the interface see [04-tui-design.md](04-tui-design.md), for
> the change log see [CORRECTIONS.md](CORRECTIONS.md), for the visual tour
> open [../demo/index.html](../demo/index.html).

## 1. The intent (unchanged)

ReleaseForge is a **local, single-binary control plane for the lifecycle of
any project on disk**: scan a folder, understand it, then test → build →
package → sign → install → draft notes → tag → release. It is a Cobra CLI
plus a Charm Bubble Tea TUI, with a mandatory local data root so heavy
artifacts never clutter the git tree. Ownership, fork status and GitHub owner
are irrelevant; detection is by filesystem markers.

Nothing in this pass changed that intent. No command was removed, no backend
pipeline was replaced, no network service was added, and no data left the
machine. What changed is **how the same work is presented and how reliably
the code behaves**.

## 2. What this pass delivered

| Area | Result |
|------|--------|
| Interface | Single viewport + fixed dashboard → **shell with 6 windows**, tab strip, generated key footer, responsive layout, overlays |
| Charm stack | `bubbles` viewport/textinput/**list/table/spinner/filepicker/help/key**, `lipgloss` layout/frames/tabs, `bubbletea` Elm routing; `huh` stays in the CLI wizard/device picker |
| Navigation | `alt+1..6` / `1..6` (window-focused) / `]` `[`, `window <n>`, `?`, `ctrl+w` focus, `esc` back |
| Discovery | `ctrl+p` fuzzy command palette, `ctrl+r` history search, `tab` completion, `ctrl+o` **terminal folder picker**, `open #n` opens a recent project |
| Help | One verb table (`internal/tui/intent.go`) generates the Help window, completion candidates, palette sources, unknown-command suggestions and `releaseforge commands` |
| Run state | Explicit phase FSM (`idle → building/testing → ok/failed/cancelled`), one `RunState` owner, spinner + elapsed clock, cancel never misreported as failure |
| Focus | Single-consumer key dispatch: typing never scrolls a pane, nav keys never type, per-window scroll/cursor preserved |
| Theme | dark / light / **mono** presets, `NO_COLOR` + `COLORFGBG` + TTY detection, full repaint on switch (including the input) |
| Layout | `tui.Compute(width,height)` breakpoints (tiny/compact/normal/wide/ultra), exact rects, ANSI-aware + wide-rune-safe clamp; no wrapping at 52/70/100/132/160/200 columns |
| Corrections | 12 documented fixes, including a panic, a focus bug, a layout overflow and 4 platform-locked tests; see [CORRECTIONS.md](CORRECTIONS.md) |
| Tests | `go build ./...`, `go vet ./...`, `go test ./...` all green — **136 passing tests** incl. new layout/intent/fuzzy/FSM/snapshot suites |
| Demo | 10 real `View()` frames → self-contained HTML player + raw ANSI captures in `demo/` |

## 3. Architecture now

```
main.go → cmd/ (Cobra)
  ├─ default / tui → internal/app (Bubble Tea shell)
  ├─ init, scan, status, build, test, release, … (scriptable twin)
  └─ commands → prints the generated TUI reference

internal/app/          TUI shell + windows + overlays
  model.go             Model, Context, RunState, theme refresh
  update.go            single-consumer router (keys/mouse/msgs)
  view.go              header · tabs · framed window · bar · help footer
  windows.go           Window interface, shared sizing, themed table/list
  window_status.go     1 Status  – facts, scan cache, recent runs, side card
  window_tests.go      2 Tests   – run history + last test log
  window_git.go        3 Git     – status / diff / log subviews
  window_build.go      4 Console – ring, viewport, spinner, follow policy
  window_logs.go       5 Logs    – list browser + preview
  window_help.go       6 Help    – generated command reference
  overlays.go          palette + folder picker
  phase.go             FSM + Focus enum
  commands.go          dispatch (generated verb table)
  jobs/scripts/scan/…  unchanged backend calls
  demo.go              scripted frame generator (dev only)

internal/tui/          pure presentation, no app state
  theme.go styles.go layout.go keys.go intent.go fuzzy.go

internal/{project,build,git,log,metrics,history,storage,config,android,github}/
                       unchanged domain packages (+2 small read helpers)
tools/demo/            ANSI → HTML demo generator (dev binary)
demo/                  generated HTML player + .ans captures
```

Design rules that still hold: any local directory is a project; scan before
act; CLI and TUI share domain packages; data root is mandatory after `init`;
passwords are never stored.

## 4. Interface reference

| Key | Action |
|-----|--------|
| `1..6` (window focused) / `alt+1..6` | switch window |
| `]` / `[` | next / previous window |
| `ctrl+w` | focus the active window (arrows scroll it) |
| `esc` | back to the bar / close an overlay / cancel a run (compat) |
| `enter` | run the command / activate a palette item |
| `↑ ↓` | command history (bar) / navigate (window) |
| `tab` | complete verb or argument from the generated table |
| `ctrl+p` / `ctrl+r` | command palette / history search |
| `ctrl+o` | folder picker (also `open` with no path) |
| `ctrl+x` | cancel a running build/test/script |
| `ctrl+l` | clear the console |
| `?` / `f1` / `help` | generated Help window |
| `r` | refresh the active window |
| `ctrl+c` / `quit` | leave |

Commands (all previous verbs preserved): `open|scan|rescan`, `recent`,
`status`, `version`, `build`, `test`, `bat`, `sign`, `install`, `devices`,
`notes`, `git`, `logs`, `ls|dir`, `pwd`, `cat`, `theme`, `window`, `help`,
`clear`, `quit`. `release`, `doctor`, `logcat` and the full release pipeline
remain CLI-side, as designed. Bare digits stay typeable in the bar so
`open 2` (recent #2) and `version --set 1.2.3` work.

## 5. Verification

```bash
go build ./...          # ok
go vet ./...            # ok
go test ./... -count=1  # ok, 136 tests across all packages
go run ./tools/demo -out demo   # regenerates demo/index.html + frames/
```

New/rewritten test suites: `internal/tui` (intent/tokenizer, completion,
fuzzy, layout breakpoints, clamp/frame, wide-rune truncation, theme
detection, generated help), `internal/app` (exact-frame snapshots at
52/70/100/160/200 cols, single-consumer dispatch, scroll preservation,
palette ranking, cancel→cancelled, theme no-arg, quoted paths, recent
index, async drain). Baseline `go test ./...` had 6 failures on Linux;
all are fixed and the suite is now platform-neutral.

## 6. Demo

`demo/index.html` is a self-contained player (no CDN, no network). Each of
its 10 frames is the real `Model.View()` output with ANSI colors translated
to HTML spans — not a mock-up. It covers Status, generated Help, Git log,
Logs browser, a live build with spinner, the command palette, the folder
picker, compact/tiny terminals and the light theme. `demo/frames/*.ans`
hold the raw captures for terminal playback. See [demo/README.md](../demo/README.md).

## 7. Known limitations (honest gaps)

- **Tool version is `0.0.2`** while some docs/examples mention `0.0.3` /
  `0.0.4`. The `VERSION` file is intentionally gitignored and machine-local
  (see `docs/sweep-2mo/26`); `internal/version/version.go` and build-time
  `-ldflags` are the truth. This pass did not bump a release.
- **No native OS folder dialog.** The vision's Windows dialog is still not
  built; `ctrl+o` is a terminal `bubbles/filepicker` instead. Typed paths and
  `open #n` remain available.
- **Windows-specific paths are not exercised on Windows here.** The adb and
  apksigner tests are now hermetic on Linux/macOS; real device/SDK flows
  still need Windows + Android hardware acceptance (docs/sweep-2mo/34).
- **No `huh` form inside the TUI** (theme select is a command; `huh` remains
  in `init` and device picking), no `glamour` markdown rendering, no
  services layer extracted from `do*` methods, no `go-git`, no glamour-style
  release-note preview. These remain sweep items 03/42/43.
- **No CI configuration** in the repo; the verification above was run
  locally with Go 1.23.9.
- TUI `release` / `doctor` / `logcat` verbs are still CLI-only by design —
  they change the world, the bar stays inspection + build/test oriented.
- Theme choice is session-only (not persisted in `global.json`).

## 8. Version truth / next steps

Priority order from the sweep archive, now that the shell exists:

1. Extract `do*` handlers into `internal/app/services` (zero tea imports)
   and share them with `cmd/` (sweep 03, 05).
2. Fill the Tests window with real JUnit/Go-test parsing (sweep 08) and the
   Git window with a diff hunk viewer (sweep 09).
3. Persist UI preferences (theme, density, mouse) in `global.json` behind a
   schema migration (sweep 25, 59).
4. Optional `glamour` markdown for release notes (sweep 43) and a `huh`
   settings form (sweep 42).
5. Hardware acceptance on Windows: adb devices, apksigner, Gradle assemble,
   installer (sweep 34, 36, 51).

The sweep docs in `docs/sweep-2mo/` are planning records; where this pass
implemented one of their tasks, a `> Delta` note at the top of that file
points here. They are not all done.
