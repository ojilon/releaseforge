# v0.0.2 doc 07: Scan cache readable + config refresh + seed

## Goal

Make `scan.json` a trusted snapshot that code actually reads, add freshness
metadata, refresh `config.json` on rescan without destroying user edits, and
honor the repo-local `.releaseforge.json` seed.

## Current state

- `WriteScan` (`internal/project/scan.go:201`) writes `cache/scan.json`, but
  no code reads it — `status`, TUI, `install`, and `release` all re-detect
  live (`audit §14`). `cache/tools.json`/`meta.json` from `docs/05` don't exist.
- `config.json` is written once (`cmd/scan.go:66`, `app.go:338`) and never
  refreshed; later scans can't add new tasks/keys.
- `.releaseforge.json` is inventoried (`scan.go:171`) but never parsed.
- `MirrorReleaseDirInRepo` will be gone (doc 01); anything referencing it
  must go here too.

## Design

Read path: `func LoadScan(dataRoot, project string) (Snapshot, error)` in
`internal/project/scan.go`; `status` and TUI header/version display prefer
cache when fresh, else live-detect (fresh = `cache/meta.json` root match,
see below). `install`/`release` keep live detection (safety-critical paths
must not trust cache) but read artifact globs from config (doc 09).

`cache/meta.json`: `{scanned_at, tool_version, root, root_mtime}` written by
every scan. Freshness rule: cache is fresh iff `meta.root` == current abs
root; mtime is informational only (no auto-rescan — explicit `scan/rescan`
still refreshes, per the v0.0.1 contract).

Config refresh with key ownership. Tool-owned keys (refreshed on every scan):
`name, root, type, version.*, build.tasks` defaults. User-owned keys (never
overwritten once present): `signing.*, github.*, artifacts.app_name`,
`editor`-style prefs. Merge rule: start from fresh detection → overlay
`.releaseforge.json` seed (no secrets allowed — warn and ignore
`signing.password`-like keys if seen) → overlay existing `config.json`
user-owned keys. New tool-owned keys appear; user edits survive. Print what
changed (`config: updated tasks.instrumented_test (was empty)` style lines).

## Files to touch

- `internal/project/scan.go` (`LoadScan`, `meta.json` write, merge function).
- `cmd/scan.go`, `app.doScan` (use merge; print change lines).
- `cmd/status.go`, TUI `statusText`/header (prefer cache when fresh).
- `internal/config/config.go` (drop mirror field fallout, if any remains).

## Steps

1. Add `meta.json` write + `LoadScan` + freshness helper + tests.
2. Implement merge with the ownership table + seed application + tests
   (temp roots, pre-seeded configs, malicious seed with password key).
3. Switch `scan`/`doScan` to merge; print per-key change lines.
4. Switch `status`/TUI display to cache-first with live fallback.
5. Delete remaining `mirror_release_dir_in_repo` references; old files load.

## Tests to add

- Merge: user `signing`/`github` preserved; new detected task added; tool
  `root` updated on move; seed applied; password-like seed keys ignored.
- Freshness: matching root fresh; moved root stale.
- `LoadScan` missing/corrupt file errors.

## Acceptance checks

- `go vet ./...`, `go test ./...` pass.
- `scan` twice: second run prints `config: unchanged`; hand-edit
  `config.json` `artifacts.app_name`, rescan, edit survives.
- Drop a `.releaseforge.json` with a custom task in a fixture project;
  rescan picks it up.

## Out of scope

Changing what scan detects (06), artifact-glob consumption (09), cache
invalidation by mtime.
