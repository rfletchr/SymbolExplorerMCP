package extractor

import "github.com/smacker/go-tree-sitter/python"

var PythonLang = &LangDef{
	Language:   python.GetLanguage(),
	Extensions: []string{".py"},
	CommentTypes: map[string]bool{
		"comment": true,
	},
	NoDescend: map[string]bool{
		"function_definition": true,
	},
	Imports: []ImportDef{
		{NodeType: "import_statement"},
		{NodeType: "import_from_statement"},
	},
	Symbols: []SymbolDef{
		{
			NodeType: "function_definition",
			Kind:     "function",
			NamePath: []string{"name"},
			DocStyle: BodyDocstring,
		},
		{
			NodeType: "class_definition",
			Kind:     "class",
			NamePath: []string{"name"},
			DocStyle: BodyDocstring,
		},
	},
}
