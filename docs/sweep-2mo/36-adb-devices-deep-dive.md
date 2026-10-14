# 36 — adb + Devices Deep Dive (doctor/run/install truth)

## Current state

`internal/android/adb.go:1` — `withSerial`, `adbStream` (inherited stdio),
`adbOut`, `Devices()` (parses `adb devices`, keeps `device` state only),
`Install(apk, serial)` (`adb install -r`). `cmd/doctor.go:24` checks
adb/gh/java/wrapper/ANDROID_HOME + lists devices. Device picker via huh exists.

Done since: `InstallCapture()` transcript variant (`signing.go`) drives TUI
`install [debug|release] [--device S]` (`app/install.go`) + `devices` verb;
`--device` required-when-ambiguous guard + tests (`signing_test.go`,
`app` suite). CLI: `install`, `devices`.

## Gaps to close

1. States beyond `device`: `unauthorized`, `offline`, `no permissions`
   must surface with FIX hints (re-plug, accept RSA prompt, udev/USB driver).
2. `adbStream` inherits stdio — runner (`12`) can't capture; add
   `InstallCapture()` returning transcript for Logs window.
3. Multi-device: every destructive adb call requires explicit serial when
   >1 device; `withSerial` already supports it — enforce at call sites.
4. `adb wait-for-device` + `get-state` before install; fail fast with
   actionable message instead of hanging.

## Snippet — guarded install

```go
devs, _ := Devices()
if len(devs) > 1 && serial == "" {
    return fmt.Errorf("multiple devices (%s) — pick with --device", strings.Join(devs, ","))
}
if st, _ := adbOut(serial, "get-state"); strings.TrimSpace(st) != "device" {
    return fmt.Errorf("device %s not ready (state %q) — run `adb devices`", serial, st)
}
return Install(apkPath, serial) // then VerifyInstalled via `adb shell pm list packages`
```

## Read

- adb docs: https://developer.android.com/tools/adb ; `adb devices -l`, `get-state`, `wait-for-device`.
- Your `adb.go:34 Devices()` + `doctor.go` checks table.

## Tasks

- [ ] Parse all device states + fix hints; `doctor` prints them.
- [x] `InstallCapture` with transcript → Logs window + metrics.
- [x] `--device` required-when-ambiguous guard + tests.

## Acceptance

```powershell
go test ./internal/android/ -v
releaseforge doctor  # unplug/plug shows unauthorized→device transitions
```
