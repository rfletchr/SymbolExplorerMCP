package main

import (
	"os"
	"path/filepath"
	"testing"
)

func makeGitIgnore(t *testing.T, content string) *gitIgnore {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".gitignore"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	g, err := loadGitIgnore(dir)
	if err != nil {
		t.Fatal(err)
	}
	// Rewrite rules so skip() receives paths relative to dir for easier testing.
	// We abuse the fact that skip() calls filepath.Rel(g.root, absPath),
	// so we pass filepath.Join(dir, rel) as the absPath.
	g.root = dir
	return g
}

func TestGitIgnore(t *testing.T) {
	tests := []struct {
		name    string
		rules   string
		path    string
		isDir   bool
		want    bool
	}{
		// Extension glob
		{"log file matched",       "*.log\n",        "app.log",            false, true},
		{"log in subdir matched",  "*.log\n",        "sub/app.log",        false, true},
		{"go file not matched",    "*.log\n",        "main.go",            false, false},

		// Directory-only pattern (trailing /)
		{"build dir matched",      "build/\n",       "build",              true,  true},
		{"build file not matched", "build/\n",       "build",              false, false},
		{"nested build matched",    "build/\n",       "pkg/build",          true,  true},
		// /build/ anchors to root — nested should not match
		{"rooted build not nested", "/build/\n",     "pkg/build",          true,  false},

		// Rooted pattern (leading /)
		{"rooted dist matched",    "/dist\n",        "dist",               true,  true},
		{"rooted dist not nested", "/dist\n",        "pkg/dist",           true,  false},

		// ** patterns
		{"any depth generated",    "**/generated\n", "generated",          true,  true},
		{"any depth nested",       "**/generated\n", "pkg/generated",      true,  true},
		{"any depth deep",         "**/generated\n", "a/b/generated",      true,  true},
		{"foo/** contents",        "foo/**\n",        "foo/bar.go",        false, true},
		{"foo/** subdir",          "foo/**\n",        "foo/sub",           true,  true},
		{"foo/** sibling",         "foo/**\n",        "other/bar.go",      false, false},

		// Negation
		{"negation keeps file",    "*.log\n!keep.log\n", "keep.log",       false, false},
		{"negation still skips",   "*.log\n!keep.log\n", "drop.log",       false, true},

		// No .gitignore match
		{"unmatched file",         "*.log\n",        "main.go",            false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := makeGitIgnore(t, tt.rules)
			abs := filepath.Join(g.root, filepath.FromSlash(tt.path))
			got := g.skip(abs, tt.isDir)
			if got != tt.want {
				t.Errorf("skip(%q, dir=%v) = %v, want %v", tt.path, tt.isDir, got, tt.want)
			}
		})
	}
}

func TestLoadGitIgnoreNotExist(t *testing.T) {
	g, err := loadGitIgnore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if g != nil {
		t.Error("expected nil when .gitignore absent")
	}
}
