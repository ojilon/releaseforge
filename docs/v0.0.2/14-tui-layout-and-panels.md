# v0.0.2 doc 14: TUI layout and panels

## Goal

Turn the single-pane TUI into a legible dashboard: bordered top strip, body,
input, footer; a state-dependent status panel; structured command output;
graceful narrow terminals. No new commands — presentation only, on top of
doc 13's architecture.

## Current state

`Model.View` (`internal/app/app.go:207`) stacks 4 unbordered lines: header,
viewport, input, help. Styles live in `internal/tui/styles.go` (5 styles,
`StatusStyle` removed in doc 01). No tabs, no panels, no state colors, no
narrow handling; footer help line is static text.

## Design

Layout (all `lipgloss`, existing deps only):

    ┌ top strip (bordered): ReleaseForge <tool> · <project> · <version> · <state pill> ┐
    │ body viewport: command output blocks                                            │
    ├ input bar: "> " + textinput                                                     │
    └ footer: context key hints (change with state)                                   ┘

- State pill: `idle` (grey) / `scanning` / `building` / `testing` (yellow) /
  `failed` (red) / `ok` (green, last command). Driven by `Model.status` +
  a new `Model.phase` set around long commands (doc 13 drain start/done).
- Command output blocks: each executed command renders as
  `> <line>` echo + body + trailing `✓/✗ <label>` — same strings as today,
  plus a faint rule line between blocks (lipgloss border, fg 8).
- Project context block: `status` content also feeds a compact always-visible
  line in the top strip second row when width allows (branch + dirty flag
  from doc 12 data, already computed in `statusText`).
- Commit sparkline: last 14 days of commit counts from scan cache as
  `▁▂▃` bars in the Git-adjacent line — pure function over
  `Snapshot.Git.RecentCommits` dates; empty when no cache. No new git calls.
- Color tokens centralized in `styles.go` (add `PillOk/PillRun/PillFail`,
  `Rule`, `PanelBorder`; keep the existing 4 styles' colors).
- Narrow fallback: width < 80 → single-line header (tool + project only),
  footer becomes `help|quit`, body unchanged. Implemented as a branch in
  `View` on `m.width` (store from `WindowSizeMsg`).
- Mouse: enable `tea.WithMouseCellMotion()` + viewport wheel — one line in
  `Run`, plus acceptance check it doesn't break non-mouse terminals.

## Files to touch

- `internal/tui/styles.go` (tokens, pill helpers, block renderer).
- `internal/app/app.go` (`View`, `phase` field, width tracking, footer text
  per state, `Run` mouse option).

## Steps

1. Centralize tokens + pill helper + tests (pure string functions).
2. Rebuild `View` with borders/panels behind a width branch; keep all text
   content identical at ≥80 cols (diff-friendly).
3. Add `phase` transitions to long-command start/done/cancel paths.
4. Sparkline pure function + tests (date buckets, empty input).
5. Mouse option + narrow-fallback manual check.

## Tests to add

- Pill renderer per phase (contains state word; no assertion on ANSI codes
  beyond `lipgloss` output containing the label).
- Block formatter: echo + body + result markers present.
- Sparkline buckets incl. empty/single-day.
- Width branch: narrow render omits footer hints (test `View` with set width
  via `WindowSizeMsg` on a headless model).

## Acceptance checks

- `go vet ./...`, `go test ./...` pass.
- Manual at ≥100 cols: bordered strip, state pill changes during `test`,
  blocks separated, footer accurate. At 70 cols: compact header, no wrap
  explosion.
- No command text changed (compare `help`/`status` output before/after).

## Out of scope

Tabs (still no tab set in v0.0.2 — panels only), charts beyond the sparkline,
themes/configurable colors, mouse text selection.
