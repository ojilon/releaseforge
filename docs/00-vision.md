# Vision

## Why ReleaseForge exists

Working across many local projects (original, forked, or inspired by upstreams) makes the same chores repeat:

- Finding what a folder *is* (Gradle? CMake? Wails? plain Go?)
- Reading git history and relating it to a release
- Version bumps in different files
- Debug builds are easy; signed releases, correct ABIs, notes, and GitHub releases are not
- Build errors and reports are hard to read in a plain terminal
- Heavy artifacts should not clutter the system drive or the git tree

ReleaseForge is a **local, single-binary** control plane: point it at **any folder on disk**, scan it, cache what it finds, and grow into test → build → package → release. Ownership or fork status does not matter.

## Principles

1. **Any local project** — detection by filesystem markers and optional config, not by GitHub owner.
2. **Scan first** — understanding the tree (git, tools, frameworks, configs, version sources) is the foundation of every later action.
3. **Local mirror storage** — dedicated data root (prefer e.g. `D:/ReleaseForgeData`) for config, per-project cache, logs, builds, releases.
4. **Cache that ages with the project** — scan results stored and refreshed; recent folders remembered.
5. **In-repo legacy scripts as the APK reference** — [`scripts/`](../scripts/) holds the Python release pipeline to reimplement and improve in Go (see [`scripts/README.md`](../scripts/README.md)). No external repo required for that logic.
6. **CLI + TUI** — Cobra for scripts; Bubble Tea for interactive use (folder open, command bar, help via `-h`).
7. **Bootstrap itself** — v0.0.1 usable for scan/navigate; v0.0.2 can release ReleaseForge itself.

## Goals (full product)

1. Scan any local path: git, history, build systems, frameworks, configs, version location.
2. Native folder picker (Windows) + recently opened projects.
3. Correct builds when configured (ABI/NDK, Gradle, Wails, CMake, …).
4. Full release pipeline: version → test → build → package → sign → notes → tag → GitHub release.
5. Install-to-device for Android APKs via adb.
6. Human-friendly live logs and persisted reports under the data root.
7. Command-bar UX with **in-app help** (`-h` / `help`).
8. Scriptable Cobra surface alongside the TUI.

## Non-goals (near term)

- Replacing CI; web UI / Electron; multi-user server mode; storing keystore passwords in plaintext.

## UX principles

- Keyboard-first; always-visible command input; help in the main pane for `-h` / `help` / `?`.
- Open project via native OS folder dialog when available.
- Never hide raw logs; confirm destructive steps; failures are first-class.

## Relationship to `scripts/`

The Python modules under `scripts/` document a complete Android release flow (version, Gradle assemble, package APKs, apksigner, notes, zip, `gh release`). ReleaseForge’s Go code should match those outcomes where applicable and improve packaging paths, logging, configurability, and multi-project support. Runtime must not depend on invoking those `.py` files.
