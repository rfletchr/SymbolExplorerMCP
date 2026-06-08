package extractor

import (
	"context"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
)

// ExtractFile parses src using def and returns all symbols found.
func ExtractFile(path string, src []byte, def *LangDef) ([]Symbol, error) {
	parser := sitter.NewParser()
	parser.SetLanguage(def.Language)
	tree, err := parser.ParseCtx(context.Background(), nil, src)
	if err != nil {
		return nil, err
	}
	return walkNode(tree.RootNode(), src, def, path), nil
}

func walkNode(n *sitter.Node, src []byte, def *LangDef, file string) []Symbol {
	var results []Symbol

	for _, sd := range def.Symbols {
		if n.Type() != sd.NodeType {
			continue
		}
		name := extractName(n, src, sd)
		if name == "" {
			continue
		}
		results = append(results, Symbol{
			Name:      name,
			Kind:      sd.Kind,
			File:      file,
			Line:      int(n.StartPoint().Row) + 1,
			EndLine:   int(n.EndPoint().Row) + 1,
			Signature: extractSignature(n, src),
			Doc:       extractDoc(n, src, def, sd),
		})
	}

	for _, id := range def.Imports {
		if n.Type() != id.NodeType {
			continue
		}
		path := extractImportPath(n, src, id)
		if path == "" {
			break
		}
		results = append(results, Symbol{
			Name:      path,
			Kind:      "import",
			File:      file,
			Line:      int(n.StartPoint().Row) + 1,
			EndLine:   int(n.EndPoint().Row) + 1,
			Signature: path,
		})
		break
	}

	if !def.NoDescend[n.Type()] {
		for i := 0; i < int(n.ChildCount()); i++ {
			results = append(results, walkNode(n.Child(i), src, def, file)...)
		}
	}
	return results
}

// extractName resolves the symbol name from node n using the given SymbolDef.
func extractName(n *sitter.Node, src []byte, sd SymbolDef) string {
	switch sd.NameStrategy {
	case DeclaratorChain:
		return followDeclarator(n, src)
	default: // FieldPath
		cur := n
		for _, step := range sd.NamePath {
			if strings.HasPrefix(step, "#") {
				cur = childByType(cur, step[1:])
			} else {
				cur = cur.ChildByFieldName(step)
			}
			if cur == nil {
				return ""
			}
		}
		return cur.Content(src)
	}
}

// followDeclarator walks "declarator" fields to the leaf identifier (C/C++).
func followDeclarator(n *sitter.Node, src []byte) string {
	cur := n.ChildByFieldName("declarator")
	for cur != nil {
		next := cur.ChildByFieldName("declarator")
		if next == nil {
			return cur.Content(src)
		}
		cur = next
	}
	return ""
}

// childByType returns the first named child of n with the given node type.
func childByType(n *sitter.Node, typ string) *sitter.Node {
	for i := 0; i < int(n.ChildCount()); i++ {
		c := n.Child(i)
		if c.Type() == typ {
			return c
		}
	}
	return nil
}

// extractDoc finds the documentation string for node n.
func extractDoc(n *sitter.Node, src []byte, def *LangDef, sd SymbolDef) string {
	switch sd.DocStyle {
	case BodyDocstring:
		return bodyDocstring(n, src)
	default: // PrecedingSibling
		return precedingSiblingDoc(n, src, def, sd.DocFilter)
	}
}

// precedingSiblingDoc collects comment siblings immediately before n.
// It stops at a blank line, a non-comment named sibling, or the start of the parent.
func precedingSiblingDoc(n *sitter.Node, src []byte, def *LangDef, filter string) string {
	// Jump through transparent wrappers (e.g. export_statement in TypeScript)
	// so we find comments before the wrapper, not inside it.
	subject := n
	if parent := n.Parent(); parent != nil {
		for _, w := range def.TransparentWrappers {
			if parent.Type() == w {
				subject = parent
				break
			}
		}
	}

	parent := subject.Parent()
	if parent == nil {
		return ""
	}

	idx := -1
	for i := 0; i < int(parent.ChildCount()); i++ {
		if parent.Child(i).Equal(subject) {
			idx = i
			break
		}
	}
	if idx <= 0 {
		return ""
	}

	var lines []string
	for i := idx - 1; i >= 0; i-- {
		sib := parent.Child(i)
		if def.CommentTypes[sib.Type()] {
			if filter != "" && !hasChildOfType(sib, filter) {
				break
			}
			lines = append([]string{sib.Content(src)}, lines...)
		} else if !sib.IsNamed() {
			// Unnamed node = whitespace. Two or more newlines means a blank line — stop.
			if strings.Count(sib.Content(src), "\n") >= 2 {
				break
			}
		} else {
			break
		}
	}
	// Trim trailing newlines from each line (e.g. Rust line_comment spans include \n).
	for i, l := range lines {
		lines[i] = strings.TrimRight(l, "\n")
	}
	return strings.Join(lines, "\n")
}

// bodyDocstring returns the first string literal in a function/class body (Python).
func bodyDocstring(n *sitter.Node, src []byte) string {
	body := n.ChildByFieldName("body")
	if body == nil {
		return ""
	}
	for i := 0; i < int(body.ChildCount()); i++ {
		child := body.Child(i)
		if !child.IsNamed() {
			continue
		}
		if child.Type() == "expression_statement" && child.ChildCount() > 0 {
			if first := child.Child(0); first.Type() == "string" {
				return first.Content(src)
			}
		}
		break // only check the first named statement
	}
	return ""
}

func hasChildOfType(n *sitter.Node, typ string) bool {
	for i := 0; i < int(n.ChildCount()); i++ {
		if n.Child(i).Type() == typ {
			return true
		}
	}
	return false
}

// extractImportPath extracts the import path string from an import node.
func extractImportPath(n *sitter.Node, src []byte, def ImportDef) string {
	if def.PathField != "" {
		child := n.ChildByFieldName(def.PathField)
		if child == nil {
			return ""
		}
		text := child.Content(src)
		if def.StripQuotes && len(text) >= 2 {
			if (text[0] == '"' && text[len(text)-1] == '"') ||
				(text[0] == '\'' && text[len(text)-1] == '\'') {
				return text[1 : len(text)-1]
			}
		}
		return text
	}
	return strings.TrimRight(n.Content(src), ";\n\r\t ")
}

// extractSignature returns the declaration text without the body.
func extractSignature(n *sitter.Node, src []byte) string {
	body := n.ChildByFieldName("body")
	var raw string
	if body != nil {
		raw = string(src[n.StartByte():body.StartByte()])
	} else {
		raw = n.Content(src)
	}
	// Normalise internal whitespace to single spaces.
	return strings.Join(strings.Fields(raw), " ")
}
