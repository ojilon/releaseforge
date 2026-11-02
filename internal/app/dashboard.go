package app

import "strings"

// The old fixed-width right-side "dashboard" is gone: its data now lives in
// the Status window (window_status.go) with a real side card, a recent-runs
// table and a generated command reference in the Help window. This file only
// keeps the small formatting helpers shared by the windows.

// shortSHA keeps git object ids readable in narrow columns.
func shortSHA(s string) string {
	s = strings.TrimSpace(s)
	if len(s) > 12 {
		return s[:12]
	}
	return s
}
