# v0.0.2 doc 09: Android build/test for real

## Goal

Make Android builds actually configurable: `--abis` applied, instrumented
tests runnable, artifacts found via scan data instead of hard-coded paths.

## Current state

- `build --abis` is accepted and ignored (`cmd/build.go:65`, verbose-only
  note). ABI filters and NDK version are never read (audit §8).
- `test instrumented` shells `connectedDebugAndroidTest` with no device
  check — it just fails deep in Gradle.
- APK paths are string constants in two places (`internal/android/sign.go:
  PackageRelease`, `cmd/install.go` fallback). Scan knows nothing about
  where outputs appear.

## Design

`--abis`: when given, append `-Paurora.abiFilters=<value>` to the Gradle
argv (project properties flow into `build.gradle` without editing files —
verify against the `aurora.abiFilters` pattern from the examples; if a
project reads it another way, doc 06's unresolved-reporting surfaces that
and `--abis` errors honestly for that project). Also record the requested
ABIs in the log header line. Never modify `gradle.properties` automatically.

Instrumented tests: before running, call the `doctor` device check from doc
10 (no devices → immediate clear error naming the fix: connect device /
`adb devices`). Kind `all` runs unit first, instrumented second, one log
each (current single-log behavior splits — index from doc 05 records both).

Artifact discovery: doc 06 records `outputs` in scan (`apk` globs +
`test-results` dir). `PackageRelease` and `install` try scan-known paths
first, constants second. Constants stay as fallback, not primary.

## Files to touch

- `cmd/build.go`, `cmd/test.go` (`--abis` → `-P` flag; device precheck).
- `internal/project/gradle.go` (record `outputs` during scan).
- `internal/android/sign.go` (`PackageRelease` accepts candidate list).
- `cmd/install.go` (use scan candidates; keep newest-dir logic fixed in 01).
- Runner interface (02): `BuildArgs` gains abis passthrough.

## Steps

1. Add `-Paurora.abiFilters` passthrough + log-header record + tests (assert
   argv, no process spawned).
2. Add device precheck to instrumented path: call `android.Devices()`
   directly here (0 → immediate error naming the fix; 1+ → proceed and let
   Gradle/adb report the rest). Doc 10 later upgrades this call site to the
   full picker — mark it with a `// TODO(doc10): use PickDevice` comment.
3. Split `all` into two logged runs.
4. Scan-recorded outputs → package/install candidate lists.

## Tests to add

- Argv contains `-P…​` exactly when `--abis` given; never touches properties
  file (fixture project file hash unchanged).
- No-device instrumented fails before spawning (mock device list).
- Candidate-list packaging prefers scan path, falls back to constants.

## Acceptance checks

- `go vet ./...`, `go test ./...` pass.
- `build --abis arm64-v8a` log header shows the `-P` flag (no SDK needed to
  verify argv — check the `$` line in the log).
- `test instrumented` with no device fails in <2s with the doctor hint.

## Out of scope

Signing (11), `run`/logcat (10), NDK validation, flavor/dimension support.
No Gradle spawned except user-invoked commands.
