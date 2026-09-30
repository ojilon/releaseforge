# apksigner and keystore

## Locate apksigner

1. Config override path
2. PATH
3. `%ANDROID_HOME%\\build-tools\\<highest>\\apksigner.bat` (Windows)
4. `$ANDROID_HOME/build-tools/<highest>/apksigner` (Unix)

## Sign

```bash
apksigner sign --ks path/to.jks --ks-key-alias ALIAS --ks-pass pass:SECRET app-release-unsigned.apk
apksigner verify --verbose app-release-unsigned.apk
```

Password must be prompted interactively; do not write to config files.
