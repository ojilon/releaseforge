# v0.0.2 doc 02: Runner interface

## Goal

Collapse the six duplicated per-type `switch` statements (in `cmd/build.go`,
`cmd/test.go`, `cmd/version.go`, `cmd/release.go`, `app.doBuild`,
`app.doTest`, `app.doVersion`) into one `Runner` interface so adding Wails or
CMake later touches one file, not six.

## Current state

Type switches exist in: `cmd/build.go:58`, `cmd/test.go:50`,
`cmd/version.go:27`, `cmd/release.go:62` (dispatch only),
`internal/app/app.go:373,417,468` (`doVersion/doBuild/doTest`). All support
exactly `android-gradle` + `go` and error otherwise. Version ops already live
in `internal/project` (`GradleVersion/SetGradleVersion/ReadVersionFile/
SetVersionFile/CurrentVersion`); build args live half in `cmd` (task
strings) and half in `internal/build` (`GoTestArgs/GoBuildArgs`).

## Design

New file `internal/project/runner.go` (domain logic stays in `internal/`):

    type TestSpec struct { Kind string; Args []string }
    type Runner interface {
        TestSpecs() []TestSpec          // valid kinds + argv (gradle tasks / go test)
        BuildArgs(variant string) ([]string, error)  // variant validated here
        Program() string               // binary to exec (wrapper path / "go")
        BinaryBaseName() string        // from doc 01 helper
        ArtifactGlobs(variant string) []string  // where outputs appear, for 09/10
        VersionRead() (code, name string, err error)
        VersionWrite(name string) (code string, err error)
    }
    func For(info Info) (Runner, error)  // returns gradleRunner / goRunner or "unsupported"

`cmd` and TUI call `project.For(info)` once, then the interface. Task strings
(`:app:testDebugUnitTest`, …) move into the gradle runner; the `:app:` prefix
stays (decided). `Program()` for gradle keeps `build.GradleWrapper` semantics.
No process changes in this doc — `build.Run` signature untouched.

## Files to touch

- New `internal/project/runner.go` (+ test).
- `cmd/build.go`, `cmd/test.go`, `cmd/version.go`, `cmd/release.go`
  (`releaseAndroid/releaseGo` keep orchestration, take a Runner),
  `internal/app/app.go` (`doBuild/doTest/doVersion` thin wrappers).
- Doc 01's `BinaryBaseName` moves here as a method.

## Steps

1. Add `runner.go` with the interface + two implementations + `For`.
2. Rewire `cmd/build.go` and `cmd/test.go` first (smallest switches).
3. Rewire `cmd/version.go`, then `releaseAndroid/releaseGo` internals.
4. Rewire TUI `doBuild/doTest/doVersion` to the same calls.
5. Delete the now-duplicate task-string literals and `goBinaryOut`/
   `projectVersionName` shims (or keep as thin wrappers if churn is high —
   prefer deletion).

## Tests to add

- `For` returns correct runner per type; error for `generic`/`wails` (for now).
- Gradle runner: valid/invalid variant + kind matrix; go runner: kind ignored.
- Version round-trip through the interface on temp dirs.

## Acceptance checks

- `go vet ./...`, `go test ./...` pass.
- `rg -n "android-gradle" --glob '*.go' cmd internal/app` returns zero hits
  (all dispatch lives in `internal/project`).
- CLI behavior unchanged: `go run . build --help`, `test`, `version` output
  identical on this repo.

## Out of scope

Async execution (03), new project types, artifact discovery changes (09),
ABI handling (09).
