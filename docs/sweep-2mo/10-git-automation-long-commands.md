# 10 — Git Automation: One-Key Wrappers for Long Commands

## Goal

Turn 80-char git invocations into reviewable one-key actions with preview
+ confirm + persisted log. Never hide the raw command.

## Command catalog (implement in this order)

| Action | Long command it wraps | Safety |
|--------|----------------------|--------|
| `gs` | `git status --short --branch` | read-only |
| `gd [file]` | `git diff --no-color -- <file>` | read-only |
| `gds` | `git diff --cached --no-color` | read-only |
| `ga <file>` | `git add -- <file>` | confirm if >10 files |
| `gc -m "msg"` | `git commit -m` (+ `--amend` variant) | confirm, pre-commit hooks respected |
| `gco <branch>` | `git checkout/switch` | confirm if dirty |
| `gcb <new>` | `git checkout -b <new>` | safe |
| `gpull --ff-only` | `git pull --ff-only` | confirm |
| `gpush` | `git push --set-upstream` | confirm + dry-run preview |
| `gtag vX` | `git tag -a vX -m` + `git push origin vX` | double confirm (see `23`) |
| `gundo` | `git reset --soft HEAD~1` (never `--hard` without `!`) | confirm + backup ref |

Every action shows: exact command → `Confirm? [y/N]` → runs via runner
(`12`) → appends to command history + log file.

## Snippet — preview-then-run

```go
type GitAction struct {
    Name string; Args []string; NeedsConfirm bool; AllowDirty bool
}
func (a GitAction) Preview() string { return "git " + strings.Join(a.Args, " ") }
func RunAction(ctx context.Context, dir string, a GitAction, confirm func(string) bool) (string, error) {
    cmd := a.Preview()
    if a.NeedsConfirm && !confirm(cmd) { return "", errCancelled }
    return runner.Capture(ctx, "git", a.Args, dir)
}
```

## Read

- https://git-scm.com/docs (status/diff/add/commit/push/pull/tag).
- Conventional Commits: https://www.conventionalcommits.org/ (for `gc` template).

## Tasks

- [ ] `internal/git/actions.go` catalog + tests (preview strings, dirty guards).
- [ ] TUI bindings in Diffs window + CLI `releaseforge git <action>` passthrough.
- [ ] Log every action to `history/commands.jsonl` + per-project log.

## Acceptance

```powershell
releaseforge git gs; releaseforge git gd --staged
# TUI Diffs: `ga`, `gc`, `gpush` all show preview first
```
