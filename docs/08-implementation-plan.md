# Implementation plan

## Phase 0 — Bootstrap (done in foundation commits)

- [x] Repo, LICENSE, README, go.mod
- [x] Cobra command tree (stubs)
- [x] internal/ package stubs
- [x] Full documentation
- [x] Example configs

## Phase 1 — Android path (highest value)

1. `storage` + `config` — data-root creation, load/save JSON.
2. `project` detect for Android Gradle; version read/write matching `scripts/version.py`.
3. `build.Runner` — run `gradlew` with live log stream + persist under data-root.
4. `test` and `build` commands wired for unit tests + assembleDebug/Release.
5. `android` package + sign (port sign.py / package.py logic) + adb install.
6. `git` notes from `git log` + `github` release via `gh`.
7. `release` command end-to-end for Conductino-Android.
8. Minimal TUI: command bar + log viewport + status header.

## Phase 2 — TUI completeness + second Android project

- Tabs (Overview, Build, Logs, Git, Metrics, Config).
- Huh forms for init, release confirm, device pick, password.
- Command history + suggestions.
- Support Wayer with same Android pipeline (multi-ABI defaults).

## Phase 3 — Wails + CMake + generic

- Wails version + `wails build`.
- CMake runner.
- Config-driven generic projects.

## Phase 4 — Intelligence & polish

- Richer notes from conventional commits / PR titles.
- Optional AI note polish (patterns from Conductino_ AI service).
- Metrics pane with simple bars/tables.
- Windows path robustness (D:, long paths, gradlew.bat).

## Phase 5 — Hardening

- Tests for version parsing, config, runner log capture.
- CI for ReleaseForge itself.
- Binary releases of the tool via its own pipeline.

## Definition of done (Phase 1)

From a clean checkout of Conductino-Android on a machine with SDK/NDK/JDK/`gh`:

```bash
releaseforge init                 # choose D:/ReleaseForgeData
releaseforge scan /path/to/Conductino-Android
releaseforge test unit
releaseforge build debug
releaseforge build release
releaseforge install debug
releaseforge release 0.0.4 --pre  # signs, notes, tag, GitHub release
```

All logs and artifacts appear under the data root; keystore password is prompted once per sign.
