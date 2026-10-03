# Release guide — general (any version, any supported project)

Every task below has two blocks: **by hand** (plain tools, no ReleaseForge)
and **with ReleaseForge** (installed `releaseforge`, or `go run .` in lane A).
Read the block that matches what you are doing; they are independent paths to
the same result.

Prefix rule: in the repo dev loop write `go run .` where this doc writes
`releaseforge`. Outside the repo (installed), `releaseforge` is on `PATH`.
All examples are PowerShell.

```powershell
# lane A (this repo, throwaway storage) — run once per terminal:
cd D:\projects\releaseforge
$env:RELEASEFORGE_DATA_ROOT = ".data"

# installed (real storage, seeded by the installer) — no flags needed:
releaseforge --version
```

Global rules, both paths:

- `init` → `scan` comes before everything else for a given storage.
- The project under work is `./` by default; use `-p <path>` or `cd` there.
- Writing a version never commits. You review and commit manually, always.
- A release finishes with a tag push. Never run the same version twice.

Supported types today: `go` and `android-gradle`. Wails / CMake / Python reuse
this shape once their runners land; until then they are scan + notes only.

---

## 1. Viewing a project — type, version, git state, paths

### 1a. With ReleaseForge

```powershell
releaseforge scan D:\path\to\proj     # detect + cache + recent (run when things change)
releaseforge status                   # tool + project + type + version + git + devices + paths
releaseforge version                  # tool version + project version
releaseforge recent                   # projects you have opened before
releaseforge logs                     # persisted build/test logs, newest first
```

`status` is the one-line health check. `scan` output ends with `Config:` and
`Cache:` paths — that is where the tool keeps what it knows.

### 1b. By hand

```powershell
Get-Content D:\path\to\proj\VERSION                    # go projects
Select-String "^app\.version(Name|Code)" D:\path\to\proj\gradle.properties   # android
git -C D:\path\to\proj branch --show-current; git -C D:\path\to\proj log --oneline -10
git -C D:\path\to\proj tag -l --sort=-creatordate | Select-Object -First 5
```

Version lives in exactly one place per type: `VERSION` file (go),
`gradle.properties` (`app.versionCode` + `app.versionName`), `wails.json`
(`info.productVersion`). Everything else is derived.

---

## 2. Testing

### 2a. With ReleaseForge

```powershell
releaseforge test                       # go: go test ./... | android: unit tests
releaseforge test all                   # android only: unit + instrumented
releaseforge logs                       # open the persisted log on failure
```

Live output streams; the full log lands under
`projects/<name>/logs/test-*.log`. The `!` lines at the end are the headline
errors.

### 2b. By hand

```powershell
cd D:\path\to\proj
go test ./...                                             # go
.\gradlew.bat :app:testDebugUnitTest                      # android unit
.\gradlew.bat :app:connectedDebugAndroidTest               # android instrumented (device needed)
```

Same commands the tool runs (see `internal/build/`). If these fail here, the
tool's `test` fails identically — fix it here first, it is faster.

---

## 3. Compiling to a binary / APK

### 3a. With ReleaseForge

```powershell
releaseforge build                    # debug: go build (+ version stamp) | gradle assembleDebug
releaseforge build release            # release: -trimpath, stripped | gradle assembleRelease
```

Go binaries land in `projects/<name>/builds/` already stamped with the
project version (`--version` on the binary proves it). Android APKs stay
under `app/build/outputs/` until packaging (section 4).

### 3b. By hand

```powershell
# go — note the version stamp; without it the binary reports "dev":
$v = (Get-Content VERSION).Trim()
go build -ldflags "-X <module>/internal/version.ToolVersion=$v" -o app.exe .
go build -trimpath -ldflags "-s -w -X <module>/internal/version.ToolVersion=$v" -o app.exe .

# android:
.\gradlew.bat assembleDebug
.\gradlew.bat assembleRelease
```

`<module>` is the `module …` line in `go.mod`. The stamp key must match the
variable the program prints — for ReleaseForge itself it is
`github.com/ojilon/releaseforge/internal/version.ToolVersion`.

---

## 4. Packaging a version folder + notes

Release folder convention: `releases/<version>/` holding the artifacts plus
`notes.md`. Under the tool it lives in the data root
(`projects/<name>/releases/<version>/`); by hand you make it wherever you
like (conventionally `dist/` or `release/<version>/` in the repo, gitignored).

### 4a. With ReleaseForge

```powershell
releaseforge notes 0.0.2        # notes.md from commits since last tag → releases/0.0.2/
releaseforge build release      # binary/APK first (previous section)
```

Packaging itself (copy + rename + zip) currently happens inside `release`
(section 6a). Standalone `package` is not a command yet — until then, zip by
hand from the release folder:

```powershell
Compress-Archive -Force releases\0.0.2\myapp-0.0.2.exe myapp-0.0.2.zip
```

### 4b. By hand

```powershell
New-Item -ItemType Directory release\0.0.2 -Force
Copy-Item app.exe release\0.0.2\myapp-0.0.2.exe
# notes: one line per commit since the last tag:
git log v0.0.1..HEAD --pretty="- %s (%h, %ad)" --date=short > release\0.0.2\notes.md
Compress-Archive release\0.0.2\myapp-0.0.2.exe myapp-0.0.2.zip
```

---

## 5. Bumping the version

### 5a. With ReleaseForge

```powershell
releaseforge version --set 0.0.2      # go: rewrites VERSION | android: name + code+1
git diff                              # review — the tool never commits
```

### 5b. By hand

```powershell
"0.0.2" | Set-Content VERSION         # go (trailing newline is fine)
# android gradle.properties: set app.versionName=0.0.2 AND increment app.versionCode by 1
```

Commit the bump yourself on both paths: `git add …; git commit -m "…";
git push`.

---

## 6. Tagging and publishing to GitHub

### 6a. With ReleaseForge — `release` (one run per version, never two)

```powershell
git status --short                    # must be clean; commit first
git push origin main
releaseforge release 0.0.2 --pre
```

It runs: version write → test → build(s) → package → notes → zip →
`tag v0.0.2` → `git push origin v0.0.2` → `gh release create` (+ `--prerelease`
from `--pre`). Android adds the sign step (keystore password prompted once,
never stored). Then attach anything the pipeline does not know about yet
(e.g. the installer — automated from v0.0.2):

```powershell
gh release upload v0.0.2 dist/installer.exe
```

Without `gh` on PATH it stops after the local tag + zip (exit 0) and prints
the exact publish command for later. A rerun with an existing tag fails on
the tag step — that is the guard working, not a bug.

### 6b. By hand — tags and GitHub release

```powershell
git tag -a v0.0.2 -m "Release 0.0.2"
git push origin v0.0.2
$n = "release\0.0.2\notes.md"
gh release create v0.0.2 --prerelease --notes-file $n myapp-0.0.2.exe myapp-0.0.2.zip
gh release upload v0.0.2 extra-asset.exe
gh release view v0.0.2                # verify assets
```

Drop `--prerelease` for a final release. If `gh` is missing, the tag push is
still the releasable state — publish from any machine with `gh` later.

---

## 7. Installing on a PC

- **ReleaseForge itself:** build `releaseforge.exe` + `installer.exe`, pack
  `dist/`, run the installer, pick drive + folder by number. Result:
  `<chosen>/ReleaseForge/{bin,data}`. Re-running updates without touching
  data. (Full walkthrough: `docs/10-bootstrap-v001.md` section 9.)
- **Your own Go project:** copy the stamped binary where it belongs; no
  registry, no services in v0.0.1 scope.
- **Android:** `adb install -r <apk>` (device connected); the tool's
  `install [debug|release]` wraps exactly this.

---

## 8. Full checklists

### 8a. All-manual release (any version, any machine with git + gh)

```powershell
git status --short; git push origin main
# bump (section 5b) → commit → push
# test (section 2b) → compile (section 3b) → package + notes (section 4b)
git tag -a vX.Y.Z -m "Release X.Y.Z"; git push origin vX.Y.Z
gh release create vX.Y.Z [--prerelease] --notes-file <notes> <assets...>
gh release view vX.Y.Z
```

### 8b. ReleaseForge release (same release, tool-driven)

```powershell
git status --short; git push origin main
# bump (section 5a) → commit → push
releaseforge test; releaseforge build release; releaseforge notes X.Y.Z
releaseforge release X.Y.Z [--pre]
gh release upload vX.Y.Z <extra-assets...>
gh release view vX.Y.Z
```

Both end in the same place: clean tree, `vX.Y.Z` on origin, GitHub release
with notes + assets.

---

## 9. Troubleshooting

| Symptom | Cause → fix |
|---|---|
| `data root not initialised` | Storage unknown here → `init`, or set `$env:RELEASEFORGE_DATA_ROOT`, or pass `--data-root`. Bare `releaseforge`/`go run .` opens the TUI, which needs storage too. |
| `unknown command: release` inside the TUI bar | TUI has no `release` verb — run it at the PowerShell prompt. |
| Second `release` fails on tag | Guard working. Already released → verify with `gh release view`. Need a redo → delete remote+local tag first, deliberately. |
| Release dies in `test` | Read the newest `logs/test-*.log` tail (or the `!` lines). Fix with plain `go test` / gradle first. |
| `gh release ...` fails but tag pushed | Publish manually per section 6b — tag + local zip is the safe state. |
| Binary reports wrong version | Built without the `-ldflags -X` stamp (section 3b), or stale `VERSION`. Rebuild. |
| Installer can't find payload | Put `releaseforge.exe` next to `installer.exe`, or pass `--bin`. |
