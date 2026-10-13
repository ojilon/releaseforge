# 49 — AAB vs APK, Splits & ABI Packaging

## Current state

APK-only (`app-debug.apk`, `app-release-*.apk`). No bundle/split story.

## Decide + document

- Debug/local: APK per ABI or universal (fast install via `39`).
- Store: **AAB** (`bundleRelease`) — Play generates splits; record
  `bundletool` version when validating locally.
- ABI splits (if APKs shipped): `splits { abi { enable true … } }` vs
  fat APK tradeoff table (size vs install simplicity); `universalApk true`
  for sideload fallback.
- Scan records `bundle:{aabTask, splits, abiList}`; manifest (`22`) lists
  per-ABI outputs + which device each fits (`36` serial → `ro.product.cpu.abi`).

## Snippet — abi of device

```powershell
adb shell getprop ro.product.cpu.abi  # arm64-v8a → install matching split
```

## Read

- Bundles: https://developer.android.com/guide/app-bundle ; splits:
  https://developer.android.com/studio/build/configure-apk-splits ; bundletool:
  https://developer.android.com/tools/bundletool

## Tasks

- [ ] `bundleRelease` runner path + split-aware manifest + device→ABI match in `install`.

## Acceptance

```powershell
releaseforge build bundle  # produces .aab; manifest lists ABIs
```
