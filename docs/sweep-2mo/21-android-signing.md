# 21 — Android Signing (apksigner, verify, hygiene)

## Current state

`internal/android/sign.go` (+`sign_test.go`): locate → sign → verify.
`signing.go` (+`signing_test.go`): `SigningConfig`, `SigningDir`
(`<data-root>/signing/<project>/`, keystore OUTSIDE the repo),
`ResolveSigning` (flags > project `signing.json` > global `signing.json` >
env), `KeystorePassword` (env `RF_KEYSTORE_PASS` only, never disk),
`SignApkCapture`/`InstallCapture` (combined-output variants for the TUI).
TUI `sign [apk] [--alias X]` (`app/sign.go`), CLI `sign` (`cmd/sign.go`,
password env else stdin prompt) + `devices` (`cmd/devices.go`).
`release --dry-run` + tag safety exist.

## Harden

1. Locate order: config path → PATH → `$ANDROID_HOME/build-tools/<newest>/apksigner[.bat]`.
   Log resolved path + version (`apksigner version`).
2. Verify mandatory: `apksigner verify --print-certs` after every sign;
   fail release on verify miss. Store cert fingerprint in release manifest.
3. Rotation: `doctor` warns if keystore <30d expiry or debug-key in release config.
4. No plaintext: keystore passwords only via env `RF_KEYSTORE_PASS` (session)
   or prompt; never config file, never logs (redact `--ks-pass`).
5. V1/V2/V3 scheme flags explicit (`--v2-signing-enabled true`); minSdk
   influences scheme choice — document.

## Snippet — verify gate

```go
if err := SignApk(apksigner, apk, ks, alias, pass); err != nil { return err }
certs, err := VerifyApk(apksigner, apk) // --print-certs
if err != nil { return fmt.Errorf("verify %s: %w", apk, err) }
manifest.Signer = certs.SHA256
```

## Read

- apksigner: https://developer.android.com/tools/apksigner
- App signing: https://developer.android.com/studio/publish/app-signing
- Your `scripts/sign.py` + `scripts/config.py` (locate/sign semantics to preserve).

## Tasks

- [ ] Print-certs capture + manifest entry + expiry check.
- [ ] `--dry-run` prints exact apksigner lines without executing.
- [ ] Redaction test: password never appears in persisted logs.

## Acceptance

```powershell
releaseforge release 9.9.9 --dry-run  # version is positional
# shows apksigner + verify lines; no password in output
```
