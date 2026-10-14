# 30 — Go Self-Host: Dogfood Releases (ReleaseForge releasing itself)

## Current state

Lane-C packaging exists: `build-installer.bat [version]` (version =
arg else `VERSION` file) builds stamped `dist/releaseforge.exe` +
`dist/installer.exe` + `dist/releaseforge-<version>.zip`.
`release <version> [--pre]` (`cmd/release.go:releaseGo`) does version →
test → stamped binary → STAMPED installer (`installer-<version>.exe`) →
notes → zip (binary+installer+notes) → tag → `gh release create`;
`--dry-run` prints the plan incl. installer lines. `VERSION` is gitignored
(local-only); `internal/version/version.go` default still needs `SyncDefault`
(see `26`). Installer smoke path: `installer-test.exe` + `--non-interactive
--dest` (proven in temp dirs).

```powershell
go vet ./...; go test ./...
.\build-installer.bat 0.0.7
releaseforge release 0.0.7 --pre --dry-run
```

## Read

- Your `RELEASE.md`, `installer/`, `docs/v0.0.2/PROGRESS.md` (done-done bar).

## Tasks

- [x] `build-installer.bat` (stamped pair + zip) — done, verified (binary
  `--version` + installer `install.json` agree).
- [x] `release` Go path builds + zips stamped installer — done.
- [ ] `release` go-path asserts `VERSION` ↔ tag ↔ binary `--version` ↔
  installer `install.json` match.
- [ ] Installer smoke test scripted (`scripts/`-style, but Go).

## Acceptance

```powershell
releaseforge release 9.9.9 --dry-run  # inside releaseforge itself (note: version is positional, not --version)
```
