# 51 — Installer (seeded data-root, smoke test)

## Current state

`installer/main.go` (+`main_test.go`) builds `installer.exe`;
`installer-test.exe` exists; seeded data-root behavior.
Lane-C flow: `build-installer.bat [version]` builds STAMPED
`dist/releaseforge.exe` + `dist/installer.exe` (`-ldflags ToolVersion`)
+ versioned zip; installer records the stamped version in `install.json`
(verified: binary `--version` ≡ installer line ≡ `install.json`).
`release <version>` Go path builds `installer-<version>.exe` the same way
and zips/publishes it with the binary + notes. Temp-dir
`--non-interactive --dest` installs proven (see `30`).

## Harden

1. Seeded root: installer proposes `D:\ReleaseForgeData` (fallback `C:`),
   writes `global.json`, never overwrites existing non-empty root without confirm.
2. Upgrade: detect prior install, migrate `config.json` schema (`59`),
   backup `history/` before touch.
3. Smoke test (automated, VM snapshot): install → `init` → `scan` fixture →
   `tui --smoke` (opens, scans, quits 0) → uninstall leaves data-root intact.
4. Sign installer exe (same hygiene as `21`); checksum published per release (`22`).

## Read

- Inno Setup docs (if Inno) or NSIS; your `installer/main.go` + `RELEASE.md`.

## Tasks

- [ ] Upgrade/migrate path + scripted smoke test + signed installer checklist.

## Acceptance

```powershell
.\installer-test.exe  # full install→smoke→uninstall, data-root preserved
```
