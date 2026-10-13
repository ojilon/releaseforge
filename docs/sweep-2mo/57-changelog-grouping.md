# 57 — Changelog Grouping (conventional commits → notes)

`git.DraftNotes` + grouped/bounded changelogs exist; grouping rules live in code.
Target: documented groups (`feat/fix/perf/docs/refactor/test/build/ci/BREAKING`),
trailer support (`Co-authored-by`, `Fixes #`), scope rendering (`feat(android):`),
length caps (subject 72, body wrap), `--since tag` default 200; `notes --json`
for scripting; golden fixtures incl. evil subjects (quotes/emoji/pipes from `11`).

Read: https://www.conventionalcommits.org/ ; https://semver.org/ ;
`git interpret-trailers` docs.
Tasks: grouping spec + golden tests + `--json`. Acceptance: `go test ./internal/git/ -run TestNotes -v`.
