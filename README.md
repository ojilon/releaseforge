# ReleaseForge

**Local release, build, test, and project-lifecycle TUI + CLI for Android and PC projects.**

Single-binary Go tool (Cobra CLI + Charm Bubble Tea TUI) that owns:

- Project scan & type detection (Gradle Android, Wails, CMake, Python, Java CLI, …)
- Version read/write in the correct source of truth
- Unit & instrumented tests with live logs and friendly error rendering
- Target-aware builds (ABI filters, NDK, Gradle tasks, `wails build`, CMake)
- APK packaging, signing (`apksigner`), install-to-device (`adb`)
- Release notes from git history (AI later)
- Tags + GitHub releases (`gh` / API)
- Live log streaming, build reports, command history & suggestions
- Configurable local data root (prefer `D:` on Windows) with install wizard
- Metrics / dashboard panes inside the TUI

Inspired by the interactive feel of Claude Code / Gemini CLI: persistent command bar, live updating panes, tabs that adapt to project state.

## Status

Foundation commit. Documentation is complete and implementation-ready. Core packages are stubbed; Phase 1 targets the Android Gradle path (Conductino-Android / Wayer).

## Quick start (once implemented)

```bash
go install github.com/ojilon/releaseforge@latest   # or build from source
releaseforge init          # first-run wizard → choose data root (e.g. D:/ReleaseForgeData)
releaseforge scan .        # detect project type + write local config
releaseforge tui           # interactive TUI
releaseforge build debug   # non-interactive CLI
releaseforge release 0.1.0 --pre
```

## Supported project families (non-forked)

| Family | Example repos | Detection markers |
|--------|---------------|-------------------|
| Android Gradle + optional NDK/CMake | Conductino-Android, Wayer | `settings.gradle`, `app/build.gradle`, `gradle.properties`, optional `backend/CMakeLists.txt` |
| Wails (Go + web frontend) | Conductino_ | `wails.json`, `main.go`, `frontend/` |
| Pure CMake / C++ tools | Android-Scaffold-Studio | root `CMakeLists.txt`, `CMakePresets.json` |
| Python companion | WayerPC | `*.py`, `requirements.txt` / `pyproject.toml` |
| Java CLI | Foundain | simple Java sources |
| Language practice | practice_go, practice_zig, practice_odin | language-specific |

All share similar needs: version tracking, test/build commands, artifact collection, release notes, tags.

## Architecture (short)

- **Cobra** — CLI surface for scripting and complex one-shot tasks (`build`, `test`, `release`, `install`, `scan`, `version`, `logs`).
- **Bubble Tea + Lip Gloss + Bubbles + Huh** — full interactive TUI with persistent command input, live log viewport, adaptive tabs, forms/wizards.
- **internal/** packages own domain logic (project detection, Gradle runner, signing, git, GitHub, storage, metrics). TUI and CLI are thin adapters.

See `docs/` for the complete specification.

## Documentation map

| Doc | Purpose |
|-----|--------|
| [docs/00-vision.md](docs/00-vision.md) | Goals, non-goals, UX principles |
| [docs/01-architecture.md](docs/01-architecture.md) | Packages, data flow, Cobra + TUI split |
| [docs/02-project-types.md](docs/02-project-types.md) | How every project family is detected and configured |
| [docs/03-release-pipeline.md](docs/03-release-pipeline.md) | End-to-end release steps |
| [docs/04-tui-design.md](docs/04-tui-design.md) | Claude-Code-like UX, tabs, command bar, live logs |
| [docs/05-config-and-storage.md](docs/05-config-and-storage.md) | Data-root layout, wizard, schemas |
| [docs/06-android-deep-dive.md](docs/06-android-deep-dive.md) | Real Gradle/CMake/signing/adb knowledge from Conductino-Android & Wayer |
| [docs/07-wails-and-pc.md](docs/07-wails-and-pc.md) | Wails, CMake, Python surfaces |
| [docs/08-implementation-plan.md](docs/08-implementation-plan.md) | Phased build order |
| [docs/reference/](docs/reference/) | Command cheatsheets |

## License

MIT — see [LICENSE](LICENSE).
