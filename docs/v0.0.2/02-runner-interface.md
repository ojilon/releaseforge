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
        Program() string                    // binary to exec (wrapper path / "go")
        TestSpecs() []TestSpec              // valid kinds + argv
        TestArgs(kind string) ([]string, error)
        BuildArgs(variant, out string) ([]string, error)  // out ignored by gradle;
                                                          // go stamps into it
        BuildOutput(dataRoot, variant string) string      // go: builds/ path; gradle: ""
        BinaryBaseName() string             // adopted from doc 01 helper
        ArtifactGlobs(variant string) []string  // repo-relative outputs (gradle APKs;
                                                // go: nil until doc 09 extends)
        VersionRead() (code, name string, err error)
        VersionWrite(name string) (code string, err error)
    }
    func For(info Info) (Runner, error)  // GradleRunner / GoRunner, else "unsupported"

Dispatch sites use a type switch (`switch r.(type)` with exported
`GradleRunner`/`GoRunner`), never a type-name string comparison, so `rg
"android-gradle"` stays out of `cmd` and `internal/app` entirely — including
`scan.go` (android seed config moves to `SeedConfig(info)` here) and
`install.go` (gradle-only gate via type assertion). `cmd` keeps output
formatting by branching on `code != ""` (gradle returns a code, go does not),
not on type.

`cmd` and TUI call `project.For(info)` once, then the interface. Task strings
(`:app:testDebugUnitTest`, …) move into the gradle runner; the `:app:` prefix
stays (decided). `Program()` for gradle keeps `build.GradleWrapper` semantics.
No process changes in this doc — `build.Run` signature untouched.

## Files to touch

- New `internal/project/runner.go` (+ test): interface, two exported
  runners, `For`, and `SeedConfig(info)` (first-scan `config.ProjectConfig`
  defaults; needs the `config` import — no cycle: config imports only storage).
- `cmd/build.go`, `cmd/test.go`, `cmd/version.go`, `cmd/release.go`
  (`releaseAndroid/releaseGo` keep orchestration, take a Runner),
  `cmd/scan.go` (use `SeedConfig`), `cmd/install.go` (gradle gate via type
  assertion), `internal/app/app.go` (`doBuild/doTest/doVersion` thin wrappers).
- Doc 01's `BinaryBaseName` becomes a Runner method (keep the package func;
  the method delegates).

## Steps

1. Add `runner.go` with the interface + two implementations + `For` +
   `SeedConfig`.
2. Rewire `cmd/build.go` and `cmd/test.go` first (smallest switches).
3. Rewire `cmd/version.go` (branch output on `code != ""`), `cmd/scan.go`
   (`SeedConfig`), `cmd/install.go` (type assertion), then
   `releaseAndroid/releaseGo` internals.
4. Rewire TUI `doBuild/doTest/doVersion` to the same calls.
5. Delete the now-duplicate task-string literals, `goBinaryOut`, and
   `projectVersionName`.

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
