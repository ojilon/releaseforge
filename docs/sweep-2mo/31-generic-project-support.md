# 31 — Generic Project Support (any folder is valid)

## Promise (from `00-vision.md`)

Any local directory opens: git + file stats + scripts inventory + recent
tracking, even with zero known markers. User supplies tasks via config.

## Design

- `generic` runner: `Build.Tasks` map in `config.json`
  (`{"build":"…","test":"…"}`); `build [task]` / `test [task]` run them.
- Seed file `.releaseforge.json` can declare tasks + version file without
  touching global config (see `25` merge rules).
- Status window shows "generic — N custom tasks" + hint to add tasks.

```json
{ "build": { "tasks": { "build": "npm run build", "test": "npm test" } } }
```

## Read

- Your `docs/02-project-types.md` generic section; `MergeConfig` seed logic.

## Tasks

- [ ] `generic` runner + `tasks list/add` commands.
- [ ] Node/Python/Rust task seeds auto-suggested at scan (lockfile → tasks).

## Acceptance

```powershell
releaseforge scan D:\some\random\folder  # type generic, tasks suggested
```
