# v0.0.2 doc 00: Overview and implementation order

## Goal

Define the work order for v0.0.2 ("solid Android Gradle support + a real TUI")
so each doc lands on top of the previous one with no rework. Phase 2
implements strictly in this order; each doc is one commit.

## Current state

Repo is at v0.0.1, shipped for the `go` type (`VERSION`, `internal/version/version.go`,
`installer/`, `docs/10-bootstrap-v001.md`, `RELEASE.md`). Android Gradle paths
exist in code (`internal/android/sign.go`, `cmd/release.go:releaseAndroid`,
`cmd/install.go`) but were never run against a real SDK (see audit §14–15).
The TUI is a single synchronous viewport (`internal/app/app.go`). Plan docs
`docs/08`, `docs/04`, `docs/09` are stale relative to the code.

## Design

Order follows dependencies: stabilize foundations (01–02), then make scan
data trustworthy (07, 06, 08), then execution (03, 04, 05), then Android
features (09, 10, 11), then history/notes (12), then TUI (13, 14), then
metrics + doc refresh last (15).

| # | Doc | Depends on |
|---|---|---|
| 01 | cleanup-and-bugfixes | — (first; everything builds on it) |
| 02 | runner-interface | 01 |
| 07 | scan-cache-and-config | 02 (Runner owns per-type scan enrichment) |
| 06 | android-scan-gradle | 07 (writes into the cache format 07 stabilizes) |
| 08 | version-and-bump | 06 (kts/properties sources found by 06) |
| 03 | async-process-and-logstream | 02 (Runner produces the processes) |
| 04 | error-parsing | 03 (consumes streamed lines + log files) |
| 05 | log-commands | 03, 04 (shows logs + error summaries) |
| 09 | android-build-test | 02, 04, 06 (tasks, ABIs, artifacts from scan) |
| 10 | adb-install-run | 09 (needs built artifacts) |
| 11 | sign-and-release-android | 09, 10, 04 (sign after build; dry-run) |
| 12 | git-integration | — (mostly independent; placed late, feeds notes) |
| 13 | tui-architecture | 03, 05 (async plumbing + log reading) |
| 14 | tui-layout-and-panels | 13 (layout on top of the new architecture) |
| 15 | metrics-and-polish | all (records what others produce; refreshes stale docs) |

Decisions already made (do not relitigate): data home is installer
`<app>/data`, `D:\` is fallback; drop `mirror_release_dir_in_repo`;
publishing stays `gh`-only; 2+ devices get a picker or `--device`; Gradle
test tasks keep the `:app:` prefix.

## Files to touch

Phase 1 touches only `docs/v0.0.2/00-*.md` … `15-*.md` plus, in Phase 2 step
zero, `docs/v0.0.2/PROGRESS.md`.

## Steps

1. Write docs 00–15 (this phase), each ≤ ~150 lines with the standard
   sections (Goal, Current state, Design, Files to touch, Step list, Tests,
   Acceptance, Out of scope).
2. Phase 2 step zero: create `PROGRESS.md` (convention below), all unchecked.
3. Implement one doc per commit in the table order; `go vet ./...` and
   `go test ./...` green before every commit; message `v0.0.2: <nn> <title>`.
4. If a doc is wrong, fix the doc first, then the code, in the same commit.

## PROGRESS.md convention

    # v0.0.2 progress
    - [ ] 01 cleanup-and-bugfixes
    - [ ] 02 runner-interface
    ...one line per doc, ticked (`[x]`) only when vet+tests pass and committed.

## Tests to add

None for doc 00. Every other doc lists its own tests.

## Acceptance checks

- `docs/v0.0.2/` contains `00`–`15`, each with all standard sections.
- No `.go` file modified in Phase 1 (`git status --short` shows only the new docs).
- Reply `approved` only after reading risks below.

## Out of scope

Wails/CMake builds, CI, autoupdate, AI notes, multi-device fan-out. No version
bump and no release until asked.

## Risks and conflicts found while planning

1. Docs `08`/`04`/`09` contradict the code; doc 15 refreshes them last so
   earlier docs must cite code, not those specs.
2. `scan.json` is currently write-only; docs 06/07/09 all assume readers
   exist — 07 must land before 06/09 are implemented.
3. TUI async (13) is the highest-risk change (deadlock/race surface); keep the
   message protocol tiny and covered by `go test -race` in CI-less local runs
   (`go test -race ./internal/app/`).
4. Low-powered dev machine: docs 06 (`--deep`) and 09/10 forbid spawning
   Gradle except on explicit user commands.
