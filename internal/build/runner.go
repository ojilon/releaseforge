// Package build runs Gradle / Wails / CMake / generic commands with live log capture.
package build

// Runner streams stdout/stderr, persists logs under data-root, and parses
// common error patterns for friendly TUI rendering.
