package extractor

import (
	"github.com/smacker/go-tree-sitter/c"
	"github.com/smacker/go-tree-sitter/cpp"
)

var CLang = &LangDef{
	Language:   c.GetLanguage(),
	Extensions: []string{".c", ".h"},
	CommentTypes: map[string]bool{
		"comment": true,
	},
	NoDescend: map[string]bool{
		"function_definition": true,
	},
	Imports: []ImportDef{
		{NodeType: "preproc_include", PathField: "path"},
	},
	Symbols: []SymbolDef{
		{
			// Name buried in declarator chain: function_definition → function_declarator → identifier.
			NodeType:     "function_definition",
			Kind:         "function",
			NameStrategy: DeclaratorChain,
		},
		{
			// typedef struct { ... } Name; — declarator field is the type_identifier directly.
			NodeType: "type_definition",
			Kind:     "type",
			NamePath: []string{"declarator"},
		},
	},
}

var CppLang = &LangDef{
	Language:   cpp.GetLanguage(),
	Extensions: []string{".cpp", ".cc", ".cxx", ".hpp"},
	CommentTypes: map[string]bool{
		"comment": true,
	},
	NoDescend: map[string]bool{
		"function_definition": true,
	},
	Imports: []ImportDef{
		{NodeType: "preproc_include", PathField: "path"},
	},
	Symbols: []SymbolDef{
		{
			NodeType:     "function_definition",
			Kind:         "function",
			NameStrategy: DeclaratorChain,
		},
		{
			NodeType: "type_definition",
			Kind:     "type",
			NamePath: []string{"declarator"},
		},
		{
			NodeType: "class_specifier",
			Kind:     "class",
			NamePath: []string{"name"},
		},
		{
			NodeType: "struct_specifier",
			Kind:     "struct",
			NamePath: []string{"name"},
		},
	},
}
