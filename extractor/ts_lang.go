package extractor

import (
	"github.com/smacker/go-tree-sitter/javascript"
	"github.com/smacker/go-tree-sitter/typescript/typescript"
)

var commentTypes = map[string]bool{
	"comment": true,
}

var tsSymbols = []SymbolDef{
	{
		NodeType: "function_declaration",
		Kind:     "function",
		NamePath: []string{"name"},
	},
	{
		NodeType: "class_declaration",
		Kind:     "class",
		NamePath: []string{"name"},
	},
	{
		NodeType: "interface_declaration",
		Kind:     "interface",
		NamePath: []string{"name"},
	},
	{
		NodeType: "type_alias_declaration",
		Kind:     "type",
		NamePath: []string{"name"},
	},
	{
		NodeType: "method_definition",
		Kind:     "method",
		NamePath: []string{"name"},
	},
	{
		NodeType: "lexical_declaration",
		Kind:     "const",
		NamePath: []string{"#variable_declarator", "name"},
	},
}

var tsNoDescend = map[string]bool{
	"function_declaration": true,
	"method_definition":    true,
}

var tsImports = []ImportDef{
	{NodeType: "import_statement", PathField: "source", StripQuotes: true},
}

var TypeScriptLang = &LangDef{
	Language:            typescript.GetLanguage(),
	Extensions:          []string{".ts", ".tsx"},
	CommentTypes:        commentTypes,
	TransparentWrappers: []string{"export_statement"},
	NoDescend:           tsNoDescend,
	Imports:             tsImports,
	Symbols:             tsSymbols,
}

var JavaScriptLang = &LangDef{
	Language:            javascript.GetLanguage(),
	Extensions:          []string{".js", ".jsx", ".mjs"},
	CommentTypes:        commentTypes,
	TransparentWrappers: []string{"export_statement"},
	NoDescend:           tsNoDescend,
	Imports:             tsImports,
	Symbols:             tsSymbols,
}
