# Bootstrap v0.0.1 — self-hosting lifecycle

Goal of v0.0.1: **ReleaseForge manages itself end-to-end.** One project
(this repo), one version source, real storage on your machine. Multi-project
features and Android pipelines reuse the same flow later; they are not the
focus now.

## 0. The four lanes: `go run`, test exe, release exe, installer

You were right to ask — the old doc mixed these up. There are four distinct
build stages, each with its own storage, and you move down the list:

| Lane | Command | Storage | Purpose |
|---|---|---|---|
| A. `go run` dev loop | `go run . <command>` with `$env:RELEASEFORGE_DATA_ROOT=".data"` | `.data/` inside the repo (gitignored, throwaway) | Fast iteration, no exe. Prove scan/version/test/build/notes here first. |
| B. Test exe | `go build -o rf-test.exe .`, run against temp storage | `%TEMP%\rf-test-data` | Prove the compiled binary + the installer flow end-to-end without touching real data. |
| C. Release exe + installer | stamped `releaseforge.exe` + `installer.exe`, packed in `dist/` | chosen by installer (e.g. `D:\Dev\ReleaseForge\data`) | The real artifacts. First-ever install on your PC. |
| D. Manual GitHub release | tag + push + `gh release create` by hand | n/a | One-time-only for v0.0.1. From v0.0.2 the tool + CI do this. |

Two facts that confuse everyone once — read these before running anything:

1. **Bare `go run .` (no subcommand) opens the TUI**, not a help screen.
   The TUI needs storage, so on a fresh setup bare `go run .` fails with
   `data root not initialised`. This is normal: run `init` first (section 12),
   then bare `go run .` works.
2. **First command in any fresh terminal sets the storage for that terminal.**
   Lane A: `$env:RELEASEFORGE_DATA_ROOT = ".data"`. Lanes B/C: pass
   `--data-root <path>` or rely on the install's seeded root. The env var only
   lives in that terminal window — a new terminal starts unset.
3. **In lane A, read every `releaseforge <cmd>` in this doc as
   `go run . <cmd>`.** Sections below drop the prefix for readability.

Rules: never point lane A or B at your real data root; never commit `.data/`,
`dist/`, or any `*.exe` (all gitignored). Lane C's exe is the only binary
that ever gets installed or uploaded.

PowerShell note: `go run .` and `go run main.go` are equivalent here
(single `package main` at root importing `cmd/`). This doc uses `go run .`.

## 1. Version source of truth

- `VERSION` at repo root holds the project version (e.g. `0.0.1`).
- `internal/version/version.go:ToolVersion` is the baked-in tool version shown
  in `--version`, `version`, `status`, and the TUI header. Bump it together
  with `VERSION` at release time.
- Release builds stamp the version so the binary reports correctly outside
  the repo (lane C):

```powershell
$v = (Get-Content VERSION).Trim()
go build -ldflags "-X github.com/ojilon/releaseforge/internal/version.ToolVersion=$v" -o releaseforge.exe .
```

- Git tag `v<version>` (e.g. `v0.0.1`) marks the release commit. Tags are
  created by `release`, never silently.
- Rule for v0.0.1: **writing a version never commits.** `version --set` and
  `release` update the `VERSION` file only; you review and commit manually.

Code: `internal/project/detect.go` (`ReadVersionFile`, `SetVersionFile`),
`cmd/version.go`, `internal/version/version.go`.

## 2. Storage check + init

Every command that needs storage calls `requireDataRoot()` (`cmd/root.go`):
resolve `--data-root` → `--config` → `$RELEASEFORGE_DATA_ROOT` → default
(`D:/ReleaseForgeData` when `D:` exists), then require
`<root>/config/global.json` to exist. When it does not, you get:

```text
data root not initialised (<path> missing) — run `releaseforge init` first
```

The TUI refuses to start the same way (`internal/app/app.go:Run`).
The installer pre-seeds `global.json`, so after a real install you never run
`init` by hand; in lanes A/B you init the throwaway root explicitly (see
section 12). Layout is created by `storage.EnsureLayout`
(`internal/storage/storage.go`): `config/`, `projects/`, `history/`,
`global-cache/`.

**Minimum viable order:** `init` → `scan` → anything else. If you just want
to poke around, run those two first — every other command assumes they have
happened at least once for that storage.

## 3. Open a project by typed path + path cache

No folder dialog in v0.0.1 — you type the path (your confirmed decision):

```bash
releaseforge scan D:\projects\releaseforge
releaseforge recent
```

`scan` also records the project in
`<root>/history/recent-projects.json` (most-recent-first, absolute cleaned
paths, deduped, capped at 30). `recent` lists it; re-scanning moves it back
to the top. Code: `internal/history/recent.go`, `cmd/recent.go`,
`cmd/scan.go`. The TUI mirrors this with `scan <path>` and `recent`.

## 4. Scan snapshot

`releaseforge scan <path>` runs `project.Scan` (`internal/project/scan.go`):

1. Resolve path (`filepath.Abs` + `Clean`), detect type (`internal/project/detect.go`).
2. Git state, read-only: branch, HEAD, origin URL, 30 recent commits, 10
   recent tags (`internal/git/notes.go`). Non-repos record `present: false`.
3. Tool markers: gradle, cmake, wails, go, node, maven, python, rust, plus
   wrapper/legacy-script hints and a config-file inventory.
4. Version hint: `VERSION` content for Go, `app.versionCode/Name` for Gradle,
   `productVersion` for Wails (best-effort).
5. Persist `projects/<name>/cache/scan.json` (schema: `docs/09-scan-foundation.md`),
   write minimal `config.json` once, touch recent.

Refresh rule: cache is trusted until you `scan`/`rescan` again. No watching.

The scan snapshot feeds everything downstream: version display, git history
view, notes drafting, and later build selection.

## 5. Version display (automatic)

```bash
releaseforge version          # tool version + project VERSION content
releaseforge status           # tool + project + type + git + devices + paths
```

The TUI header always shows `ReleaseForge <tool> · <project> · <version>`.
Git history for the open project comes from the scan cache; raw commands
remain available.

## 6. Test and build this repo

```bash
releaseforge test             # go test ./...  (log under projects/releaseforge/logs/)
releaseforge build            # go build + version stamp → projects/releaseforge/builds/releaseforge.exe
releaseforge build release    # -trimpath, stripped symbols → builds/releaseforge-release.exe
releaseforge logs             # newest-first persisted logs
```

Runners stream live lines and persist them (`internal/build/runner.go`,
`internal/build/golang.go`, `internal/log/stream.go`). Go-specific args and
the `ToolVersion` stamp live in `GoBuildArgs`; module path is read from
`go.mod`. In lane A these write under `.data/`; the repo stays clean.

## 7. Notes from git history

```bash
releaseforge notes            # draft to stdout (merges commits since last tag)
releaseforge notes 0.0.1      # write .../releases/0.0.1/notes.md
```

Template sections (Highlights / Changes / Bug fixes / Known issues) are filled
with one line per commit since the previous tag (`internal/git/notes.go:
DraftNotes`). TUI: `notes [version]`.

## 8. `release` command — when to use it and when NOT to

```bash
releaseforge release 0.0.1 --pre
```

**State rules (this is where confusion comes from):**

- **Lanes A/B (throwaway storage): do NOT run `release`.** It writes the
  `VERSION` file, creates a real git tag, and pushes it to origin. Test the
  pipeline pieces individually instead (`test`, `build`, `notes`) — that is
  what lanes A/B are for.
- **Lane C/D (real install, real storage): this is the one allowed run.**
  Either as the automated path in section 13, option 1, or not at all if you
  pick the fully-manual option 2. Never both — running it twice creates the
  tag twice (the second run fails on the existing tag, harmlessly).
- **Precondition:** the tree is committed up to and including the version
  bump (`VERSION` + `internal/version/version.go` committed and pushed).
  `release` warns on a dirty tree but continues, so check `git status`
  yourself first.

Order (tag comes last, per your confirmation):

1. Warn on dirty tree (non-fatal).
2. Write `VERSION` = `0.0.1` (no commit).
3. `go test ./...` (skip with `--skip-tests`, not recommended).
4. `go build -trimpath` with stamp → `releases/0.0.1/releaseforge-0.0.1.exe`.
5. Draft `releases/0.0.1/notes.md` from history.
6. Zip the binary → `releases/0.0.1/releaseforge-0.0.1.zip`.
7. Tag `v0.0.1`, `git push origin v0.0.1`.
8. `gh release create v0.0.1 --notes-file … <binary> <zip>` (with `--pre` →
   `--prerelease`).

Edge behaviour: **without `gh` on PATH the command stops after the local
tag + zip and exits 0**, printing tag, notes, and artifact paths plus the
exact `gh release create` line to run later. Push failures abort with the
manual retry hint; nothing is force-pushed. Code: `cmd/release.go`
(`releaseGo`, `publishTagAndRelease`), `cmd/notes.go`.

## 9. Installer: build, package, first-ever install

The installer (`installer/main.go`, same Charm stack) installs a
`releaseforge.exe` payload sitting next to it (`--bin` overrides). It asks
for a drive by number, then a folder by number (or new folder, or drive
root), and creates `<chosen>/ReleaseForge/{bin,data}`. Re-runs overwrite
`bin/`, top up missing dirs/seed files, refresh `install.json`, and **never
touch `projects/` or `history/`**.

Build and package (lane C):

```powershell
$v = (Get-Content VERSION).Trim()
go build -ldflags "-X github.com/ojilon/releaseforge/internal/version.ToolVersion=$v" -o releaseforge.exe .
go build -o installer.exe ./installer
New-Item -ItemType Directory dist -Force
Copy-Item releaseforge.exe, installer.exe dist/
```

First-ever installation on your PC:

```powershell
cd dist
.\installer.exe
# pick drive number (e.g. D:\), pick folder number (e.g. Dev)
$env:Path += ";D:\Dev\ReleaseForge\bin"
releaseforge scan D:\projects\releaseforge
releaseforge status
```

Result: `D:\Dev\ReleaseForge\bin\releaseforge.exe`,
`D:\Dev\ReleaseForge\data\{config,projects,history,global-cache,install.json}`.
`install.json` records version + timestamp for the future 0.0.3 updater.
Done once; every later version updates through this same installer.

## 10. Seeing the resources

- `releases/<version>/` under `projects/releaseforge/` holds binary, zip,
  notes — the "specific folder" from your flow.
- `releaseforge logs` shows persisted build/test logs.
- `releaseforge recent` shows known project paths.
- `releaseforge status` shows paths to config, cache, and data root.

## 11. Code map (where to read)

| Flow step | Code |
|---|---|
| Storage guard | `cmd/root.go:requireDataRoot`, `internal/storage/storage.go` |
| Init wizard | `cmd/init.go` |
| Installer | `installer/main.go` (+ `main_test.go`) |
| Scan + cache | `cmd/scan.go`, `internal/project/scan.go`, `internal/project/detect.go` |
| Recent cache | `internal/history/recent.go`, `cmd/recent.go` |
| Version | `VERSION`, `internal/version/version.go`, `cmd/version.go` |
| Test / build | `cmd/test.go`, `cmd/build.go`, `internal/build/` |
| Logs | `cmd/logs.go`, `internal/log/stream.go` |
| Notes | `cmd/notes.go`, `internal/git/notes.go` |
| Release | `cmd/release.go`, `internal/github/release.go`, `internal/android/sign.go` (later) |
| TUI | `internal/app/app.go`, `internal/tui/styles.go` |

## 12. Full first session, lane by lane

Lane A — `go run` dev loop (throwaway `.data/`, repo root).
Run from `D:\projects\releaseforge` — `.data` resolves relative to where you
stand. Do NOT run `release` here (section 8):

```powershell
cd D:\projects\releaseforge
$env:RELEASEFORGE_DATA_ROOT = ".data"
go run . init --non-interactive
go run . scan D:\projects\releaseforge
go run . version
go run . status
go run . recent
go run . test
go run . build
go run . notes 0.0.1
go run . logs
go run .            # bare form = TUI. Only works AFTER init above.
```

Lane B — test exe + installer dry-run (temp storage only):

```powershell
Remove-Item env:RELEASEFORGE_DATA_ROOT
go build -o rf-test.exe .
.\rf-test.exe init --data-root $env:TEMP\rf-test-data --non-interactive
.\rf-test.exe --data-root $env:TEMP\rf-test-data scan D:\projects\releaseforge
go build -o installer-test.exe ./installer
.\installer-test.exe --bin .\rf-test.exe --dest $env:TEMP\rf-test-install --non-interactive
$env:TEMP\rf-test-install\ReleaseForge\bin\releaseforge.exe --version
```

Lane C — release exe + real first install: section 9.

Lane D — manual 0.0.1 publish: section 13.

## 13. Manual 0.0.1 GitHub release (one-time-only)

This is the single manual release. From v0.0.2 the tool + CI do it
(`release` command, pre-release shell, local artifact upload).
Pick ONE option — option 1 uses the `release` command, option 2 is fully
manual. Both assume lanes A–C already passed and the tree is clean.

Common first steps (both options):

```powershell
# 1. clean tree, version files already at 0.0.1 (VERSION + internal/version/version.go)
git status --short            # must be clean apart from intended files
git add VERSION internal/version/version.go
git commit -m "Release 0.0.1"
git push origin main          # or your default branch

# 2. lane C artifacts ready in dist/: releaseforge.exe, installer.exe,
#    plus notes from: releaseforge notes 0.0.1
#    plus a zip of the stamped binary:
Compress-Archive -Force dist/releaseforge.exe dist/releaseforge-0.0.1.zip
```

Option 1 — `release` command (recommended):

```powershell
releaseforge release 0.0.1 --pre
# does: test → build → notes → zip → tag v0.0.1 → push tag → gh release create
# then attach the installer by hand (automated from v0.0.2):
gh release upload v0.0.1 dist/installer.exe
```

Option 2 — fully manual (full control, or `gh` missing):

```powershell
# 3. tag + push tag
git tag -a v0.0.1 -m "Release 0.0.1"
git push origin v0.0.1

# 4. publish (pre-release for 0.0.1)
$n = "$env:TEMP\rf-rel\0.0.1\notes.md"   # path printed by `notes 0.0.1`
gh release create v0.0.1 --prerelease --notes-file $n `
  dist/releaseforge.exe dist/installer.exe dist/releaseforge-0.0.1.zip
```

If `gh` is missing, stop after the tag push — tag + local zip under the data
root is the releasable state; publish from any machine with `gh` later.

## 14. What v0.0.2 automates (not now)

- CI workflow creating the pre-release shell on tag push.
- `release upload <version>` for pushing local artifacts afterwards.
- `release` growing installer-payload + zip steps for the Go type.
- v0.0.3: `update check` / auto-install from local file or GitHub using
  `data/install.json`.
