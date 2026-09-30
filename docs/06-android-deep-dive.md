# Android deep-dive

Concrete APK release behaviour is defined by the **in-repo** Python reference under [`scripts/`](../scripts/). Read those sources when implementing Go. Summary below for orientation.

## Version (`gradle.properties`)

Implemented in [`scripts/version.py`](../scripts/version.py):

- Keys: `app.versionCode`, `app.versionName`
- Read via multiline regex
- Set: `versionName` ← argument; `versionCode` ← current int + 1

## Orchestration

[`scripts/release.py`](../scripts/release.py) steps:

1. `set_version(version)`
2. `testDebugUnitTest`
3. `assembleDebug`
4. `assembleRelease`
5. `package_release(version)` — see package.py
6. `sign_apk(...)` on the renamed release APK in `release/<version>/`
7. `create_notes(version, release_type)`
8. `create_zip(version)`
9. `gh release create v<version> --title ... --notes-file ...` (+ `--prerelease`) + attach `*.apk` and `*.zip`

Gradle command resolution: prefer `gradlew` / `gradlew.bat` under project root; optional system gradle.

## Package paths ([`scripts/package.py`](../scripts/package.py))

| Source | Destination pattern in Python |
|--------|-------------------------------|
| `app/build/outputs/apk/debug/app-debug.apk` | `release/<ver>/<Name>-<ver>-debug.apk` |
| `app/build/outputs/apk/release/app-release-unsigned.apk` | `release/<ver>/<Name>-<ver>-release.apk` |

Python hard-codes `Conductino-Study` as `<Name>`. **Go must use `artifacts.app_name` from project config (or scan).**

## Signing ([`scripts/sign.py`](../scripts/sign.py))

1. Custom apksigner path from config if set
2. `PATH`
3. `$ANDROID_HOME` / `$ANDROID_SDK_ROOT` → `build-tools/<newest>/apksigner(.bat)`
4. Fallback command name

```text
apksigner sign --ks <keystore> --ks-key-alias <alias> --ks-pass pass:<pwd> <apk>
apksigner verify --verbose <apk>
```

Password via interactive prompt only.

## Config ([`scripts/config.py`](../scripts/config.py))

Optional `scripts/config.json` keys: `keystore_path`, `keystore_alias`, `gradle_user_home`, `apksigner_path`. Map into ReleaseForge project `signing` + build settings in `config.json` under the data root — not into committed secrets.

## Notes & zip

- [`scripts/notes.py`](../scripts/notes.py) — markdown template only; improve with git log in Go
- [`scripts/zip_release.py`](../scripts/zip_release.py) — zip all `*.apk` in the version dir

## Typical Gradle / native context (apps this pipeline targets)

- Unit: `./gradlew :app:testDebugUnitTest`
- Instrumented: `./gradlew :app:connectedDebugAndroidTest`
- Often NDK + CMake (`backend/CMakeLists.txt` or `native/`), ABI filters in `gradle.properties`
- JDK 17, Gradle wrapper, `adb` for install (install is **not** in the Python scripts; add in Go `install` command)

## Error surfaces to parse in the Go runner

- Gradle compilation (`e: file:line`)
- Test failures (JUnit under `app/build/test-results/`)
- CMake / NDK link errors
- apksigner verification failures
