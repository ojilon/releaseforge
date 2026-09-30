# Release pipeline

Canonical flow for an Android Gradle project (mirrors and generalises Conductino-Android `scripts/release.py`).

## Steps

1. **Resolve project** — load config, detect type, confirm working tree is clean enough (warn on dirty).
2. **Version** — set `versionName`, increment `versionCode` in `gradle.properties` (or equivalent).
3. **Tests** — run unit tests (optional flag to skip). Fail → stop, surface report.
4. **Build debug** — `assembleDebug` (or configured task). Capture log.
5. **Build release** — `assembleRelease`. Capture log.
6. **Package** — copy APKs into `<data-root>/projects/<name>/releases/<version>/` (and optional repo `release/<version>/` mirror). Rename consistently.
7. **Sign** — `apksigner sign` + `verify` on the release APK. Prompt for keystore password.
8. **Notes** — generate markdown from git log since previous tag (template sections: Highlights, Changes, Bug fixes, Known issues). User can edit before continue.
9. **Zip** — archive APKs (and any extra assets).
10. **Tag** — annotated tag `v<version>`.
11. **Publish** — `gh release create` (or API) with notes file and artifacts; `--prerelease` when requested.
12. **Record** — update metrics and history under data-root.

## CLI shape

```bash
releaseforge release 0.1.0 --pre
releaseforge release 0.1.0 --skip-tests
releaseforge release 0.1.0 --notes-only
```

## TUI shape

Command bar: `release 0.1.0 --pre`  
Wizard (Huh) confirms version, pre-release flag, whether to install after, and shows a notes editor pane before publish.

## Wails / other types

Same orchestration; build step uses `wails build` or CMake; signing/install steps are no-ops or replaced by installer packaging as configured.

## Safety

- Never overwrite an existing GitHub release tag without explicit force.
- Never store keystore passwords.
- Prefer dry-run / confirm for publish step in TUI.
