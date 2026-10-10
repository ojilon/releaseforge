# 02 — Scan and Detect Rethink

## Current state

- `internal/project/detect.go:47` — first-strong-match wins; records only
  primary `Type`, secondary markers partially in scan.
- `internal/project/scan.go:71` — `os.Stat` probes, `app/build.gradle`
  string-contains for `com.android.application`, bounded CMake search
  (`FindCmakeFiles`), `ScanGradle` for modules/flavours/ABIs.
- Freshness (`scan.go:285`) = root-path equality only. No mtime/hash.

## Problems

1. Shallow evidence: secondary tools appended ad hoc; no lockfile matrix.
2. No invalidation: editing `build.gradle.kts` doesn't mark cache stale.
3. CMake search bounds undocumented; monorepos mis-detected.
4. Version-source detection (`DetectVersionSource`) fragile on `.kts`.

## Target

- `ScanResult{Snapshot, Evidence[], StaleReasons[]}` with per-tool evidence.
- `Fresh()` → root + mtime + content-hash of key manifests.
- `rescan --explain` prints why cache was stale.
- Fuzz-tested detectors on synthetic trees.

## Snippet — evidence-first probe

```go
type Probe struct {
    ID string; Files []string; // relative candidates
    Match func(root string, found []string) (ToolHit, bool)
}
var probes = []Probe{
    {ID: "gradle", Files: []string{"settings.gradle","settings.gradle.kts","gradlew","gradlew.bat"}},
    {ID: "cmake", Files: []string{"CMakeLists.txt","backend/CMakeLists.txt","native/CMakeLists.txt"}},
    {ID: "wails", Files: []string{"wails.json"}},
    {ID: "go", Files: []string{"go.mod"}},
}
for _, p := range probes {
    var found []string
    for _, f := range p.Files {
        if fileExists(info.Root, f) { found = append(found, f) }
    }
    if hit, ok := p.Match(info.Root, found); ok { snap.Tools = append(snap.Tools, hit) }
}
```

## Read (then come back and implement)

- `docs/09-scan-foundation.md`, `docs/02-project-types.md` (repo).
- Git docs on `git status --porcelain` for dirty-tree signals you can reuse
  for staleness: https://git-scm.com/docs/git-status
- `filepath.WalkDir` + `fnmatch` patterns; Go `os.ReadDir` semantics.

## Tasks

- [ ] Extract probes table from `Scan()`; add node/maven/python/rust/cargo hits.
- [ ] Hash key manifests (sha1 of first 64KB) into `meta.json`.
- [ ] `Fresh()` compares root + hash; `scan --explain` lists stale files.
- [ ] Golden tests: android/wails/cmake/go/generic fixtures under `testdata/`.

## Acceptance

```powershell
go test ./internal/project/ -run 'TestScan|TestDetect|TestFresh' -v
releaseforge scan D:\path\to\repo; releaseforge scan --explain D:\path\to\repo
```
