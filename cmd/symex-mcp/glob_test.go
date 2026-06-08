package main

import "testing"

func TestMatchGlob(t *testing.T) {
	tests := []struct {
		pattern string
		path    string
		want    bool
	}{
		// Default catch-all
		{"**/*", "foo.go", true},
		{"**/*", "pkg/foo.go", true},

		// Extension filter
		{"**/*.go", "foo.go", true},
		{"**/*.go", "sub/foo.go", true},
		{"**/*.go", "foo.py", false},

		// Directory prefix
		{"extractor/**", "extractor/foo.go", true},
		{"extractor/**", "extractor/sub/foo.go", true},
		{"extractor/**", "cmd/foo.go", false},

		// Combined prefix + extension
		{"extractor/**/*.go", "extractor/foo.go", true},
		{"extractor/**/*.go", "extractor/sub/foo.go", true},
		{"extractor/**/*.go", "extractor/foo.py", false},
		{"extractor/**/*.go", "cmd/foo.go", false},

		// No ** — exact filepath.Match
		{"*.go", "foo.go", true},
		{"*.go", "sub/foo.go", false},
	}

	for _, tt := range tests {
		got := matchGlob(tt.pattern, tt.path)
		if got != tt.want {
			t.Errorf("matchGlob(%q, %q) = %v, want %v", tt.pattern, tt.path, got, tt.want)
		}
	}
}
