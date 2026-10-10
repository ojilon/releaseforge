# 14 — ADVANCED: Prebuilt `.so` Handoff (Gradle stops compiling C++)

> Horizon: weeks 5–6. Do NOT start before `13` + `15` are solid.
> Idea: ReleaseForge compiles native code itself; Gradle only packages
> the resulting `.so`s. Faster, reproducible, cacheable.

## Why

Gradle's `externalNativeBuild { cmake { … } }` recompiles C++ on every
flavour/ABI, hides flags, and couples Android plugin upgrades to NDK
breakage. Prebuilts invert control: native toolchain (CMake today, Zig
tomorrow — see `16`) produces `lib/<abi>/libfoo.so`; Gradle's
`sourceSets { main { jniLibs.srcDirs = ['libs'] } }` just picks them up.

## Design

```
native/ (CMake or Zig)
  └─ out/<abi>/libfoo.so   ← built + stripped + hashed by ReleaseForge
android/app/libs/<abi>/libfoo.so  ← copied (or jniLibs.srcDirs → native/out)
  └─ assembleRelease packages them; NO externalNativeBuild block
```

- `rf native-build --abis arm64-v8a,armeabi-v7a --mode release`:
  configure → build → strip (`llvm-strip`) → `sha256` manifest
  (`native-manifest.json`: abi → {file, sha, git_rev, flags}).
- `rf gradle-sync-libs`: copies manifest-pinned `.so`s into
  `app/libs/<abi>/`; writes `app/libs/MANIFEST.json`; Gradle task
  `preBuild` asserts manifest match (fail loudly on drift).
- `build.gradle` toggle: `ext.usePrebuiltLibs = true` removes
  `externalNativeBuild` block (keep commented reference + doctor check).
- Scan records `hints.prebuilt_so = true`, per-ABI `.so` presence + staleness
  (native src mtime vs `.so` mtime).

## Snippet — Gradle side

```groovy
android {
    sourceSets { main { jniLibs.srcDirs += ['libs'] } }
    // externalNativeBuild REMOVED when usePrebuiltLibs=true
}
preBuild.dependsOn verifyPrebuiltLibs // checks libs/MANIFEST.json sha256
```

## Read

- `jniLibs` + packaging: https://developer.android.com/studio/projects/gradle-external-native-builds (then do the OPPOSITE: prebuilts)
- `llvm-strip`, `sha256sum` man pages; reproducible-builds.org.
- Read `15` (CMake) and `16` (Zig) first — handoff is toolchain-agnostic.

## Tasks

- [ ] `internal/native/` package: build → strip → manifest → sync-libs.
- [ ] `verifyPrebuiltLibs` Gradle snippet + template in `configs/`.
- [ ] `build release --use-prebuilt` e2e; ABI-missing error message lists
      exactly which `.so` is stale.

## Acceptance

```powershell
rf native-build --abis arm64-v8a --mode release
rf gradle-sync-libs
releaseforge build release --use-prebuilt  # no C++ compile in Gradle log
```
