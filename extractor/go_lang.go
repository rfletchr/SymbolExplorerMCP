package extractor

import "github.com/smacker/go-tree-sitter/golang"

var GoLang = &LangDef{
	Language:   golang.GetLanguage(),
	Extensions: []string{".go"},
	CommentTypes: map[string]bool{
		"comment": true,
	},
	TransparentWrappers: []string{"const_declaration", "var_declaration"},
	NoDescend: map[string]bool{
		"function_declaration": true,
		"method_declaration":   true,
	},
	Imports: []ImportDef{
		{NodeType: "import_spec", PathField: "path", StripQuotes: true},
	},
	Symbols: []SymbolDef{
		{
			NodeType: "function_declaration",
			Kind:     "function",
			NamePath: []string{"name"},
		},
		{
			NodeType: "method_declaration",
			Kind:     "method",
			NamePath: []string{"name"},
		},
		{
			// type_spec is an unnamed child (no field label), so use '#' prefix.
			NodeType: "type_declaration",
			Kind:     "type",
			NamePath: []string{"#type_spec", "name"},
		},
		{
			// Match const_spec directly so grouped consts each emit a symbol.
			// TransparentWrappers causes doc lookup to jump up to const_declaration.
			NodeType: "const_spec",
			Kind:     "const",
			NamePath: []string{"name"},
		},
		{
			NodeType: "var_spec",
			Kind:     "var",
			NamePath: []string{"name"},
		},
	},
}
