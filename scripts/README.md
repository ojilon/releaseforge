# Legacy reference: Android release scripts (Python)

These files are a **working reference implementation** of an Android release pipeline (originally used with a Gradle Android app). They are **not** executed by ReleaseForge at runtime.

Implementers reimplement this behaviour in Go (`internal/android`, `internal/build`, `internal/git`, release orchestration), with improvements documented in `docs/`.

## Files

| File | Role |
|------|------|
| [release.py](release.py) | Orchestrator: version → unit tests → assembleDebug/Release → package → sign → notes → zip → `gh release create` |
| [version.py](version.py) | Read/set `app.versionCode` / `app.versionName` in `gradle.properties` (code increments by 1 on set) |
| [package.py](package.py) | Copy debug + unsigned release APKs into `release/<version>/` with renamed files |
| [sign.py](sign.py) | Locate `apksigner`, prompt keystore password, sign + verify |
| [notes.py](notes.py) | Create markdown template under `release_notes/v<version>.md` |
| [zip_release.py](zip_release.py) | Zip APKs in the version folder |
| [config.py](config.py) | Load optional `scripts/config.json` (keystore path, alias, gradle user home, apksigner path) |

## Assumptions in the Python code (important)

- Project root is the parent of `scripts/` (`ROOT = Path(__file__).resolve().parent.parent`).
- Debug APK: `app/build/outputs/apk/debug/app-debug.apk`
- Release unsigned: `app/build/outputs/apk/release/app-release-unsigned.apk`
- Gradle via `./gradlew` (or system `gradle` with a flag)
- GitHub publish via `gh release create`
- App display name in artifact filenames was hard-coded as `Conductino-Study-…` in package/zip — **Go port must use project config `artifacts.app_name` (or scan-derived name), not a fixed string**

## What to improve in Go

1. **Data root** — default artifacts/logs under ReleaseForge data-root; optional in-repo `release/` mirror only if config says so.
2. **Live logs** — stream Gradle/apksigner output to TUI/CLI and persist under `projects/<name>/logs/`.
3. **No hard-coded app names** — from project config / scan.
4. **Structured errors** — headline lines from Gradle failures + path to full log.
5. **Password** — prompt only; never store in config files.
6. **Multi-project** — same pipeline driven by per-project `config.json`, not a single repo layout.
7. **Notes** — fill from `git log` (see `docs/09-scan-foundation.md` / git helpers), not only empty templates.
8. **Windows** — `gradlew.bat`, path handling, apksigner `.bat` (already partially handled in `sign.py`).

## Mapping to Go packages

| Python | Go target |
|--------|-----------|
| `version.py` | `internal/project` version R/W for `android-gradle` |
| `package.py` | `internal/android` package helpers |
| `sign.py` | `internal/android` sign/verify |
| `notes.py` + git | `internal/git` notes generation |
| `zip_release.py` | `internal/android` or small archive util |
| `release.py` | Cobra `release` + shared orchestrator |
| `config.py` | `internal/config` project `signing` section |

Read the `.py` sources directly when implementing; do not depend on an external repo.
