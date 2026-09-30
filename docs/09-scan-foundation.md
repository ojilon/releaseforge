# Scan foundation (v0.0.1 core)

Scanning is the **basis** of ReleaseForge. Until a project is scanned, the tool cannot responsibly choose build tasks, version files, or release behaviour. Implement this thoroughly before packaging/signing work.

## User-facing flow

1. Ensure data root exists (`init` if needed).
2. **Open folder**
   - TUI command `open` / button: native Windows folder dialog when possible; else path prompt (Huh or command bar).
   - CLI: `releaseforge scan [path]` (default `.` or interactive).
3. Run scan pipeline on the absolute path.
4. Persist:
   - `projects/<name>/cache/scan.json`
   - update `config.json` (minimal or full)
   - push entry to `history/recent-projects.json`
5. Show results in TUI panes (Overview, Git, Tools) or CLI summary.

## Native folder picker (Windows)

Goal: same comfort as OpenCode-style “open folder”.

Options for implementers (pick one, document choice in code):

- PowerShell `FolderBrowserDialog` / `System.Windows.Forms` invoked from Go.
- A small helper or existing Go binding for Win32 `IFileDialog`.
- Fallback: user pastes path in the command bar (`open D:\Dev\MyRepo`).

Non-Windows: path entry + optional `zenity`/`kdialog` later; not required for v0.0.1 if the primary machine is Windows.

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
  - `gradle.properties` key/value for `app.versionName`, `app.versionCode`, `ndkVersion`, `aurora.abiFilters`
  - `wails.json` productVersion
  - `go.mod` module path
- List top-level scripts: `scripts/*.py`, `Makefile`, etc.
- For Android, note presence of `scripts/release.py` (legacy pipeline to absorb later).

### 5. Write cache + config

- `EnsureProjectLayout`
- Write `cache/scan.json` (full snapshot)
- Write or merge `config.json` with at least: `type`, `name`, `root`, detected version file if known
- Update recent projects list

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

## CLI output (v0.0.1)

```text
Project: SomeProject
Root:    D:\Dev\SomeProject
Type:    android-gradle
Git:     main @ abc1234 (30 commits loaded)
Tools:   gradle, cmake, ndk
Version: 0.0.3_2 (code 4) from gradle.properties
Cache:   D:\ReleaseForgeData\projects\SomeProject\cache\scan.json
```

## TUI presentation (v0.0.1)

- **Overview** — type, path, version, last scan time
- **Git** — recent commits / tags from cache
- **Tools** — tools + frameworks + config files list
- Command bar accepts: `open`, `scan`, `rescan`, `recent`, `status`, `-h` / `help`, `quit`

## Help in the command bar

When the user submits `-h`, `help`, or `?`:

- Do not shell out only to Cobra; render a **help view in the main pane** listing TUI commands and short descriptions.
- Include examples: `open`, `scan`, `recent`, `status`, `quit`.
- Outer process still supports `releaseforge --help` via Cobra for CLI users.

## Implementation packages (suggested)

| Package | Responsibility |
|---------|----------------|
| `internal/storage` | paths, layout (exists) |
| `internal/config` | global/project JSON (exists) |
| `internal/project` | Detect + Scan orchestration |
| `internal/git` | local log/branch/tags (expand beyond notes stub) |
| `internal/history` | recent-projects + command history |
| `internal/app` / `internal/tui` | open dialog, panes, help view |
| `cmd/scan.go` | wire CLI |

## Acceptance criteria (v0.0.1)

- [ ] `init` creates data root on chosen drive
- [ ] User can select a folder (dialog or path) and run scan
- [ ] Git repos show branch + recent commits in CLI and TUI
- [ ] Non-git folders still show tool detection
- [ ] Gradle/CMake/Wails/Go markers detected when present
- [ ] Results written under data-root project cache
- [ ] Recent projects list updates and is reopenable
- [ ] `-h` / `help` in TUI shows command help in the display pane
- [ ] Binary is installable/usable on the author’s Windows machine without implementing full release yet
