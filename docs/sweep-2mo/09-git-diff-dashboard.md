# 09 — Git Diff Dashboard (status, diff, log graph)

## Current state

`internal/git/` gives branch/head/commits/tags/origin; `doNotes:808`
drafts notes; scan caches 30 commits + 10 tags. No diff viewer, no staging,
no graph.

## Target

Diffs window (`3`):

- Left: `git status --porcelain` file list (staged/unstaged/untracked icons).
- Right: `git diff [--staged] -- <file>` with word-diff + hunk navigation
  (`j/k`, `[/]` prev/next hunk, `s` stage hunk via `git apply --cached`).
- Bottom: `git log --graph --oneline -20` + selected commit preview.
- Keys never rewrite history by default; destructive ops need `!` + confirm.

Use `git` binary first (porcelain-stable), go-git later only if needed.
Large diffs: cap at 200KB shown + `…truncated (N bytes)` + open-in-pager key.

## Snippet

```go
// porcelain v1 parse (stable contract)
out, _ := exec.Command("git","status","--porcelain=v1","-z").Output()
// entries split on \x00; XY status + path
// diff per file:
diff, _ := exec.Command("git","diff","--no-color","--",file).Output()
```

Hunk nav: split on `^@@ ` headers; store offsets for jump.

## Read (then implement)

- Porcelain: https://git-scm.com/docs/git-status#_porcelain_format_v1
- Diff: https://git-scm.com/docs/git-diff; log graph: https://git-scm.com/docs/git-log
- Deliberately AVOID go-git for diff rendering initially — CLI output is the contract.

## Tasks

- [ ] `internal/git/porcelain.go` parser + tests (renames, quotes, unicode).
- [ ] Diffs window: list + diff + log graph panes; hunk nav.
- [ ] `s` stage / `u` unstage hunk (confirm); refresh status after.

## Acceptance

```powershell
go test ./internal/git/ -v
# TUI Diffs: modify a file → appears; `s` stages; log graph renders
```
