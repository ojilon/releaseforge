package tui

import (
	"sort"
	"strings"
	"unicode"
)

// Match is a ranked candidate from the command palette.
type Match struct {
	Index int
	Score int
}

// Score fuzzy-matches query against target. Returns -1 when the query is not
// a subsequence of the target. Bonuses: consecutive matches, word boundaries,
// prefix match. Penalties: gaps and long targets. Case-insensitive.
//
// Pure and dependency-free by design (docs/sweep-2mo/44): no new module is
// needed for a palette with a few hundred entries.
func Score(query, target string) int {
	q := []rune(strings.ToLower(strings.TrimSpace(query)))
	t := []rune(target)
	if len(q) == 0 {
		return 0
	}
	if len(q) > len(t) {
		return -1
	}
	score, qi, prev, streak := 0, 0, -2, 0
	for ti, r := range t {
		if qi >= len(q) {
			break
		}
		if unicode.ToLower(r) != q[qi] {
			continue
		}
		points := 1
		if ti == prev+1 {
			streak++
			points += streak * 2
		} else {
			streak = 0
			points -= (ti - prev - 2) / 2 // gap penalty
		}
		if ti == 0 || isBoundary(t[ti-1]) {
			points += 4
		}
		prev, qi = ti, qi+1
		score += points
	}
	if qi < len(q) {
		return -1
	}
	if strings.HasPrefix(strings.ToLower(target), strings.ToLower(query)) {
		score += 12
	}
	score -= len(t) / 12
	return score
}

func isBoundary(r rune) bool {
	return r == ' ' || r == '/' || r == '\\' || r == '-' || r == '_' || r == '.' || r == ':'
}

// Rank returns up to limit matches, best first, stable for equal scores.
func Rank(query string, targets []string, limit int) []Match {
	if limit <= 0 {
		limit = len(targets)
	}
	var out []Match
	for i, t := range targets {
		if s := Score(query, t); s >= 0 {
			out = append(out, Match{Index: i, Score: s})
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Score != out[j].Score {
			return out[i].Score > out[j].Score
		}
		return out[i].Index < out[j].Index
	})
	if len(out) > limit {
		out = out[:limit]
	}
	return out
}
