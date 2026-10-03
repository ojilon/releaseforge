# Scan foundation (v0.0.1 core, current through v0.0.2)

> v0.0.2 status: implemented as specified here, plus `cache/meta.json`
> freshness (`internal/project/scan.go: Meta/LoadMeta/Fresh`), config merge
> with key ownership (tool-owned refreshed, user-owned preserved), repo-local
> `.releaseforge.json` seed, Gradle deep parse (`gradle` object in scan.json),
> bounded CMake inventory (`cmake_files`), and cache-first display
> (`CachedVersion` in `status` and the TUI header). Folder dialog was
> deliberately never built — typed paths only (see below).

Scanning is the **basis** of ReleaseForge. Until a project is scanned, the tool cannot responsibly choose build tasks, version files, or release behaviour. Implement this thoroughly before packaging/signing work.

## User-facing flow

1. Ensure data root exists (`init` if needed).
2. **Open folder** — type the path (CLI arg or TUI command bar). No native
   folder dialog exists; that was a deliberate deferral, still in force.
   - TUI: `open <path>` / `scan <path>` (aliases `scan`, `rescan`).
   - CLI: `releaseforge scan [path]` (default: current project).
3. Run scan pipeline on the absolute path.
4. Persist:
   - `projects/<name>/cache/scan.json` + `projects/<name>/cache/meta.json`
   - merge (not just write) `config.json`: tool-owned keys refreshed,
     user-owned keys (`signing`, `github`, `artifacts.app_name`, custom
     tasks) preserved; `.releaseforge.json` seed applied (secrets refused)
   - push entry to `history/recent-projects.json`
5. Show results in the CLI summary or the TUI viewport (`scan` output includes
   a `Config: …​ (written|updated|unchanged)` marker).

## Native folder picker (Windows)

Decision (unchanged since v0.0.1): typed paths only. A native dialog was
evaluated and deferred — `scan <path>` / `open <path>` plus the `recent`
list cover the flow with zero new dependencies. Revisit only with a concrete
usability complaint, not speculatively.

Non-Windows: path entry only; no `zenity`/`kdialog` plans.

## Scan pipeline (ordered steps)

### 1. Resolve path

- `filepath.Abs` + `Clean`
- Must be a directory; error if not
- Derive default `name` = `storage.ProjectNameFromPath`

### 2. Git detection

- Is `.git` present (file or directory)?
- If yes, run read-only git commands (or go-git):
  - `rev-parse --is-inside-work-tree`
  - current branch
  - `log -n 30 --oneline` (configurable N)
  - recent tags (`tag -l --sort=-creatordate | head`)
  - optional: remote `origin` URL (for later GitHub owner/repo guess)
- If not a git repo: record `git: false` and continue (still useful for tool detection).

Do **not** require network. History is local only in v0.0.1.

### 3. Marker / tool detection

Walk or stat known files (do not need a full recursive inventory of every file in v0.0.1; depth-limited is fine):

| Signal | Meaning |
|--------|--------|
| `settings.gradle(.kts)`, `gradlew`, `app/build.gradle(.kts)` | Android / Gradle |
| `gradle.properties` | version + Android props |
| `CMakeLists.txt` | CMake |
| `wails.json` | Wails |
| `go.mod` | Go |
| `package.json`, `pnpm-lock.yaml`, `yarn.lock` | Node frontend |
| `pom.xml` | Maven |
| `pyproject.toml`, `requirements.txt` | Python |
| `Cargo.toml` | Rust |
| `.releaseforge.json` | existing tool seed |

Primary `type` from `docs/02-project-types.md` order. Additional signals go into `tools[]` / `frameworks[]`.

### 4. Config & dependency hints

- Parse lightly where cheap:
  - `gradle.properties` full key map with line numbers (`app.versionName`,
    `app.versionCode`, `ndkVersion`, `*abiFilters`, …)
  - `wails.json` productVersion
  - `go.mod` module path
  - Android deep parse (v0.0.2): settings modules, `android{}` fields with
    `{value,file,line}` provenance, wrapper `distributionUrl`, redacted
    `sdk.dir` presence, `libs.versions.toml` versions, manifest
    package + launcher presence, bounded `cmake_files` list
- List top-level scripts: `scripts/*.py`, etc.
- Unresolved `property()`/`extra[]` references are recorded as `unresolved`,
  never guessed.

### 5. Write cache + config

- `EnsureProjectLayout`
- Write `cache/scan.json` (full snapshot) + `cache/meta.json`
  (`scanned_at`, `tool_version`, `root`, `root_mtime`)
- Merge `config.json` per the ownership rule above (print
  `written|updated|unchanged`)
- Update recent projects list

Cache is trusted until the next explicit `scan`/`rescan` (freshness = root
match in `meta.json`); safety-critical paths (build/release/install) still
re-detect live, display paths (`status`, TUI header) prefer the cache.

## `cache/scan.json` schema (v1)

```json
{
  "scanned_at": "2026-09-30T12:00:00Z",
  "root": "D:/Dev/SomeProject",
  "name": "SomeProject",
  "type": "android-gradle",
  "git": {
    "present": true,
    "branch": "main",
    "head": "abc1234",
    "origin_url": "https://github.com/org/repo.git",
    "recent_commits": [
      { "sha": "abc1234", "subject": "Fix crash", "date": "2026-09-29" }
    ],
    "recent_tags": ["v0.0.3"]
  },
  "tools": [
    { "id": "gradle", "evidence": ["gradlew", "settings.gradle"] },
    { "id": "cmake", "evidence": ["backend/CMakeLists.txt"] },
    { "id": "ndk", "evidence": ["ndkVersion in gradle.properties"] }
  ],
  "frameworks": [
    { "id": "android", "evidence": ["com.android.application"] }
  ],
  "configs": [
    { "path": "gradle.properties", "role": "version+android" },
    { "path": "app/build.gradle", "role": "android-module" }
  ],
  "version": {
    "file": "gradle.properties",
    "name": "0.0.3_2",
    "code": "4"
  },
  "hints": {
    "legacy_release_scripts": true,
    "has_wrapper": true
  }
}
```

Unknown fields should be ignored by older readers; bump `scanned_at` on every refresh.

Since v0.0.2 the snapshot also carries `gradle` (deep parse, android-gradle
only) and `cmake_files` (bounded inventory), and `cache/meta.json` sits
beside it. See `internal/project/gradle.go` and `Snapshot` in
`internal/project/scan.go` for the authoritative shapes.

## CLI output (v0.0.1)

```text
Project: SomeProject
Root:    D:\Dev\SomeProject
Type:    android-gradle
Git:     main @ abc1234 (30 commits loaded)
Tools:   gradle, cmake, ndk
Version: 0.0.3_2 (code 4) from gradle.properties
Config:  D:\ReleaseForgeData\projects\SomeProject\config.json (updated)
Cache:   D:\ReleaseForgeData\projects\SomeProject\cache\scan.json
```

## TUI presentation

Single viewport, no panes or tabs. The same data surfaces through the
command bar: `open|scan [path]`, `rescan`, `recent`, `status` (includes
`[cache]` provenance on the version line), `version`, `help`, plus
`build`/`test`/`notes`/`logs`/`clear`/`quit`. Panes/tabs remain explicitly
out of scope (see `docs/04-tui-design.md`).

## Help in the command bar

When the user submits `-h`, `help`, or `?`:

- Do not shell out only to Cobra; render a **help view in the main pane** listing TUI commands and short descriptions.
- Include examples: `open`, `scan`, `recent`, `status`, `quit`.
- Outer process still supports `releaseforge --help` via Cobra for CLI users.

## Implementation packages (suggested)

| Package | Responsibility | State |
|---------|----------------|-------|
| `internal/storage` | paths, layout | done |
| `internal/config` | global/project JSON | done |
| `internal/project` | Detect + Scan orchestration + gradle deep parse | done |
| `internal/git` | branch, log, tags, origin, changelog grouping | done |
| `internal/history` | recent-projects + persisted command history | done |
| `internal/app` / `internal/tui` | command bar, viewport, help view, panels | done (no panes/tabs) |
| `cmd/scan.go` | wire CLI (`--deep` opt-in runs `gradlew :app:properties`) | done |

## Acceptance criteria

All met as of v0.0.2 (checked against the implementation):

- [x] `init` creates data root on chosen drive
- [x] User types a folder path and runs scan (dialog explicitly deferred)
- [x] Git repos show branch + recent commits in CLI and TUI
- [x] Non-git folders still show tool detection
- [x] Gradle/CMake/Wails/Go markers detected when present
- [x] Results written under data-root project cache (+ `meta.json`)
- [x] Recent projects list updates; reopen via `scan <path from recent>`
- [x] `-h` / `help` in TUI shows command help in the display pane
- [x] Binary is installable/usable on Windows (`installer.exe` + `init`/seed)
