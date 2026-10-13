# 54 — Node Frontend Support (npm/pnpm, lockfiles, Wails frontend)

Scan records `package.json` + lockfile (`pnpm-lock.yaml`/yarn/package-lock)
as hints today. Target: package-manager detect (lockfile wins),
`{build: "<pm> run build", test: "<pm> test"}` seeds, `node_modules`
excluded from hashing (`02`), frontend version from `package.json`;
Wails builds (`20`) run `npm install` check first with `--frozen-lockfile` in CI.

Read: npm/pnpm/yarn docs. Tasks: pm detect + frozen-install flag + fixture.
Acceptance: scan prints `node(pnpm)` + seeded tasks.
