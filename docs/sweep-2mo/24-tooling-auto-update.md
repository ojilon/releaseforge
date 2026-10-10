# 24 — Tooling Auto-Update (active web connection)

## Goal

`rf update-check` + `rf update-tools` keep JDK/Gradle/SDK/NDK/Zig/Odin fresh
over an ACTIVE web connection, with pins + rollback. Never silent-upgrade
during a release.

## Design

- `tools.json` pins: `{jdk, gradle, build_tools, ndk, zig, odin}` versions +
  download URLs + sha256. Lives in data-root (user-editable), seeded in `configs/`.
- `update-check`: GET version endpoints (Gradle releases API, Zig
  `https://ziglang.org/download/index.json`, Android `sdkmanager --list`,
  Odin GitHub releases), diff vs pins, render table (current → latest).
- `update-tools --apply`: downloads to data-root `tooling/<tool>/<ver>/`,
  verifies sha, flips `current` symlink, keeps previous (`rollback` restores).
- Releases record `tool_versions{}` in manifest (`22`) for forensics.
- Offline mode: `--offline` uses pins only; updater refuses during `release`
  unless `--allow-tool-change`.

## Snippet — pin table

```go
type ToolPin struct {
    Name, Current, Latest, URL, SHA string
    UpdateCmd []string // e.g. {"sdkmanager","build-tools;34.0.0"}
}
```

## Read

- Gradle compat matrix: https://docs.gradle.org/current/userguide/compatibility.html
- Zig download index: https://ziglang.org/download/index.json
- `sdkmanager`: https://developer.android.com/tools/sdkmanager
- Odin releases: https://github.com/odin-lang/Odin/releases

## Tasks

- [ ] `internal/update/` checker + pin file + `update-check` CLI/TUI view.
- [ ] `update-tools --apply/--rollback` with sha verify + symlink flip.
- [ ] Manifest records tool versions; `doctor` surfaces drift.

## Acceptance

```powershell
releaseforge update-check
releaseforge update-tools --dry-run
```
