# 55 — Rust / Cargo Support (toolchain, targets, audit)

`Cargo.toml` recorded as hint only. Target: `cargo build/test --locked`,
toolchain pin (`rust-toolchain.toml`), target matrix note (aarch64-linux-android
via `cargo-ndk` — contrast with Zig path `16`), `cargo audit` in `doctor`;
version from `Cargo.toml [package] version` R/W; release packs binaries + checksums.

Read: https://doc.rust-lang.org/cargo/ ; cargo-ndk docs.
Tasks: runner + version R/W + audit hook. Acceptance: fixture `cargo test` runs.
