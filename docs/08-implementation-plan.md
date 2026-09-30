# Implementation plan (agent guide)

Follow this order. **Do not skip the scan foundation** to chase signing early.

**Android APK release behaviour:** read **[`scripts/`](../scripts/)** in this repository ([`scripts/README.md`](../scripts/README.md)). That is the complete legacy reference; no other repo is required.

---

## Version 0.0.1 — Scan, storage, navigate (CURRENT TARGET)

Goal: installable binary the author can run daily to **open any local folder**, see what it is, see git history, and keep a cache under a data root (prefer `D:`).

### Already done

- [x] Repo scaffold, Cobra stubs, docs foundation
- [x] `internal/storage` — data-root paths, `EnsureLayout`, Windows `D:` default
- [x] `internal/config` — global/project load/save, `DiscoverDataRoot`
- [x] `cmd/init` — partial implementation exists; finish wiring to storage/config
- [x] **`scripts/`** — Python reference for later Android release port

### 0.0.1 work items (detailed)

1. **Finish `init`** — layout + `global.json`; idempotent.
2. **Recent projects (`internal/history`)** — `history/recent-projects.json`.
3. **Git read helpers (`internal/git`)** — branch, head, commits, tags, origin (best-effort).
4. **Scan (`internal/project`)** — `docs/02-project-types.md` + `docs/09-scan-foundation.md`; write `cache/scan.json`.
5. **CLI `scan`** — human summary + cache path.
6. **Minimal TUI** — command bar; `open`, `scan`, `recent`, `status`, `help`/`-h`, `quit`.
7. **Windows folder picker** — with path fallback.
8. **Tool version `0.0.1`** in status/header.

### 0.0.1 definition of done

```bash
releaseforge init
releaseforge scan D:\path\to\any\repo
releaseforge tui
# open folder, see git + tools, type help, reopen from recent
```

---

## Version 0.0.2 — Bootstrap ReleaseForge + Android pipeline from scripts/

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

## Later phases

- adb install, Wails/CMake builds, metrics, optional AI notes, CI, dogfood releases

## Agent rules

1. Prefer updating docs when behaviour changes.
2. Domain logic in `internal/`; cmd and TUI stay thin.
3. Any local path is valid — never filter by GitHub owner or fork flag.
4. For APK packaging/signing details, **open `scripts/*.py` in this repo**.
5. v0.0.1 acceptance must pass before large release-pipeline coding.
