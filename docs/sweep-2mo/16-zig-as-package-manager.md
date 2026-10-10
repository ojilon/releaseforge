# 16 — ADVANCED: Zig as Package Manager (replaces CMake target building)

> Horizon: weeks 5–6. Zig owns **target building**; CMake becomes legacy
> fallback. Non-GC, fast, cacheable — see `32`.

## Why Zig (not CMake) for targets

- Real cross-compile: `-Dtarget=aarch64-linux-android` without toolchains.
- Content-hashed cache (`~/.cache/zig`) — incremental native builds are free.
- `build.zig` is code: ABIs × modes as a matrix, not Gradle flavour soup.
- One toolchain to update via `24` updater (`zig fetch`, version pin).

## Design

```
native/
  build.zig          ← libfoo static/shared per ABI×mode
  build.zig.zon      ← package manifest (deps pinned)
  src/*.zig (+ cinterop for existing C++)
```

- `rf zig-build --abis arm64-v8a,x86_64 --mode release` shells
  `zig build -Dtarget=… -Doptimize=ReleaseFast`, collects
  `zig-out/lib/libfoo.so` per ABI → feeds `14` handoff manifest.
- `build.zig.zon` is the package manager: `zig fetch --save <url>` pins
  native deps; scan records `zig_version`, `zon_deps[]`, per-target staleness.
- Interop first: `zig build-lib` compiles existing C/C++ (`addCSourceFiles`)
  — no rewrite needed to adopt.
- CMake stays as `--toolchain cmake` escape hatch until Zig parity.

## Snippet — build.zig sketch

```zig
const std = @import("std");
pub fn build(b: *std.Build) void {
    const target = b.standardTargetOptions(.{});
    const optimize = b.standardOptimizeOption(.{});
    const lib = b.addSharedLibrary(.{
        .name = "foo", .root_source_file = b.path("src/foo.zig"),
        .target = target, .optimize = optimize,
    });
    lib.addCSourceFiles(.{ .files = &.{ "src/native.c" }, .flags = &.{"-O2"} });
    lib.linkLibC();
    b.installArtifact(lib);
}
```

## Read (then implement)

- Zig guide + `zig build`: https://ziglang.org/learn/ ; build system:
  https://ziglang.org/learn/build-system/
- `build.zig.zon` / package manager: https://ziglang.org/learn/package-manager/
- Andrew Kelley talks on cross-compilation; `zig targets` JSON for ABI matrix.

## Tasks

- [ ] Vendor `zig` version pin in config; `doctor` checks `zig version`.
- [ ] `internal/zig/` runner: matrix build → collect `.so`s → manifest.
- [ ] `testdata/zig-hello/` fixture; `14` handoff accepts Zig or CMake `.so`s.

## Acceptance

```powershell
zig version; rf zig-build --abis aarch64-linux-android --mode release
# zig-out per ABI + manifest; Gradle packages with NO cmake block
```
