# TUI design

## Inspiration

Claude Code / Gemini CLI / OpenCode: **persistent command input**, live panes, open-folder flow, minimal chrome.

## Layout

```
┌─ ReleaseForge 0.0.1  ·  SomeProject  ·  android-gradle  ·  main ─┐
│ Tabs:  Overview │ Git │ Tools │ Logs │ Config                      │
├────────────────────────────────────────────────────────────────────┤
│  (active pane — scan summary, commit list, tool list, help text)   │
├────────────────────────────────────────────────────────────────────┤
│ > help                                                   [history] │
└────────────────────────────────────────────────────────────────────┘
```

## Tabs (v0.0.1 minimum)

| Tab | Content |
|-----|--------|
| Overview | Name, root, type, version, last scan |
| Git | Recent commits and tags from scan cache |
| Tools | Detected tools, frameworks, config files |
| Logs | Later: build logs; v0.0.1 may show scan messages |
| Config | Data-root path, project config path (read-only first) |

Add Build / Metrics when builds exist (0.0.2+).

## Command bar commands (v0.0.1)

| Input | Action |
|-------|--------|
| `open` | Native folder picker (or path prompt) → scan |
| `open <path>` | Scan given path |
| `scan` / `rescan` | Rescan current project |
| `recent` | List recent projects; optional `recent <n>` to reopen |
| `status` | Short status of current project + data root |
| `help` / `-h` / `?` | Show help **in the main pane** |
| `quit` / `exit` / `q` | Leave TUI |

Unknown commands: short error + hint to type `help`.

## Help view

Rendered in the main pane (Viewport), not a one-line toast. Include command table above and note that CLI flags are available via `releaseforge --help` outside the TUI.

## Folder open

1. Prefer OS folder browser on Windows.
2. On cancel, stay on previous project.
3. On success: scan → update recent list → switch Overview to new project.

## Libraries

- bubbletea, lipgloss, bubbles (textinput, viewport, list), huh (init form, path fallback)
