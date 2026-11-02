# TUI design (current — windowed shell, 2026-11-02)

> Supersedes the v0.0.2 "single viewport + dashboard" design. That UI
> (bordered header, one ring-backed viewport, fixed 34-col dashboard,
> hand-written help) is gone. The intent is unchanged: a **persistent
> command bar plus live panes** for local scan/build/test/release work.
> See [STATE.md](STATE.md) for what is built and [CORRECTIONS.md](CORRECTIONS.md)
> for why. Visual tour: [../demo/index.html](../demo/index.html).

## Inspiration

Claude Code / Gemini CLI / OpenCode for the always-visible command input;
`bubbletea` Elm architecture for routing; `bubbles` for every interactive
widget; `lipgloss` for frames, tabs and exact geometry. No custom widget
where a Charm component exists.

## Layout (as built, 132 × 36)

```
╭─ ReleaseForge 0.0.2 · atlas-app · v0.0.4 · ready   ● idle   Status ─────╮
│ master · clean · █  ·  bar focus  ·  theme dark                          │
╰─────────────────────────────────────────────────────────────────────────╯
 1 Status   2 Tests   3 Git   4 Console   5 Logs   6 Help
╭─────────────────────────────────────────────────────────────────────────╮
│ Project ─────────────────────────────────────   ╭─ At a glance ───────╮ │
│ project     atlas-app                            │ phase    idle        │ │
│ root        D:\Dev\atlas-app                     │ theme    dark        │ │
│ version     v0.0.4                              │ branch   main        │ │
│ …                                               ╰──────────────────────╯ │
│ Recent runs ──────────────────────────────────────────────────────────  │
│  when        kind   variant  result  time       (bubbles/table)          │
╰─────────────────────────────────────────────────────────────────────────╯
╭─ ❯ type a command · tab completes · ctrl+p palette · ctrl+o folder ──────╮
╰─────────────────────────────────────────────────────────────────────────╯
enter run · tab complete · ctrl+p command palette · ctrl+o pick folder · …
```

- **Header frame** (4 lines normal, 3 compact, 1 tiny): tool + project +
  version + status, phase pill, active-window chip, context line
  (branch · clean/dirty · commit sparkline · focus · theme).
- **Tab strip**: numbered windows, active tab highlighted. Compact terminals
  collapse it to numbers + the active label; tiny terminals drop it.
- **Body frame**: the active window, rendered into an exact content rect.
  Wide layouts (≥120 cols) can split it into a primary pane + side card.
- **Command bar frame**: always focused by default, `↑/↓` history (persisted
  to `history/commands.jsonl`), `tab` completion.
- **Footer**: generated from `bubbles/key` bindings through
  `bubbles/help` (`ShortHelpView`) — automatically truncates with `…` on
  narrow terminals, and swaps in `ctrl+x cancel` while a run is active.

Everything is sized by `internal/tui/layout.go:Compute`. `Clamp` truncates
long lines with `ansi.Truncate` (wide-rune/emoji safe) and pads to the exact
rect, so no block ever wraps. Snapshot tests assert exact width and height
at 50, 70, 100, 160 and 200 columns.

## Windows

| # | Window | Content | Charm component |
|---|--------|---------|-----------------|
| 1 | **Status** | project facts, cached scan, git/upstream, last build, recent runs, live side card | `bubbles/table` + lipgloss cards |
| 2 | **Tests** | run history from `metrics.jsonl` + tail of the last test log | `bubbles/table` + `viewport` |
| 3 | **Git** | `s` status · `d` diff · `l` log, branch/upstream/tag summary | `bubbles/table` + `viewport` |
| 4 | **Console** | command echo, command output, streamed build/test/script lines | `viewport` + ring + `bubbles/spinner` |
| 5 | **Logs** | persisted logs browser + preview | `bubbles/list` + `viewport` |
| 6 | **Help** | generated command reference + key cheat sheet | `bubbles/table` |

Each window is an independent component (`Window` interface in
`internal/app/windows.go`) with its own viewport/cursor/offset. Switching
never resets scroll position. Only the active window receives keys and mouse
messages. `r` refreshes its cached data.

The Console keeps the original follow policy: it sticks to the bottom only
when the user was already at the bottom (`f` re-follows, `g`/`G` jump).

## Overlays

| Key | Overlay | Behaviour |
|-----|---------|-----------|
| `ctrl+p` | fuzzy command palette | `bubbles/list` + `textinput`, ranked by `tui.Rank` over commands, recent projects, logs, windows and history |
| `ctrl+r` | history search | same overlay, history-only source; selection fills the bar |
| `ctrl+o` / `open` (no path) | folder picker | `bubbles/filepicker` with `DirAllowed=true`, `FileAllowed=false`; `enter` selects a directory, `s` uses the directory you are in, `esc` closes |

Overlays are centered inside the body frame with a double border, so header,
tabs, bar and footer stay visible. This is the in-terminal answer to the
vision's folder picker; a native Windows dialog is still not built.

## Keys

| Key | Action |
|-----|--------|
| `1..6` (window focused), `alt+1..6` | switch window |
| `]` / `[` (also `alt+right/left`) | next / previous window |
| `ctrl+w` | toggle window focus (arrows then scroll the pane) |
| `esc` | close overlay / return focus to bar / clear input / cancel a run (compat) |
| `enter` | run command / activate palette item |
| `↑ ↓` | history (bar) / navigate (window) |
| `tab` | complete verb or argument |
| `ctrl+p` / `ctrl+r` | palette / history search |
| `ctrl+o` | folder picker |
| `ctrl+x` | cancel run |
| `ctrl+l` | clear console |
| `?` / `f1` / `help` | Help window |
| `r` | refresh active window |
| `ctrl+c` / `quit` | leave |

Bare digits are deliberately **not** global window shortcuts while the bar
is focused: `open 2`, `build release 0.0.4` and `version --set 1.2.3` must
be typeable. `alt+1..6` (and `1..6` once the window owns focus) switch panes.

## Command bar

`internal/tui/intent.go:Verbs` is the single grammar. It drives:

- dispatch (`commands.go:exec`),
- generated Help window and `releaseforge commands` CLI output,
- `tab` completion (verbs, subcommands, recent paths, theme names),
- palette candidates,
- "did you mean …" suggestions for unknown verbs.

All v0.0.2 verbs still work: `open|scan|rescan`, `recent`, `status`,
`version`, `build`, `test`, `bat`, `sign`, `install`, `devices`, `git`,
`ls|dir`, `pwd`, `cat`, `notes`, `logs`, `clear`, `help`, `quit`. Added:
`theme`, `window`, and recent-index / folder-picker forms of `open`.
`release`, `doctor` and `logcat` stay CLI-only by design.

`open "D:\Dev\My App"` works: the tokenizer honors quotes and keeps Windows
backslashes literal. `open 2` opens recent entry #2. Error lines are styled
but never hidden; raw output always lands in the Console and in the
persisted log under the data root.

## Run state

`phase.go` is an explicit FSM:

```
idle → building|testing → ok | failed | cancelled → idle
```

A `RunState` owns the handle, kind, variant, start time, labels and cancel
flag. `ctrl+x` marks it cancelled *before* killing the process, so a late
`doneMsg` renders `■ cancelled`, never `✗ failed`. The Console shows a
`bubbles/spinner` plus elapsed seconds while busy; metrics are recorded once
on completion.

## Theme

`internal/tui/theme.go` defines `dark`, `light` and `mono`. Startup calls
`DetectTheme`: `RELEASEFORGE_THEME` → `NO_COLOR` / `TERM=dumb` →
`COLORFGBG` → terminal background (TTY only) → dark. `SetTheme` rebuilds
every derived style *and* repaints the focused `textinput` and palette input.
`mono` emits no color at all. `theme dark|light|mono` switches at runtime;
the choice is session-only.

## Libraries actually used

| Library | Where |
|---------|-------|
| `bubbletea` | `Model.Init/Update/View`, `tea.Cmd` streaming, `tea.Tick`, `WithAltScreen`, `WithMouseCellMotion` |
| `bubbles/textinput` | command bar + palette query |
| `bubbles/viewport` | Console, Tests preview, Git detail, Logs preview |
| `bubbles/table` | Status recent runs, Tests history, Git log, Help commands |
| `bubbles/list` | Logs browser, command/history palette |
| `bubbles/filepicker` | folder picker overlay |
| `bubbles/spinner` | live build indicator |
| `bubbles/key` + `bubbles/help` | all bindings + generated footer |
| `bubbles/progress` | available, not yet wired (no determinate progress source) |
| `lipgloss` | frames, tabs, chips, pills, rules, cards, `JoinHorizontal/Vertical`, `Place` |
| `huh` | CLI `init` wizard + device picker (unchanged; not embedded in the TUI) |

## Not built (deliberately)

- Native OS folder dialog; the terminal picker is the substitute.
- `glamour` markdown rendering for notes (sweep 43).
- `huh` settings form inside the TUI (sweep 42).
- Deterministic progress bar for builds (Gradle/Go expose no reliable %).
- Extracted `services` layer / `go-git` (sweep 03).
- Persisted UI preferences, CI config, metrics charts.

## Acceptance checks

```bash
go test ./internal/app/ -run 'TestView|TestSingleConsumer|TestWindowSwitch|TestPalette|TestCancel' -v
go test ./internal/tui/ -v
go run ./tools/demo -out demo       # 10 frames, exact geometry
```
