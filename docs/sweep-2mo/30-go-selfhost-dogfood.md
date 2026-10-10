# 30 — Go Self-Host: Dogfood Releases (ReleaseForge releasing itself)

## Current state

v0.0.1 self-hosted tag + GitHub pre-release + installer (`installer/`,
`RELEASE.md`, `VERSION`); `release` supports `go` type
(`go test ./...`, `go build`).

## Keep sharp

- `VERSION` file is truth; `release --version X` writes it, tags `vX`,
  attaches `dist/releaseforge.exe` (+ checksums + notes).
- Installer (`installer/`) seeds data-root; test install in a VM snapshot
  before every minor (`installer-test.exe` exists — use it).
- `go vet + go test` gate is non-skippable; add `gofmt -l` gate.

```powershell
go vet ./...; go test ./...
go build -o dist/releaseforge.exe .
releaseforge release --version <next> --dry-run
```

## Read

- Your `RELEASE.md`, `installer/`, `docs/v0.0.2/PROGRESS.md` (done-done bar).

## Tasks

- [ ] `release` go-path asserts `VERSION` ↔ tag ↔ binary `--version` match.
- [ ] Installer smoke test scripted (`scripts/`-style, but Go).

## Acceptance

```powershell
releaseforge release --dry-run --version 9.9.9  # inside releaseforge itself
```
