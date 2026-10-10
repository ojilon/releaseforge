# 00 — How To Use This Folder

One workstream per day (or per 2–3 days). Do not binge all 35 files.

## Daily loop

1. Read the target file fully (e.g. `07-window-navigation-tabs-dashboard.md`).
2. Open every `Read:` link inside it — papers, git docs, Charm docs.
3. Reproduce the **Current state** snippet locally (`go vet`, run TUI).
4. Implement only that file's **Tasks**; check its **Acceptance**.
5. Run `go vet ./...; go test ./...` — must be green.
6. Commit with message `sweep(<area>): <what>` and tick the box in
   `35-eight-week-plan.md`.

## Conventions used in every file

- `Current state` — file:line refs, honest about what is stub.
- `Problem` — why AI-generated code hurts here.
- `Target` — concrete end state, no hand-waving.
- `Snippet` — minimal Go you can paste and adapt.
- `Read:` — docs/papers to read BEFORE coding, then come back.
- `Tasks` — ordered checklist.
- `Acceptance` — shell commands that prove it.

## Commit hygiene

- One workstream per commit. Never mix TUI refactor with Zig spikes.
- Keep `docs/sweep-2mo/` updated when behaviour changes (same commit).
- If a snippet in a guide goes stale, fix the guide first.
