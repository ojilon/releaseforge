# 60 — Data-Root Backup & Restore (D: is precious)

Target: `rf backup [--to PATH]` (zips `global.json/history/projects/*/config+cache+metrics`,
EXCLUDES `logs/native-cache/releases` by default with `--full` opt-in),
`rf restore --from ZIP` (refuses non-empty root without `--force` + confirm),
`doctor` shows root size by subtree (`logs` vs `releases` vs `native-cache`
so you know what to prune); scheduled-task/Task-Scheduler one-liner documented.

Acceptance: backup→wipe→restore round-trips `recent` + configs.
