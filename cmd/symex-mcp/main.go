package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"


	minimcp "github.com/rfletchr/MiniMCP"

	"indexmcp/symex/extractor"
)

func main() {
	d := minimcp.NewDispatcher()
	s := minimcp.NewServer("indexmcp", "0.1.0")

	sharedProps := minimcp.Properties{
		"root":          {Type: "string", Description: "Absolute path to the project root directory or a single file."},
		"use_gitignore": {Type: "boolean", Description: "Respect the project's .gitignore when walking the tree. Recommended for most projects."},
		"glob":          {Type: "string", Description: `File glob filter e.g. "extractor/**/*.go". Defaults to "**/*" (all files).`},
	}

	s.AddTool(minimcp.Tool{
		Name: "directory_summary",
		Description: "Return a directory-level overview of the project: one line per subdirectory " +
			"showing file count and symbol counts by kind. Always call this first on an unfamiliar " +
			"project to understand layout without hitting response size limits. " +
			"Then use file_summary with glob to drill into specific directories.",
		InputSchema: &minimcp.Schema{
			Type: "object",
			Properties: minimcp.Properties{
				"root":          sharedProps["root"],
				"use_gitignore": sharedProps["use_gitignore"],
				"glob":          sharedProps["glob"],
			},
			Required:             []string{"root"},
			AdditionalProperties: minimcp.Ptr(false),
		},
	}, handleDirectorySummary)

	s.AddTool(minimcp.Tool{
		Name: "file_summary",
		Description: "Return a file-level overview: each source file with a count of symbols by kind. " +
			"Results are paged — use offset and limit to step through large directories. " +
			"Narrow scope with glob (e.g. \"comfy/**\") after using directory_summary to identify " +
			"the relevant subdirectory.",
		InputSchema: &minimcp.Schema{
			Type: "object",
			Properties: minimcp.Properties{
				"root":          sharedProps["root"],
				"use_gitignore": sharedProps["use_gitignore"],
				"glob":          sharedProps["glob"],
				"offset":        {Type: "integer", Description: "Number of files to skip (default 0)."},
				"limit":         {Type: "integer", Description: "Maximum files to return (default 50)."},
			},
			Required:             []string{"root"},
			AdditionalProperties: minimcp.Ptr(false),
		},
	}, handleFileSummary)

	s.AddTool(minimcp.Tool{
		Name: "index",
		Description: "Return a symbol index grouped by file. " +
			"Use 'outline' (default) for names and kinds only, or 'full' for signatures and doc strings. " +
			"Narrow scope with 'glob' to keep output token-efficient on large projects.",
		InputSchema: &minimcp.Schema{
			Type: "object",
			Properties: minimcp.Properties{
				"root":          sharedProps["root"],
				"use_gitignore": sharedProps["use_gitignore"],
				"glob":          sharedProps["glob"],
				"detail": {
					Type:        "string",
					Enum:        []interface{}{"outline", "full"},
					Description: "outline: names and kinds only (default). full: signatures and doc strings.",
				},
			},
			Required:             []string{"root"},
			AdditionalProperties: minimcp.Ptr(false),
		},
	}, handleIndex)

	s.AddTool(minimcp.Tool{
		Name: "find_symbols",
		Description: "Locate symbols by name using a regular expression. Returns name, kind, file, " +
			"line, signature, and doc for each match — never bodies. " +
			"Use ^New to match prefix, ^New$ for exact, (?i)new for case-insensitive.",
		InputSchema: &minimcp.Schema{
			Type: "object",
			Properties: minimcp.Properties{
				"root":          sharedProps["root"],
				"use_gitignore": sharedProps["use_gitignore"],
				"glob":          sharedProps["glob"],
				"pattern":       {Type: "string", Description: "Regular expression matched against symbol names."},
				"kind":          {Type: "string", Description: "Filter by symbol kind e.g. function, method, type, class, const, var, import."},
			},
			Required:             []string{"root", "pattern"},
			AdditionalProperties: minimcp.Ptr(false),
		},
	}, handleFindSymbols)

	s.AddTool(minimcp.Tool{
		Name: "read_symbol",
		Description: "Return a named symbol's content. Use 'file' to disambiguate when the " +
			"same name appears in multiple files.",
		InputSchema: &minimcp.Schema{
			Type: "object",
			Properties: minimcp.Properties{
				"root":          sharedProps["root"],
				"use_gitignore": sharedProps["use_gitignore"],
				"name":          {Type: "string", Description: "Exact symbol name to read."},
				"file":          {Type: "string", Description: "Restrict to this file path when the name is ambiguous."},
				"detail": {
					Type:        "string",
					Enum:        []interface{}{"signature", "full"},
					Description: "signature: declaration only, no body (default). full: complete source including body.",
				},
			},
			Required:             []string{"root", "name"},
			AdditionalProperties: minimcp.Ptr(false),
		},
	}, handleReadSymbol)

	s.AddTool(minimcp.Tool{
		Name: "find_importers",
		Description: "Find all files that import a given package or module, grouped by import path. " +
			"Pattern is a regular expression matched against the import path. " +
			"Use extractor to match any path containing it, /extractor$ to anchor to the end, " +
			"^react to match react and react-dom etc. " +
			"Useful for impact analysis: which files depend on a given package.",
		InputSchema: &minimcp.Schema{
			Type: "object",
			Properties: minimcp.Properties{
				"root":          sharedProps["root"],
				"use_gitignore": sharedProps["use_gitignore"],
				"glob":          sharedProps["glob"],
				"pattern":       {Type: "string", Description: "Regular expression matched against import paths."},
			},
			Required:             []string{"root", "pattern"},
			AdditionalProperties: minimcp.Ptr(false),
		},
	}, handleFindImporters)

	s.Register(d)

	useHTTP := flag.Bool("http", false, "use HTTP transport instead of stdio")
	addr := flag.String("addr", ":3002", "listen address for HTTP transport")
	flag.Parse()

	if *useHTTP {
		http.Handle("/mcp", minimcp.NewHTTPHandler(d))
		log.Printf("indexmcp listening on %s", *addr)
		if err := http.ListenAndServe(*addr, nil); err != nil {
			log.Fatal(err)
		}
		return
	}

	minimcp.ServeStdio(d, os.Stdin, os.Stdout)
}

// --- arg structs ---

type baseArgs struct {
	Root         string `json:"root"`
	UseGitignore bool   `json:"use_gitignore"`
	Glob         string `json:"glob"`
}

func (a baseArgs) skipFunc() func(string, bool) bool {
	if !a.UseGitignore {
		return nil
	}
	g, err := loadGitIgnore(a.Root)
	if err != nil || g == nil {
		return nil
	}
	return g.skip
}

func (a baseArgs) globPattern() string {
	if a.Glob == "" {
		return "**/*"
	}
	return a.Glob
}

type directorySummaryArgs struct{ baseArgs }

type fileSummaryArgs struct {
	baseArgs
	Offset int `json:"offset"`
	Limit  int `json:"limit"`
}

type indexArgs struct {
	baseArgs
	Detail string `json:"detail"`
}

type findSymbolsArgs struct {
	baseArgs
	Pattern string `json:"pattern"`
	Kind    string `json:"kind,omitempty"`
}

type readSymbolArgs struct {
	baseArgs
	Name   string `json:"name"`
	File   string `json:"file,omitempty"`
	Detail string `json:"detail,omitempty"`
}

type findImportersArgs struct {
	baseArgs
	Pattern string `json:"pattern"`
}

// --- handlers ---

func handleDirectorySummary(raw json.RawMessage) (string, bool) {
	var args directorySummaryArgs
	if err := json.Unmarshal(raw, &args); err != nil {
		return "invalid arguments: " + err.Error(), true
	}
	syms, unscraped, err := walkPath(args.Root, args.skipFunc(), args.globPattern())
	if err != nil {
		return err.Error(), true
	}
	if len(syms) == 0 && len(unscraped) == 0 {
		return "no files found", false
	}
	return formatDirectorySummary(syms, unscraped), false
}

func handleFileSummary(raw json.RawMessage) (string, bool) {
	var args fileSummaryArgs
	if err := json.Unmarshal(raw, &args); err != nil {
		return "invalid arguments: " + err.Error(), true
	}
	limit := args.Limit
	if limit <= 0 {
		limit = 50
	}
	syms, unscraped, err := walkPath(args.Root, args.skipFunc(), args.globPattern())
	if err != nil {
		return err.Error(), true
	}
	if len(syms) == 0 && len(unscraped) == 0 {
		return "no files found", false
	}
	return formatFileSummary(syms, unscraped, args.Offset, limit), false
}

func handleIndex(raw json.RawMessage) (string, bool) {
	var args indexArgs
	if err := json.Unmarshal(raw, &args); err != nil {
		return "invalid arguments: " + err.Error(), true
	}
	syms, unscraped, err := walkPath(args.Root, args.skipFunc(), args.globPattern())
	if err != nil {
		return err.Error(), true
	}
	if len(syms) == 0 && len(unscraped) == 0 {
		return "no files found", false
	}
	detail := args.Detail
	if detail == "" {
		detail = "outline"
	}
	return formatIndex(syms, unscraped, detail), false
}

func handleFindSymbols(raw json.RawMessage) (string, bool) {
	var args findSymbolsArgs
	if err := json.Unmarshal(raw, &args); err != nil {
		return "invalid arguments: " + err.Error(), true
	}
	re, err := regexp.Compile(args.Pattern)
	if err != nil {
		return fmt.Sprintf("invalid pattern: %s", err), true
	}
	syms, _, err := walkPath(args.Root, args.skipFunc(), args.globPattern())
	if err != nil {
		return err.Error(), true
	}

	var matches []extractor.Symbol
	for _, s := range syms {
		if args.Kind != "" && s.Kind != args.Kind {
			continue
		}
		if re.MatchString(s.Name) {
			matches = append(matches, s)
		}
	}
	if len(matches) == 0 {
		return fmt.Sprintf("no symbols matching %q found", args.Pattern), false
	}
	return formatSymbols(matches), false
}

func handleReadSymbol(raw json.RawMessage) (string, bool) {
	var args readSymbolArgs
	if err := json.Unmarshal(raw, &args); err != nil {
		return "invalid arguments: " + err.Error(), true
	}
	syms, _, err := walkPath(args.Root, args.skipFunc(), args.globPattern())
	if err != nil {
		return err.Error(), true
	}

	var matches []extractor.Symbol
	for _, s := range syms {
		if s.Name != args.Name {
			continue
		}
		if args.File != "" && s.File != args.File {
			continue
		}
		matches = append(matches, s)
	}

	switch len(matches) {
	case 0:
		return fmt.Sprintf("no symbol named %q found", args.Name), false
	case 1:
		m := matches[0]
		if args.Detail == "full" {
			src, err := readLines(m.File, m.Line, m.EndLine)
			if err != nil {
				return err.Error(), true
			}
			return fmt.Sprintf("// %s:%d\n%s", m.File, m.Line, src), false
		}
		// Default: signature only — no file read needed.
		var b strings.Builder
		fmt.Fprintf(&b, "// %s:%d\n", m.File, m.Line)
		if m.Doc != "" {
			fmt.Fprintf(&b, "%s\n", m.Doc)
		}
		fmt.Fprintf(&b, "%s\n", m.Signature)
		return b.String(), false
	default:
		var b strings.Builder
		fmt.Fprintf(&b, "%q is ambiguous — %d matches. Use 'file' to disambiguate:\n", args.Name, len(matches))
		for _, m := range matches {
			fmt.Fprintf(&b, "  %s:%d  %s\n", m.File, m.Line, m.Signature)
		}
		return b.String(), false
	}
}

func handleFindImporters(raw json.RawMessage) (string, bool) {
	var args findImportersArgs
	if err := json.Unmarshal(raw, &args); err != nil {
		return "invalid arguments: " + err.Error(), true
	}
	re, err := regexp.Compile(args.Pattern)
	if err != nil {
		return fmt.Sprintf("invalid pattern: %s", err), true
	}
	syms, _, err := walkPath(args.Root, args.skipFunc(), args.globPattern())
	if err != nil {
		return err.Error(), true
	}

	type loc struct {
		file string
		line int
	}
	var order []string
	byImport := map[string][]loc{}
	for _, s := range syms {
		if s.Kind != "import" {
			continue
		}
		if !re.MatchString(s.Name) {
			continue
		}
		if _, ok := byImport[s.Name]; !ok {
			order = append(order, s.Name)
		}
		byImport[s.Name] = append(byImport[s.Name], loc{s.File, s.Line})
	}

	if len(order) == 0 {
		return fmt.Sprintf("no imports matching %q found", args.Pattern), false
	}

	var b strings.Builder
	for _, imp := range order {
		fmt.Fprintf(&b, "%s\n", imp)
		for _, l := range byImport[imp] {
			fmt.Fprintf(&b, "  %s:%d\n", l.file, l.line)
		}
	}
	return b.String(), false
}

// --- helpers ---

// walkPath walks root and returns extracted symbols and any files that were
// visited but not recognised by the extractor (no language parser available).
func walkPath(root string, skip func(string, bool) bool, glob string) ([]extractor.Symbol, []string, error) {
	info, err := os.Stat(root)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot access %q: %w", root, err)
	}
	if !info.IsDir() {
		mtime := info.ModTime()
		if syms, ok := cacheGet(root, mtime); ok {
			return syms, nil, nil
		}
		syms, err := extractor.ExtractPath(root)
		if err != nil {
			return nil, nil, err
		}
		cachePut(root, mtime, syms)
		return syms, nil, nil
	}

	var all []extractor.Symbol
	var unscraped []string
	err = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		name := d.Name()
		isDir := d.IsDir()

		if name[0] == '.' || name == "vendor" || name == "node_modules" {
			if isDir {
				return filepath.SkipDir
			}
			return nil
		}
		if skip != nil && skip(path, isDir) {
			if isDir {
				return filepath.SkipDir
			}
			return nil
		}
		if isDir {
			return nil
		}

		rel, _ := filepath.Rel(root, path)
		rel = filepath.ToSlash(rel)
		if !matchGlob(glob, rel) {
			return nil
		}

		if extractor.LangForFile(path) == nil {
			unscraped = append(unscraped, fmt.Sprintf("[%s] %s", fileClass(path), rel))
			return nil
		}

		info, err := d.Info()
		if err != nil {
			unscraped = append(unscraped, fmt.Sprintf("[error: %s] %s", err.Error(), rel))
			return nil
		}
		mtime := info.ModTime()

		if syms, ok := cacheGet(path, mtime); ok {
			if len(syms) == 0 {
				unscraped = append(unscraped, fmt.Sprintf("[no symbols] %s", rel))
			} else {
				all = append(all, syms...)
			}
			return nil
		}

		syms, err := extractor.ExtractPath(path)
		if err != nil {
			unscraped = append(unscraped, fmt.Sprintf("[error: %s] %s", err.Error(), rel))
			return nil
		}
		cachePut(path, mtime, syms)
		if len(syms) == 0 {
			unscraped = append(unscraped, fmt.Sprintf("[no symbols] %s", rel))
			return nil
		}
		all = append(all, syms...)
		return nil
	})
	return all, unscraped, err
}

// firstDir returns the top-level directory component of a slash-separated path,
// with a trailing slash. Root-level files (no slash) return "./".
func firstDir(path string) string {
	if i := strings.IndexByte(path, '/'); i >= 0 {
		return path[:i+1]
	}
	return "./"
}

func formatDirectorySummary(syms []extractor.Symbol, unscraped []string) string {
	type dirEntry struct {
		files int
		kinds map[string]int
	}
	var order []string
	byDir := map[string]*dirEntry{}
	seenFiles := map[string]bool{}

	for _, s := range syms {
		if s.Kind == "import" {
			continue
		}
		dir := firstDir(s.File)
		if _, ok := byDir[dir]; !ok {
			byDir[dir] = &dirEntry{kinds: map[string]int{}}
			order = append(order, dir)
		}
		if !seenFiles[s.File] {
			seenFiles[s.File] = true
			byDir[dir].files++
		}
		byDir[dir].kinds[s.Kind]++
	}

	maxLen := 0
	for _, d := range order {
		if len(d) > maxLen {
			maxLen = len(d)
		}
	}

	var b strings.Builder
	for _, dir := range order {
		e := byDir[dir]
		kinds := make([]string, 0, len(e.kinds))
		for k := range e.kinds {
			kinds = append(kinds, k)
		}
		sort.Strings(kinds)
		parts := make([]string, 0, len(kinds))
		for _, k := range kinds {
			parts = append(parts, fmt.Sprintf("%d %s", e.kinds[k], k))
		}
		fileWord := "files"
		if e.files == 1 {
			fileWord = "file"
		}
		fmt.Fprintf(&b, "%-*s  %d %s   %s\n", maxLen, dir, e.files, fileWord, strings.Join(parts, ", "))
	}

	// Unscraped: aggregate by directory showing count
	if len(unscraped) > 0 {
		if b.Len() > 0 {
			b.WriteByte('\n')
		}
		b.WriteString("Other files:\n")
		var uOrder []string
		uByDir := map[string]int{}
		for _, entry := range unscraped {
			path := entry
			if i := strings.Index(entry, "] "); i >= 0 {
				path = entry[i+2:]
			}
			dir := firstDir(path)
			if _, ok := uByDir[dir]; !ok {
				uOrder = append(uOrder, dir)
			}
			uByDir[dir]++
		}
		for _, dir := range uOrder {
			fmt.Fprintf(&b, "  %s  %d\n", dir, uByDir[dir])
		}
	}
	return b.String()
}

func formatFileSummary(syms []extractor.Symbol, unscraped []string, offset, limit int) string {
	var order []string
	byFile := map[string]map[string]int{}
	for _, s := range syms {
		if s.Kind == "import" {
			continue
		}
		if _, ok := byFile[s.File]; !ok {
			byFile[s.File] = map[string]int{}
			order = append(order, s.File)
		}
		byFile[s.File][s.Kind]++
	}

	total := len(order)
	start := offset
	if start > total {
		start = total
	}
	end := start + limit
	if end > total {
		end = total
	}
	page := order[start:end]

	maxLen := 0
	for _, f := range page {
		if len(f) > maxLen {
			maxLen = len(f)
		}
	}

	var b strings.Builder
	if total > limit || offset > 0 {
		fmt.Fprintf(&b, "Files %d-%d of %d:\n\n", offset+1, offset+len(page), total)
	}
	for _, file := range page {
		counts := byFile[file]
		kinds := make([]string, 0, len(counts))
		for k := range counts {
			kinds = append(kinds, k)
		}
		sort.Strings(kinds)
		parts := make([]string, 0, len(kinds))
		for _, k := range kinds {
			parts = append(parts, fmt.Sprintf("%d %s", counts[k], k))
		}
		fmt.Fprintf(&b, "%-*s  %s\n", maxLen, file, strings.Join(parts, ", "))
	}

	remaining := total - (offset + len(page))
	if remaining > 0 {
		fmt.Fprintf(&b, "\n%d more files — use offset: %d to continue\n", remaining, offset+len(page))
	}

	// Unscraped only on the last page
	if remaining == 0 && len(unscraped) > 0 {
		b.WriteByte('\n')
		b.WriteString("Other files:\n")
		for _, f := range unscraped {
			fmt.Fprintf(&b, "  %s\n", f)
		}
	}
	return b.String()
}

func formatIndex(syms []extractor.Symbol, unscraped []string, detail string) string {
	type fileGroup struct {
		file    string
		imports []string
		symbols []extractor.Symbol
	}
	var groups []fileGroup
	fileIdx := map[string]int{}
	for _, s := range syms {
		idx, ok := fileIdx[s.File]
		if !ok {
			idx = len(groups)
			fileIdx[s.File] = idx
			groups = append(groups, fileGroup{file: s.File})
		}
		if s.Kind == "import" {
			groups[idx].imports = append(groups[idx].imports, s.Name)
		} else {
			groups[idx].symbols = append(groups[idx].symbols, s)
		}
	}

	var b strings.Builder
	for i, g := range groups {
		if i > 0 {
			b.WriteByte('\n')
		}
		fmt.Fprintf(&b, "### %s\n\n", g.file)
		if len(g.imports) > 0 {
			fmt.Fprintf(&b, "imports: %s\n\n", strings.Join(g.imports, ", "))
		}
		for _, s := range g.symbols {
			switch detail {
			case "full":
				if s.Doc != "" {
					for _, line := range strings.Split(s.Doc, "\n") {
						fmt.Fprintf(&b, "%s\n", line)
					}
				}
				fmt.Fprintf(&b, "%s\n\n", s.Signature)
			default: // outline
				fmt.Fprintf(&b, "%-10s %-40s  %d\n", s.Kind, s.Name, s.Line)
			}
		}
	}
	if len(unscraped) > 0 {
		if b.Len() > 0 {
			b.WriteByte('\n')
		}
		b.WriteString("Other files:\n")
		for _, f := range unscraped {
			fmt.Fprintf(&b, "  %s\n", f)
		}
	}
	return b.String()
}

func formatSymbols(syms []extractor.Symbol) string {
	var b strings.Builder
	for _, s := range syms {
		fmt.Fprintf(&b, "%s:%d  (%s) %s\n", s.File, s.Line, s.Kind, s.Signature)
		if s.Doc != "" {
			for _, line := range strings.Split(s.Doc, "\n") {
				fmt.Fprintf(&b, "  %s\n", line)
			}
		}
		b.WriteByte('\n')
	}
	return b.String()
}

func fileClass(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return "unknown"
	}
	defer f.Close()
	buf := make([]byte, 512)
	n, _ := f.Read(buf)
	ct := http.DetectContentType(buf[:n])
	if i := strings.IndexByte(ct, ';'); i != -1 {
		ct = strings.TrimSpace(ct[:i])
	}
	return ct
}

func readLines(path string, start, end int) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read %s: %w", path, err)
	}
	lines := strings.Split(string(data), "\n")
	if start < 1 {
		start = 1
	}
	if end > len(lines) {
		end = len(lines)
	}
	return strings.Join(lines[start-1:end], "\n"), nil
}
