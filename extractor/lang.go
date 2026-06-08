package extractor

import sitter "github.com/smacker/go-tree-sitter"

// DocStyle controls how doc comments are located for a symbol.
type DocStyle int

const (
	// PrecedingSibling: walk backwards through siblings collecting comment nodes.
	// Used by Go, C, Rust, TypeScript.
	PrecedingSibling DocStyle = iota
	// BodyDocstring: first statement in the body block is a string literal.
	// Used by Python.
	BodyDocstring
)

// NameStrategy controls how the symbol name is extracted from a node.
type NameStrategy int

const (
	// FieldPath: follow a chain of named fields (and optionally typed children).
	// Path steps starting with '#' are matched by node type; others by field name.
	FieldPath NameStrategy = iota
	// DeclaratorChain: follow "declarator" fields to the leaf identifier.
	// Used for C/C++ where the name is buried in a nested declarator.
	DeclaratorChain
)

// ImportDef describes how to extract import/include paths from one node type.
type ImportDef struct {
	NodeType    string
	PathField   string // named field containing the path; if empty, use full node text
	StripQuotes bool   // strip surrounding " or ' from the extracted value
}

// SymbolDef describes how to extract one class of symbol from the AST.
type SymbolDef struct {
	NodeType     string
	Kind         string
	NameStrategy NameStrategy
	NamePath     []string // used when NameStrategy == FieldPath
	DocStyle     DocStyle
	// DocFilter: if set, only accept comment siblings that contain a child of this type.
	// Used by Rust to require outer_doc_comment_marker (/// vs //).
	DocFilter string
}

// LangDef describes a language's grammar and symbol extraction rules.
type LangDef struct {
	Language *sitter.Language
	// Extensions maps file extensions (e.g. ".go") to this language.
	Extensions []string
	// CommentTypes is the set of node type names that represent comments.
	CommentTypes map[string]bool
	// TransparentWrappers lists node types (e.g. "export_statement") that wrap a
	// declaration without changing its doc-comment ownership. When looking for
	// preceding comments, the extractor jumps up through these nodes.
	TransparentWrappers []string
	// NoDescend is the set of node types whose children are not walked.
	// Use this to prevent recursing into function bodies, which would pick up
	// local variable declarations as top-level symbols.
	NoDescend map[string]bool
	Imports   []ImportDef
	Symbols   []SymbolDef
}
