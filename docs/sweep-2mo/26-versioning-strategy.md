# 26 — Versioning Strategy (every project type)

| Type | File | Keys |
|------|------|------|
| android-gradle | `gradle.properties` | `app.versionCode` (+1 on set), `app.versionName` (=X) |
| android kts | `app/build.gradle.kts` | `versionCode = N`, `versionName = "X"` R/W |
| go (+VERSION) | `VERSION` | single string |
| wails | `wails.json` | `info.productVersion` |
| cmake/generic | config-declared | `version.file` in `config.json` |

- `version [--set X|bump X|--code-only]` UX stays; `ValidateVersion` guards
  SemVer-ish (`v` prefix optional, no spaces).
- Every `VersionWrite` records old→new in metrics + notes header.
- `release` bumps BEFORE tests (mirrors `scripts/version.py`).

## Read

- SemVer: https://semver.org/ ; Gradle versioning docs.
- Your `scripts/version.py`, `internal/project/version*.go`.

## Tasks

- [ ] Kts R/W fuzz tests (comments, multiline `versionCode`).
- [ ] `version --explain` shows file + keys + next code.

## Acceptance

```powershell
go test ./internal/project/ -run TestVersion -v
```
