# v0.0.2 doc 08: Version read/write for kts + bump command

## Goal

Version handling that covers Kotlin DSL build files, adds an explicit `bump`
flow, and validates input — so Android projects without `gradle.properties`
versions stop erroring out.

## Current state

`GradleVersion/SetGradleVersion` (`internal/project/detect.go:106-156`)
only understand `gradle.properties` (`app.versionCode/Name` regexes, code+1
semantics from `scripts/version.py`). `version --set` fails for kts-based
versions and for `go` without a `VERSION` file. No `bump` verb, no
`--code-only`, no validation beyond non-empty.

## Design

Version-source detection (in the Runner from doc 02, implemented here),
first hit wins: (1) `gradle.properties` with both keys (existing behavior,
code+1); (2) `version = "…"` / `version = '…'` assignment in
`app/build.gradle.kts`, then root `build.gradle.kts` (write = replace the
literal, no code concept — report `code: ""`); (3) go `VERSION` file
(existing). `Info` gains `VersionSource string` (`properties|kts|file|none`)
so display logic stops guessing. The `Runner` interface (doc 02) gains one
additive method, `BumpCode() (string, error)`: properties implements code+1,
kts/go return a clear "unsupported" error. CLI and TUI call all three verbs
(`--set`, `bump`, `--code-only`) through the Runner; `release` needs no
further changes because `GradleRunner.VersionWrite` dispatches on
`VersionSource` internally.

New command shape (same `version` command, additive flags):

    version                    show (unchanged output per source)
    version --set X            write name only (gradle-properties also code+1, as today)
    version bump X             alias of --set with clearer intent for releases
    version --code-only        properties only: code+1, name untouched

Validation (all sources): reject empty/whitespace-only; reject embedded
whitespace/newlines; warn (not fail) when the name is not semver-like
(`X.Y.Z…​`, since Gradle names are free-form). kts write preserves quote
style and indentation; refuses when the assignment is not a plain literal
(unresolved → honest error naming file:line, per doc 06 data).

## Files to touch

- `internal/project/` (new `version.go`: move gradle + VERSION funcs out of
  `detect.go`, add kts read/write + source detection).
- `cmd/version.go` (`bump` positional-or-flag, `--code-only`).
- TUI `doVersion` (same verbs through the Runner).
- `cmd/release.go` (use source-aware write; behavior identical for
  properties/VERSION projects).

## Steps

1. Move existing version code to `version.go` untouched + tests move too.
2. Add kts literal read/write + source detection + tests (fixture kts).
3. Add validation helper + tests (empty, whitespace, non-semver warn).
4. Add `bump` + `--code-only` to CLI and TUI.
5. Rewire `release` to source-aware write; confirm properties/VERSION
  projects behave byte-identically.

## Tests to add

- kts read/write round-trip (double/single quotes, indentation kept).
- Non-literal kts version → clean error with file:line.
- Validation matrix; `--code-only` on kts/go → clear "unsupported" error.

## Acceptance checks

- `go vet ./...`, `go test ./...` pass.
- Fixture kts project: `version` shows source `kts`; `--set` rewrites only
  the literal; `version --code-only` errors cleanly.
- This repo: `version` output unchanged.

## Out of scope

Catalog/toml version sources, auto-commit (still forbidden), tag creation.
