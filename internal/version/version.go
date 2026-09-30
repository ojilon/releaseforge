// Package version owns the ReleaseForge tool version.
//
// Source of truth is the VERSION file at repo root. ToolVersion defaults to
// the current release and can be stamped at build time:
//
//	go build -ldflags "-X github.com/ojilon/releaseforge/internal/version.ToolVersion=0.0.2" .
package version

// ToolVersion is the current tool release. Bump with VERSION + git tag.
var ToolVersion = "0.0.1"
