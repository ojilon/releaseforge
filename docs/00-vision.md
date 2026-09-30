# Vision

## Why ReleaseForge exists

Manual release work across multiple Android and PC projects is error-prone and repetitive:

- Version bumps live in different files (`gradle.properties`, `wails.json`, …).
- Debug APKs are easy; signed release APKs, correct ABIs, packaging, notes and GitHub releases are not.
- Gradle errors and build reports are hard to read in a plain terminal.
- Installing a just-built APK to a phone is a separate ad-hoc step.
- Release notes are written from memory instead of git history.
- Heavy artifacts (APKs, logs, reports) should not live on the system drive.

ReleaseForge is a **local, single-binary** tool that owns this lifecycle for all non-forked projects under the user’s account, with a professional interactive experience.

## Goals

1. **One tool for many project shapes** — Android Gradle (+ NDK/CMake), Wails, pure CMake, Python, simple Java CLIs, language practice repos.
2. **Correct builds** — honour ABI filters, NDK version, Gradle tasks, Wails output paths; surface failures clearly.
3. **Full release pipeline** — version → test → build → package → sign → notes → tag → GitHub release.
4. **Install-to-device** — after a successful debug or signed release, push to phone via adb.
5. **Human-friendly logs** — live streaming, colour, parsed Gradle errors, persisted reports under a chosen data root.
6. **Claude-Code / Gemini-CLI style UX** — persistent command bar, live panes, adaptive tabs, command history and suggestions.
7. **User-controlled storage** — first-run wizard chooses drive/path (prefer D:); tool manages builds, logs, releases, cache there.
8. **Scriptable + interactive** — Cobra CLI for automation and complex one-shots; Bubble Tea TUI for daily work.
9. **Git-history notes first** — AI-assisted notes are a later enhancement, not a blocker.

## Non-goals (v1)

- Replacing CI (GitHub Actions still useful); this is the **local** control plane.
- Hosting a web UI or Electron shell (pure terminal).
- Managing secrets beyond keystore password prompts and env-based tokens.
- Multi-user / team server mode.

## UX principles

- Keyboard-first, dense but readable.
- Always-visible command input; results and live logs update elsewhere.
- Never hide raw logs — make them easy to scroll and re-open.
- Prefer confirmation for destructive or irreversible steps (sign, tag push, release publish).
- Failures are first-class: show the failing task, the key error lines, and the path to the full log/report.
