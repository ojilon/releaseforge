# 06 — Command Bar, Keys, Help

## Current state

`model.go:New()` placeholder lists verbs; `commands.go:parse` lowercases
first field; `update.go:onKey` handles ctrl+c / esc-cancel / up-down /
enter / ctrl+l. `bat` streams async like build (`scripts.go:startScript`,
`pendingRun.kind="script"`). Help is a static string (`commands.go:helpText`).
Verbs now: status/scan/recent/version/build [variant] [version]/test/
sign/install/devices/git/ls/dir/pwd/cat/bat/notes/logs/clear/theme/quit.

## Problems

- `esc` STILL overloaded (cancel run AND would-be unfocus); up/down history
  has no search; no `tab`-completion; help goes stale when verbs change
  (already happened twice — `helpText` updated by hand).

## Target

- Grammar: `verb [subverb] [--flags] [args]`; `parse()` → `Intent` struct
  with completion candidates.
- Keys: `enter` run, `up/down` history, `ctrl+r` reverse-search history,
  `tab` complete verb, `ctrl+l` clear, `ctrl+x` cancel run, `esc` back/focus-pop.
  Reserve `ctrl+c` quit (confirm if running).
- Help rendered FROM the intent table (single source), in Help window AND
  `--help`. Never a second hand-written string.

## Snippet

```go
type Intent struct{ Verb, Sub string; Args []string; Flags map[string]string }
var verbs = []Verb{
    {Name: "build", Subs: []string{"debug","release"}, Help: "go build / gradle assemble, optional version first"},
    {Name: "test", Subs: []string{"unit","instrumented","all"}, Help: "run tests"},
    {Name: "sign", Subs: []string{}, Help: "sign APK (data-root keystore)"},
    {Name: "bat", Subs: []string{}, Help: "run .bat/.ps1/.sh/.py with live logs"},
}
func Complete(prefix string) []string { /* filter verbs+subs */ }
```

## Read

- `internal/history/history.go`, `history_test.go` — extend for search.
- Bubbles textinput + `suggestion` patterns.

## Tasks

- [ ] `intent.go` + completion + `help` generated from `verbs` table.
- [ ] Rebind: esc=back, ctrl+x=cancel; update footer + help + docs.
- [ ] `ctrl+r` history search overlay.

## Acceptance

```powershell
go test ./internal/app/ -run TestParse -v
# TUI: type `b<tab>` → `build `; `help` lists same verbs as `--help`
```
