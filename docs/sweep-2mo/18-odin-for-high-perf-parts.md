# 18 — MATURE: Odin for High-Performance Non-Network Parts

> Horizon: near-complete (week 7+). Odin handles **hot, non-network,
> non-GC** code (diff hashing, log parsing, image/zip crunching). Go keeps
> networking, orchestration, TUI. See `32`.

## Why Odin (maybe, not must)

- Manual memory + SOA-friendly, fast compile, C-interop without cgo pain.
- Good fit: file walkers, sha pipelines, JUnit/Gradle log parsers,
  APK zip alignment checks — CPU-bound, parallelizable, no sockets.
- "Maybe": only if profiling proves Go is the bottleneck. Otherwise skip.

## Integration shape (no rewrite)

```
internal/hot/        ← Go interface + fallback pure-Go impl
  hot.go: HashFiles(), ParseGradleLog(), ZipAlignCheck()
native/odin/
  hot.odin           ← optional accelerated impl
  build.bat          ← odin build -build-mode:dll → hot.dll
```

- Go `hot.go` has build tag `purego` default; `odinhot` tag loads
  `hot.dll` via `syscall.NewLazyDLL` (Windows) / `dlopen` elsewhere.
- Runner auto-detects DLL; `doctor` reports `odin: dll|purego`; benchmarks
  decide default.
- Never put networking, TLS, or `gh` calls in Odin — Go owns I/O.

## Snippet — Go seam

```go
// internal/hot/hot.go
func HashFiles(files []string) (map[string]string, error) {
    if accel.Available() { return accel.HashFiles(files) }
    return purego.HashFiles(files) // sha256, parallel
}
```

```odin
// native/odin/hot.odin
package hot
import "core:crypto/hash"
@(export) hash_files :: proc() -> i32 { /* batched sha over mmap */ return 0 }
```

## Read

- Odin docs + `core:` libs: https://odin-lang.org/docs/overview/
- GingerBill intro + performance notes; `build-mode:dll` examples.
- Profile first: Go `pprof` (see `32`) — no Odin without numbers.

## Tasks

- [ ] `internal/hot/` seam + pure-Go impl + bench.
- [ ] Spike `hot.dll` for `HashFiles` only; compare bench.
- [ ] Decision log: keep/drop Odin per benchmark (write it in this file).

## Acceptance

```powershell
go test -bench=. ./internal/hot/ -benchtime=3x
# odin available → same results, ≥2x on 10k-file tree or DROP odin
```
