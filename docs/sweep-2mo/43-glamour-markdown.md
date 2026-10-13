# 43 — Glamour Markdown (notes + help rendered beautifully)

## Goal

`notes`, `help`, changelogs render as styled markdown in viewport via
`github.com/charmbracelet/glamour` (already anticipated in `go.mod:42`
comment) — readable without leaving the TUI.

## Design

- `internal/tui/md.go`: `Render(md string, width int) string` with
  auto-style (dark/light via `termenv.HasDarkBackground`), width-clamped,
  `NO_COLOR` fallback to plain.
- `notes [version]` output + Help window (`07`) go through it; raw `.md`
  path still printed for scripting (`--raw` flag).
- Cap render at 200KB input; larger → plain + truncation note (perf guard).

## Snippet

```go
import "github.com/charmbracelet/glamour"
r, _ := glamour.NewTermRenderer(glamour.WithAutoStyle(), glamour.WithWordWrap(width))
out, err := r.Render(md)
```

## Read

- Glamour: https://github.com/charmbracelet/glamour ; your `git.DraftNotes` format.

## Tasks

- [ ] Add glamour dep + `md.go` + snapshot tests (70/100 cols, NO_COLOR).
- [ ] Wire notes/help; `--raw` passthrough for pipes.

## Acceptance

```powershell
go test ./internal/tui/ -run TestMarkdown -v
# TUI: `notes` renders headings/lists; pipes still get raw md with --raw
```
