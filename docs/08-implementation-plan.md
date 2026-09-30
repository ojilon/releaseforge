# Implementation plan (agent guide)

Follow this order. **Do not skip the scan foundation** to chase signing early.

---

## Version 0.0.1 — Scan, storage, navigate (CURRENT TARGET)

Goal: installable binary the author can run daily to **open any local folder**, see what it is, see git history, and keep a cache under a data root (prefer `D:`).

### Already done

- [x] Repo scaffold, Cobra stubs, docs foundation
- [x] `internal/storage` — data-root paths, `EnsureLayout`, Windows `D:` default
- [x] `internal/config` — global/project load/save, `DiscoverDataRoot`
- [x] `cmd/init` — partial implementation exists; finish wiring to storage/config

### 0.0.1 work items (detailed)

1. **Finish `init`**
   - Create layout, write `global.json`, print path.
   - Idempotent if already initialized.

2. **Recent projects (`internal/history`)**
   - Read/write `history/recent-projects.json`.
   - API: `ListRecent`, `Touch(path, name, type)`, max 30.

3. **Git read helpers (`internal/git`)**
   - Branch, head, recent commits, recent tags, origin URL (best-effort).
   - Prefer invoking `git` CLI for v0.0.1 speed of implementation; go-git optional later.
   - Graceful when not a repo.

4. **Scan orchestration (`internal/project`)**
   - Implement detection order from `docs/02-project-types.md`.
   - Collect tools/frameworks/configs/version hints.
   - Write `cache/scan.json` + minimal `config.json`.
   - Call `EnsureProjectLayout`.
   - Full schema: `docs/09-scan-foundation.md`.

5. **CLI `scan`**
   - `releaseforge scan [path]` prints human summary and cache path.
   - Exit non-zero on missing path / unreadable dir.

6. **Minimal TUI**
   - Command bar + one main viewport.
   - Commands: `open`, `open <path>`, `scan`, `recent`, `status`, `help`/`-h`, `quit`.
   - Help rendered in pane.
   - Overview/Git/Tools can be simple text sections or tab switches.

7. **Windows folder picker**
   - Wire `open` with dialog + path fallback.

8. **Version metadata**
   - Embed or hardcode tool version `0.0.1` in `version` / status header.

### 0.0.1 definition of done

```bash
releaseforge init
releaseforge scan D:\path\to\any\repo
releaseforge tui
# open folder via dialog, see git + tools, type help, reopen from recent
```

Works on forked and non-forked trees alike.

---

## Version 0.0.2 — Bootstrap ReleaseForge + start project builds

Goal: ReleaseForge can **test/build/package a release of itself** (Go binary + notes + tag + GitHub release), proving the pipeline. Begin applying the same ideas to an Android project by absorbing the Python scripts’ behaviour.

### Work items

1. **Build runner** — generic command runner with log capture to data-root `logs/`.
2. **Self project type** — detect this repo (`go.mod` module `github.com/ojilon/releaseforge`), tasks: `go test ./...`, `go build`.
3. **Notes from git** — between tags or since date.
4. **`release` for self** — version bump strategy for this repo (e.g. VERSION file or tag-only), attach binary, `gh release create`.
5. **Port Conductino-Android script behaviour** (see below) behind Android project type — can be partial (version + assemble + package) if sign comes right after.

### Absorb Python tool (Conductino-Android `scripts/`)

Reimplement in Go with improvements:

| Python module | Go responsibility |
|---------------|-------------------|
| `version.py` | Read/set `app.versionCode` / `app.versionName` |
| `package.py` | Copy APKs into versioned release dir with stable names |
| `sign.py` | apksigner locate + sign + verify + password prompt |
| `notes.py` | Template + git-history fill |
| `zip_release.py` | Zip artifacts |
| `release.py` | Orchestrate order; stream logs; write to data-root |
| `config.py` | Map into project `config.json` signing section |

Improvements required:

- Live log streaming into TUI/CLI
- Artifacts default under data-root (optional in-repo mirror)
- Structured error headlines from Gradle output
- No plaintext password storage

---

## Later phases

- Full Android sign + adb install
- Wails / CMake first-class builds
- Metrics pane, richer suggestions
- Optional AI notes
- CI for ReleaseForge; dogfood releases

---

## Agent rules

1. Prefer updating docs when behaviour changes.
2. Keep domain logic in `internal/`; cmd and TUI stay thin.
3. Any local path is valid input — never filter by GitHub owner or fork flag.
4. v0.0.1 acceptance must pass before large release-pipeline coding.
