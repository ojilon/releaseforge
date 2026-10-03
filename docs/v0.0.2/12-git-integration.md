# v0.0.2 doc 12: Git integration (status, changelog, bounded notes)

## Goal

Richer local-git insight and release notes that stay bounded and grouped:
branch/dirty/ahead-behind, commits since last tag, conventional-commit
grouping, and a cap so huge histories can't flood notes or the TUI.

## Current state

`internal/git/notes.go` shells out for branch/head/tags/log/clean/origin
[DONE]. `DraftNotes` dumps every commit since the last tag unbounded
(`notes.go:143`) — with no tags at all, `LogSince` returns the entire
history. No ahead/behind, no grouping, no cap.

## Design

- `AheadBehind(dir)` in `internal/git`: `git status -sb` parse for branch +
  dirty already covered; ahead/behind via `git rev-list --left-right
  --count HEAD...@{u}` (no upstream → `("", "")`, not an error).
- `Changelog(dir, max=100)`: `LogSince(LatestTag)` capped at 100 commits;
  group subjects by conventional prefix (`^(\w+)([!])?:`) into
  Features/Fixes/Docs/Chores/Other (case-insensitive `feat|fix|docs|refactor|
  chore|test|perf|build|ci|style`; everything else → Other; `!` marks
  breaking and is kept in the subject). Returns sections + `truncated bool` +
  total count; body prints `…​and N more (see git log)` when truncated.
- `notes`/`release` use `Changelog` instead of raw `LogSince` (same template
  sections, Changes now grouped). `status` gains an
  `upstream: ahead X, behind Y (or no-upstream)` line.
- Cap constant `MaxChangelogCommits = 100` next to `MaxRecent` style.

## Files to touch

- `internal/git/notes.go` (`AheadBehind`, `Changelog`, cap; `DraftNotes`
  takes grouped sections — additive overload, keep old signature for tests).
- `cmd/notes.go`, `cmd/release.go:writeNotes`, `cmd/status.go` (upstream line).

## Steps

1. `AheadBehind` + tests (temp repo, no-upstream + ahead/behind cases need a
   second clone — use local bare remote fixture, no network).
2. `Changelog` grouping + cap + tests (synthetic commit lists + temp repo).
3. Switch notes/release/status call sites; keep template headers identical.

## Tests to add

- Grouping matrix (feat/fix/FEAT/no-prefix/`!`/scope like `feat(api):`).
- Cap: 150 commits → 100 shown + truncated flag + count.
- `AheadBehind`: no-upstream empty; ahead 2 / behind 1 via local remotes.

## Acceptance checks

- `go vet ./...`, `go test ./...` pass.
- On this repo: `notes` output shows grouped sections; `status` shows an
  upstream line (or `no-upstream`).

## Out of scope

AI polish (still later), PR-title fetching (network), tag creation (11).
