# 45 — Windows Terminal Quirks (coninput, pwsh, ANSI, sizes)

## Why this file exists

Windows console is the daily driver here: `coninput` (already indirect),
`go-isatty`, `termenv`, alt-screen + mouse (`run.go:Run`), pwsh quoting,
`<80`-col fallback + `>=100` dashboard (`view.go`), `bat` scripts run via
`cmd.exe /C` (`scripts.go:scriptProgram`). AI code usually tests on Linux only.

## Checklist

1. isatty: non-tty (piped) → no alt-screen, plain output, `--json` safe.
   Gate `Run()` on `isatty.IsTerminal`.
2. Resize: `WindowSizeMsg` storms during drag — debounce layout (`04`
   `layout.go`), never realloc viewport per pixel.
3. Unicode: runewidth truncation (`04`), never byte-slice; test CJK/emoji.
4. Keys: esc-sequences vs alt-keys in Windows Terminal vs conhost; `esc`
   STILL cancels runs (the `06` rebind to `ctrl+x` not done) so cancel
   survives terminal variance for now; document when it changes.
5. pwsh quoting: `--set "v1.0 (2)"` forms documented; Cobra args quoted in help.
6. `NO_COLOR` + light-background: theme auto-detect (`04`), snapshot both.

## Read

- coninput, go-isatty, termenv, lipgloss Windows issues; Windows Terminal
  key-binding docs: https://learn.microsoft.com/windows/terminal/.

## Tasks

- [ ] isatty gate + resize debounce + key-matrix test doc.
- [ ] Snapshot tests: dark/light × 70/100/160 × NO_COLOR on/off.

## Acceptance

```powershell
releaseforge status | Select-Object -First 5   # piped: no ANSI, no hang
# resize drag: no flicker, footer intact
```
