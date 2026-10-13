# 37 — Emulator Automation (create, boot, wait, test)

## Goal

`rf emu` manages AVDs for instrumented tests without Android Studio:
list → create → cold-boot → `wait-for-device` → unlock → run tests → snapshot.

## Commands (wrap, don't hide)

| Action | Long command |
|--------|-------------|
| `emu list` | `emulator -list-avds` |
| `emu create NAME --abi arm64-v8a --api 34` | `avdmanager create avd -n NAME -k "system-images;android-34;google_apis;arm64-v8a"` |
| `emu boot NAME` | `emulator -avd NAME -no-snapshot -wipe-data &` then `adb wait-for-device` + `sys.boot_completed` poll |
| `emu unlock` | `adb shell input keyevent 82` |
| `emu kill` | `adb emu kill` |

Boot-complete poll (up to 5 min):

```powershell
adb wait-for-device
do { $b = adb shell getprop sys.boot_completed } until ($b -match '1')
```

Snapshot cold-boot (`-no-snapshot`) for deterministic instrumented runs;
record AVD + API + ABI in test metrics (`08`).

## Read

- Emulator CLI: https://developer.android.com/studio/run/emulator-commandline
- avdmanager: https://developer.android.com/tools/avdmanager

## Tasks

- [ ] `internal/android/emu.go` wrapper + boot poll + timeout.
- [ ] `test --emulator NAME` spins AVD, runs, tears down (flag, confirm).

## Acceptance

```powershell
rf emu list; rf emu boot <avd>  # reaches boot_completed, unlocks
```
