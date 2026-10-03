# v0.0.2 doc 11: Sign and Android release hardening

## Goal

Make the Android sign + release path trustworthy: a working apksigner search,
safe password handling, notes in the zip, a no-side-effects dry run, and tag
safety.

## Current state

- `FindApksigner` (`internal/android/sign.go:19`) order custom → PATH →
  SDK newest → fallback is right, but the Python reference it was ported from
  has an inverted `if not sdk` + `build_tools` (vs `build-tools`) bug —
  verify the Go SDK branch actually works (it looks correct; prove with a test).
- Password via huh masked input (`cmd/release.go:145`), never stored [DONE].
- Zip omits `notes.md`; `release --dry-run` doesn't exist; re-running a
  released version fails late at `git tag` instead of early.
- Python publishes with `--title` (`scripts/release.py:50`); Go omits it
  (fine — document the choice, don't re-add).

## Design

1. Harden `FindApksigner`: fixture test with a fake SDK tree
   (`build-tools/34.0.0/apksigner[.bat]`, `35.0.0/…​`) asserting newest wins;
   PATH-lookup test with fixture bin dir. No behavior change if already right.
2. Zip includes `notes.md` alongside APKs (Go) — matches "release folder =
   the releasable unit" from `RELEASE.md`.
3. `release --dry-run`: print the ordered plan (version write preview with
   diff of the changed lines, test/build argv, package destinations, tag
   name, gh argv) and exit 0 touching nothing — no VERSION write, no tag,
   no processes. Especially valuable on the Celeron (no Gradle spawned).
4. Tag safety: before tagging, `git rev-parse --verify refs/tags/<tag>`;
   if it exists, abort early with "already released — see `gh release view`"
   (covers both the double-run case and pre-existing tags).
5. Keep `gh`-missing graceful stop (already implemented,
   `cmd/release.go:272`); keep `gh`-only publishing (decided).

## Files to touch

- `internal/android/sign.go` (tests; fix only if fixtures prove a bug).
- `cmd/release.go` (`--dry-run`, tag-exists precheck, zip += notes).
- `internal/git/notes.go` (tiny `TagExists` helper).

## Steps

1. Fixture SDK tree + PATH tests for `FindApksigner`.
2. Add `TagExists`; call it first in `publishTagAndRelease` (both types).
3. Add `--dry-run` printing the full plan; assert zero side effects in test
   (temp git repo + temp data root: no tag, no VERSION change, no zip).
4. Include `notes.md` in both zips (android + go).
5. Document the `--title` omission in code comment (one line).

## Tests to add

- SDK newest-wins; custom-path precedence; fallback string.
- Dry-run on temp repo: no filesystem/git changes; output contains all 8
  android steps (or 6 go steps).
- Existing tag → early clean error before any version write.

## Acceptance checks

- `go vet ./...`, `go test ./...` pass.
- `release 0.0.2 --dry-run` on this repo prints plan, changes nothing
  (`git status --short` clean apart from pre-existing).
- Second `release <same>` fails fast on existing tag (use a scratch version
  like `0.0.0-dry` — still creates no tag thanks to dry-run-first habit).

## Out of scope

Play-track uploads, key rotation, password managers, APK signature-scheme
choices. No Gradle spawned by dry-run.
