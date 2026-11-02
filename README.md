# ReleaseForge

**Local release, build, test, and project-lifecycle TUI + CLI for any project on disk.**

Single-binary Go tool (Cobra CLI + Charm Bubble Tea TUI). Point it at a
folder; it scans, caches, and adapts — original, forked, or inspired repos
alike. Nothing is sent anywhere: config, caches, logs, builds and releases
live in a dedicated local data root.

> **Current state (2026-11-02):** the TUI has been rethought as a six-window
> shell with a fuzzy command palette, folder picker, generated help, an
> explicit run state machine and a responsive, themeable layout. The intent
> is unchanged. See [docs/STATE.md](docs/STATE.md),
> [docs/04-tui-design.md](docs/04-tui-design.md) and
> [docs/CORRECTIONS.md](docs/CORRECTIONS.md).
> **Watch the [demo](demo/index.html)** — 10 frames captured from the real
> `View()` output.

## Quick start

```bash
go build -o releaseforge .
./releaseforge init --non-interactive --data-root .data
./releaseforge scan /path/to/project       # or: ctrl+o inside the TUI
./releaseforge tui                        # also: ./releaseforge (no args)
./releaseforge commands                   # print the TUI command reference
```

In the TUI, type in the always-visible bar. `help` / `?` opens the generated
reference; `1..6` or `alt+1..6` switch windows; `ctrl+p` opens the fuzzy
palette; `ctrl+o` opens the folder picker; `ctrl+w` focuses a window so
arrows scroll it; `ctrl+x` cancels a run; `ctrl+c` quits.

## What it owns

- **Scan** any local folder: git, history, tools (Gradle, CMake, Wails, …),
  frameworks, configs, version sources — cached under the data root.
- **Local data root** (prefer `D:` on Windows) with per-project mirror cache,
  logs, builds, releases and reports.
- **Six TUI windows**: Status, Tests, Git, Console (live logs), Logs browser,
  Help. Each keeps its own scroll/cursor state.
- **Typed commands + palette**: tab completion, history search, recent
  projects (`open 2`), folder picker, generated help.
- **Async build/test/script** with streamed output, spinner, elapsed time,
  cancel, structured error reports and persisted logs.
- **Android APK pipeline**: sign with the data-root keystore, install via adb,
  devices, notes from git history. Reimplements and improves the legacy
  Python reference under [`scripts/`](scripts/) (see
  [`scripts/README.md`](scripts/README.md)).
- **Release notes**, tags and GitHub releases via the Cobra CLI.

## Interface at a glance

```
╭─ ReleaseForge · atlas-app · v0.0.4 · ready   ● idle   Status ───────────╮
│ main · clean · ▁▂▅  ·  bar focus  ·  theme dark                         │
╰────────────────────────────────────────────────────────────────────────╯
 1 Status   2 Tests   3 Git   4 Console   5 Logs   6 Help
╭─ (active window: facts, tables, logs, generated help, overlays …) ─────╮
╰────────────────────────────────────────────────────────────────────────╯
╭─ ❯ type a command · tab completes · ctrl+p palette · ctrl+o folder ─────╮
╰────────────────────────────────────────────────────────────────────────╯
enter run · tab complete · ctrl+p palette · ctrl+o folder · ? help · …
```

Built with `bubbletea`, `bubbles` (textinput, viewport, table, list,
filepicker, spinner, key, help), `lipgloss`, `huh` and `cobra`. Wide-rune
safe, `NO_COLOR` aware, exact geometry at 50–200 columns.

## Verification

```bash
go build ./...
go vet ./...
go test ./... -count=1     # 136 tests across all packages, all green
go run ./tools/demo -out demo
```

## Documentation map

| Doc | Purpose |
|-----|---------|
| [docs/STATE.md](docs/STATE.md) | **Honest current state**, architecture, verification, gaps, roadmap |
| [docs/CORRECTIONS.md](docs/CORRECTIONS.md) | Corrections log (16 fixes + baseline → now) |
| [docs/00-vision.md](docs/00-vision.md) | Goals, any-project scope, non-goals |
| [docs/01-architecture.md](docs/01-architecture.md) | Packages, data flow |
| [docs/02-project-types.md](docs/02-project-types.md) | Detection markers |
| [docs/03-release-pipeline.md](docs/03-release-pipeline.md) | Release steps + links to `scripts/` |
| [docs/04-tui-design.md](docs/04-tui-design.md) | **Current windowed interface**, keys, windows, theme |
| [docs/05-config-and-storage.md](docs/05-config-and-storage.md) | Data-root, cache |
| [docs/06-android-deep-dive.md](docs/06-android-deep-dive.md) | APK details from `scripts/` |
| [docs/07-wails-and-pc.md](docs/07-wails-and-pc.md) | Other project shapes |
| [docs/08-implementation-plan.md](docs/08-implementation-plan.md) | Original agent plan 0.0.1 → 0.0.2 |
| [docs/09-scan-foundation.md](docs/09-scan-foundation.md) | Scan algorithm + cache schema |
| [docs/v0.0.3/02-tui-redesign.md](docs/v0.0.3/02-tui-redesign.md) | Design note for this pass |
| [docs/sweep-2mo/README.md](docs/sweep-2mo/README.md) | Long-term planning archive (with Delta notes) |
| [RELEASE.md](RELEASE.md) | Step-by-step release guide |
| [demo/README.md](demo/README.md) | Demo contents + regeneration |
| [scripts/README.md](scripts/README.md) | Legacy Python reference index |

## License

MIT — see [LICENSE](LICENSE).
