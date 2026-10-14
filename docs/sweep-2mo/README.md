# ReleaseForge 2-Month Sweep — Master Index

> Dedicated planning folder for the slow, sure rewrite pass.
> Location: `docs/sweep-2mo/`. Read in numeric order. Every file is
> standalone but cross-links. Code refs are `path:line` from repo root;
> v0.0.2-era refs (`internal/app/app.go`, …) carry a `> Delta` note where
> the code moved on (split TUI, theme, new commands).

## Why this folder exists

Up to this point the project was AI-built and it shows: scan + TUI +
Gradle/Android + self-host works, but the TUI logic (was
`internal/app/app.go:172 Update`, `exec` at `:383` — since split into
`internal/app/*.go`, one file per job) had no structure, styling was ad-hoc
(now `internal/tui/theme.go`), and forward ideas (prebuilt `.so` handoff,
Zig package manager, Odin/Win32 native window, test/diff dashboards,
one-key git automation, updater) have no design home. First dashboard panel
+ theme + sign/install/bat commands are built (see `01` deltas); the
rest of this sweep fixes the remainder over ~8 weeks without breaking daily use.

## Reading order

| # | File | What you get |
|---|------|--------------|
| 00 | `00-how-to-use-this-folder.md` | How to work one file per day, verify, commit |
| 01 | `01-current-codebase-map.md` | Serious file-by-file map with line refs |
| 02 | `02-scan-and-detect-rethink.md` | Scan pipeline today vs where it must go |
| 03 | `03-tui-architecture-rethink.md` | Why the single-viewport model must die |
| 04 | `04-bubbletea-bubbles-lipgloss-cookbook.md` | Charm stack done right, snippets |
| 05 | `05-cobra-cli-design.md` | CLI as scriptable twin of the TUI |
| 06 | `06-command-bar-keys-help.md` | Command bar grammar, keys, help |
| 07 | `07-window-navigation-tabs-dashboard.md` | Multi-window TUI switching design |
| 08 | `08-test-dashboard.md` | Unit + instrumented dashboard |
| 09 | `09-git-diff-dashboard.md` | Diff/commit dashboard |
| 10 | `10-git-automation-long-commands.md` | One-key wrappers for long git invocations |
| 11 | `11-git-study-guide.md` | Git docs + papers to read, then re-implement |
| 12 | `12-build-runner-async-logs.md` | Async runner, cancel, log streaming |
| 13 | `13-gradle-deep-dive.md` | Gradle + NDK + ABI today |
| 14 | `14-prebuilt-so-handoff.md` | **Advanced:** hand Gradle `.so`s instead of compiling C++ |
| 15 | `15-cmake-cpp-pipeline.md` | CMake discovery, configure, build |
| 16 | `16-zig-as-package-manager.md` | **Advanced:** Zig owns target building, not CMake |
| 17 | `17-zig-build-targets.md` | `build.zig`, cross targets, caching |
| 18 | `18-odin-for-high-perf-parts.md` | **Mature:** Odin for non-network hot paths |
| 19 | `19-win32-native-window.md` | **Near-complete:** own native window via Win32 |
| 20 | `20-wails-pc-builds.md` | Wails/desktop builds |
| 21 | `21-android-signing.md` | apksigner, verify, keystore hygiene |
| 22 | `22-bundling-packaging-releases.md` | APK + binary zips, layouts |
| 23 | `23-pushing-to-production.md` | Tags, `gh release`, promotion gates |
| 24 | `24-tooling-auto-update.md` | Active web updater for SDK/NDK/tooling |
| 25 | `25-config-storage-cache.md` | Data root, config merge, cache freshness |
| 26 | `26-versioning-strategy.md` | versionCode/Name, VERSION, wails.json |
| 27 | `27-logging-metrics.md` | Ring, persisted logs, metrics |
| 28 | `28-error-parsing-reporting.md` | Structured Gradle/CMake/Go errors |
| 29 | `29-java-support.md` | Java/Gradle + Maven CLI projects |
| 30 | `30-go-selfhost-dogfood.md` | ReleaseForge releasing itself |
| 31 | `31-generic-project-support.md` | Any-folder promise, task seeds |
| 32 | `32-perf-non-network-split.md` | GC vs non-GC split (Go vs Zig/Odin) |
| 33 | `33-security-secrets.md` | Secrets, keystore, no-plaintext rules |
| 34 | `34-testing-strategy.md` | Go tests, hardware acceptance |
| 35 | `35-eight-week-plan.md` | Week-by-week sweep schedule |

## Wave 2 (added on request — depth + missing areas)

| # | File | What you get |
|---|------|--------------|
| 36 | `36-adb-devices-deep-dive.md` | Device states, guarded install |
| 37 | `37-emulator-automation.md` | AVD create/boot/test |
| 38 | `38-logcat-viewer.md` | Filtered logcat + crash jump |
| 39 | `39-install-run-ux.md` | Manifest-driven install/run |
| 40 | `40-tui-state-machine.md` | Explicit phase FSM |
| 41 | `41-focus-routing-mouse-scroll.md` | Single-consumer focus, scroll |
| 42 | `42-huh-forms-wizards.md` | Shared forms package |
| 43 | `43-glamour-markdown.md` | Styled notes/help |
| 44 | `44-fuzzy-finder-palette.md` | ctrl+p palette |
| 45 | `45-windows-terminal-quirks.md` | coninput/pwsh/ANSI/size |
| 46 | `46-build-release-dashboards.md` | Build phases + release checklist |
| 47 | `47-doctor-checks.md` | Full tool matrix |
| 48 | `48-jdk-gradle-tuning.md` | Daemon/memory/cache |
| 49 | `49-aab-vs-apk-splits.md` | Bundles + ABI splits |
| 50 | `50-r8-proguard-shrinking.md` | Rules + mapping archive |
| 51 | `51-installer.md` | Upgrade + smoke test |
| 52 | `52-play-publish.md` | Tracks + rollout gates |
| 53 | `53-python-support.md` | pyproject/venv/tasks |
| 54 | `54-node-frontend-support.md` | npm/pnpm + frozen lock |
| 55 | `55-rust-cargo-support.md` | Toolchain + audit |
| 56 | `56-cmake-presets-ninja-cache.md` | Presets + Ninja + ccache |
| 57 | `57-changelog-grouping.md` | Conventional grouping spec |
| 58 | `58-git-hooks-templates.md` | Opt-in hooks + template |
| 59 | `59-config-schema-migration.md` | Versioned config |
| 60 | `60-dataroot-backup-restore.md` | Backup/restore |
| 61 | `61-plugin-new-project-types.md` | Plugin registry |
| 62 | `62-sbom-provenance-licenses.md` | SBOM + provenance |
| 63 | `63-crash-telemetry.md` | Opt-in, local-first |
| 64 | `64-troubleshooting-playbook.md` | Error → fix catalog |
| 65 | `65-glossary-faq.md` | Terms + FAQ |
| 66 | `66-sweep-retro-template.md` | Weekly + final retro |

## The 3 horizons

- **Now (weeks 1–4):** TUI refactor, scan hardening, dashboards (test/diff),
  git automation, error parsing, packaging/signing cleanup. No new languages.
- **Advanced (weeks 5–6):** Prebuilt `.so` handoff (`14`), Zig target
  building (`16`, `17`), updater (`24`).
- **Mature / near-complete (weeks 7–8+):** Native Win32 window (`19`),
  Odin hot paths (`18`, `32`). Do NOT start these early — they depend on
  a solid TUI + build core.

## Ground rules for the sweep

1. One file = one workstream. Finish its Acceptance before moving on.
2. Domain logic in `internal/`; `cmd/` and TUI stay thin (see `01`).
3. Every change: `go vet ./...` + `go test ./...` green before commit.
4. Update the relevant sweep doc when behaviour changes.
5. Never break `releaseforge init/scan/tui` daily flow.
