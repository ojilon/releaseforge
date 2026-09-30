# Gradle command reference

```bash
# Unit tests (Robolectric)
./gradlew :app:testDebugUnitTest

# Instrumented (device/emulator)
./gradlew :app:connectedDebugAndroidTest

# APKs
./gradlew assembleDebug
./gradlew assembleRelease

# Clean
./gradlew clean

# With custom Gradle user home
./gradlew -g /path/to/gradle-home assembleDebug
```

Windows: use `gradlew.bat`.
