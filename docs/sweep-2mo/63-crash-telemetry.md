# 63 — Crash Reports & Telemetry (opt-in, local-first)

No telemetry today (good). Target: stays OFF by default; `init --telemetry`
opts in; local crash files (`crashes/<ts>.log` with stack + version + phase
from `40`, NO paths beyond basename, NO secrets — scrubbed per `27`);
`rf crash-send` explicitly uploads (prints payload first); metrics (`27`)
remain local-only otherwise.

Tasks: crash writer + scrub test + explicit send. Acceptance: panic writes redacted local file, nothing leaves disk unasked.
