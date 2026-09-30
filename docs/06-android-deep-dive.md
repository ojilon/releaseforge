# Android deep-dive (from real repos)

Knowledge extracted from **Conductino-Android** and **Wayer** so implementers do not need GitHub access.

## Conductino-Android

### Version (`gradle.properties`)

```
app.versionCode=4
app.versionName=0.0.3_2
ndkVersion=29.0.14206865
aurora.abiFilters=arm64-v8a
```

`scripts/version.py` behaviour:

- Read code + name via regex on `gradle.properties`.
- Set name to argument; set code to `current + 1`.

### App module (`app/build.gradle`)

- `compileSdk 36`, `minSdk 26`, `targetSdk 34` (properties also list higher targets).
- `ndkVersion "29.0.14206865"`.
- `versionCode` / `versionName` from project properties.
- ABI filters from `aurora.abiFilters` (comma-separated).
- External native build: CMake path `../backend/CMakeLists.txt`, version 3.22.1.
- Release: minify + ProGuard.
- Unit tests: Robolectric; `includeAndroidResources = true`.

### Native (`backend/CMakeLists.txt`)

- `option(BUILD_AURORA_CORE ... OFF)` — native core optional.
- C11, SQLite amalgamation, curl, lexbor when enabled.
- Shared lib `aurora_core` with JNI bridge.

### Existing Python release tool (`scripts/`)

| File | Role |
|------|------|
| `release.py` | Orchestrator: version → test → assembleDebug → assembleRelease → package → sign → notes → zip → `gh release create` |
| `version.py` | Read/set versionCode + versionName |
| `package.py` | Copy debug + unsigned release APKs into `release/<version>/` with renamed files |
| `sign.py` | Find apksigner, prompt password, sign + verify |
| `notes.py` | Create markdown template under `release_notes/v<version>.md` |
| `zip_release.py` | Zip APKs in the version folder |
| `config.py` | Load `scripts/config.json` (keystore_path, alias, gradle_user_home, apksigner_path) |

### apksigner resolution (sign.py)

1. Custom path from config if exists.
2. `shutil.which("apksigner")`.
3. `$ANDROID_HOME` or `$ANDROID_SDK_ROOT` → `build-tools/<newest>/apksigner(.bat)`.
4. Fallback command name.

Sign:

```
apksigner sign --ks <keystore> --ks-key-alias <alias> --ks-pass pass:<pwd> <apk>
apksigner verify --verbose <apk>
```

### Tests

```bash
./gradlew :app:testDebugUnitTest
./gradlew :app:connectedDebugAndroidTest
```

Docs: `docs/TESTING.md` (Robolectric unit tests, instrumented BrowserActivityTest).

### CI notes

GitHub Actions build debug + unsigned release, upload artifacts; separate auto-release workflow exists. Local tool is the control plane for signed releases and richer notes.

## Wayer (Android)

- Same Gradle + native pattern; native under `native/` (C++23).
- Multi-ABI default documented: `arm64-v8a, armeabi-v7a, x86_64`.
- `docs/RELEASE_AND_BUILD.md`, `keystore.properties.example`.
- Unit tests via `./gradlew :app:testDebugUnitTest`.

## Common environment (developer machine)

- Android SDK + command-line tools
- NDK (r26+ / 29.x as pinned)
- JDK 17
- Gradle wrapper preferred (`gradlew` / `gradlew.bat`)
- `adb` in PATH
- `gh` authenticated for publish
- Keystore file present (never commit)

## Error surfaces to parse

- Gradle compilation errors (`e: file:line`)
- Test failures (JUnit XML under `app/build/test-results/`)
- CMake / NDK link errors
- `UnsatisfiedLinkError` (native not built)
- apksigner verification failures

The build runner should capture full console output, persist it, and extract a short “headline” error list for the TUI card.
