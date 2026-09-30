# Architecture

## High-level split

```
┌─────────────────────────────────────────────────────────────┐
│  main.go → cmd (Cobra)                                      │
│    ├─ default / tui  → internal/app (Bubble Tea root)       │
│    ├─ init, scan, status, … (build/release from 0.0.2+)     │
│    └─ all commands call the same internal packages          │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│  internal/                                                  │
│    storage/   data-root paths, layout (implemented)         │
│    config/    global + project JSON (implemented)           │
│    history/   recent-projects + command history             │
│    project/   Detect + Scan → cache/scan.json               │
│    git/       local branch, log, tags (read-only first)     │
│    build/     runners (from 0.0.2)                          │
│    android/   APK, apksigner, adb (after scan foundation)   │
│    github/    gh / API release create                       │
│    log/       live stream + persist                         │
│    metrics/   later                                         │
│    tui/       styles, shared widgets, help view             │
│    app/       Bubble Tea model (command bar, panes, open)   │
└─────────────────────────────────────────────────────────────┘
```

## Design rules

1. **Any local directory** is a valid project root. No filter by GitHub owner, fork status, or a fixed repo list.
2. **Scan before act** — build/release code paths consume `cache/scan.json` + `config.json`, not ad-hoc re-detection every time (but `rescan` refreshes).
3. **Cobra + TUI share domain packages** — no duplicated pipelines.
4. **Data root is mandatory after init** — caches, logs, and releases live there; git trees stay clean.

## v0.0.1 data flow (open + scan)

1. User runs TUI or `scan`.
2. `open` resolves a path (folder dialog or typed path).
3. `project.Scan` → git + markers + version hints.
4. `storage.EnsureProjectLayout` + write `cache/scan.json` + minimal config.
5. `history.Touch` updates recent projects.
6. UI reads cache for Overview / Git / Tools panes.
7. `help` / `-h` renders help text into the main viewport.

## Later: release data flow

See `docs/03-release-pipeline.md`. Runners stream logs via `internal/log`; Android signing stays in `internal/android`.

## Error handling

- Scan errors: missing path, unreadable directory — clear message, non-zero CLI exit.
- Git missing: not an error; `git.present = false`.
- Build/release (later): structured result with headline lines + path to full log.
