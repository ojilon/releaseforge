# 20 — Wails / PC Builds (desktop companion to Android)

## Current state

`wails.json` detection, `productVersion` read, type `wails`. No build runner.

## Target

- `wails build` / `wails dev` via runner (`12`); frontend lockfile (npm/pnpm)
  recorded in scan; `wails.json` `info.productVersion` writable via
  `version --set` (same UX as Gradle bump).
- `release` supports `wails` type: version → test (`go test` + `npm test`
  if present) → `wails build` → zip installer/exe → notes → tag → `gh release`.
- This is also the Phase-2 native-window host candidate (see `19`).

## Snippet

```go
func (r WailsRunner) BuildArgs(variant, out, _ string) ([]string, error) {
    return []string{"build", "-clean"}, nil // program: "wails"
}
```

## Read

- Wails CLI: https://wails.io/docs/reference/cli ; `wails.json` schema.
- Your `docs/07-wails-and-pc.md`.

## Tasks

- [ ] Wails runner + version R/W + `doctor` (wails doctor passthrough).
- [ ] `release` path for `wails` with data-root artifacts.

## Acceptance

```powershell
releaseforge build release  # inside a wails tree → wails build
```
