package main

import (
	"path/filepath"
	"strings"
)

// matchGlob reports whether the slash-separated relative path matches pattern.
// Supports * (within a segment) and ** (across segments).
func matchGlob(pattern, path string) bool {
	pattern = filepath.ToSlash(pattern)
	path = filepath.ToSlash(path)

	if pattern == "" || pattern == "**" || pattern == "**/*" {
		return true
	}

	if !strings.Contains(pattern, "**") {
		ok, _ := filepath.Match(pattern, path)
		return ok
	}

	// Split on first ** and match prefix + suffix independently.
	idx := strings.Index(pattern, "**")
	prefix := strings.TrimSuffix(pattern[:idx], "/")
	suffix := strings.TrimPrefix(pattern[idx+2:], "/")

	remainder := path
	if prefix != "" {
		if !strings.HasPrefix(path, prefix+"/") {
			return false
		}
		remainder = strings.TrimPrefix(path, prefix+"/")
	}

	if suffix == "" || suffix == "*" {
		return true
	}

	// Try matching suffix against each trailing segment of remainder.
	parts := strings.Split(remainder, "/")
	for i := range parts {
		sub := strings.Join(parts[i:], "/")
		if ok, _ := filepath.Match(suffix, sub); ok {
			return true
		}
	}
	return false
}
