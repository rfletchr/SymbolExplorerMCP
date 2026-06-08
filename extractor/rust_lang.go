package extractor

import "github.com/smacker/go-tree-sitter/rust"

var RustLang = &LangDef{
	Language:   rust.GetLanguage(),
	Extensions: []string{".rs"},
	CommentTypes: map[string]bool{
		"line_comment":  true,
		"block_comment": true,
	},
	NoDescend: map[string]bool{
		"function_item": true,
		"struct_item":   true,
		"enum_item":     true,
		"trait_item":    true,
	},
	Imports: []ImportDef{
		{NodeType: "use_declaration"},
	},
	Symbols: []SymbolDef{
		{
			NodeType:  "function_item",
			Kind:      "function",
			NamePath:  []string{"name"},
			DocFilter: "outer_doc_comment_marker",
		},
		{
			NodeType:  "struct_item",
			Kind:      "struct",
			NamePath:  []string{"name"},
			DocFilter: "outer_doc_comment_marker",
		},
		{
			NodeType:  "enum_item",
			Kind:      "enum",
			NamePath:  []string{"name"},
			DocFilter: "outer_doc_comment_marker",
		},
		{
			NodeType:  "trait_item",
			Kind:      "trait",
			NamePath:  []string{"name"},
			DocFilter: "outer_doc_comment_marker",
		},
		{
			NodeType:  "type_item",
			Kind:      "type",
			NamePath:  []string{"name"},
			DocFilter: "outer_doc_comment_marker",
		},
	},
}
