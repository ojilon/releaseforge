# Config and storage

## Why a dedicated data root

- Keeps git working trees free of APKs, logs, and caches.
- Survives moving or re-cloning the source repo.
- Lets the user put heavy data on `D:` (or any path) instead of `C:`.
- Holds **per-project mirror state** that updates as scans and builds run.

## First run

`releaseforge init` (CLI or TUI form):

1. Propose default root (`D:/ReleaseForgeData` on Windows if `D:` exists, else home-based path).
2. Allow path edit or browse.
3. Create layout via `storage.EnsureLayout`.
4. Write `config/global.json`.

On later startups, if the data root is missing, prompt to run init or pass `--data-root`.

Env override: `RELEASEFORGE_DATA_ROOT`.

## Layout

```
<data-root>/                          # e.g. D:/ReleaseForgeData
  config/
    global.json
  projects/
    <sanitized-name>/
      config.json                     # durable project settings
      cache/
        scan.json                     # last full scan snapshot
        tools.json                    # optional extracted tool versions
        meta.json                     # last opened, last scan time, hash of root
      builds/                         # optional copied artifacts
      logs/                           # persisted command logs
      releases/                       # versioned release folders
      reports/                        # test/build reports
  history/
    commands.jsonl                    # global command history
    recent-projects.json              # ordered list of opened folders
  global-cache/
```

`storage.SanitizeProjectName` already exists for safe directory names.

## Recent projects (`history/recent-projects.json`)

```json
{
  "items": [
    {
      "path": "D:/Dev/Conductino-Android",
      "name": "Conductino-Android",
      "type": "android-gradle",
      "last_opened": "2026-09-30T12:00:00Z",
      "last_scan": "2026-09-30T12:01:00Z"
    }
  ],
  "max": 30
}
```

- Updated whenever the user opens/scans a folder.
- TUI “Open recent” list + CLI `releaseforge scan` without args can offer recent paths.
- Deduplicate by absolute cleaned path; move to front on re-open.

## Global config

```json
{
  "data_root": "D:/ReleaseForgeData",
  "github_token_env": "GITHUB_TOKEN",
  "preferred_device": null,
  "editor": "",
  "ai": { "enabled": false, "provider": "", "model": "" }
}
```

(Implemented in `internal/config`.)

## Project config

Written/updated by scan; editable later. Example types: `android-gradle`, `wails`, `cmake`, `python`, `java-cli`, `go`, `generic`.

See `configs/example-*.json`. Validation already enforces type + name + root + version.file for known schemas; **v0.0.1 scan may write a richer scan snapshot first** and only a minimal config until version keys are confirmed.

## Scan cache (`cache/scan.json`)

Authoritative for “what we last knew about this tree”. Schema in `docs/09-scan-foundation.md`. Refresh on explicit `scan` / `rescan` or when the TUI opens a project and the root mtime/identity changed.

## Repo-local optional seed

`.releaseforge.json` in the project may override defaults (no secrets). Data-root copy remains the working source of truth for the tool.
