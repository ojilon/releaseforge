# TUI design

## Inspiration

Claude Code and modern Gemini-style CLIs: a **persistent command input**, live updating regions, and minimal chrome. Not a single static screen — panes and tabs change with project state.

## Layout (conceptual)

```
┌─ ReleaseForge  ·  Conductino-Android  ·  v0.0.3_2  ·  device: R58M… ─┐
│ Tabs:  Overview │ Build │ Logs │ Git │ Metrics │ Config                 │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  (active pane content — e.g. live Gradle log viewport, or metrics)      │
│                                                                         │
├─────────────────────────────────────────────────────────────────────────┤
│ > build release --abis arm64-v8a                              [history] │
└─────────────────────────────────────────────────────────────────────────┘
```

- **Header** — project name, current version, optional device, last build status.
- **Tabs** — switch with number keys or click (BubbleZone later).
- **Main pane** — Viewport (Bubbles) for logs, tables for metrics/git, forms for config.
- **Command bar** — always focused unless a modal form is open. Supports history (↑/↓), fuzzy suggestions from history + known verbs.

## Tabs

| Tab | Content |
|-----|--------|
| Overview | Scan summary, version, last build/test, quick actions |
| Build | Task list, live log, progress, ABI/NDK controls |
| Logs | Historical logs under data-root, open/re-tail |
| Git | Recent commits, tags, draft notes |
| Metrics | Simple counts/bars (builds, failures, releases) |
| Config | Project + global settings, keystore path, data-root |

## Live logs

- Build/test runners push line events to the app model.
- Viewport auto-scrolls unless the user scrolls up.
- Error lines highlighted (Lip Gloss); summary card on failure with path to full log and report.

## Commands (examples)

```
scan
status
build debug
build release --abis arm64-v8a,armeabi-v7a
test unit
test instrumented
install debug
install release
version
version --set 0.1.0
release 0.1.0 --pre
logs
notes
git log
```

Unknown input shows help snippet; Tab completes from history + verbs.

## Forms (Huh)

- `init` wizard (data-root path).
- Release confirmation (version, pre-release, edit notes).
- Keystore password (secure input).
- Device picker when multiple adb devices.

## Libraries

- bubbletea — model/update/view
- lipgloss — styles and layout
- bubbles — textinput, viewport, list, table, spinner, progress, help
- huh — multi-step forms and selects
