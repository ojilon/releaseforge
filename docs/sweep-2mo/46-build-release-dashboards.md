# 46 — Build & Release Dashboards (the windows that earn the name)

## Goal

Build (`4`) + a Release overview complete the `07` shell: live progress,
artifacts, and publish gates — not just streamed text.

## Build window (`4`)

- Header: task + elapsed + progress (Gradle `--console=plain` phase parse:
  `:app:compile…`, `:app:package…`; Go package counter).
- Body: follow-tail (policy `Follow` from `41`) + finding pills from `28`
  (top errors pinned above tail).
- Footer: `ctrl+x` cancel, `r` rerun, `l` jump to Logs, artifact paths on success.

## Release window (part of `23` gates)

- Checklist: version ✓, tests ✓, assemble ✓, sign+verify ✓, manifest ✓,
  tag-new ✓ — each row links to its log; `publish` blocked until all green.
- Assets table from manifest (`22`): name + size + sha (truncated).
- `gh release` URL after publish; `promote` action inline.

## Snippet — phase parse

```go
var gradlePhase = regexp.MustCompile(`^> Task :(\S+)`)
```

## Read

- `12` runner + `28` findings + `22` manifest as data sources.

## Tasks

- [ ] Phase parser + checklist model + asset table (bubbles `table`).
- [ ] `r` rerun + `publish` confirm wired to `23` gates.

## Acceptance

```powershell
# TUI Build: failing build pins top error + log path; Release: publish blocked until green
```
