# 58 — Git Hooks & Commit Templates (quality at the source)

Target: `rf git install-hooks` writes `.git/hooks/commit-msg` (conventional-commit
lint) + `pre-push` (vet+fast tests) as OPT-IN, skippable with `--no-verify`
documented; `commit.template` (`.gitmessage`) with type/scope/body hints;
hooks never block `--dry-run` flows; `doctor` reports hook presence.

Snippet hook (commit-msg): grep `^(feat|fix|docs|refactor|test|build|ci|chore)(\(.+\))?(!)?: .{1,72}$` or warn (not fail by default — warn-first month).

Read: https://git-scm.com/docs/githooks ; conventional commits.
Tasks: hook templates + installer + opt-out docs. Acceptance: bad message warns with example fix.
