# Vision

## Why ReleaseForge exists

Working across many local projects (original, forked, or inspired by upstreams such as a-java-ide → Java_ide, plus Android/desktop tools) makes the same chores repeat:

- Finding what a folder *is* (Gradle? CMake? Wails? plain Go?)
- Reading git history and relating it to a release
- Version bumps in different files
- Debug builds are easy; signed releases, correct ABIs, notes, and GitHub releases are not
- Build errors and reports are hard to read in a plain terminal
- Heavy artifacts should not clutter the system drive or the git tree

ReleaseForge is a **local, single-binary** control plane: point it at **any folder on disk**, scan it, cache what it finds, and grow into test → build → package → release. Ownership or fork status does not matter.

## Principles

1. **Any local project** — no whitelist of “our” repos. Detection is by filesystem markers and optional config, not by GitHub owner.
2. **Scan first** — understanding the tree (git, tools, frameworks, configs, version sources) is the foundation of every later action.
3. **Local mirror storage** — a dedicated data root (prefer non-system drive, e.g. `D:/ReleaseForgeData`) holds config, per-project cache, logs, builds, releases. The project git repo stays clean.
4. **Cache that ages with the project** — scan results are stored and refreshed; the tool remembers recent folders and last-known state.
5. **Absorb existing automation** — the Conductino-Android Python `scripts/` release pipeline is a concrete reference to reimplement and improve in Go, not throw away.
6. **CLI + TUI** — Cobra for scripts and complex flags; Bubble Tea for daily interactive use (OpenCode-like folder open, persistent command bar, help via `-h`).
7. **Bootstrap itself** — v0.0.1 is usable for scan/navigate; v0.0.2 can release ReleaseForge itself.

## Goals (full product)

1. Scan any local path: git presence, history summary, build systems, frameworks, key config files, version location.
2. Native folder picker (Windows) + list of recently opened projects.
3. Correct builds when configured (ABI/NDK, Gradle tasks, Wails, CMake, …).
4. Full release pipeline: version → test → build → package → sign → notes → tag → GitHub release.
5. Install-to-device for Android APKs via adb.
6. Human-friendly live logs and persisted reports under the data root.
7. Claude-Code / Gemini-CLI style UX: command bar, live panes, adaptive tabs, history, **in-app help**.
8. Scriptable Cobra surface alongside the TUI.

## Non-goals (near term)

- Replacing CI (GitHub Actions remain useful).
- Web UI / Electron (pure terminal).
- Multi-user server mode.
- Storing keystore passwords or secrets in plaintext.

## UX principles

- Keyboard-first; dense but readable.
- Always-visible command input; results and live content update in panes above.
- Typing `-h`, `help`, or `?` in the command bar shows command help in the display pane (not only Cobra `--help` on the outer process).
- Open project via native OS folder dialog when available; fall back to path entry.
- Never hide raw logs — easy to scroll and re-open from the data root.
- Confirm destructive steps (sign, tag push, publish).
- Failures are first-class: failing task, headline errors, path to full log/report.

## Relationship to existing Python release scripts

Conductino-Android ships `scripts/release.py` and helpers (`version`, `package`, `sign`, `notes`, `zip_release`, `config`). ReleaseForge should:

- Reproduce that pipeline in Go with the same observable outcomes (version bump rules, APK layout, apksigner, `gh release create`).
- Improve it: live logs, central data-root artifacts, multi-project config, TUI confirmations, better error parsing, no dependency on ad-hoc Python for day-to-day use.
- Remain compatible with optional in-repo `release/` mirrors when a project config asks for them.
