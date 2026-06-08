package main

import "testing"

func TestScore(t *testing.T) {
	tests := []struct {
		pattern string
		name    string
		want    int
	}{
		// Exact
		{"ExtractFile", "ExtractFile", 5},
		// Case-insensitive exact
		{"extractfile", "ExtractFile", 4},
		// Prefix
		{"Extract", "ExtractFile", 3},
		// Suffix
		{"File", "ExtractFile", 3},
		// Substring
		{"tractF", "ExtractFile", 2},
		// Initialism
		{"ef", "ExtractFile", 1},
		{"wn", "walkNode", 1},
		{"pf", "parse_file", 1},
		{"psd", "precedingSiblingDoc", 1},
		// No match
		{"xyz", "ExtractFile", 0},
		{"xyz", "walkNode", 0},
	}

	for _, tt := range tests {
		got := score(tt.pattern, tt.name)
		if got != tt.want {
			t.Errorf("score(%q, %q) = %d, want %d", tt.pattern, tt.name, got, tt.want)
		}
	}
}

func TestExtractInitials(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"ExtractFile", "EF"},
		{"walkNode", "wN"},
		{"parse_file", "pf"},
		{"precedingSiblingDoc", "pSD"},
		{"Config", "C"},
		{"GoLang", "GL"},
	}

	for _, tt := range tests {
		got := extractInitials(tt.name)
		if got != tt.want {
			t.Errorf("extractInitials(%q) = %q, want %q", tt.name, got, tt.want)
		}
	}
}
