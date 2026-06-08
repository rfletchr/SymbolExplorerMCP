// symex prints symbols extracted from source files or directories.
// Usage: symex [-json] <path> [path ...]
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rfletchr/SymbolExplorerMCP/extractor"
)

func main() {
	jsonOut := flag.Bool("json", false, "output JSON instead of text")
	flag.Parse()

	if flag.NArg() == 0 {
		fmt.Fprintln(os.Stderr, "usage: symex [-json] <path> [path ...]")
		os.Exit(1)
	}

	var all []extractor.Symbol
	for _, arg := range flag.Args() {
		syms, err := walk(arg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		all = append(all, syms...)
	}

	if *jsonOut {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.Encode(all)
		return
	}

	printText(all)
}

func walk(root string) ([]extractor.Symbol, error) {
	info, err := os.Stat(root)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return extractor.ExtractPath(root)
	}

	var all []extractor.Symbol
	err = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == "vendor" || d.Name() == "node_modules" || d.Name()[0] == '.' {
				return filepath.SkipDir
			}
			return nil
		}
		syms, err := extractor.ExtractPath(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warn: %v\n", err)
			return nil
		}
		all = append(all, syms...)
		return nil
	})
	return all, err
}

func printText(syms []extractor.Symbol) {
	curFile := ""
	for _, s := range syms {
		if s.File != curFile {
			fmt.Printf("\n// %s\n", s.File)
			curFile = s.File
		}
		fmt.Printf("%-10s %s\n", s.Kind, s.Signature)
		if s.Doc != "" {
			for _, line := range strings.Split(s.Doc, "\n") {
				fmt.Printf("           %s\n", line)
			}
		}
	}
}
