# v0.0.3 design note — rethinking the TUI without changing the intent

> Implemented 2026-11-02. Companion to [../04-tui-design.md](../04-tui-design.md)
> (the new spec), [../STATE.md](../STATE.md) (state) and
> [../CORRECTIONS.md](../CORRECTIONS.md) (fixes). This file records the
> decisions and trade-offs so the next pass does not have to re-litigate them.

## Brief

"Open choices, but maintain the sole intent of the app: make and document
corrections, test, make a demo, document the state of the project, rethink
the interface, make good use of Charm / lipgloss / bubbles. It does not need
to stay rigid; improve how it is done."

Interpretation: the *what* is fixed — a local, single-binary control plane
for scan → test → build → sign → install → notes → release over any folder
on disk, with a persistent command bar. The *how* was open. The repo's own
`docs/sweep-2mo/` archive (03, 04, 06, 07, 40, 41, 44) already described
the desired direction, so this pass implements a coherent slice of it
instead of inventing a different product.

## Decisions

### 1. Shell + windows, not a bigger viewport

The old UI had one transcript plus a fixed 34-column dashboard. Build output
drowned git state, and there was no place to put a table, a spinner or a
browser. The new `Model` is a shell (header, tabs, framed body, command bar,
key footer) that hosts six independent `Window` components. Each owns its
viewport/cursor; switching preserves scroll. This is sweep 03 + 07.

**Trade-off:** more files (`window_*.go`) but each is small and testable, and
no component reaches back into `Model` — they receive an immutable
`Context` snapshot.

### 2. One consumer per key

Sweep 41's focus bug was real: `Update` sent every key to both the input and
the viewport. The router now intercepts global shortcuts, then dispatches to
exactly one of bar / active window / palette / picker. Mouse wheel still
reaches the active window.

**Decision:** bare `1..6` switch windows *only when the window owns focus*;
`alt+1..6` works everywhere. Sweep 07 wanted plain digits globally, but
that makes `open 2`, `build release 0.0.4` and `version --set 1.2.3`
impossible to type — the command bar is the app's primary instrument and
wins. `]` / `[` and `window <n>` are additional paths.

### 3. A folder picker, in the terminal

The vision asked for a native Windows folder dialog; `docs/04` recorded it as
deliberately unbuilt. `bubbles/filepicker` with `DirAllowed=true` gives the
same capability in-terminal, cross-platform, with typed paths preserved as a
fallback. `open` with no path now opens it instead of doing nothing useful.

### 4. One verb table, many surfaces

`internal/tui/intent.go:Verbs` generates the Help window, completion, palette
candidates, suggestions and the CLI `releaseforge commands`. A new verb can
no longer leave help stale. The CLI twin is a print, not a second grammar.

### 5. Generated geometry

`tui.Compute(width,height)` centralizes breakpoints (tiny/compact/normal/
wide/ultra). `Clamp` (ANSI-aware, wide-rune safe) + `Frame` force exact
rects. This kills the ad-hoc narrow-terminal branch and the dashboard width
mismatch. Snapshot tests assert exact width/height at five sizes.

### 6. Explicit run state

Sweep 40's FSM replaces the string `phase` and five scattered fields.
`cancelled` is a first-class outcome; `ctrl+x` is the explicit cancel while
`esc` keeps a backwards-compatible cancel as a last resort. A `RunState`
owns the handle and labels.

### 7. Charm, properly

Used: `textinput`, `viewport`, `table`, `list`, `filepicker`, `spinner`,
`key`, `help` from bubbles; lipgloss frames/tabs/chips/`Join*`/`Place`;
bubbletea `tea.Tick`, `tea.Batch`, `WithAltScreen`, `WithMouseCellMotion`.
`huh` remains the CLI wizard/device picker and is **not** embedded as a TUI
form: a form needs its own focus stack, and the command bar already covers
those settings. `glamour` was not added (no markdown surface yet).
`bubbles/progress` was left unwired because Gradle/Go expose no trustworthy
percentage — a fake bar would be worse than a spinner.

### 8. Theme as a real subsystem

Dark / light / mono; `DetectTheme` honors `RELEASEFORGE_THEME`, `NO_COLOR`,
`TERM=dumb`, `COLORFGBG` and TTY background; `SetTheme` repaints *all*
derived styles plus both text inputs. Mono emits no color at all.

## What was deliberately left alone

- All backend packages (`project`, `build`, `git`, `log`, `metrics`,
  `android`, `github`) except two small read helpers (`metrics.Recent`,
  `git.StatusPorcelain`/`DiffStat`).
- All Cobra commands and the release pipeline.
- The rule that TUI `release` / `doctor` / `logcat` are CLI-only.
- Passwords never stored; data-root layout unchanged.
- Machine-local `VERSION` / tool-version truth (no release was cut).

## Verification & deliverables

```
go build ./...   go vet ./...   go test ./... -count=1   (136 tests)
go run ./tools/demo -out demo   (10 real frames)
```

- [../04-tui-design.md](../04-tui-design.md) — current interface spec
- [../STATE.md](../STATE.md) — honest state, architecture, gaps, roadmap
- [../CORRECTIONS.md](../CORRECTIONS.md) — 16 corrections with tests
- [../../demo/index.html](../../demo/index.html) — visual tour

## Open items for the next pass

Services layer extraction (sweep 03/05), JUnit/Go test parsing in window 2
(sweep 08), diff hunk viewer (sweep 09), persisted UI preferences (sweep
25/59), glamour notes (sweep 43), huh settings form (sweep 42), Windows
hardware acceptance (sweep 34/36/51).
