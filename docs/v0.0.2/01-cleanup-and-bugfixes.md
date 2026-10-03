# v0.0.2 doc 01: Cleanup and bugfixes

## Goal

Fix known bugs and remove dead code before new work builds on it. Pure
stabilization; no behavior changes except the bugfixes listed.

## Current state

- `cmd/install.go:46`: `if e.IsDir() && strings.Contains(e.Name(), ".") || e.IsDir()`
  is true for every directory (`&&` binds tighter than `||`), so the `.`
  filter and "newest" selection are broken.
- Dead code: `notImplemented` (`cmd/root.go:83`, zero callers),
  `asyncResult` (`internal/app/app.go`, never referenced), `StatusStyle`
  (`internal/tui/styles.go:16`, unused), `GradleArgs`
  (`internal/build/runner.go:163`, unused), `Stream.Subscribe` + fan-out
  (no callers; doc 03 re-adds subscription deliberately).
- Kept deliberately (reserved, not dead): `logMsg`/`doneMsg`/`running` and
  their `Update` branches (doc 03 sends them for real), `HistoryFile`
  (doc 13 needs it for `commands.jsonl` persistence).
- Garbled TUI help (`internal/app/app.go:298`): `show or set version (gradle
  VERSION-free: VERSION file)`.
- Binary base name `releaseforge` hard-coded in three places: `goBinaryOut`
  (`cmd/test.go:92`), `releaseGo` output (`cmd/release.go:205`), TUI `doBuild`
  (`internal/app/app.go:429-432`) — wrong for any other Go project.
- `SanitizeProjectName` (`internal/storage/storage.go:208`) truncates bytes
  (`s[:64]`), can split UTF-8; also counts bytes, not runes.
- `config.go:301` formatting glitch (collapsed brace).
- `mirror_release_dir_in_repo` (`internal/config/config.go:108`) is never
  acted upon — decision made: drop the field.

## Design

Surgical edits only. Drop the field with a tolerant loader (ignore unknown
JSON keys on read — `encoding/json` already does — so old `config.json`
files keep loading). Binary naming moves to a per-project helper owned by
`internal/project` so doc 02 can adopt it (`BinaryBaseName(info)` returning
the sanitized project name; `releaseforge` stays correct for this repo).
Truncation becomes rune-safe with the same 64 limit and same fallback.

## Files to touch

- `cmd/install.go` (precedence fix + keep behavior otherwise)
- `cmd/root.go`, `internal/app/app.go`, `internal/tui/styles.go`,
  `internal/build/runner.go`, `internal/log/stream.go` (dead code removal;
  keep `Subscribe`? No — remove it; doc 03 re-adds a subscription design
  deliberately)
- `internal/storage/storage.go` (rune-safe truncation)
- `internal/storage/storage.go` (rune-safe truncation + test)
- `internal/config/config.go` (drop field, formatting fix)
- `cmd/test.go`, `cmd/release.go`, `internal/app/app.go` (binary-name helper)
- `internal/project/detect.go` (new `BinaryBaseName` + test)
- `configs/example-android-gradle.json` (drop the removed field from the example)
- New `cmd/install_test.go` (pure dir-pick function test)

## Steps

1. Fix the `install.go:46` condition to the intended "directory whose name
   compares newest" logic; keep the glob fallback.
2. Delete dead symbols listed above; `go vet` must stay clean.
3. Rewrite the garbled help line.
4. Add `project.BinaryBaseName` (sanitized name) + use it in the three
   call sites; verify this repo still yields `releaseforge`.
5. Rune-safe `SanitizeProjectName`; extend its test with multibyte names.
6. Remove `mirror_release_dir_in_repo`; confirm old configs still load.
7. Fix `config.go` formatting.

## Tests to add

- `install`-level dir-pick test (temp `releases/` with `0.0.1/`, `0.0.2/`,
  junk file; assert newest dir chosen). Put selection logic in a pure
  function to make this testable without adb.
- `BinaryBaseName` cases (normal, spaces, multibyte, empty→fallback).
- Old `config.json` containing `mirror_release_dir_in_repo` still loads.

## Acceptance checks

- `go vet ./...` and `go test ./...` pass.
- `rg -n "notImplemented\(|StatusStyle|GradleArgs|asyncResult|mirror_release" --glob '*.go' internal cmd` returns only comments/docs.
- `go run . install --help` unchanged; behavior identical except newest-dir pick.

## Out of scope

Runner interface (02), any TUI behavior change, any scan-format change.
