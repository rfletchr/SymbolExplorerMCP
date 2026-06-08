package extractor

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var registry = map[string]*LangDef{}

func init() {
	for _, def := range []*LangDef{GoLang, PythonLang, CLang, CppLang, RustLang, TypeScriptLang, JavaScriptLang} {
		for _, ext := range def.Extensions {
			registry[ext] = def
		}
	}
}

// LangForFile returns the LangDef for the given file path, or nil if unsupported.
func LangForFile(path string) *LangDef {
	ext := strings.ToLower(filepath.Ext(path))
	return registry[ext]
}

// ExtractPath reads and extracts symbols from a single file.
// Returns nil, nil for unsupported file types.
func ExtractPath(path string) ([]Symbol, error) {
	def := LangForFile(path)
	if def == nil {
		return nil, nil
	}
	src, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	syms, err := ExtractFile(path, src, def)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	return syms, nil
}
