# 17 — Zig Build Targets Matrix (ABIs × modes)

Companion to `16`. Concrete matrix + caching contract.

## Matrix

| ABI | Zig target | Notes |
|-----|-----------|-------|
| arm64-v8a | `aarch64-linux-android` | primary |
| armeabi-v7a | `arm-linux-androideabi` | 32-bit compat |
| x86_64 (emu) | `x86_64-linux-android` | emulator/test |
| host tools | `x86_64-windows-gnu` / `native` | codegen helpers |

Modes: `Debug` (tests) / `ReleaseSafe` (store) / `ReleaseFast` (bench).
Default store = `ReleaseSafe` + `llvm-strip` + sha manifest.

## Caching

- Zig cache (`-Dcache`) under data-root `native-cache/zig/` (override
  `ZIG_GLOBAL_CACHE_DIR`) so `D:` holds weight, not `C:`.
- Manifest (`native-manifest.json`): `{abi, target, mode, sha256, zig_version,
  src_hash}`. `gradle-sync-libs` refuses mismatched `zig_version`.
- `scan` records `zig.targets[]` + per-target `stale: bool`
  (src mtime/hash vs `.so`).

## Snippet — matrix driver (Go)

```go
targets := map[string]string{
    "arm64-v8a": "aarch64-linux-android",
    "x86_64":    "x86_64-linux-android",
}
for abi, triple := range targets {
    run("zig","build","-Dtarget="+triple,"-Doptimize="+mode)
    collect("zig-out/lib/libfoo.so", "out/"+abi+"/libfoo.so")
}
```

## Read

- `zig targets` output; `ZIG_GLOBAL_CACHE_DIR` semantics.
- Reproducible builds: `-Dstrip`, hash pinning.

## Tasks

- [ ] Matrix table in config (`native.abis[]`, `native.mode`).
- [ ] Cache dir override + `scan --explain` staleness per ABI.
- [ ] Bench: Zig vs CMake clean/incremental times into `metrics`.

## Acceptance

```powershell
rf zig-build --abis arm64-v8a,x86_64 --mode ReleaseSafe
cat native-manifest.json
```
