# Implementation plan (current through v0.0.2)

> v0.0.2 status: the plan below is historical. v0.0.1 shipped (scan,
> storage, navigate, plus self-hosting release + installer), and v0.0.2
> followed the per-doc plan in `docs/v0.0.2/` (see `PROGRESS.md`, all
> ticked). The rules at the bottom still apply.

Follow this order. **Do not skip the scan foundation** to chase signing early.

**Android APK release behaviour:** read **[`scripts/`](../scripts/)** in this repository ([`scripts/README.md`](../scripts/README.md)). That is the complete legacy reference; no other repo is required.

---

## Version 0.0.1 — Scan, storage, navigate (SHIPPED)

Goal: installable binary the author can run daily to **open any local folder**, see what it is, see git history, and keep a cache under a data root (prefer `D:`).

### Done (all verified)

- [x] Repo scaffold, Cobra stubs, docs foundation
- [x] `internal/storage` — data-root paths, `EnsureLayout`, Windows `D:` default
- [x] `internal/config` — global/project load/save, `DiscoverDataRoot` (+ exe-sibling discovery)
- [x] `cmd/init` — layout + `global.json`, idempotent
- [x] **`scripts/`** — Python reference for later Android release port
- [x] Recent projects (`internal/history`) — `history/recent-projects.json`
- [x] Git read helpers (`internal/git`) — branch, head, commits, tags, origin
- [x] Scan (`internal/project`) — `docs/02-project-types.md` + `docs/09-scan-foundation.md`; `cache/scan.json`
- [x] CLI `scan` — human summary + cache path
- [x] Minimal TUI — command bar; `open`, `scan`, `recent`, `status`, `help`/`-h`, `quit` (typed paths only; folder picker deliberately deferred)
- [x] Tool version `0.0.1` in status/header
- [x] Beyond plan: Go test/build runners, `release` (go+android), `notes`, `logs`, `install`, installer with seeded data root, self-hosted v0.0.1 tag + GitHub pre-release

### 0.0.1 definition of done (all passed)

```bash
releaseforge init
releaseforge scan D:\path\to\any\repo
releaseforge tui
# open folder, see git + tools, type help, reopen from recent
```

---

## Version 0.0.2 — Solid Android Gradle support + a real TUI (SHIPPED)

Implemented per-doc in `docs/v0.0.2/` (`00`–`15`, `PROGRESS.md` all ticked):

- Cleanup (install dir-pick bug, dead code, binary naming, UTF-8 truncation,
  dropped `mirror_release_dir_in_repo`), Runner interface, async
  process/log-stream with cancel+timeouts, structured Gradle/CMake/NDK error
  parsing + JUnit summaries, `logs show|last|tail` + Gradle verbosity flags,
  Gradle deep scan + bounded CMake search, readable scan cache + config merge
  + seed, kts version R/W + `bump`/`--code-only`, real `--abis` + instrumented
  precheck + scan-driven artifacts, adb `doctor`/`run`/`logcat` + device
  picker, apksigner verification + `release --dry-run` + tag safety + notes
  in zips, grouped/bounded changelogs + upstream tracking, async TUI
  architecture (ring buffer, persisted history), bordered panels + state pill
  + sparkline, metrics capture + `last build` in status.
- This old 0.0.2 sketch below is superseded; kept for archeology. The
  `scripts/` → Go port table is done except live-hardware acceptance
  (real SDK/device runs).

### Original 0.0.2 sketch (historical, do not execute)

### Self-host

1. Generic build runner + log capture to data-root.
2. This repo as a Go project: `go test ./...`, `go build`.
3. Notes from git; tag + `gh release create` with binary.

### Port [`scripts/`](../scripts/) to Go (improved)

| Python | Go target | Improve |
|--------|-----------|--------|
| [`version.py`](../scripts/version.py) | android-gradle version R/W | same semantics |
| [`package.py`](../scripts/package.py) | `internal/android` | data-root output; **configurable app name** |
| [`sign.py`](../scripts/sign.py) | `internal/android` | same locate/sign/verify; password prompt |
| [`notes.py`](../scripts/notes.py) | `internal/git` | fill from git log |
| [`zip_release.py`](../scripts/zip_release.py) | archive helper | data-root paths |
| [`release.py`](../scripts/release.py) | `release` command | live logs, structured errors |
| [`config.py`](../scripts/config.py) | project `signing` in config.json | no secrets in git |

Do **not** shell out to these Python files in production. They are documentation-by-code.

---

## Later phases (still future)

- Wails/CMake builds, CI, dogfood releases beyond v0.0.1's, optional AI notes
- Live-hardware acceptance: real SDK builds, instrumented tests on device,
  signed APK end-to-end, `run`/`logcat` against a phone

## Agent rules

1. Prefer updating docs when behaviour changes.
2. Domain logic in `internal/`; cmd and TUI stay thin.
3. Any local path is valid — never filter by GitHub owner or fork flag.
4. For APK packaging/signing details, **open `scripts/*.py` in this repo**.
5. v0.0.1 acceptance must pass before large release-pipeline coding. (Done;
   same rule applies per-doc in v0.0.2+: vet + tests green before every commit.)
