# Wails and other PC projects

## Conductino_ (Wails v2)

- `wails.json`: `name`, `outputfilename`, `info.productVersion`, frontend dir, build dir.
- Go module at repo root; frontend is pnpm + Vite + React/TS.
- Commands:
  - Dev: `wails dev` (and `pnpm dev` for pure browser mock)
  - Build: `wails build` → artifacts under `build/bin/`
- Version for releases: `info.productVersion` in `wails.json` (also mirrored in planning docs).
- Existing `docs/release-prep/` already discusses tags, local storage layout, installer drive choice, CI — ReleaseForge data-root design should stay compatible.

## Android-Scaffold-Studio

- Pure CMake project generating Android-related scaffolds.
- Root `CMakeLists.txt`, `cmake/` modules, `src/`, templates.
- Build via CMake presets / configure + build. Version strategy to be declared in project config (CMake variable or VERSION file).

## WayerPC

- Python companion to Wayer Android.
- Config-driven test/run commands; packaging is lightweight.

## Foundain / practice_*

- Minimal CLIs or language drills.
- Tag-based releases and simple “run tests” commands are enough for v1.

## Generic project interface

Every type implements:

- `Name()`, `Root()`, `Type()`
- `CurrentVersion() (code, name, error)`
- `SetVersion(name string) error`
- `Test(kind) → Result`
- `Build(variant, opts) → Result`
- `ArtifactPaths(variant) []string`
- Optional: `Sign`, `Install`, `Package`

Android and Wails provide full implementations; others start with config-driven command maps.
