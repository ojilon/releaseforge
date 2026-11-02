# Corrections log — 2026-11-02 pass

Every item below is a defect found by reading/running the code, with the fix,
the reason, and the test that now pins it. The intent of the app did not
change: these are corrections to *how* the existing work is done.

Legend: **B** = before, **A** = after, **T** = test/evidence.

---

## C1 · `theme` with no argument crashed the TUI

- **B** `internal/app/commands.go`: `case "theme": tui.SetTheme(args[0])` —
  typing `theme` (or picking it from the palette) panicked with
  `index out of range [0] with length 0`, killing the whole program.
- **A** The verb table declares `theme [dark|light|mono]`. No argument prints
  the active preset and the available names; `tui.SetTheme` returns `bool`
  and an unknown name is reported instead of silently ignored.
- **T** `TestThemeCommandWithoutArgsDoesNotPanic` (app), `TestSetThemeAndRepaint` (tui).

## C2 · Every key went to two consumers

- **B** `update.go:Update` ended with both `m.input.Update(msg)` and
  `m.viewport.Update(msg)` for *every* message. Typing could scroll the log,
  arrow keys could appear in the bar, and there was no way to scroll a pane
  with the keyboard (sweep 41 called this out).
- **A** `update.go` is now a single-consumer router. Global shortcuts are
  handled first; the remaining keys go to **exactly one** of: command bar,
  active window, palette, folder picker. Mouse wheel is routed to the active
  window. Each window owns its own viewport/cursor.
- **T** `TestSingleConsumerDispatch`, `TestWindowSwitchKeepsScroll`.

## C3 · Layout arithmetic overflowed and wrapped

- **B** `model.go:layout()` used magic heights; the dashboard was rendered at
  `DashBoxStyle.Width(34)` (which adds padding + borders) while layout only
  subtracted `34+3`. On wide terminals the header/footer/body drifted. The
  narrow path was an unrelated branch in `View()`.
- **A** `internal/tui/layout.go:Compute(width,height)` owns all breakpoints
  and rects; `Clamp` + `Frame` (ANSI-aware, wide-rune safe) force every block
  to an exact `width × height`. Windows render into `Layout.Content`.
- **T** `TestViewFitsTerminal` asserts every line has exactly the terminal
  width and every frame exactly the terminal height at 50/70/100/160/200
  columns; `TestComputeBreakpoints`, `TestClampAndFrameExact`.

## C4 · Help was a second, stale source of truth

- **B** `helpText()` was a hand-written string that had already drifted twice
  (`theme`, `devices`, `bat` missing or wrong). It was also only echoed into
  the log viewport.
- **A** `internal/tui/intent.go:Verbs` is the single table. It generates the
  Help window (`bubbles/table`), tab completion, palette candidates,
  unknown-verb suggestions, and the new CLI `releaseforge commands`.
- **T** `TestHelpTextGeneratedFromVerbs`, `TestCompletionsGenerated`.

## C5 · `recent` was a dead end

- **B** `recent` printed a numbered list, but the number was not usable; you
  had to retype the full path (with spaces this was impossible anyway, C9).
- **A** `open 2` / `open #2` resolves the recent entry; `ctrl+p` lists recent
  projects as fuzzy-searchable actions. The Status window shows recent runs.
- **T** `TestOpenRecentByIndex`, `TestPaletteRanksCommands`.

## C6 · `esc` was overloaded and cancel lost its meaning

- **B** `esc` cancelled a run *and* was the intended unfocus key; a
  `doneMsg` arriving after `Cancel()` produced `phase = "failed"`, so a
  deliberate stop looked like a build failure (sweep 40).
- **A** `phase.go` defines an explicit FSM with `PhaseCancelled` and a
  validated transition table. `cancelRun()` marks the `RunState` before
  killing the process; the late `doneMsg` reads that flag. `ctrl+x` is the
  explicit cancel, `esc` closes overlays / returns focus / cancels only as a
  backwards-compatible last resort.
- **T** `TestCancelYieldsCancelledPhase`, `TestWaitLineDrainsAndCancel`,
  `TestViewPhasePill` equivalent via `TestPillCancelled`.

## C7 · Async state was scattered across five fields

- **B** `handle`, `running`, `pendingRun{kind,variant,started}`, `pendingOK`,
  `pendingErr` had to be kept in sync by hand in `jobs.go`/`scripts.go`.
- **A** One `RunState` struct owns the handle, kind, variant, start time,
  labels and cancel flag; `installRun()` is the single mutation point.
- **T** `TestWaitLineDrainsAndCancel`, `TestCancelYieldsCancelledPhase`.

## C8 · History and clear left stale state

- **B** `historyDown()` did nothing when `histIndex == len(history)-1`, so the
  last command could never be cleared; `ctrl+l` replaced the ring but left
  `viewport.YOffset` at the old value.
- **A** Moving past the newest entry clears the input; `buildWindow.Clear()`
  resets the ring, content **and** offset.
- **T** `TestHistoryDownClearsInput`.

## C9 · Paths with spaces were unopenable

- **B** `parse()` used `strings.Fields`, so `open "D:\Dev\My App"` became four
  tokens and the scan failed.
- **A** `tui.Tokenize` honors single/double quotes and escaped quotes while
  keeping Windows backslashes literal; an unterminated quote is a clear error
  rather than a mangled path.
- **T** `TestTokenize`, `TestCompletionAndQuotedPath`.

## C10 · Theme switching was half-applied and blind to the environment

- **B** `SetTheme` rebuilt chrome styles but the focused `textinput` kept its
  old colors; there was no `NO_COLOR`, no light-background detection and no
  monochrome option.
- **A** `theme.go` has dark/light/mono presets, `DetectTheme`
  (`RELEASEFORGE_THEME > NO_COLOR/TERM=dumb > COLORFGBG > TTY background`),
  and `PaintInput` is called on every switch. `fg`/`bg` helpers suppress all
  color in mono mode.
- **T** `TestDetectTheme`, `TestSetThemeAndRepaint`, `TestThemeNames`, frame
  10 of the demo (`theme light`), frame 09 (tiny terminal).

## C11 · Windows paths were misparsed on non-Windows hosts

- **B** `storage.ProjectNameFromPath("D:\Dev\Conductino-Android\")` used
  `filepath.Base`, which only knows `/` on Linux, yielding
  `D--Dev-Conductino-Android-`. A data-root config copied between machines
  produced a bogus project name.
- **A** Both separators are normalized before taking the last segment.
- **T** `TestProjectNameFromPath` (4 cases, both styles on any OS).

## C12 · apksigner detection assumed a single SDK filename

- **B** `FindApksigner` looked only for `apksigner.bat` on Windows or
  `apksigner` elsewhere, so mixed/older SDK installs were missed and the
  test failed on Linux.
- **A** `apksignerCandidates()` prefers the native launcher but accepts both
  names in every `build-tools/<version>` directory.
- **T** `TestFindApksignerSdkNewest` now creates both files and expects the
  native one on the current OS.

## C13 · Four android tests required Windows

- **B** `fakeAdb` wrote an `adb.bat` script; `Devices`, `Install`, `LaunchApp`
  and `LogcatTail` tests failed on Linux/macOS with
  `exec: "adb": executable file not found in $PATH`.
- **A** The fixture now installs a POSIX `sh` script on Unix (same canned
  responses, args log via `RF_TEST_ADB_LOG`) and keeps the `.bat` on Windows.
- **T** `go test ./internal/android/` passes on Linux; no adb/SDK needed.

## C14 · Installer test assumed `.exe`

- **B** `TestEnsureTreeFreshAndRerun` hardcoded `releaseforge.exe` even though
  `installer/main.go` already had a platform-aware `exeName()`.
- **A** The test uses `exeName("releaseforge")`.
- **T** `go test ./installer/` passes on Linux.

## C15 · Script launching was Windows-only

- **B** `scriptProgram` always chose `cmd.exe`, `powershell.exe` and
  `python`, so `bat *.sh` / `bat *.py` failed on Linux with a raw exec error.
- **A** Platform-aware launchers (`pwsh`/`sh`/`python3`/`python`, `bash` on
  Windows) with actionable messages when a runtime is missing; Windows-only
  extensions fail clearly off Windows.
- **T** Covered by `go vet` + build; behavior is documented in the verb help.

## C16 · Smaller consistency fixes

- `git diff` in the TUI now calls `git.DiffStat` instead of duplicating an
  `exec.Command` in `gitcmd.go`.
- `statusText()` gained `git: branch · clean|dirty` and the data root.
- Byte-based truncation helpers replaced with `ansi.Truncate` (CJK/emoji
  safe): `TestTruncateWideRunes`.
- `metrics.Recent` and `git.StatusPorcelain` were added as small read helpers
  so dashboards do not re-parse files or shell out ad hoc.
- `go.mod`/`go.sum`: `github.com/sahilm/fuzzy` (required by
  `bubbles/list`) is now recorded; `x/ansi` and `termenv` are direct deps.

---

## Baseline → now

| Check | Before | Now |
|-------|--------|-----|
| `go build ./...` | ok | ok |
| `go vet ./...` | ok | ok |
| `go test ./...` | **6 failures** (installer, android ×4, storage) | **136 tests pass**, all packages |
| `theme` (no arg) | panic | prints current preset |
| terminal widths verified | ad hoc | 50/70/100/160/200 cols, exact frames |
| demo | none | 10 real frames + HTML player |
