# 50 — R8 / ProGuard Shrinking (rules without fear)

## Goal

Release builds shrink safely: `minifyEnabled` + `shrinkResources` +
`proguard-rules.pro` per flavor; failures parse into `28` findings with
`-keep` suggestions.

## Checklist

1. `buildTypes.release { minifyEnabled true; shrinkResources true;
   proguardFiles getDefaultProguardFile('proguard-android-optimize.txt'),
   'proguard-rules.pro' }` — record in scan.
2. Mapping file (`mapping.txt`) archived per release (`22` layout) for
   deobfuscation; never dropped.
3. Missing-class errors (`can't find referenced class`) → suggest `-keep`
   or `-dontwarn` with exact rule snippet (from corpus `28`).
4. JNI/Native: `-keepclasseswithmembernames class * { native <methods>; }`
   when prebuilt `.so`s (`14`) present.

## Read

- Shrink: https://developer.android.com/build/shrink-code ; R8 troubleshooting.

## Tasks

- [ ] Scan records minify/shrink flags + rules files; mapping archived in manifest.
- [ ] R8 error corpus + `-keep` fix hints in `28`.

## Acceptance

```powershell
# release build with minify → mapping.txt in releases/<ver>/, deobfuscate documented
```
