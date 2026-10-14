# 07 — Window Navigation: Tabs / Dashboard Shell

## Current state

Single viewport + command bar + first dashboard panel (`docs/04-tui-design.md`
— tabs explicitly "future, not built"). `dashboard.go:buildDash` renders
name/type/version/branch/git head/tag/tools from scan cache; `view.go:View`
shows it right of the logs at `width >= 100` (`dashVisible`), hidden on
narrow screens. No tab strip yet.

## Target — 6 windows, `1..6` + `tab`

| Key | Window | Content |
|-----|--------|---------|
| `1` | Status | Project, type, version, data-root, upstream, last build (`metrics.LastBuildLine`) |
| `2` | Tests | `08` dashboard (unit/instrumented, JUnit summary) |
| `3` | Diffs | `09` dashboard (git status/diff/log graph) |
| `4` | Build | Runner output, progress, cancel, artifact paths |
| `5` | Logs | Persisted logs browser (`logs show/tail`) |
| `6` | Help | Generated help + key cheatsheet |

Shell chrome: header (tool+project+version+phase pill+context line),
window tab-strip, body (active window rect), command bar, footer (active
window keys). Router from `03` owns switching; each window keeps its own
scroll/cursor state (switching never loses position).

## Snippet — tab strip

```go
tabs := []string{"Status","Tests","Diffs","Build","Logs","Help"}
for i, t := range tabs {
    st := theme.TabInactive
    if i == activeIdx { st = theme.TabActive }
    rendered = append(rendered, st.Render(fmt.Sprintf("%d %s", i+1, t)))
}
strip := lipgloss.JoinHorizontal(lipgloss.Top, rendered...)
```

Narrow terminals: strip collapses to `1..6` numbers + active title.

## Read

- Lipgloss `JoinHorizontal`, `Place`; Bubbles `tabs` patterns (if added,
  prefer own router — fewer deps).
- `docs/v0.0.2/*` doc 14 (bordered panels + state pill) — keep visuals.

## Tasks

- [ ] Router + tab strip + `1..6/tab/esc` (behind `RF_TABS=1` flag first).
- [ ] Per-window scroll state; `Tests`/`Diffs` skeletons (real content in `08`/`09`).
- [ ] Snapshot tests for 70/100/160-col layouts.

## Acceptance

```powershell
RF_TABS=1 releaseforge tui
# press 1..6, tab, esc — focus returns, scroll positions preserved
```
