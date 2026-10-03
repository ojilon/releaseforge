# v0.0.2 doc 06: Android scan (Gradle deep parse)

## Goal

Scan an Android Gradle project deeply enough that builds, versions, and ABIs
stop being hard-coded: every value recorded with its source file and line,
unknowns reported honestly instead of guessed.

## Current state

`project.Scan` (`internal/project/scan.go:107-195`) does existence checks
only, plus three content reads: `app/build.gradle` substring for the android
plugin, `gradle.properties` substring for `ndkVersion`, `wails.json` product
version. It never parses `build.gradle[.kts]` structure, wrapper properties,
`local.properties`, `libs.versions.toml`, or the manifest. CMake search covers
root/`backend`/`native` but not `app/src/main/cpp/`. ABI filters and NDK
version are invisible to the tool.

## Design

New `internal/project/gradle.go` (stdlib only: `bufio/regexp/encoding/json`):

- `gradle.properties`: full `key=value` map with line numbers (skip `#`
  comments/blank lines); known keys surfaced: `app.versionCode`,
  `app.versionName`, `ndkVersion`, `*.abiFilters`.
- `settings.gradle[.kts]`: `include` module list (`include ":app"`).
- `app/build.gradle[.kts]`: `namespace`/`applicationId`, `compileSdk`,
  `minSdk`, `targetSdk`, `ndkVersion`, `abiFilters` block, `externalNativeBuild
  cmake path …​`, `versionCode`/`versionName` source (literal vs `property()`).
  Groovy + Kotlin syntaxes via tolerant regexes, not a grammar.
- `gradle/wrapper/gradle-wrapper.properties`: `distributionUrl` (Gradle
  version).
- `local.properties`: `sdk.dir` presence (value redacted to `<set>` in
  scan.json — machine-local path, not a secret but not portable).
- `gradle/libs.versions.toml`: `[versions]` table, raw lines.
- `AndroidManifest.xml` basics: `package` attribute if present + launcher
  activity presence (for doc 10's launch step).
- Every extracted value stored as `{value, file, line}`; anything written as
  `property()/extra[]/project.` reference is recorded as
  `{value:"", unresolved:"<expr>", file, line}` — never guessed.
- CMake search: current three locations + `app/src/main/cpp/**` and
  `app/src/**/CMakeLists.txt`, bounded (max depth 4 from root, max 20 hits).
- `scan --deep`: runs `gradlew :app:properties` and stores stdout raw under
  `cache/properties.txt`. Explicit user opt-in only (spawns Gradle; Celeron
  rule). Never run automatically.

Results extend `Snapshot` additively: `gradle` object
(`properties{}, modules[], android{}, wrapper{}, manifest{}`), `cmakeFiles[]`.
Old readers ignore unknown fields (JSON contract from `docs/09`).

## Files to touch

- New `internal/project/gradle.go` (+ fixtures under `testdata/gradle/`).
- `internal/project/scan.go` (call parsers, extend `Snapshot`).
- `cmd/scan.go` (`--deep` flag; prints Gradle version + module count line).

## Steps

1. Properties parser (values + lines + comments skipped) + tests.
2. Groovy extractors (namespace/applicationId/sdks/ndk/abifilters/cmake
   path/version source) + fixture tests (groovy sample).
3. KTS extractors (same fields, `=`/`set()` syntax) + fixture tests.
4. Wrapper/local/toml/manifest readers + tests.
5. CMake bounded search + tests (temp tree).
6. `--deep` flag + raw capture; docs note it spawns Gradle.

## Tests to add

- Fixture-based: groovy sample, kts sample, properties with comments,
  unresolved references, missing files (all-empty result, no error).
- CMake search bounds (deep nesting beyond limit ignored; >20 capped).
- Redaction: `sdk.dir` value never lands in scan.json.

## Acceptance checks

- `go vet ./...`, `go test ./...` pass.
- Against a real Android tree (or fixture copy): `scan` prints module list
  + Gradle version; `scan.json` shows `gradle.properties:line` provenance
  for versionCode; unknown `property()` refs appear as `unresolved`, never
  as invented values.

## Out of scope

Writing any Gradle file (08 owns version writes), executing builds (09),
dependency-graph resolution, manifest merging.
