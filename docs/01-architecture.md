# Architecture

## High-level split

```
┌─────────────────────────────────────────────────────────────┐
│  main.go → cmd (Cobra)                                      │
│    ├─ default / tui  → internal/app (Bubble Tea root)       │
│    ├─ init, scan, build, test, release, install, …          │
│    └─ all commands call the same internal packages          │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│  internal/                                                  │
│    project/   detect type, version R/W, Project interface   │
│    build/     runners (Gradle, Wails, CMake, generic)       │
│    android/   APK paths, apksigner, adb, ABI helpers        │
│    git/       log, tags, notes from history                 │
│    github/    gh CLI / API release create                   │
│    config/    global + per-project JSON                     │
│    storage/   data-root paths                               │
│    log/       live stream + persist                         │
│    history/   command history + suggestions                 │
│    metrics/   simple stats for dashboard pane               │
│    tui/       Lip Gloss styles, shared Bubbles widgets      │
│    app/       Bubble Tea model (tabs, command bar, panes)   │
└─────────────────────────────────────────────────────────────┘
```

## Why Cobra + TUI together

- **Cobra** gives a professional, composable CLI for scripts, CI helpers, and complex multi-flag operations (`release 0.1.0 --pre --skip-tests`).
- **Bubble Tea** gives the interactive daily experience (live logs, tabs, forms via Huh).
- Domain logic lives only in `internal/`; both surfaces are thin adapters. No duplicated pipelines.

## Data flow (release example)

1. CLI/TUI resolves project path and loads `Project` via `project.Detect`.
2. `config` + `storage` supply paths and signing settings.
3. `project.SetVersion` updates the correct file.
4. `build.Runner` executes tests and assemble tasks; `log` streams and persists.
5. `android` packages APKs, signs, optionally installs.
6. `git` drafts notes from commits since last tag.
7. `github` creates the release and attaches artifacts.
8. Metrics and history are updated under the data root.

## Error handling

- Runners return structured results: exit code, key error lines, path to full log, path to HTML/XML reports when present.
- TUI renders a friendly card + “open full log”; CLI prints the same summary and non-zero exit.

## Concurrency

- Long-running builds run in goroutines; Bubble Tea receives progress/log messages via `tea.Cmd` / channels.
- Only one heavy build per project at a time (simple lock in the runner).
