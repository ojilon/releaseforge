# Project types

ReleaseForge must handle all non-forked repositories the user actively works on. Detection is by marker files (order matters).

## Detection order

1. **Android Gradle** — `settings.gradle` or `settings.gradle.kts` + `app/build.gradle` (or `.kts`) + `gradle.properties`
2. **Wails** — `wails.json` + Go module
3. **Pure CMake** — root `CMakeLists.txt` (and often `CMakePresets.json`)
4. **Python** — `pyproject.toml` or `requirements.txt` + substantial `*.py`
5. **Java CLI** — simple Java sources without Android markers
6. **Generic** — fallback: run user-defined commands from config

## Android Gradle (+ optional NDK/CMake)

**Examples:** Conductino-Android, Wayer

**Version source of truth**

- File: `gradle.properties`
- Keys: `app.versionCode=<int>`, `app.versionName=<semver or string>`
- On set: increment `versionCode` by 1, set `versionName` to the supplied value (same behaviour as existing `scripts/version.py`).

**Important properties (Conductino-Android / Wayer style)**

```
app.versionCode=4
app.versionName=0.0.3_2
ndkVersion=29.0.14206865
aurora.abiFilters=arm64-v8a          # or arm64-v8a,armeabi-v7a,x86_64
```

**Standard tasks**

| Intent | Gradle task |
|--------|-------------|
| Unit tests | `:app:testDebugUnitTest` |
| Instrumented | `:app:connectedDebugAndroidTest` |
| Debug APK | `assembleDebug` |
| Release (unsigned) | `assembleRelease` |

**Artifact paths (typical)**

- Debug: `app/build/outputs/apk/debug/app-debug.apk`
- Release unsigned: `app/build/outputs/apk/release/app-release-unsigned.apk`

**Native**

- CMake at `backend/CMakeLists.txt` or `native/` (Wayer).
- Option `BUILD_AURORA_CORE` may gate the native library.
- NDK version and ABI filters must be editable from the tool and re-applied on build.

**Signing**

- Keystore path + alias from project config (defaults historically: `conductino-release.jks`, alias `conductino`).
- Tool locates `apksigner` (custom path → PATH → `$ANDROID_HOME/build-tools/<newest>/`).
- Password prompted; never stored in plaintext.

**Packaging convention (from existing scripts)**

- Directory: `release/<version>/`
- Names: `<AppName>-<version>-debug.apk`, `<AppName>-<version>-release.apk`
- Zip of APKs; notes under `release_notes/v<version>.md`

## Wails (desktop)

**Example:** Conductino_

- Version: `wails.json` → `info.productVersion`
- Build: `wails build` (output under `build/bin/`)
- Dev: `wails dev` (frontend via pnpm)
- Frontend lockfile is `pnpm-lock.yaml` (prefer pnpm).
- Already has planning docs under `docs/release-prep/` (tags, local storage, installer drive) — ReleaseForge should align with those ideas for data-root and versioning.

## Pure CMake / C++ tools

**Example:** Android-Scaffold-Studio

- Root `CMakeLists.txt`, optional presets.
- Build: configure + `cmake --build`.
- Version may live in CMake variables or a VERSION file — config must declare the source.

## Python companion

**Example:** WayerPC

- Tests/build via `pytest` / custom scripts declared in config.
- Packaging may be a zip or simple copy of scripts.

## Java CLI / practice repos

**Examples:** Foundain, practice_go, practice_zig, practice_odin

- Lightweight: version file or tag-only releases, simple test/build commands from config.

## Per-project config

After `scan`, the tool writes a project config (under data-root and optionally a repo-local `.releaseforge.json`) describing type, version keys, tasks, artifact globs, signing, and GitHub owner/repo. See `docs/05-config-and-storage.md` and `configs/`.
