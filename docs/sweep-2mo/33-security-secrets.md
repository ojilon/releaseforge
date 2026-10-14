# 33 — Security: Secrets, Keystores, Hygiene

## Non-negotiables (already partly enforced — keep)

1. Keystore passwords: `RF_KEYSTORE_PASS` env (session-only) or prompt;
   NEVER `signing.json`, NEVER `config.json`, NEVER `.releaseforge.json`
   seed, NEVER logs (see `looksSecret` in `scan.go`, scrub in `27`).
   Keystore lives in `<data-root>/signing/<project>/` (outside the repo);
   TUI signs via `SignApkCapture` (combined output, no `pass:` leak —
   apksigner prints no secret).
2. `git` safe.directory + ssh verification for automation (`10` actions log
   actor + timestamp).
3. `gh` token via `gh auth` (env `GH_TOKEN`), never stored by us.
4. Release zips contain no keystores, no `local.properties`, no `*.jks`.
   Add a `deny-list` check before zip (`22`).
5. Updater (`24`) verifies sha256 + cosign/sigstore where available before flip.

## Snippet — zip deny-list

```go
var deny = []string{".jks",".keystore","local.properties","google-services.json"}
```

Huh password prompt already exists — reuse, don't rebuild.

## Read

- OWASP secrets management cheatsheet; `gh auth` docs; apksigner key rotation docs.
- Your `scripts/config.py` (must never hold secrets — enforce in tests).

## Tasks

- [ ] Deny-list zip gate + redaction tests + `doctor --security` report.
- [ ] Audit `configs/` samples for secret-shaped keys (CI grep).

## Acceptance

```powershell
Select-String -Pattern 'password|ks-pass|token' configs/ docs/sweep-2mo/  # only docs prose
go test ./internal/... -run 'Secret|Redact|Scrub' -v
```
