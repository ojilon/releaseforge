# ReleaseForge

**Local release, build, test, and project-lifecycle TUI + CLI for any project on disk.**

Single-binary Go tool (Cobra CLI + Charm Bubble Tea TUI). It does **not** care whether a repo is forked, original, or inspired by something else (Cooda, Java_ide from a-java-ide, Conductino-*, Wayer, practice repos, …). Point it at a folder; it scans, caches, and adapts.

## What it owns (full vision)

- **Scan** a local folder: detect git, history, build tools (Gradle, CMake, Wails, …), frameworks, configs, version sources
- **Local data root** (prefer `D:` on Windows) with per-project mirror cache that updates over time
- Folder picker (native Windows dialog) + history of recently opened projects
- Tests, target-aware builds, packaging, signing, install-to-device
- Release notes from git history → tags → GitHub releases
- Live logs, command history, in-app help (`-h` / `help` in the command bar)
- Absorb and improve the existing Conductino-Android Python `scripts/` release pipeline

## Version roadmap (bootstrapping)

| Version | Goal |
|---------|------|
| **0.0.1** | Local storage + scan + navigate. Pick a folder, detect git/tools/deps, show history, cache scan results, help in the TUI. Installable and usable immediately for exploration. |
| **0.0.2** | Self-bootstrap: ReleaseForge can package, test, and release **itself** (and start applying the same pipeline to Android/other projects). |
| later | Full Android sign/install, multi-type builds, metrics, optional AI notes. |

## Quick start (target after 0.0.1)

```bash
go build -o releaseforge.exe .
./releaseforge init          # first run: choose data root (e.g. D:/ReleaseForgeData)
./releaseforge tui           # or just ./releaseforge
# In TUI: open → native folder picker → scan runs → Overview + Git + Tools panes
# Type -h or help in the command bar for commands
```

## Architecture (short)

- **Cobra** — scriptable CLI (`scan`, `init`, `status`, later `build` / `release`, …)
- **Bubble Tea + Lip Gloss + Bubbles + Huh** — interactive TUI (command bar, panes, forms)
- **internal/** — domain only (storage, config, project scan, git, build, …). CLI and TUI are thin adapters.

## Documentation map

| Doc | Purpose |
|-----|--------|
| [docs/00-vision.md](docs/00-vision.md) | Goals (any local project), non-goals, UX |
| [docs/01-architecture.md](docs/01-architecture.md) | Packages, data flow |
| [docs/02-project-types.md](docs/02-project-types.md) | Detection markers (flexible, not limited to one org) |
| [docs/03-release-pipeline.md](docs/03-release-pipeline.md) | Full release steps + Python-tool absorption |
| [docs/04-tui-design.md](docs/04-tui-design.md) | Command bar, help, folder open, panes |
| [docs/05-config-and-storage.md](docs/05-config-and-storage.md) | Data-root, per-project cache/mirror |
| [docs/06-android-deep-dive.md](docs/06-android-deep-dive.md) | Gradle/CMake/signing from real Android work |
| [docs/07-wails-and-pc.md](docs/07-wails-and-pc.md) | Wails / CMake / Python / generic |
| [docs/08-implementation-plan.md](docs/08-implementation-plan.md) | **Agent follow this** — detailed 0.0.1 then 0.0.2 |
| [docs/09-scan-foundation.md](docs/09-scan-foundation.md) | Scan algorithm, cache schema, git + tools detection |
| [docs/reference/](docs/reference/) | Command cheatsheets |

## License

MIT — see [LICENSE](LICENSE).
