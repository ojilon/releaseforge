# 59 — Config Schema & Migration (versioned, never breaks)

`config.json` has no schema version today. Target: `"schema": 2` field,
`Migrate(v1→v2…)` pure functions + tests, `config validate` (see `25`),
installer backup-then-migrate (`51`), unknown fields preserved (forward-compat),
documented tool-owned vs user-owned table (from `scan.go:315`).

Read: JSON schema versioning practices; your `configs/example-*.json` as migration fixtures.
Tasks: schema field + migrators + validate. Acceptance: v1 config migrates losslessly, user keys survive.
