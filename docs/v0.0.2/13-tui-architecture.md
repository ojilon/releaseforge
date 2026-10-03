# v0.0.2 doc 13: TUI architecture (async, ring buffer, history)

## Goal

Rebuild the TUI update loop around async work: commands return `tea.Cmd`,
lines stream through messages, output lives in a bounded ring, history
persists — while keeping every existing command working.

## Current state

`internal/app/app.go`: single `viewport`, string-concat `appendLine`
(O(n²)), blocking `doBuild/doTest` in the Enter handler, dead
`logMsg/doneMsg/asyncResult/running`, no tabs, arrow keys stolen by history
so viewport keyboard scroll is dead except PgUp/PgDn, history in-memory only
(`history/commands.jsonl` never written).

## Design

Message protocol (tiny on purpose):

    type lineMsg struct{ text string }     // one streamed line
    type doneMsg struct{ label string; err error }  // reuses existing type, now actually sent
    // doneMsg.err == nil → "✓ label"; else red "✗ label: err"

`exec` stays a pure parser returning `(tea.Cmd, bool quit)`: for instant
commands (`status/scan/version/recent/logs/notes/help/clear`) it returns a
`tea.Cmd` producing the already-computed string (or prints directly — keep
the current append path, just move it into the Cmd for uniformity); for
long commands (`build/test`) it calls `build.Start` (doc 03) and returns a
drain Cmd: loop on `Handle.Lines`, emit `lineMsg`, watch `Handle.Done`,
then emit `doneMsg`. `esc` while `running` calls `Handle.Cancel` and marks
`cancelled` (doc 03 provides Cancel; this doc wires the key).

Output store: replace `SetContent(View()+…​)` with `log.Ring` (doc 03,
cap 2000) + render tail into the viewport on each batch (batch every ~50ms
via `tea.Tick` to avoid re-render per line on fast Gradle output). Viewport
keeps its own scroll position; `GotoBottom` only when already at bottom
(standard tail-follow).

History: on every Enter, append the line to `history/commands.jsonl`
(best-effort, cap file at 500 lines on write); load at `New`. Keep the
in-memory Up/Down behavior identical.

Keymap: keep everything; add `esc` = cancel running (else no-op),
`pgup/pgdn` unchanged, `ctrl+l` = clear. Document in the footer (doc 14
owns footer text; this doc implements the bindings).

## Files to touch

- `internal/app/app.go` (Update/exec rework; `do*` handlers return
  `(string, tea.Cmd)` or split into instant vs long paths).
- `internal/log/` (use `Ring` from doc 03).
- `internal/history/` (tiny `AppendCommand/LoadCommands` for commands.jsonl).

## Steps

1. History persistence first (smallest, independent) + tests.
2. Extract a pure `parse(line)` (verb + args) with a table test; keep
   instant commands synchronous (append directly) — wrapping them in Cmds
   would add churn for zero behavior change.
3. Convert `build`/`test` to `Start` + drain Cmd + `esc` cancel.
4. Ring-backed rendering with tick batching + tail-follow rule.
5. `go test -race ./internal/app/` in acceptance (new goroutines).

## Tests to add

- `parse` table: every verb, aliases (`q`, `?`, `-h`, `open`, `rescan`),
  unknown verb error.
- History file round-trip + 500-line cap.
- Drain Cmd against a fake `Handle` (canned lines + result) → message order.
- Race test for cancel-during-stream.

## Acceptance checks

- `go vet ./...`, `go test ./...`, `go test -race ./internal/app/` pass.
- Manual: `test` streams visibly, `esc` cancels within ~2s, Up/Down history
  survives restart, long output doesn't grow memory (ring cap observable in
  code, not manually).
- All pre-existing commands produce identical text output.

## Out of scope

Visual layout changes (14), tabs, progress bars, mouse (14 decides).
