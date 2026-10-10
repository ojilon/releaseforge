# 35 — Eight-Week Sweep Schedule (slowly and surely)

> One workstream at a time. Tick boxes here as you land commits.
> Every week ends with `go vet + go test` green + daily-driver `tui` check.

## Weeks 1–2 — TUI + scan foundations

- [ ] `00` working loop adopted; `01` map verified against tree
- [ ] `02` probes + hash freshness; `25` config schema + validate
- [ ] `03` router + intent extraction; `04` theme/layout/truncate
- [ ] `05` CLI twins + `--json`; `06` keys (esc=back, ctrl+x=cancel) + generated help
- [ ] `34` fixtures started; daily `tui` flow unbroken

## Weeks 3–4 — Dashboards + git + errors + packaging

- [ ] `07` tab shell (flagged) ; `08` Tests window + JUnit
- [ ] `09` Diffs window + `10` git actions + `11` NUL-parse rewrite
- [ ] `12` watchdog + `.exit`; `27` scrub + tail -f; `28` error corpus
- [ ] `13` flavours/ABIs/doctor; `21` verify-gate; `22` manifest+checksums; `23` gates
- [ ] `26` kts fuzz; `29` maven runner; `31` generic tasks; `30` dogfood check

## Weeks 5–6 — Advanced (prebuilts + Zig + updater)

- [ ] `15` cmake runner; `14` `.so` handoff e2e (Gradle packages, never compiles)
- [ ] `16`+`17` Zig matrix + cache + manifest; `24` update-check/apply/rollback
- [ ] `32` benches recorded; `08`/`28` re-benched (Zig vs CMake numbers)

## Weeks 7–8 — Near-complete (native window + Odin decision)

- [ ] `19` win32-gdi spike → Wails/walk decision; file-drop install works
- [ ] `18` Odin DLL spike → keep/drop with numbers in the doc
- [ ] `20` wails release path; `33` security audit; hardware acceptance run
- [ ] Sweep retro: update `README.md` map + `docs/01-architecture.md` + `04-tui-design.md`

## Do-NOT-start-early list

`14`, `16`, `17`, `18`, `19` before weeks 5–7. If tempted, re-read `01` +
`32` and fix a dashboard instead.
