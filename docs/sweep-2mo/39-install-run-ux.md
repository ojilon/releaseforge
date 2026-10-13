# 39 — Install / Run UX (device as first-class target)

## Current state

`Install()` + `adbStream` install/launch; `cmd/install.go`, `run.go` exist.
Flow: pick signed APK from manifest (`22`) → pick device (huh) → install →
launch → optionally tail logcat (`38`).

## Target flow (Build window → device)

1. `install [--device ID] [--apk PATH]`: defaults APK from release manifest
   (no glob guessing); `--apk` overrides with existence check (`adb.go:56`).
2. `run [--device ID]`: `adb shell am start -n <pkg>/<activity>`; package +
   activity from scan (`ScanGradle` applicationId + manifest parse).
3. Post-install: offer `logcat --pid <pkg>` tail (`t` key) + `uninstall` (`u`, confirm).
4. Failure map: `INSTALL_FAILED_UPDATE_INCOMPATIBLE` → suggest uninstall;
   `SIGNATURE_MISMATCH` → warn different keystore (see `21`); `NO_DEVICES`
   → `doctor`.

## Snippet — launch intent

```go
// adb shell monkey -p <pkg> -c android.intent.category.LAUNCHER 1  // fallback
exec.Command("adb", withSerial(serial, []string{"shell","am","start","-n",pkg+"/"+activity})...)
```

## Read

- `am start`: https://developer.android.com/tools/adb#am ; `pm install` codes.
- Your `cmd/install.go`, `cmd/run.go`, `cmd/logcat.go` audit.

## Tasks

- [ ] Manifest-driven APK default + activity resolution at scan.
- [ ] install→run→logcat chained keys in Build window; failure map + tests.

## Acceptance

```powershell
releaseforge install --device <id>  # installs manifest APK
releaseforge run --device <id>      # launches, prints am output
```
