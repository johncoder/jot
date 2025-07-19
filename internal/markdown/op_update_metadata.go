package markdown

import "github.com/gomarkdown/markdown/ast"

// OpUpdateMetadata updates frontmatter or heading attributes.
type OpUpdateMetadata struct {
	Selector SelectorPath // Optional: for heading metadata
	Metadata map[string]interface{}
}

func (op *OpUpdateMetadata) Apply(doc ast.Node) error {
	if doc == nil {
		return nil // nothing to update
	}
	head, _ := op.Selector.FindHeadingNode(doc)
	if head == nil {
		return nil // heading not found, nothing to update
	}
	// TODO: Update metadata/frontmatter in AST (requires custom extension or comment node)
	return nil
}
