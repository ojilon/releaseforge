# 01 — Current Codebase Map (serious, line-anchored)

> Generated from a read of the tree at v0.0.2-era. Verify with
> `Get-ChildItem -Recurse internal cmd | Select-Object FullName`.
>
> > Delta (v0.0.7-era): `internal/app/app.go` is gone — split into one file
> > per job (see below). New: `cmd/sign.go`, `cmd/devices.go`,
> > `internal/android/signing.go`, `internal/tui/theme.go`,
> > `build-installer.bat`. `VERSION` is gitignored (local-only truth).

## Entry

- `main.go:1` — delegates to `cmd.Execute()`. Nothing else. Keep it that way.
- `cmd/root.go` — Cobra root, wires `--data-root`, subcommands. All
  subcommands (`scan.go`, `build.go`, `test.go`, `release.go`, `install.go`,
  `sign.go`, `devices.go`, `tui.go`, `doctor.go`, `logcat.go`, `logs.go`,
  `notes.go`, `recent.go`, `status.go`, `version.go`, `run.go`, `init.go`)
  must call the same `internal/*` packages the TUI calls. If you find
  duplicated pipeline logic in `cmd/`, move it into `internal/`.

## `internal/project/` — detection + scan + runners

| File | Owns |
|------|------|
| `detect.go:47 Detect()` | Marker order: android-gradle → wails → cmake → python → java-cli → go → generic. Returns `Info{Root,Name,Type,VersionFile,…}` |
| `scan.go:71 Scan()` | Ordered pipeline resolve → git → markers → hints. Writes `Snapshot` (schema v1). Gradle deep scan via `ScanGradle`, CMake via `FindCmakeFiles` |
| `scan.go:212 WriteScan()` | Persists `cache/scan.json` + `meta.json`. Overwrites both every scan |
| `scan.go:315 MergeConfig()` | Tool-owned keys from detection; user keys (signing, github, appName, custom tasks) survive. Seed `.releaseforge.json` honoured unless secret-looking |
| `runner.go`, `gradle.go`, `version.go` | `Runner` interface: `Program()`, `BuildArgs()`, `TestArgs()`, `VersionRead/Write`, `BuildOutput()`. Go stamp reads `VERSION` (`BuildArgs`); missing file → `dev` stamp |
| `build.go` (`cmd` + TUI `jobs.go:doBuild`) | `build [variant] [version]` — optional version is written (`VersionWrite`, never committed) BEFORE args are computed, so the stamp uses it |

Current weakness: `Scan()` does shallow `os.Stat` probes plus one-level
CMake search; no content hashing, no freshness invalidation beyond root
path compare (`Fresh()` at `scan.go:285`). See `02`.

## `internal/app/` — TUI frontend, one file per job (split from `app.go`)

- `model.go` — `Model` struct + `New()` + `refreshProject()` + `layout()`.
  Adds `height`, `dash` (right-panel text) next to the old fields.
- `run.go` — `Run()` + `waitLine` (batch ≤100) + `finishRun` + `doneMsg`
  (now carries `summary` from `Report.Summary`).
- `update.go` — `Update` → `onResize/onDone/onKey/historyUp/historyDown`.
  `onDone` renders failures via `tui.RenderReport`.
- `view.go` — `View`: header pill + context + log pane (+dashboard when
  `width >= 100` via `JoinHorizontal`) + left-bordered input + footer.
- `commands.go` — `parse` (tested) + `onEnter` + `exec` string-switch.
  Verbs: help/clear/status/scan/recent/version/build/test/logs/notes +
  sign/install/devices/git/ls/dir/pwd/cat/bat/run/theme/quit.
- `dashboard.go` — `buildDash()`: name/type/version/branch/git head/tag/
  tools from scan cache. `scan.go`, `version.go`, `jobs.go` (`doBuild`
  takes `variant, setVer`; `doTest`), `logs.go`, `notes.go`,
  `sign.go`, `install.go` (TUI `InstallCapture` path), `gitcmd.go`,
  `files.go`, `scripts.go` (`bat` streams via `build.Start`, allowlist
  `.bat/.cmd/.ps1/.sh/.py`).

Problem (still open): no windows/tabs, no focus routing; `exec()`
string-switch instead of router; `running`+`handle`+`pendingRun` still
three fields. See `03`, `06`, `40`.

## `internal/tui/` — theme + styles

Central `theme.go` (`Theme` struct, `DarkTheme`/`LightTheme`, `Current`,
`SetTheme`, `PaintInput`) — edit colors ONLY there. `styles.go` derives
`HeaderStyle/BarStyle` (thick LEFT border only)/`LogBoxStyle` (thin)/
`DashBoxStyle`/`Pill/Error/Help/Rule` + `RenderReport` (fail-red summary,
dimmed `hint:` lines). Good start; missing: `NO_COLOR` path,
width-clamped tables, runewidth-safe truncation. See `04`.

## `internal/build/` — async runner + errors

- `runner.go` — `Start(program, args, Options{Dir,LogPath,Project,Timeout})`
  returns `Handle{Lines chan, Done chan}`. Cancel via `handle.Cancel()`.
- `errors.go` — Gradle/CMake/NDK regex parsing + JUnit summaries →
  `Report{Items,Summary}`. `Format()` items + `Summary` shown in TUI
  (`tui.RenderReport`) and CLI (`printReport`). Needs corpus tests. See `12`, `28`.
- `golang.go` — `GoBuildArgs(dir,out,stamp,variant)` stamps
  `<module>/internal/version.ToolVersion`; `release`/`build-installer.bat`
  stamp both main + installer binaries.

## `internal/android/` — APK truth

- `sign.go` — apksigner locate (config → PATH → `$ANDROID_HOME/build-tools/<newest>`), sign, verify. `signing.go` — `SigningConfig`,
  `SigningDir` (`<data-root>/signing/<project>/`), `ResolveSigning`
  (flags > project file > global file > env), `KeystorePassword` (env
  `RF_KEYSTORE_PASS` only, never disk), `SignApkCapture`/`InstallCapture`
  (combined-output variants for the TUI log pane).
- `adb.go` — `doctor/run/logcat` + device picker.
- `cmd/sign.go`, `cmd/devices.go`, `cmd/install.go` — CLI twins of the
  TUI `sign/install/devices` verbs.
- Reference: `scripts/*.py` (version/package/sign/notes/zip/release).
  Go must match outcomes, improve paths/logging. See `13`, `21`.

## `internal/git/`, `github/`, `log/`, `metrics/`, `config/`, `storage/`, `history/`

- `git/` — `IsRepo/CurrentBranch/IsClean/Log/LogSince/RecentTags`,
  `DraftNotes`, grouped/bounded changelogs. Read-only first; automation in `10`.
- `github/` — `gh release create` wrapper. Promotion gates in `23`.
- `log/` — `Ring` (cap 2000) + `LogPath/List/Newest/Resolve/Tail/Head` + index.
- `metrics/` — `Append` + `LastBuildLine` (surfaced in `statusText`).
- `config/` — global + project JSON; `storage/` — data-root layout
  (`D:` default on Windows), `EnsureProjectLayout`, `ScanFile`, `LogsDir`,
  `ReleaseVersionDir`. `history/` — recent + `commands.jsonl`.

## `docs/` today

`00-vision` → `10-bootstrap-v001`, plus `reference/`, `v0.0.2/` (00–15 +
`PROGRESS.md` all ticked), `v0.0.3/`. This sweep folder does not replace
them; it maps the next 2 months.
