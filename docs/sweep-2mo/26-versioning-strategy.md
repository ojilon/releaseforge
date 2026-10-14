# 26 — Versioning Strategy (every project type)

| Type | File | Keys |
|------|------|------|
| android-gradle | `gradle.properties` | `app.versionCode` (+1 on set), `app.versionName` (=X) |
| android kts | `app/build.gradle.kts` | `versionCode = N`, `versionName = "X"` R/W |
| go (+VERSION) | `VERSION` (**gitignored, local-only**) | single string |
| wails | `wails.json` | `info.productVersion` |
| cmake/generic | config-declared | `version.file` in `config.json` |

- `version [--set X|bump X|--code-only]` UX stays; `ValidateVersion` guards
  SemVer-ish (`v` prefix optional, no spaces).
- `build [variant] [version]` writes the version FIRST (`VersionWrite`,
  never committed) so the stamp uses it (`cmd/build.go`, `jobs.go:doBuild`).
- `VERSION` is gitignored: each machine bumps its own; fresh clones start
  without it (Detect still types `go`; stamp falls back to `dev` until
  `version --set` recreates it).
- KNOWN DRIFT: `internal/version/version.go:ToolVersion` default (`0.0.2`)
  is the fallback for UNSTAMPED dev builds (`go run .`) and lags `VERSION`.
  Stamped builds (`build-installer.bat`, `release`) are unaffected.
  Proposed: `version.SyncDefault(root, ver)` rewriting the `var ToolVersion`
  line, called by `version --set`, `build … <ver>`, `releaseGo` — bat needs
  no change.
- Every `VersionWrite` records old→new in metrics + notes header.
- `release` bumps BEFORE tests (mirrors `scripts/version.py`).

## Read

- SemVer: https://semver.org/ ; Gradle versioning docs.
- Your `scripts/version.py`, `internal/project/version*.go`.

## Tasks

- [ ] Kts R/W fuzz tests (comments, multiline `versionCode`).
- [ ] `version --explain` shows file + keys + next code.
- [ ] `SyncDefault`: keep `version.go:ToolVersion` default in step with
  `VERSION` on every write path (see drift note above).

## Acceptance

```powershell
go test ./internal/project/ -run TestVersion -v
```
