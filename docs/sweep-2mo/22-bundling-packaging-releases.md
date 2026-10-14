# 22 — Bundling & Packaging Releases (APKs, zips, layouts)

## Current state

`scripts/package.py` + `zip_release.py` reference; Go port writes under
data-root `releases/<version>/` with config `appName`; notes in zips.
Go releases now also build + zip `installer-<version>.exe`
(`cmd/release.go:releaseGo`; same build as `build-installer.bat` step 2).
Lane-C artifacts: `build-installer.bat` → `dist/releaseforge.exe` +
`dist/installer.exe` + `dist/releaseforge-<version>.zip`.

## Target layout (data-root, canonical)

```
projects/<name>/releases/<version>/
  <App>-<version>-debug.apk
  <App>-<version>-release.apk        (signed)
  <App>-<version>-release-unsigned.apk (kept for forensics)
  libs/MANIFEST.json                 (from 14: which .so went in)
  notes.md  manifest.json  checksums.sha256
  <App>-<version>.zip                (apks + notes + manifest)
```

- `manifest.json`: `{version, versionCode, git_rev, abis[], signer_sha, built_at,
  tool_version}`. `checksums.sha256` over all files.
- App name from config (`artifacts.app_name`), never hard-coded.
- `zip` deterministic: sorted entries, fixed mtime (reproducible).
- `install`/`run` consume the signed APK path from manifest (no glob guessing).

## Snippet — manifest write

```go
type ReleaseManifest struct {
    Version, Code, GitRev, Tool string
    ABIs []string; SignerSHA string; BuiltAt string
}
sha := sha256.SumOf(files...); write("checksums.sha256", sha)
```

## Read

- `zip` spec (determinism), `sha256sum` format; reproducible-builds.org.
- Your `scripts/package.py`, `zip_release.py` (name conventions to preserve).

## Tasks

- [ ] Manifest + checksums + deterministic zip + `notes.md` in every zip.
- [ ] `release --dry-run` prints layout without writing.
- [ ] Scan-driven artifact discovery (no hardcoded `app-debug.apk`).

## Acceptance

```powershell
releaseforge release 9.9.9 --dry-run  # version is positional
# lists exact files above; rerun → identical sha256
```
