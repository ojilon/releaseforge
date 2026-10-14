# 05 — Cobra CLI Design (scriptable twin)

## Current state

`cmd/` has one file per verb; `root.go` owns `--data-root`. Good shape.
New twins since v0.0.2: `sign.go` (data-root keystore, `--apk/--keystore/
--alias/--apksigner`), `devices.go`. `build` takes `[variant] [version]`
(version written first, never committed).
Risk: CLI and TUI diverge (TUI `commands.go:exec` string-switch vs Cobra flags).

## Target

- Single source of truth: `internal/app/services.go` functions called by
  BOTH `cmd/*` and TUI router. Cobra files only parse flags → call service
  → print report. (Still open — TUI verbs live in `internal/app/*.go`.)
- Every TUI verb has a CLI twin: `status/scan/recent/version/build/test/
  notes/logs/release/install/sign/devices/doctor/run/logcat`.
  TUI-only (no CLI twin yet): `git/ls/dir/pwd/cat/bat/clear/theme`.
- `--json` on read verbs (`scan/status/version/logs`) for scripting.
- Consistent exit codes: 0 ok, 1 usage, 2 infra (missing SDK), 3 build/test failure.

## Snippet

```go
var scanCmd = &cobra.Command{
    Use: "scan [path]", Short: "Detect project, write scan cache",
    Args: cobra.MaximumNArgs(1),
    RunE: func(cmd *cobra.Command, args []string) error {
        svc := app.NewServices(dataRoot())
        rep, err := svc.Scan(firstOr(args, "."))
        if jsonOut { return printJSON(rep) }
        fmt.Println(rep.Human())
        return err
    },
}
```

## Read

- Cobra: https://github.com/spf13/cobra (RunE, arg validators, help templates).
- Your own `cmd/*.go` — audit for duplicated `project.Detect/Scan` blocks.

## Tasks

- [ ] Audit `cmd/` for inline domain logic; move into services.
- [ ] Add `--json` to scan/status/version; document exit codes in `--help`.
- [ ] TUI `exec()` delegates to same services (no second implementation).

## Acceptance

```powershell
releaseforge scan --json D:\x | ConvertFrom-Json | Select-Object name,type
releaseforge status --json | ConvertFrom-Json
```
