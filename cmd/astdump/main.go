// astdump prints the tree-sitter AST for a source file.
// Usage: astdump [-depth N] <file>
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/c"
	"github.com/smacker/go-tree-sitter/cpp"
	"github.com/smacker/go-tree-sitter/golang"
	"github.com/smacker/go-tree-sitter/python"
	"github.com/smacker/go-tree-sitter/rust"
	"github.com/smacker/go-tree-sitter/typescript/typescript"
)

var langByExt = map[string]*sitter.Language{
	".go":  golang.GetLanguage(),
	".py":  python.GetLanguage(),
	".c":   c.GetLanguage(),
	".h":   c.GetLanguage(),
	".cpp": cpp.GetLanguage(),
	".cc":  cpp.GetLanguage(),
	".rs":  rust.GetLanguage(),
	".ts":  typescript.GetLanguage(),
	".tsx": typescript.GetLanguage(),
}

func main() {
	maxDepth := flag.Int("depth", 8, "max depth to print (-1 for unlimited)")
	flag.Parse()

	if flag.NArg() != 1 {
		fmt.Fprintln(os.Stderr, "usage: astdump [-depth N] <file>")
		os.Exit(1)
	}

	path := flag.Arg(0)
	src, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	ext := strings.ToLower(filepath.Ext(path))
	lang, ok := langByExt[ext]
	if !ok {
		fmt.Fprintf(os.Stderr, "unsupported extension: %s\n", ext)
		os.Exit(1)
	}

	parser := sitter.NewParser()
	parser.SetLanguage(lang)
	tree, err := parser.ParseCtx(context.Background(), nil, src)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	printNode(tree.RootNode(), src, "", *maxDepth, 0)
}

func printNode(n *sitter.Node, src []byte, fieldName string, maxDepth, depth int) {
	if maxDepth >= 0 && depth > maxDepth {
		return
	}

	indent := strings.Repeat("  ", depth)
	label := ""
	if fieldName != "" {
		label = fieldName + ": "
	}

	start := n.StartPoint()
	end := n.EndPoint()
	pos := fmt.Sprintf("[%d:%d - %d:%d]", start.Row+1, start.Column, end.Row+1, end.Column)

	if n.ChildCount() == 0 {
		// Leaf node — print content
		content := n.Content(src)
		if len(content) > 60 {
			content = content[:60] + "…"
		}
		fmt.Printf("%s%s%s %s %q\n", indent, label, n.Type(), pos, content)
	} else {
		fmt.Printf("%s%s%s %s\n", indent, label, n.Type(), pos)
		for i := 0; i < int(n.ChildCount()); i++ {
			child := n.Child(i)
			childField := n.FieldNameForChild(i)
			printNode(child, src, childField, maxDepth, depth+1)
		}
	}
}
