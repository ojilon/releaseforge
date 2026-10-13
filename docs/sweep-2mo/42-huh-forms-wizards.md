# 42 — huh Forms & Wizards (init, device picker, passwords done once)

## Current state

huh is a dep; init form + device picker + password prompt exist but each
hand-rolls validation + theme.

## Target — one forms package

- `internal/tui/forms.go`: `InitForm(dataRoot)`, `DevicePick(devs)`,
  `PasswordPrompt(label)`, `Confirm(cmdPreview)` — shared theme (from `04`),
  shared validators, all headless-testable (feed answers, assert values).
- Rules: passwords `echoMode.Password`, never echoed, never logged;
  device picker shows `serial + model + state` (from `36`);
  init form writes `global.json` + seeds layout (idempotent, rerunnable).
- CLI twins: `--no-input` fails fast in scripts instead of hanging on a prompt.

## Snippet

```go
var device, password string
form := huh.NewForm(
    huh.NewGroup(huh.NewSelect(&device).Title("Device").Options(opts...)),
    huh.NewGroup(huh.NewInput(&password).Title("Keystore password").EchoMode(huh.EchoModePassword)),
)
```

## Read

- huh: https://github.com/charmbracelet/huh (forms, validation, themes).

## Tasks

- [ ] Extract forms package + headless tests; `--no-input` on all prompts.
- [ ] Confirm-preview form reused by git actions (`10`) + publish gates (`23`).

## Acceptance

```powershell
go test ./internal/tui/ -run TestForms -v
releaseforge init  # rerun keeps values, validates, never duplicates
```
