# 62 — SBOM, Provenance & Licenses (know what you ship)

Target: per release (`22` manifest gains `sbom.json`): native deps
(`build.zig.zon` pins from `16`, gradle deps via `dependencies --configuration
releaseRuntimeClasspath`), license list (`LICENSE` scan + `cargo audit`-style
deny rules later), build provenance (tool versions from `24`, git_rev, dirty?);
`doctor --licenses` fails on GPL-incompatible surprises (config allow-list).

Read: SPDX-lite, SLSA provenance (slsa.dev), reproducible-builds.org.
Tasks: SBOM writer + provenance block + license gate. Acceptance: every release zip contains sbom.json.
