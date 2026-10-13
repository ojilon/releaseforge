# 52 — Play Publish (tracks, staged rollout, gates)

## Goal

After `23` GitHub release, an OPTIONAL Play track promotion with the same
gate discipline — never silent.

## Design

- `rf play --track internal --aab <path>`: wraps `fastlane supply` or
  Play Developer API (`edit.insert → bundle.upload → edit.commit`);
  service-account JSON via env `RF_PLAY_SA` (never config file — see `33`).
- Rollout: `internal → closed → production` with `--rollout 0.1` fractions;
  `play status` shows current track versions; `play halt` stops rollout.
- Gates reuse `23`: sign+verify ✓, AAB ✓ (`49`), mapping archived ✓ (`50`),
  release notes from `DraftNotes` ✓.
- `--dry-run` prints API calls; no network in tests (recorded fixtures).

## Read

- Play API + tracks: https://developers.google.com/android-publisher ;
  fastlane supply: https://docs.fastlane.tools/actions/supply/

## Tasks

- [ ] `internal/play/` wrapper + dry-run + status/halt; SA auth docs.

## Acceptance

```powershell
rf play --dry-run --track internal --aab releases\<ver>\app.aab
```
