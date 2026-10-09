# v0.0.3 doc 01: CLI UX fixes from field use (no code yet)

## Goal

Fix the three UX defects found running installed v0.0.1 against real
workflows: silently-ignored path arguments, `init` resolving a different
default root than every other command, and no memory of the last scanned
project. Docs only in this step; implementation comes after approval.

## Current state

- `init` has no Cobra `Args` validator (`cmd/init.go`), so
  `releaseforge init D:\projects\releaseforge` silently ignores the path and
  initializes storage instead. Same silent-ignore class: `test <path>` and
  `build <path>` parse the path as kind/variant and fail with a confusing
  message (`version <path>` already errors via its usage guard).
- `init` defaults to `storage.DefaultRoot()` (`cmd/init.go`), bypassing the
  exe-sibling discovery in `config.DiscoverDataRoot` that all other commands
  use. Result on an installed machine: `init` reported `D:\ReleaseForgeData`
  while `scan`/`status` used `D:\Dev\ReleaseForge\data` — two live roots.
- Every command defaults to `./` (`cmd/root.go: -p/--project, "."`). `scan`
  writes cache but changes nothing about what bare `status`/`test` look at,
  and nothing tells the user. Field symptom: `status` run from `$HOME`
  reported a `generic` project instead of the just-scanned repo.
- Minor: `status` prints the global `config:` path, not the project config.

## Design

1. **Reject stray positionals.** `init` gets `Args: cobra.NoArgs` (clear
   error naming `--data-root`). `test`/`build` keep their `[kind]`/`[variant]`
   arg but their unknown-value errors append `(did you mean -p <path>?)`.
   `version` keeps its usage error; verify it already fires for a lone path.
2. **One default root everywhere.** `init`'s wizard default becomes
   `DiscoverDataRoot("", "")` (flags → env → exe-sibling → machine default)
   instead of `DefaultRoot()`. Re-running `init` on an installed machine
   then reports the install's data dir as already initialised. No other
   `init` behavior changes.
3. **Remember the last project — explicitly, never magically.** `scan`
   writes `history/last-project.json` (`{path, name, at}`). New helper
   `currentProjectDir(cmd)` in `cmd`: `--project` explicitly set (via
   `cmd.Flags().Changed("project")`) wins; else if cwd detects to a
   non-`generic` type, cwd wins silently; else if a last-project exists,
   use it AND print `using last project <path> (pass -p to override)`;
   else cwd (fails naturally downstream). TUI startup applies the same rule
   when `-p` is default. `status` prints the project `config.json` path.
4. Nothing here changes scan/build/test/release semantics, storage layout,
   or version handling.

## Files to touch

- `cmd/init.go` (NoArgs + discovered default), `cmd/root.go`
  (`currentProjectDir` helper), `cmd/test.go` + `cmd/build.go` (hint text),
  `cmd/scan.go` (write last-project), `cmd/status.go` (project config path),
  `internal/history/` (last-project load/save + test), TUI startup path.

## Steps

1. `history/last-project.json` load/save + tests (missing file, corrupt file).
2. `init`: NoArgs + discovered default; test default resolution with a
   seeded install layout (reuse `dataRootFromExePath` semantics).
3. `scan` writes last-project on success.
4. `currentProjectDir` + wire into all project commands; tests for the
   precedence matrix (explicit flag > non-generic cwd > last-project+notice >
   cwd fallback).
5. `test`/`build` hint text; `status` project-config path.
6. TUI startup fallback when `-p` is default.

## Tests to add

- Last-project round-trip, missing/corrupt tolerance.
- Precedence matrix via temp dirs (generic cwd + seeded last-project, etc.).
- `init <path>` errors; discovered default prefers exe-sibling layout.
- Notice line asserted on stderr/stdout in fallback case.

## Acceptance checks

- `go vet ./...`, `go test ./...` pass.
- `releaseforge init D:\x` → immediate usage error, nothing created.
- Fresh machine flow: install → `scan <repo>` → bare `status` from `$HOME`
  prints the `using last project …​` notice and the repo's data.
- `init` on the installed machine reports the install's data dir, no second
  root.

## Out of scope

Shell completions, `cd` integration, changing scan/build/test semantics,
version bump, release.
