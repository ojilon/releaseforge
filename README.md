# ReleaseForge

**Local release, build, test, and project-lifecycle TUI + CLI for any project on disk.**

Single-binary Go tool (Cobra CLI + Charm Bubble Tea TUI). Point it at a folder; it scans, caches, and adapts — original, forked, or inspired repos alike.

## What it owns (full vision)

- **Scan** a local folder: git, history, tools (Gradle, CMake, Wails, …), frameworks, configs, version sources
- **Local data root** (prefer `D:` on Windows) with per-project mirror cache
- Folder picker + history of recently opened projects
- Tests, builds, packaging, signing, install-to-device (roadmap)
- Release notes from git → tags → GitHub releases
- Live logs, command history, in-app help (`-h` / `help`)
- Android APK pipeline: reimplement and improve the reference under [`scripts/`](scripts/) (see [`scripts/README.md`](scripts/README.md))

## Version roadmap

| Version | Goal |
|---------|------|
| **0.0.1** | Local storage + scan + navigate (picker, git, tools, cache, help) |
| **0.0.2** | Self-bootstrap release + start Go port of `scripts/` Android pipeline |
| later | Full sign/install, multi-type builds, metrics, optional AI notes |

## Quick start (target after 0.0.1)

```bash
go build -o releaseforge.exe .
./releaseforge init
./releaseforge tui
# open → folder picker → scan → type help
```

## Documentation map

| Doc | Purpose |
|-----|--------|
| [docs/00-vision.md](docs/00-vision.md) | Goals, any-project scope |
| [docs/01-architecture.md](docs/01-architecture.md) | Packages, data flow |
| [docs/02-project-types.md](docs/02-project-types.md) | Detection markers |
| [docs/03-release-pipeline.md](docs/03-release-pipeline.md) | Release steps + **links to `scripts/`** |
| [docs/04-tui-design.md](docs/04-tui-design.md) | Command bar, help, open |
| [docs/05-config-and-storage.md](docs/05-config-and-storage.md) | Data-root, cache |
| [docs/06-android-deep-dive.md](docs/06-android-deep-dive.md) | APK details from **`scripts/`** |
| [docs/07-wails-and-pc.md](docs/07-wails-and-pc.md) | Other project shapes |
| [docs/08-implementation-plan.md](docs/08-implementation-plan.md) | Agent plan 0.0.1 → 0.0.2 |
| [docs/09-scan-foundation.md](docs/09-scan-foundation.md) | Scan algorithm + cache schema |
| [scripts/README.md](scripts/README.md) | **Legacy Python reference index** |

## License

MIT — see [LICENSE](LICENSE).
