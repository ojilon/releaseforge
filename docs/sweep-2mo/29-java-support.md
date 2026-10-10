# 29 — Java / JVM-CLI Support (Gradle + Maven)

ReleaseForge handles Gradle+Java+C++ today via Android; plain Java CLI
(`pom.xml`, non-Android `build.gradle`, `src/main/java`) is detected
(`detect.go:89`) but thinly served.

## Targets

- Gradle Java: `build` / `test` tasks (`build`, `test`), shadow-jar artifact
  discovery (`build/libs/*.jar`), version from `build.gradle` or config.
- Maven: `mvn -q package`, `mvn -q test`; wrapper `mvnw` preferred.
- `release` for `java-cli`: version → test → package jar → zip + checksums
  → tag → `gh release` (same gates as `23`, minus signing).

## Read

- Gradle Java plugin: https://docs.gradle.org/current/userguide/java_plugin.html
- Maven lifecycle: https://maven.apache.org/guides/introduction/introduction-to-the-lifecycle.html

## Tasks

- [ ] Maven runner + jar discovery + version R/W.
- [ ] `java-cli` release path + fixtures.

## Acceptance

```powershell
releaseforge build  # in a maven fixture → mvn package
```
