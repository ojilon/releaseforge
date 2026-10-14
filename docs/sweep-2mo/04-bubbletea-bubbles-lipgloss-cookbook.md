# 04 — Bubble Tea / Bubbles / Lipgloss Cookbook (make it legit)

## Current state

`internal/tui/theme.go` + `styles.go` + `view.go:View()` — `Theme` struct
(`DarkTheme`/`LightTheme`, `Current`, `SetTheme`, `PaintInput`), bordered
header (`PanelBorder`), `Pill(phase)`, sparkline, left-bordered input
(`BarStyle`), thin log border (`LogBoxStyle`), dashboard border
(`DashBoxStyle`), state footer. Correct libraries (bubbletea 1.3.6,
bubbles, lipgloss 1.1.0, huh) but:

- Theme object exists; `NO_COLOR` / light-background auto-detect still missing.
- Narrow path is `<80` branch + dashboard at `>=100`; no `layout.go`,
  sizing math in `model.go:layout()`.
- `viewport` + `textinput` wired manually; mouse + alt-screen flags fixed.
- Truncation uses byte slices in places (see v0.0.2 cleanup: UTF-8 bug).

## Target

- `internal/tui/theme.go`: DONE (`Theme{HeaderFg/HeaderBg/Text/Dim/Border/
  Accent/OK/Busy/Fail}` + `SetTheme("dark"|"light")`; every style derives
  from it via `refreshStyles()`). Still open: `DetectTheme(env, bg)`.
- `internal/tui/layout.go`: `Layout(width,height) → HeaderH/BodyH/FooterH`;
  windows receive exact rects; replace `model.go:layout()` magic numbers.
- Runewidth-safe `Truncate(s, n)`, `Table(rows, widths)` helpers with tests.

## Snippets

```go
// theme
var Default = Theme{ Accent: lipgloss.Color("12"), OK: lipgloss.Color("10"),
    Err: lipgloss.Color("9"), Muted: lipgloss.Color("8") }
func (t Theme) Header() lipgloss.Style {
    return lipgloss.NewStyle().Bold(true).Foreground(t.Accent)
}
// respect NO_COLOR
if os.Getenv("NO_COLOR") != "" { lipgloss.SetColorProfile(termenv.Ascii) }

// safe truncate (never split graphemes)
import "github.com/mattn/go-runewidth"
func Truncate(s string, n int) string { return runewidth.Truncate(s, n, "…") }

// window chrome
top := theme.Panel.Render(header + " " + theme.Pill(phase) + "\n" + ctx)
```

Bubbles patterns to adopt:

- `textinput`: always-focused bar, `SetValue("")` on enter, history via
  `history.LoadCommands/AppendCommand` (already done — keep).
- `viewport`: `GotoBottom()` only when `AtBottom()` (already in `render()`);
  extract to `FollowViewport` wrapper.
- `huh` (already dep): init form, device picker, password prompt — reuse,
  don't hand-roll forms.
- `key.Binding` + `help.Model` for footer instead of string literals.

## Read

- Lipgloss: https://github.com/charmbracelet/lipgloss (layout, `Place`, `JoinVertical`).
- Bubbles examples: https://github.com/charmbracelet/bubbles/tree/master/examples
- huh: https://github.com/charmbracelet/huh
- `go-runewidth`: https://github.com/mattn/go-runewidth (why byte-truncation breaks emoji/CJK).

## Tasks

- [x] Add `theme.go` (done — presets + `SetTheme` + `PaintInput` + tests via `styles_test.go`).
- [ ] Add `layout.go`, `truncate.go` + unit tests (incl. CJK/emoji).
- [ ] Replace string footers with `key.Binding` help view.
- [ ] Narrow (`<80`), normal, wide (`>140`) snapshots via `termenv` fake width tests.

## Acceptance

```powershell
go test ./internal/tui/ -v
# manual: resize terminal to 70/100/160 cols; header never wraps, footer stays one line
```
