# Project types (detection)

ReleaseForge detects **any local directory**. Fork/origin status is irrelevant. Optional examples below are only illustrations; detection is marker-based.

## Detection order

Evaluate markers from the project root (and light one-level checks where noted). First strong match wins; secondary markers are still recorded in the scan report (e.g. Gradle project that also has CMake).

1. **Android Gradle** — (`settings.gradle` | `settings.gradle.kts`) and (`app/build.gradle` | `app/build.gradle.kts` | any `*/build.gradle` with Android plugin) and usually `gradle.properties` or Gradle wrapper.
2. **Wails** — `wails.json` present (typically with a Go module).
3. **Pure CMake** — root `CMakeLists.txt` (optional `CMakePresets.json`), without Android Gradle winning first.
4. **Python** — `pyproject.toml` or `requirements.txt` (or `setup.py`) with substantial project layout.
5. **Java / JVM CLI** — `pom.xml` or `build.gradle` without Android application plugin, or simple `src/main/java` layouts.
6. **Go module** — `go.mod` without Wails.
7. **Generic** — none of the above; still record git, file stats, and any scripts found; user supplies tasks via config.

Always also record:

- Whether `.git` exists (and basic history summary).
- Wrapper scripts (`gradlew`, `gradlew.bat`, `mvnw`, …).
- Lockfiles / manifests (`package.json`, `pnpm-lock.yaml`, `Cargo.toml`, …) as **framework/tool hints**, even when not the primary type.

## Android Gradle (+ optional NDK/CMake)

**Illustrative trees:** Conductino-Android, Wayer, and similar app modules elsewhere.

**Version source of truth (common pattern)**

- File: `gradle.properties`
- Keys: `app.versionCode=<int>`, `app.versionName=<string>`
- On set (Conductino-Android `scripts/version.py` behaviour to preserve): increment `versionCode` by 1; set `versionName` to the supplied value.

**Other properties often present**

```
ndkVersion=...
aurora.abiFilters=arm64-v8a   # or comma-separated list
```

**Standard tasks**

| Intent | Typical Gradle task |
|--------|---------------------|
| Unit tests | `:app:testDebugUnitTest` |
| Instrumented | `:app:connectedDebugAndroidTest` |
| Debug APK | `assembleDebug` |
| Release (unsigned) | `assembleRelease` |

**Artifact paths (typical)**

- Debug: `app/build/outputs/apk/debug/app-debug.apk`
- Release unsigned: `app/build/outputs/apk/release/app-release-unsigned.apk`

**Native**

- CMake may live at `backend/CMakeLists.txt`, `native/`, or under `app/src/main/cpp/`.
- NDK version and ABI filters must be readable and later editable from the tool.

**Signing**

- Keystore path + alias from project config; password prompted only at sign time.
- Locate `apksigner`: config path → PATH → `$ANDROID_HOME/build-tools/<newest>/`.

**Packaging convention (from Conductino-Android scripts)**

- Directory: `release/<version>/` (and/or data-root releases mirror)
- Names: `<AppName>-<version>-debug.apk`, `<AppName>-<version>-release.apk`
- Zip of APKs; notes file for GitHub release body

## Wails

- Version: `wails.json` → `info.productVersion` (or equivalent).
- Build: `wails build`; dev: `wails dev`.
- Frontend may use pnpm/npm; record lockfile in scan.

## CMake / C++

- Root `CMakeLists.txt`; build via configure + `cmake --build`.
- Version source declared in project config when non-obvious.

## Python / Java / Go / generic

- Config-driven tasks after scan seeds a `type` and suggested commands.
- Generic projects still get git history, file inventory, and recent-project tracking.

## Per-project config

After a successful scan, write:

- `<data-root>/projects/<name>/config.json` — stable tool config
- `<data-root>/projects/<name>/cache/scan.json` — last scan snapshot (refreshable)
- Optional repo-local `.releaseforge.json` seed (no secrets)

See `docs/05-config-and-storage.md` and `docs/09-scan-foundation.md`.
