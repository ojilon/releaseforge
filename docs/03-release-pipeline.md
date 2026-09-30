# Release pipeline

Full pipeline applies from **v0.0.2+**. v0.0.1 only needs scan + cache (see `docs/09-scan-foundation.md`).

## Canonical steps (Android Gradle)

Matches and improves Conductino-Android `scripts/release.py`:

1. Resolve project + load config/scan cache
2. Set version (`versionName`, increment `versionCode`)
3. Unit tests (optional skip)
4. `assembleDebug` / `assembleRelease`
5. Package APKs into data-root `releases/<version>/` (optional repo mirror)
6. Sign release APK (`apksigner`) + verify
7. Notes from git history (+ template sections)
8. Zip artifacts
9. Annotated tag `v<version>`
10. `gh release create` with notes + assets (`--prerelease` optional)
11. Update metrics/history

## Self-host pipeline (ReleaseForge 0.0.2)

1. `go test ./...`
2. `go build -o dist/releaseforge.exe` (or platform matrix later)
3. Notes from git since last tag
4. Tag + GitHub release attaching the binary

## Safety

- Confirm publish in TUI
- No password files
- Do not overwrite existing tags without explicit force
