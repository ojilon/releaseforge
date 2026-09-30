# Release pipeline

Full pipeline applies from **v0.0.2+**. v0.0.1 only needs scan + cache (see `docs/09-scan-foundation.md`).

## Canonical reference in this repo

Legacy Python implementation lives under **[`scripts/`](../scripts/)** (see [`scripts/README.md`](../scripts/README.md)).

Agents and implementers must read those files for concrete behaviour (version bump rules, APK paths, apksigner invocation, `gh release create` argument shape). **Do not** assume access to any other GitHub repository for this logic.

Go reimplementation must **improve** on that reference (data-root artifacts, live logs, configurable app name, structured errors) — not call the Python scripts from production paths.

## Canonical steps (Android Gradle)

Order mirrors [`scripts/release.py`](../scripts/release.py):

1. Resolve project + load config/scan cache
2. Set version — same rules as [`scripts/version.py`](../scripts/version.py) (`versionName` set, `versionCode` += 1 in `gradle.properties`)
3. Unit tests (optional skip) — e.g. `:app:testDebugUnitTest`
4. `assembleDebug` then `assembleRelease`
5. Package APKs — logic of [`scripts/package.py`](../scripts/package.py), but output under data-root `releases/<version>/` by default; optional in-repo mirror; **app name from config**, not hard-coded
6. Sign + verify — [`scripts/sign.py`](../scripts/sign.py) (apksigner locate, password prompt)
7. Notes — start from [`scripts/notes.py`](../scripts/notes.py) template; **fill Changes from git history**
8. Zip — [`scripts/zip_release.py`](../scripts/zip_release.py)
9. Annotated tag `v<version>`
10. `gh release create` with notes file + artifacts (`--prerelease` optional) — see `publish_github_release` in `release.py`
11. Update metrics/history under data-root

## Self-host pipeline (ReleaseForge 0.0.2)

1. `go test ./...`
2. `go build -o dist/releaseforge.exe` (or platform matrix later)
3. Notes from git since last tag
4. Tag + GitHub release attaching the binary

## Safety

- Confirm publish in TUI
- No password files (`scripts/config.json` must never hold secrets; gitignored if created)
- Do not overwrite existing tags without explicit force
