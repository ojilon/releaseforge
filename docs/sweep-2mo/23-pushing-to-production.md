# 23 — Pushing to Production (tags, gh release, gates)

## Current state

`internal/github/release.go`, `release` command with tag safety + notes;
`--prerelease` flag; metrics/history update.

## Gates (no silent publish)

1. `pre-publish` checklist rendered in TUI: version bumped? tests green?
   apksigner verify ok? manifest + checksums present? tag `vX` not exists
   (or `--force` with double confirm)?
2. Tag: annotated `git tag -a v<version> -m "…"`; push tag; on push fail,
   offer `git push origin vX` retry (never auto-force).
3. `gh release create vX --title --notes-file notes.md [--prerelease] <assets>`:
   assets = zip + signed APK + checksums. Capture URL into metrics + status.
4. Promote: `prerelease → release` is a separate confirmed action
   (`release promote vX` → `gh release edit --prerelease=false`).
5. Rollback note: releases are immutable; document yank (new patch, not delete).

## Snippet — tag safety

```go
if git.TagExists(dir, "v"+ver) && !force {
    return fmt.Errorf("tag v%s exists — use --force (double confirm)", ver)
}
```

## Read

- `gh release create/edit`: https://cli.github.com/manual/gh_release_create
- `git tag`: https://git-scm.com/docs/git-tag ; SemVer: https://semver.org/

## Tasks

- [ ] Publish checklist UI + `--dry-run` full preview (tag + gh lines).
- [ ] `release promote` subcommand; URL persisted to metrics.
- [ ] Live-hardware acceptance still required (real `gh` + device).

## Acceptance

```powershell
releaseforge release --dry-run --version 9.9.9
releaseforge release promote --dry-run v9.9.8
```
