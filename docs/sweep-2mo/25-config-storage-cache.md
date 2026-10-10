# 25 — Config, Storage, Cache (data-root truth)

## Current state

`internal/storage/` (`EnsureProjectLayout`, `D:` default),
`internal/config/` (global/project JSON, `DiscoverDataRoot`),
`scan.go:315 MergeConfig`, `meta.json` freshness by root path.

## Rules (do not regress)

1. Data-root mandatory after `init`; git trees stay clean (builds/logs/
   releases under `D:/ReleaseForgeData` or chosen root).
2. Tool-owned vs user-owned keys (`scan.go:315`): detection overwrites
   type/root/version/build-defaults; signing/github/appName/custom tasks survive.
3. Seed `.releaseforge.json` honoured unless secret-looking (`looksSecret`).
4. Cache invalidation moves to hash-based (`02`); `rescan` always full refresh.

Layout:

```
<root>/global.json  history/  projects/<name>/{config.json,cache/{scan.json,meta.json},logs/,releases/,native-cache/}
```

## Read

- Your `docs/05-config-and-storage.md`, `docs/09-scan-foundation.md`.
- XDG / Windows known-folders for future non-Windows ports.

## Tasks

- [ ] JSON-schema version field in `config.json` + migrator.
- [ ] `config validate` command (used by `doctor`).
- [ ] Hash-based freshness (with `02`).

## Acceptance

```powershell
releaseforge init; releaseforge config-validate; releaseforge scan --explain .
```
