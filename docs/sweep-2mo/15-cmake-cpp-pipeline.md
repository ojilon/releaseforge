# 15 — CMake / C++ Pipeline (discovery → configure → build)

## Current state

`FindCmakeFiles` bounded search; `cmake` ToolHit; no configure/build runner;
Gradle owns C++ via `externalNativeBuild` (to be replaced by `14` later).

## Target (weeks 2–4, before prebuilts)

- Discovery: walk depth ≤3 for `CMakeLists.txt`, record
  `backend/`, `native/`, `app/src/main/cpp/` hits; parse
  `cmake_minimum_required` + `project(foo VERSION x)` for version hints.
- Configure: `cmake -S <src> -B <build>/<abi> -DCMAKE_BUILD_TYPE=Release
  -DANDROID_ABI=<abi> -DANDROID_PLATFORM=android-24
  -DCMAKE_TOOLCHAIN_FILE=$NDK/build/cmake/android.toolchain.cmake`
- Build: `cmake --build <build> --config Release --parallel`.
- Presets: honour `CMakePresets.json` when present (`--preset`).
- Scan stores `cmake_files[]`, `cmake_min_version`, `presets[]`.

## Snippet

```go
cfg := exec.Command("cmake","-S",src,"-B",buildDir,
    "-DCMAKE_BUILD_TYPE="+mode,
    "-DANDROID_ABI="+abi,
    "-DCMAKE_TOOLCHAIN_FILE="+ndkToolchain)
bld := exec.Command("cmake","--build",buildDir,"--config",mode,"--parallel")
```

## Read

- CMake tutorial + `cmake --help`: https://cmake.org/documentation/
- Android toolchain file: https://developer.android.com/ndk/guides/cmake#android_toolchain
- `CMakePresets.json` schema: https://cmake.org/cmake/help/latest/manual/cmake-presets.7.html

## Tasks

- [ ] `internal/cmake/` runner (configure+build) on top of `12` runner.
- [ ] Preset detection + `--preset` passthrough.
- [ ] Fixture C++ project under `testdata/cmake-hello/` for CI.

## Acceptance

```powershell
go test ./internal/cmake/ -v
releaseforge cmake --configure --abi arm64-v8a
```
