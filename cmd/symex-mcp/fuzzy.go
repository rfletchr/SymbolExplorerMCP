package main

import (
	"strings"
	"unicode"
)

// score returns how well pattern matches name. Higher is better, 0 means no match.
// 5 exact
// 4 case-insensitive exact
// 3 prefix or suffix (case-insensitive)
// 2 substring (case-insensitive)
// 1 initialism (e.g. "ef" matches "ExtractFile")
func score(pattern, name string) int {
	if name == pattern {
		return 5
	}
	pl := strings.ToLower(pattern)
	nl := strings.ToLower(name)
	if nl == pl {
		return 4
	}
	if strings.HasPrefix(nl, pl) || strings.HasSuffix(nl, pl) {
		return 3
	}
	if strings.Contains(nl, pl) {
		return 2
	}
	if initialsMatch(pl, name) {
		return 1
	}
	return 0
}

// initialsMatch reports whether pattern matches the first letter of each word
// in name, where words are split by camelCase boundaries and underscores.
// e.g. "ef" matches "ExtractFile", "wn" matches "walkNode", "pf" matches "parse_file"
func initialsMatch(pattern, name string) bool {
	initials := extractInitials(name)
	pl := strings.ToLower(pattern)
	il := strings.ToLower(initials)
	return il == pl || strings.HasPrefix(il, pl)
}

func extractInitials(name string) string {
	var b strings.Builder
	afterUnderscore := false
	for i, r := range name {
		if r == '_' {
			afterUnderscore = true
			continue
		}
		if i == 0 || afterUnderscore || unicode.IsUpper(r) {
			b.WriteRune(r)
		}
		afterUnderscore = false
	}
	return b.String()
}
