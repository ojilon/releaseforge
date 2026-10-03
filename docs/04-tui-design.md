# TUI design (current through v0.0.2)

> v0.0.2 status: single viewport + command bar, implemented. There are no
> tabs and no panes; the tab sketches below are retained as explicitly
> future, not as built UI. Folder picker was deliberately never built.

## Inspiration

Claude Code / Gemini CLI / OpenCode: **persistent command input**, live panes, open-folder flow, minimal chrome.

## Layout (as built)

```
┌─ ReleaseForge 0.0.1 · SomeProject · v0.0.3_2 (4) · ready  ● idle ─┐
│ context: main · clean · ▁▂▅                                     │
├────────────────────────────────────────────────────────────────────┤
│  (viewport — command echo, output blocks, ─── rule separators)    │
├────────────────────────────────────────────────────────────────────┤
│ > help                                                             │
│ help · status · scan [path] · … (state-dependent footer)           │
└────────────────────────────────────────────────────────────────────┘
```

- Top strip (bordered): tool + project + version + status + state pill, plus
  a context line (branch · dirty/clean · commit sparkline). Narrow terminals
  (<80 cols) fall back to a one-line header and `help|quit` footer.
- Body: one scrollable viewport (ring-backed, 2000 lines; mouse wheel on).
- Input: always-focused bar with `↑/↓` history (persisted to
  `history/commands.jsonl`). `esc` cancels a running build/test, `ctrl+l`
  clears.
- Footer: key hints that change with phase (`esc cancels` while running).

## Tabs (future, not built)

| Tab | Content |
|-----|--------|
| Overview | Name, root, type, version, last scan |
| Git | Recent commits and tags from scan cache |
| Tools | Detected tools, frameworks, config files |
| Logs | Build logs |
| Config | Data-root path, project config path |

Equivalent coverage today, without tabs: `status` (project + git + last
build), `scan` output, `logs show|tail`, `recent`, `version`. Build /
Metrics tabs stay future.

## Command bar commands

| Input | Action |
|-------|--------|
| `open <path>` / `scan [path]` / `rescan` | Scan path (typed only, no dialog) |
| `recent` | List recent projects |
| `status` | Tool + project + version (`[cache]` provenance) + data root + upstream + last build |
| `version [--set X \| bump X \| --code-only]` | Show or set version |
| `build debug\|release` | Async Gradle/Go build, streamed, `esc` cancels |
| `test [kind]` | Async tests (`--device` for instrumented) |
| `notes [version]` | Draft notes from grouped changelog |
| `logs [show\|last\|tail]` | List/read/tail persisted logs |
| `clear` / `ctrl+l` | Clear viewport |
| `help` / `-h` / `?` | Show help **in the main pane** |
| `quit` / `exit` / `q` | Leave TUI |

Unknown commands: short error + hint to type `help`. There is no `release`,
`install`, `run`, or `doctor` verb — those are CLI-only, deliberately (they
change the world; the bar is for inspection plus build/test).

## Help view

Rendered in the main pane (Viewport), not a one-line toast. Include command table above and note that CLI flags are available via `releaseforge --help` outside the TUI.

## Folder open

Typed paths only (`open D:\Dev\MyRepo`), then scan → recent → header
refresh. No OS dialog (see `docs/09-scan-foundation.md` for the decision).

## Libraries

- bubbletea (+ `WithAltScreen`, `WithMouseCellMotion`), lipgloss, bubbles (textinput, viewport), huh (init form, device picker, password prompt)
