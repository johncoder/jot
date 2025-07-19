package markdown

import (
	"github.com/gomarkdown/markdown/ast"
)

// OpReplace replaces the first paragraph after the heading node with new content.
type OpReplace struct {
	Selector SelectorPath
	Content  string
}

func (op *OpReplace) Apply(doc ast.Node) error {
	node, _ := op.Selector.FindHeadingNode(doc)
	if node == nil {
		return nil // node not found, nothing to replace after
	}
	parent := node.GetParent()
	// Only allow if parent is the document root
	docRoot, isDoc := parent.(*ast.Document)
	if !isDoc {
		return nil // only replace as sibling at the document root
	}
	children := docRoot.GetChildren()
	idx := -1
	for i, c := range children {
		if c == node {
			idx = i
			break
		}
	}
	if idx == -1 || idx+1 >= len(children) {
		return nil // no paragraph to replace
	}
	// Find the first paragraph after the heading
	replaceIdx := -1
	for i := idx + 1; i < len(children); i++ {
		if _, ok := children[i].(*ast.Paragraph); ok {
			replaceIdx = i
			break
		}
	}
	if replaceIdx == -1 {
		return nil // no paragraph to replace
	}
	p := &ast.Paragraph{}
	textNode := &ast.Text{}
	textNode.Literal = []byte(op.Content)
	ast.AppendChild(p, textNode)
	children[replaceIdx] = p
	docRoot.SetChildren(children)
	return nil
}
