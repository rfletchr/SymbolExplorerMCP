package extractor

import (
	"testing"
)

// sym is a partial Symbol used in test expectations.
// Fields left as zero/empty are not checked.
type sym struct {
	Name      string
	Kind      string
	Line      int
	Signature string
	Doc       string
}

func check(t *testing.T, lang *LangDef, src string, want []sym) {
	t.Helper()
	got, err := ExtractFile("test", []byte(src), lang)
	if err != nil {
		t.Fatalf("extract error: %v", err)
	}
	if len(got) != len(want) {
		t.Errorf("got %d symbols, want %d:", len(got), len(want))
		for i, s := range got {
			t.Logf("  [%d] %s %q sig=%q doc=%q", i, s.Kind, s.Name, s.Signature, s.Doc)
		}
		return
	}
	for i, w := range want {
		g := got[i]
		if g.Name != w.Name {
			t.Errorf("[%d] Name: got %q want %q", i, g.Name, w.Name)
		}
		if g.Kind != w.Kind {
			t.Errorf("[%d] Kind: got %q want %q", i, g.Kind, w.Kind)
		}
		if w.Line != 0 && g.Line != w.Line {
			t.Errorf("[%d] Line: got %d want %d", i, g.Line, w.Line)
		}
		if w.Signature != "" && g.Signature != w.Signature {
			t.Errorf("[%d] Signature: got %q want %q", i, g.Signature, w.Signature)
		}
		if w.Doc != "" && g.Doc != w.Doc {
			t.Errorf("[%d] Doc: got %q want %q", i, g.Doc, w.Doc)
		}
	}
}

// --- Go ---

func TestGoFunction(t *testing.T) {
	check(t, GoLang, `package p
// ParseFile reads and parses a source file.
func ParseFile(path string) (*AST, error) { return nil, nil }
`, []sym{
		{Name: "ParseFile", Kind: "function",
			Signature: "func ParseFile(path string) (*AST, error)",
			Doc:       "// ParseFile reads and parses a source file."},
	})
}

func TestGoMethod(t *testing.T) {
	check(t, GoLang, `package p
// Run executes the parser.
func (c *Config) Run() error { return nil }
`, []sym{
		{Name: "Run", Kind: "method",
			Signature: "func (c *Config) Run() error",
			Doc:       "// Run executes the parser."},
	})
}

func TestGoType(t *testing.T) {
	check(t, GoLang, `package p
// Config holds parser settings.
type Config struct { MaxDepth int }
`, []sym{
		{Name: "Config", Kind: "type", Doc: "// Config holds parser settings."},
	})
}

func TestGoMultiLineDoc(t *testing.T) {
	check(t, GoLang, `package p
// ParseFile reads and parses a source file.
// It returns an error if the file cannot be read.
func ParseFile(path string) error { return nil }
`, []sym{
		{Name: "ParseFile", Kind: "function",
			Doc: "// ParseFile reads and parses a source file.\n// It returns an error if the file cannot be read."},
	})
}

func TestGoBlankLineSeparatesDoc(t *testing.T) {
	// A blank line between comment and declaration means no doc.
	check(t, GoLang, `package p
// This comment is detached.

func ParseFile() {}
`, []sym{
		{Name: "ParseFile", Kind: "function", Doc: ""},
	})
}

func TestGoGroupedConsts(t *testing.T) {
	check(t, GoLang, `package p
// Depth limits for the parser.
const (
	MaxDepth = 8
	MinDepth = 1
)
`, []sym{
		{Name: "MaxDepth", Kind: "const", Doc: "// Depth limits for the parser."},
		{Name: "MinDepth", Kind: "const", Doc: "// Depth limits for the parser."},
	})
}

func TestGoNoDoc(t *testing.T) {
	check(t, GoLang, `package p
func undocumented() {}
`, []sym{
		{Name: "undocumented", Kind: "function", Doc: ""},
	})
}

func TestGoLocalVarsNotExtracted(t *testing.T) {
	// Variables and consts declared inside function bodies must not appear.
	check(t, GoLang, `package p
func process() {
	var buf strings.Builder
	const limit = 10
	result := buf.String()
	_ = result
	_ = limit
}
`, []sym{
		{Name: "process", Kind: "function"},
	})
}

// --- Python ---

func TestPythonFunction(t *testing.T) {
	check(t, PythonLang, `def parse_file(path: str) -> list:
    """Read and parse the given source file."""
    return []
`, []sym{
		{Name: "parse_file", Kind: "function",
			Doc: `"""Read and parse the given source file."""`},
	})
}

func TestPythonClass(t *testing.T) {
	check(t, PythonLang, `class Config:
    """Holds parser configuration."""
    def __init__(self, max_depth: int = 8):
        self.max_depth = max_depth
`, []sym{
		{Name: "Config", Kind: "class", Doc: `"""Holds parser configuration."""`},
		{Name: "__init__", Kind: "function"},
	})
}

func TestPythonNoDocstring(t *testing.T) {
	check(t, PythonLang, `def undocumented():
    return None
`, []sym{
		{Name: "undocumented", Kind: "function", Doc: ""},
	})
}

// --- Rust ---

func TestRustFunction(t *testing.T) {
	check(t, RustLang, `/// Parses the given source file.
pub fn parse_file(path: &str) -> Result<(), ()> { Ok(()) }
`, []sym{
		{Name: "parse_file", Kind: "function",
			Signature: "pub fn parse_file(path: &str) -> Result<(), ()>",
			Doc:       "/// Parses the given source file."},
	})
}

func TestRustRegularCommentNotDoc(t *testing.T) {
	// A plain // comment should not be attached as doc (only /// is).
	check(t, RustLang, `// This is not a doc comment.
pub fn parse_file() {}
`, []sym{
		{Name: "parse_file", Kind: "function", Doc: ""},
	})
}

func TestRustStruct(t *testing.T) {
	check(t, RustLang, `/// Holds parser configuration.
pub struct Config {
    pub max_depth: i32,
}
`, []sym{
		{Name: "Config", Kind: "struct", Doc: "/// Holds parser configuration."},
	})
}

func TestRustImplMethod(t *testing.T) {
	check(t, RustLang, `pub struct Config {}
impl Config {
    /// Executes the parser.
    pub fn run(&self) {}
}
`, []sym{
		{Name: "Config", Kind: "struct"},
		{Name: "run", Kind: "function", Doc: "/// Executes the parser."},
	})
}

// --- TypeScript ---

func TestTypeScriptFunction(t *testing.T) {
	check(t, TypeScriptLang, `/** Parses the given source file. */
export function parseFile(path: string): void {}
`, []sym{
		{Name: "parseFile", Kind: "function",
			Doc: "/** Parses the given source file. */"},
	})
}

func TestTypeScriptInterface(t *testing.T) {
	check(t, TypeScriptLang, `/** Holds parser configuration. */
export interface Config {
    maxDepth: number;
}
`, []sym{
		{Name: "Config", Kind: "interface", Doc: "/** Holds parser configuration. */"},
	})
}

func TestTypeScriptClassMethod(t *testing.T) {
	check(t, TypeScriptLang, `class Parser {
    /** Executes the parser. */
    run(): void {}
}
`, []sym{
		{Name: "Parser", Kind: "class"},
		{Name: "run", Kind: "method", Doc: "/** Executes the parser. */"},
	})
}

// --- C ---

func TestCFunction(t *testing.T) {
	check(t, CLang, `/* Parses the given source file. */
int parse_file(const char *path) { return 0; }
`, []sym{
		{Name: "parse_file", Kind: "function",
			Doc: "/* Parses the given source file. */"},
	})
}

func TestCTypedef(t *testing.T) {
	check(t, CLang, `/* Holds parser configuration. */
typedef struct { int max_depth; } Config;
`, []sym{
		{Name: "Config", Kind: "type", Doc: "/* Holds parser configuration. */"},
	})
}

// --- imports ---

func TestGoImports(t *testing.T) {
	check(t, GoLang, `package p
import (
	"fmt"
	"os"
)
func Foo() {}
`, []sym{
		{Name: "fmt", Kind: "import"},
		{Name: "os", Kind: "import"},
		{Name: "Foo", Kind: "function"},
	})
}

func TestCIncludes(t *testing.T) {
	check(t, CLang, `#include <stdio.h>
#include "mylib.h"
int main() { return 0; }
`, []sym{
		{Name: "<stdio.h>", Kind: "import"},
		{Name: "\"mylib.h\"", Kind: "import"},
		{Name: "main", Kind: "function"},
	})
}

func TestTypeScriptImports(t *testing.T) {
	check(t, TypeScriptLang, `import React from 'react';
import { useState } from 'react';
function App() {}
`, []sym{
		{Name: "react", Kind: "import"},
		{Name: "react", Kind: "import"},
		{Name: "App", Kind: "function"},
	})
}

func TestPythonImports(t *testing.T) {
	check(t, PythonLang, `import os
from os import path
def read_file(p): pass
`, []sym{
		{Name: "import os", Kind: "import"},
		{Name: "from os import path", Kind: "import"},
		{Name: "read_file", Kind: "function"},
	})
}

func TestRustImports(t *testing.T) {
	check(t, RustLang, `use std::collections::HashMap;
use std::io::{self, Write};
fn process() {}
`, []sym{
		{Name: "use std::collections::HashMap", Kind: "import"},
		{Name: "use std::io::{self, Write}", Kind: "import"},
		{Name: "process", Kind: "function"},
	})
}
