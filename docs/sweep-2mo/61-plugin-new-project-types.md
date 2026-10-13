# 61 — Plugin Model for New Project Types (no fork of Detect)

`Detect()` if-chain (`detect.go:47`) doesn't scale. Target: `ProjectPlugin`
interface (`ID, Markers[], Detect(root)->(confidence, Info), Runner()`)
+ registry; built-ins (android/wails/cmake/python/java/go/generic) are just
plugins; confidence-ordered (strong markers beat weak), ties → deterministic
order + `scan --explain` shows scores; third-party via config-declared tasks
first (no dynamic loading until v2 — keep single binary).

Snippet: `type Plugin interface { ID() string; Score(root string) int; Info(root string) Info }`.
Tasks: extract built-ins as plugins + score tests. Acceptance: fixture matrix resolves same types as today.
