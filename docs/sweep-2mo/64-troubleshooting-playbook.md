# 64 — Troubleshooting Playbook (errors → fixes)

Living catalog fed by `28` findings. Format per entry: symptom (log excerpt),
cause, fix command, prevention (`doctor` check). Seed entries: SDK-licenses,
NDK-missing, JDK-mismatch, daemon-OOM, no-device, unauthorized-device,
keystore-expiry, tag-exists, gh-auth-missing, gradle-offline-cache-miss,
zig-version-drift, cmake-preset-missing. Each entry links to its sweep doc.
`doctor` prints matching entry IDs. Update this file with every new corpus case.

Acceptance: top-10 failures each resolve via one copied command.
