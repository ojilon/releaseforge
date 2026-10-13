# 48 — JDK & Gradle Tuning (daemon, memory, cache)

## Goal

Document + automate the 5 settings that decide whether Gradle is fast or
miserable; `doctor` + `update-tools` (`24`) enforce them.

## The five

1. JDK: 17 for AGP 8.x (record exact `java -version` in scan; warn on 21+quirk).
2. `gradle.properties`: `org.gradle.jvmargs=-Xmx4g -Dfile.encoding=UTF-8`,
   `org.gradle.caching=true`, `org.gradle.parallel=true`, `kotlin.incremental=true`.
3. Daemon: leave ON locally; `--no-daemon --offline` only in CI/dry-run.
4. `GRADLE_USER_HOME` under data-root (`D:`) so caches don't fill `C:`.
5. `--console=plain -q` for log parsing (`28`); `--info/--stacktrace` on demand flags.

## Snippet — env + args

```powershell
$env:GRADLE_USER_HOME = "D:\ReleaseForgeData\gradle-home"
gradlew assembleRelease --console=plain -q
```

Scan records `gradle.{version,agp,jdk,caching}`; `doctor` diffs vs recommended.

## Read

- Gradle performance: https://docs.gradle.org/current/userguide/performance.html
- AGP/JDK compat: https://developer.android.com/build/jdk

## Tasks

- [ ] Scan records JDK/AGP/caching flags; `doctor` warns on drift.
- [ ] `GRADLE_USER_HOME` default under data-root (documented, overridable).

## Acceptance

```powershell
releaseforge doctor  # flags Xmx/caching/daemon drift with fix lines
```
