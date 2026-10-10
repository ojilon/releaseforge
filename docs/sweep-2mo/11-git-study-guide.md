# 11 — Git Study Guide (read, then re-implement + integrate)

> Do not skim. Read → take notes → come back and rewrite `internal/git/`.

## Week-1 reads (porcelain contracts)

1. `git status` porcelain: https://git-scm.com/docs/git-status — sections
   PORCELAIN FORMAT v1, `-z` планировка. Implement `porcelain.go` from this.
2. `git log` pretty formats: https://git-scm.com/docs/git-log — `--pretty=format:%H%x00%s%x00%ad`,
   `--graph`, `--since`. Your `Log/LogSince` must use NUL separators (subjects contain spaces).
3. `git diff`: https://git-scm.com/docs/git-diff — `--no-color`, `--stat`,
   `--numstat` for dashboards; `--check` for whitespace errors.
4. Pro Git Ch. 2–3 (free): https://git-scm.com/book/en/v2 — branching,
   staging, remotes. Ch. 7 (reset/checkout), Ch. 10 (internals: objects, refs).

## Week-2 reads (history → releases)

5. `git tag`: https://git-scm.com/docs/git-tag — annotated vs lightweight;
   why releases need annotated (`23` tag safety).
6. `git notes` / trailers: https://git-scm.com/docs/git-interpret-trailers —
   for Changes grouping in `DraftNotes`.
7. GitHub `gh release create`: https://cli.github.com/manual/gh_release_create
8. Conventional Commits + SemVer: https://www.conventionalcommits.org/,
   https://semver.org/ — grouping + `notes` templates.

## Papers / deep cuts (optional but worth it)

- "Software Configuration Management" — Bosch et al. (release pipeline theory).
- Rachlis dirty-tree UX studies → why preview-then-confirm beats silent `add -A`.
- libgit2 architecture docs (if you ever consider go-git): https://libgit2.org/

## Re-implementation checklist (after reading)

- [ ] NUL-separated log parser; rename/copy detection flags documented.
- [ ] `Changelog{Groups[feat/fix/docs/…], Breaking[]}` from trailers + prefixes.
- [ ] `LatestTag` + `LogSince(tag)` bounded (default 200, flag to raise).
- [ ] Upstream tracking (`@{u}`, ahead/behind) surfaced in Status window.
- [ ] Replace every `strings.Split(line," ")` log parse with NUL parse + test
      containing `"`quotes`"`, unicode, `|` pipes in subjects.

## Acceptance

```powershell
go test ./internal/git/ -v
# craft evil commit subjects with spaces/quotes/emoji → dashboard still parses
```
