# v0.0.2 doc 10: adb — doctor, devices, install, run, logcat

## Goal

Own the device side: one `doctor` health check, a 2+ device picker, working
install for both variants, one-command `run`, and a `logcat` tail.

## Current state

`android.Devices/Install` (`internal/android/sign.go:144-180`) parse
`adb devices` and run `adb [-s serial] install -r`. `install` command
(`cmd/install.go`) auto-picks only with exactly one device, else passes
empty serial (adb then fails confusingly with 2+ devices). No `doctor`, no
`run`, no logcat, no launch support. `status` prints devices when present.

## Design

`doctor` command (new, `cmd/doctor.go`, read-only, fast):

    doctor                  checks, one line each with ok/missing/fix hint:
                            adb on PATH (+ version), gh on PATH, java on PATH,
                            gradlew[.bat] present, ANDROID_HOME set (warn only),
                            devices listed (0/1/N)

Device selection helper in `internal/android` (reused by `install`,
instrumented tests from doc 09, `run`, `logcat`):

    func PickDevice(preferred string) (string, error)
    // --device wins; else 1 device → it; 0 → "connect a device" error;
    // 2+ → huh select picker (TTY) or "use --device" error (non-TTY).

`install` keeps its artifact lookup (fixed in 01, scan-aware in 09) and uses
`PickDevice`. New commands:

    run [debug]             build → install → launch; launch = adb shell
                            monkey -p <package> -c android.intent.category.LAUNCHER 1
                            (package id from doc 06 manifest/namespace data; clear
                            error when unknown — no guessing)
    logcat [-n 200] [--pid name]   adb logcat -d tail; --pid filters by
                            package via pidof when available, else grep fallback

All adb invocations go through one helper (`adbRun(serial, args…)`) so
`-s` wiring and logging stay consistent. Timeouts from doc 03 apply
(`logcat -d` is bounded; document that streaming logcat is out of scope).

## Files to touch

- New `cmd/doctor.go`, `cmd/run.go` (or fold `run` into `install.go`? separate
  file, thin), extend `cmd/install.go` (picker), extend `cmd/` with `logcat`.
- `internal/android/` (new `adb.go`: `PickDevice`, `adbRun`, `LaunchApp`,
  `LogcatTail`; move `Devices/Install` there from `sign.go`).

## Steps

1. Move adb funcs to `adb.go` + add `PickDevice` + `adbRun` + tests with a
   fake `adb` on PATH (shell script fixture echoing canned output).
2. `doctor` command + tests (fixture PATH with/without tools).
3. `install` via picker; non-TTY 2+ devices → `--device` error.
4. `run` = Runner build + install + `LaunchApp`.
5. `logcat` with `-n`/`--pid`.

## Tests to add

- Picker matrix (0/1/N devices, TTY/non-TTY — huh part manual, logic pure).
- Fake-adb fixture tests for install args, launch args, logcat filter.
- `doctor` output lines for present/missing tools.

## Acceptance checks

- `go vet ./...`, `go test ./...` pass.
- Without adb: `doctor` reports missing with fix hint; `install` fails fast
  naming `doctor`.
- With a real device (author's phone): `run debug` builds, installs,
  launches; `logcat -n 20` prints lines.

## Out of scope

Streaming logcat, screenshots, wireless pairing, multi-device fan-out,
instrumented-test device orchestration beyond the precheck.
