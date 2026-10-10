# 06 — Command Bar, Keys, Help

## Current state

`app.go:93` placeholder lists verbs; `parse:374` lowercases first field;
`Update:229` handles ctrl+c / esc / up-down / enter / ctrl+l. Help is a
static string (`helpText:485`).

## Problems

- `esc` overloaded (cancel run AND would-be unfocus); up/down history has
  no search; no `tab`-completion; help goes stale when verbs change.

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
    {Name: "build", Subs: []string{"debug","release"}, Help: "gradle assemble / go build"},
    {Name: "test", Subs: []string{"unit","instrumented","all"}, Help: "run tests"},
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
