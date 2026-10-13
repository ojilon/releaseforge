# 47 — Doctor Checks (every tool, every hint)

## Current state

`cmd/doctor.go:24` — 5 checks (adb/gh/java/wrapper/ANDROID_HOME) + devices.
First-line version strings; no SDK/NDK/CMake/Zig/Odin coverage.

## Target matrix

| Check | How | Fail hint |
|-------|-----|-----------|
| JDK 17+ | `java -version` stderr parse | winget/temurin link |
| Gradle wrapper | `gradlew --version` | wrapper jar present? |
| ANDROID_HOME/SDK | `sdkmanager --list_installed` | paths to set |
| build-tools | newest dir scan | `sdkmanager "build-tools;34.0.0"` |
| NDK | `ndkVersion` vs installed | exact install line |
| CMake/Ninja | `--version` | |
| Zig | `zig version` vs pin (`24`) | |
| Odin | `odin version` (mature only) | |
|Keystore| expiry + debug-key guard (`21`)| rotate cmd |
| Licenses | `sdkmanager --licenses` status | accept line |

`doctor --fix-dry` prints install lines without running; `doctor --json`
for scripting. Each hint is copy-pasteable on Windows (pwsh).

## Read

- sdkmanager: https://developer.android.com/tools/sdkmanager ; JDK compat matrix.

## Tasks

- [ ] Checks table (not inline ifs) + `--json` + `--fix-dry` + tests with fake PATH.

## Acceptance

```powershell
releaseforge doctor --json | ConvertFrom-Json
```
