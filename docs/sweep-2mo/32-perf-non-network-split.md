# 32 — Performance Split: GC vs Non-GC (Go vs Zig/Odin)

> Principle for mature stages: **Go owns orchestration + network;
> Zig/Odin own hot non-network CPU work.** Never invert.

| Layer | Language | Examples |
|-------|----------|----------|
| CLI/TUI/orchestration, git, `gh`, updater, adb | Go (GC fine) | `cmd/`, `app/`, `git/`, `update/` |
| Target building, packaging, hashing | Zig | `16`, `17`, `14` handoff |
| Hot parsing/hashing/zip crunch (iff proven) | Odin (optional) | `18` DLL |

## Rule

No Odin/Zig without a Go `pprof`/`bench` number proving the win.
Instrument first:

```powershell
go test -bench=. -benchmem ./internal/hot/ ./internal/build/
go tool pprof -top <profile>
```

Threshold: ≥2× on realistic trees (10k files, 100MB logs) or stay Go.
Network/TLS stays Go ALWAYS (certs, proxies, auth — GC languages win).

## Read

- Go pprof: https://go.dev/blog/pprof ; bench: https://pkg.go.dev/testing
- Zig vs C benchmarks (take with salt); Odin `core:time` + context allocators.
- "The Performance of Open Source Applications" (free,aosabook.org) — measurement discipline.

## Tasks

- [ ] Bench harness for hash/parse/zip paths + CI bench baseline.
- [ ] Decision log in `18` (keep/drop Odin with numbers).

## Acceptance

```powershell
go test -bench=. ./internal/... -benchtime=3x  # baseline recorded
```
