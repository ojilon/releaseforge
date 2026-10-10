# 01 — Current Codebase Map (serious, line-anchored)

> Generated from a read of the tree at v0.0.2-era. Verify with
> `Get-ChildItem -Recurse internal cmd | Select-Object FullName`.

## Entry

- `main.go:1` — delegates to `cmd.Execute()`. Nothing else. Keep it that way.
- `cmd/root.go` — Cobra root, wires `--data-root`, subcommands. All
  subcommands (`scan.go`, `build.go`, `test.go`, `release.go`, `install.go`,
  `tui.go`, `doctor.go`, `logcat.go`, `logs.go`, `notes.go`, `recent.go`,
  `status.go`, `version.go`, `run.go`, `init.go`) must call the same
  `internal/*` packages the TUI calls. If you find duplicated pipeline
  logic in `cmd/`, move it into `internal/`.

## `internal/project/` — detection + scan + runners

| File | Owns |
|------|------|
| `detect.go:47 Detect()` | Marker order: android-gradle → wails → cmake → python → java-cli → go → generic. Returns `Info{Root,Name,Type,VersionFile,…}` |
| `scan.go:71 Scan()` | Ordered pipeline resolve → git → markers → hints. Writes `Snapshot` (schema v1). Gradle deep scan via `ScanGradle`, CMake via `FindCmakeFiles` |
| `scan.go:212 WriteScan()` | Persists `cache/scan.json` + `meta.json`. Overwrites both every scan |
| `scan.go:315 MergeConfig()` | Tool-owned keys from detection; user keys (signing, github, appName, custom tasks) survive. Seed `.releaseforge.json` honoured unless secret-looking |
| `runner.go`, `gradle.go`, `version.go` | `Project` interface: `Program()`, `BuildArgs()`, `TestArgs()`, `VersionRead/Write`, `BuildOutput()` |

Current weakness: `Scan()` does shallow `os.Stat` probes plus one-level
CMake search; no content hashing, no freshness invalidation beyond root
path compare (`Fresh()` at `scan.go:285`). See `02`.

## `internal/app/app.go` — TUI god-model (the rethink target)

- `Model:56` — project context + textinput + viewport + ring + history +
  phase + async `build.Handle`. Single struct owns everything.
- `Update:172` — `WindowSizeMsg` → viewport sizing; `linesMsg/doneMsg` →
  async pump; `KeyMsg` → ctrl+c quit, esc cancel, up/down history,
  enter `exec()`, ctrl+l clear. Input + viewport updated at `:281` every key.
- `exec:383` — string-switch verbs (`help/scan/build/test/logs/notes/…`).
  Pure `parse()` at `:374` is tested; `doBuild:648` / `doTest:683` start
  async runs via `build.Start`.
- `waitLine:441` — first-line blocking, drain up to 100, else `finishRun:465`.
- `View:288` — header pill + context line + viewport + input bar + footer.
  Narrow `<80` fallback at `:293`.

Problem: no windows/tabs, no focus routing, business logic (`doScan`,
`doVersion`, log tail math) inline in the view model. See `03`.

## `internal/tui/styles.go` — Lipgloss theme

Central `BarStyle`, `PanelBorder`, `Pill(phase)`, `Sparkline()`,
`ErrorStyle`, `HelpStyle`, `Rule()`. Good start; missing: adaptive
light/dark + no-color (`NO_COLOR`) paths, width-clamped tables,
truncation-safe runewidth handling. See `04`.

## `internal/build/` — async runner + errors

- `runner.go` — `Start(program, args, Options{Dir,LogPath,Project,Timeout})`
  returns `Handle{Lines chan, Done chan}`. Cancel via `handle.Cancel()`.
- `errors.go` — Gradle/CMake/NDK regex parsing + JUnit summaries →
  `Report.Format()`. Needs corpus tests. See `12`, `28`.

## `internal/android/` — APK truth

- `sign.go` — apksigner locate (config → PATH → `$ANDROID_HOME/build-tools/<newest>`), sign, verify. Password only via prompt.
- `adb.go` — `doctor/run/logcat` + device picker.
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
