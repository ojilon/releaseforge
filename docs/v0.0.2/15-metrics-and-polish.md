# v0.0.2 doc 15: Metrics, status polish, and doc refresh

## Goal

Close v0.0.2 with visible history (build durations), an honest `status`,
tests for anything new, and docs that match the code — so the next version
starts from truth.

## Current state

- `internal/metrics` is a one-line stub, imported nowhere; `reports/` and
  command history exist on disk but are never read back.
- `status` shows no last-build info despite its Short text claiming it.
- Docs `04` (tabs, folder picker), `08` (0.0.1 plan, 0.0.2 sketch), `09`
  (schema drift) contradict the code (audit §14). `RELEASE.md` +
  `docs/10-bootstrap-v001.md` are current.

## Design

`internal/metrics` (stdlib only): `Record{At, Project, Kind, Variant,
Success, DurationMs}` + `Append(dataRoot, rec)` / `Last(dataRoot, project,
kind)` over `projects/<name>/metrics.jsonl` (append, cap 200 lines on
write, best-effort — never fail a build). `build`/`test` (CLI + TUI via the
Runner from doc 02) record start/end around `Run`/`Start`-drain. `status`
prints `last build: <variant> <ok|FAILED> <duration> <ago>` (or `no builds
yet`). No dashboard UI in v0.0.2 — data capture only, surfaced in `status`
and TUI `status` text.

Doc refresh (same commit, docs half): rewrite the stale parts of `docs/04`
(single-pane reality + panels from doc 14, drop tabs-or-mark-future),
`docs/08` (0.0.2 work list → point at `docs/v0.0.2/`, mark 0.0.1 items done
truthfully), `docs/09` (schema = `Snapshot` + `meta.json` + merge rules from
doc 07). Keep `docs/10` and `RELEASE.md` untouched unless behavior they
describe changed — then fix them too (rule: docs follow code).

## Files to touch

- New `internal/metrics/metrics.go` (+ test).
- `cmd/build.go`, `cmd/test.go`, TUI `doBuild/doTest`, `cmd/status.go`
  (+ TUI `statusText`).
- `docs/04-tui-design.md`, `docs/08-implementation-plan.md`,
  `docs/09-scan-foundation.md` (refresh only).

## Steps

1. Metrics record/append/last + cap + tests.
2. Wire record calls into CLI build/test (both types, success + failure).
3. Wire into TUI long commands (record on `doneMsg`).
4. `status` last-build line (CLI + TUI); `no builds yet` default.
5. Refresh docs 04/08/09; re-check every claim against code.
6. Full test pass + `RELEASE.md`/`docs/10` consistency sweep.

## Tests to add

- Append/last/cap; corrupt-line tolerance (skip bad JSON lines).
- `status` shows last build after a fixture build; `no builds yet` fresh.
- No test for prose — instead a checklist in PROGRESS.md commit message
  naming which doc sections changed.

## Acceptance checks

- `go vet ./...`, `go test ./...` pass.
- On this repo: run `test`, then `status` shows `last build: …​ ok …​ ago`.
- `rg -n "tabs|Overview.*Git.*Tools" docs/04-tui-design.md` shows no
  present-tense tab claims (future-marked or removed).
- Final `git log --oneline` shows `v0.0.2: 01 …​` through `v0.0.2: 15 …​`
  in order.

## Out of scope

Metrics UI/panels, trend analysis, CI, version bump, release. Do not bump
the tool version or cut a release until asked.
