# 13 — Gradle Deep Dive (Android + NDK + ABI)

## Current state

`internal/project/gradle.go` (`ScanGradle`), `runner.go` (`BuildArgs`,
`TestArgs`), `version.go` (kts R/W), `FindCmakeFiles`, `ndk` hint from
`gradle.properties`. Tasks: `:app:testDebugUnitTest`,
`:app:connectedDebugAndroidTest`, `assembleDebug/Release`. Artifacts under
`app/build/outputs/apk/*/`.

## What to harden (weeks 1–3)

- Wrapper resolution: `./gradlew[.bat]` > `gradle` on PATH; record version
  (`--version` parse) in scan.
- Flavours + buildTypes: parse `app/build.gradle[.kts]` for
  `productFlavors/buildTypes`; `build --flavor X` support.
- ABI filters: `ndkVersion`, `abiFilters` (`arm64-v8a`, `armeabi-v7a`,
  `x86_64`) from `gradle.properties` + `ndk{ abiFilters }` blocks; `--abis`
  flag already exists — verify end-to-end.
- Verbosity flags (`--info/--stacktrace`) plumbed to runner; structured
  error parsing stays in `28`.
- Daemon + caching: `--offline` mode, `org.gradle.caching` hint in doctor.

## Snippet — BuildArgs shape

```go
func (r GradleRunner) BuildArgs(variant, out, abis string) ([]string, error) {
    task := "assembleDebug"; if variant=="release" { task="assembleRelease" }
    args := []string{task, "--console=plain", "-q"}
    if abis != "" { args = append(args, "-Pandroid.injected.invoked.from.ide=false") }
    return args, nil
}
```

## Read

- Gradle Android plugin: https://developer.android.com/build
- NDK + ABI: https://developer.android.com/ndk/guides/abis
- CMake for Android: https://developer.android.com/ndk/guides/cmake
- Your `scripts/*.py` — version bump, package, sign arg shapes.

## Tasks

- [ ] Record wrapper version + flavours + ABIs in scan snapshot.
- [ ] `--abis` + `--flavor` e2e on a real Android tree.
- [ ] `doctor` checks: JDK, SDK, NDK, `ANDROID_HOME`, licenses.

## Acceptance

```powershell
releaseforge doctor; releaseforge build debug --abis arm64-v8a
```
