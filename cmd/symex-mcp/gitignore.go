package main

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

type gitIgnore struct {
	root  string
	rules []ignoreRule
}

type ignoreRule struct {
	pattern string
	negate  bool
	dirOnly bool
	rooted  bool // leading / — anchored to repo root
}

func loadGitIgnore(root string) (*gitIgnore, error) {
	f, err := os.Open(filepath.Join(root, ".gitignore"))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	g := &gitIgnore{root: root}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		r := ignoreRule{}
		if strings.HasPrefix(line, "!") {
			r.negate = true
			line = line[1:]
		}
		if strings.HasSuffix(line, "/") {
			r.dirOnly = true
			line = strings.TrimSuffix(line, "/")
		}
		if strings.HasPrefix(line, "/") {
			r.rooted = true
			line = line[1:]
		}
		r.pattern = line
		g.rules = append(g.rules, r)
	}
	return g, sc.Err()
}

// skip reports whether absPath should be excluded from the walk.
func (g *gitIgnore) skip(absPath string, isDir bool) bool {
	rel, err := filepath.Rel(g.root, absPath)
	if err != nil || rel == "." {
		return false
	}
	rel = filepath.ToSlash(rel)

	matched := false
	for _, r := range g.rules {
		if r.dirOnly && !isDir {
			continue
		}
		if r.matches(rel) {
			matched = !r.negate
		}
	}
	return matched
}

func (r ignoreRule) matches(rel string) bool {
	pat := r.pattern

	// foo/** — everything under a directory
	if strings.HasSuffix(pat, "/**") {
		prefix := strings.TrimSuffix(pat, "/**")
		return rel == prefix || strings.HasPrefix(rel, prefix+"/")
	}

	// **/foo — match at any depth (treat as unanchored)
	if strings.HasPrefix(pat, "**/") {
		return matchUnanchored(strings.TrimPrefix(pat, "**/"), rel)
	}

	// pattern with interior slash or rooted — match against full relative path
	if r.rooted || strings.ContainsRune(pat, '/') {
		ok, _ := filepath.Match(pat, rel)
		return ok
	}

	// bare pattern — match against each path segment and sub-path
	return matchUnanchored(pat, rel)
}

// matchUnanchored tries pattern against every suffix of the slash-separated rel path.
func matchUnanchored(pattern, rel string) bool {
	parts := strings.Split(rel, "/")
	for i := range parts {
		if ok, _ := filepath.Match(pattern, parts[i]); ok {
			return true
		}
		if ok, _ := filepath.Match(pattern, strings.Join(parts[i:], "/")); ok {
			return true
		}
	}
	return false
}
