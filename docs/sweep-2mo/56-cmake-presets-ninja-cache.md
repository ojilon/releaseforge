# 56 — CMake Presets, Ninja, ccache/sccache (fast native loop)

Companion to `15`. Target: `CMakePresets.json` configure/build/test presets
honoured (`--preset`), Ninja preferred (`-G Ninja`, `ninja --version` in
`doctor`), `CMAKE_C_COMPILER_LAUNCHER=ccache/sccache` supported with cache
dir under data-root `native-cache/`; scan records `generator`, `presets[]`,
`launcher`; metrics capture configure vs build times.

Read: https://cmake.org/cmake/help/latest/manual/cmake-presets.7.html ;
https://ninja-build.org/manual.html ; ccache/sccache docs.
Tasks: preset runner + launcher env + fixture. Acceptance: preset build uses cache on rebuild.
