# Config and storage

## Data root (managed storage)

Chosen once via `releaseforge init` (Huh wizard). Prefer a non-system drive:

```
D:/ReleaseForgeData/          # example
  config/
    global.json
  projects/
    Conductino-Android/
      config.json             # scanned/edited project config
      builds/
      logs/
      releases/
        0.1.0/
          *.apk
          *.zip
          notes.md
      reports/
      cache/
    Conductino_/
      ...
  history/
    commands.jsonl
  global-cache/
```

The git repo of each project stays clean; heavy artifacts live under the data root. Optional thin mirror paths inside the repo (e.g. `release/`) can still be written if the project config requests it (for compatibility with existing scripts).

## Global config (`config/global.json`)

```json
{
  "data_root": "D:/ReleaseForgeData",
  "github_token_env": "GITHUB_TOKEN",
  "preferred_device": null,
  "editor": "",
  "ai": {
    "enabled": false,
    "provider": "",
    "model": ""
  }
}
```

## Project config

Written by `scan`, editable in TUI Config tab or as JSON.

See `configs/example-android-gradle.json` and `configs/example-wails.json`.

Key fields:

- `type` — `android-gradle` | `wails` | `cmake` | `python` | `java-cli` | `generic`
- `root` — absolute path to the project
- `version` — file + keys or path expression
- `build.tasks` — map of logical names → commands/tasks
- `signing` — keystore, alias, optional apksigner path
- `artifacts` — globs / fixed paths for debug & release outputs
- `github.owner` / `github.repo`
- `abi_filters` / `ndk_version` where relevant

## Repo-local optional file

`.releaseforge.json` in the project root can override or seed the data-root copy so clones share defaults without committing secrets.
