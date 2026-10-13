# 44 — Fuzzy Finder & Command Palette (open anything in 3 keys)

## Goal

`ctrl+p` palette: fuzzy-search projects (recent + scan cache), files
(diff list), commands (verbs from `06`), devices — one input, ranked list,
enter acts.

## Design

- `internal/tui/fuzzy.go`: pure scorer (subsequence + boundary bonus),
  tested on 10k candidates (<5ms). No new dep needed; keep simple, or
  vendor `sahilm/fuzzy` only if bench wins.
- Sources: recent projects (`history`), `git status` files (`09`),
  verbs (`06`), devices (`36`). Each row: `icon + label + hint + action`.
- TUI: overlay list over active window; `up/down/enter/esc`; command bar
  `open <fuzzy>` equivalent for CLI parity (`recent --fuzzy q`).

## Snippet — scorer contract

```go
func Score(query, target string) int // -1 no match; higher better
func Rank(query string, targets []string, top int) []string
```

Bonus: word-boundary, basename, exact-prefix; penalty: long gaps.

## Read

- fzf exact/fuzzy matching docs (algorithm inspiration, not dep).
- Your `history.LoadRecent` + porcelain file list as sources.

## Tasks

- [ ] Scorer + bench + golden rankings; palette overlay; CLI `--fuzzy`.

## Acceptance

```powershell
go test ./internal/tui/ -run TestFuzzy -v
go test -bench=. ./internal/tui/ -benchtime=3x  # <5ms per 10k
```
