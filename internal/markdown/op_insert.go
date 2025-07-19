package markdown

import (
	"github.com/gomarkdown/markdown/ast"
)

// OpInsert inserts content at a specific offset among siblings after the heading node.
type OpInsert struct {
	Selector SelectorPath
	Offset   int // Offset after the heading node
	Content  string
}

func (op *OpInsert) Apply(doc ast.Node) error {
	node, _ := op.Selector.FindHeadingNode(doc)
	if node == nil {
		return nil // node not found, nothing to insert after
	}
	parent := node.GetParent()
	// Only allow if parent is the document root
	docRoot, isDoc := parent.(*ast.Document)
	if !isDoc {
		return nil // only insert as sibling at the document root
	}
	p := &ast.Paragraph{}
	textNode := &ast.Text{}
	textNode.Literal = []byte(op.Content)
	ast.AppendChild(p, textNode)
	children := docRoot.GetChildren()
	idx := -1
	for i, c := range children {
		if c == node {
			idx = i
			break
		}
	}
	if idx == -1 {
		return nil // node not found in parent's children
	}
	insertAt := idx + 1 + op.Offset
	if insertAt < 0 {
		insertAt = 0
	}
	if insertAt > len(children) {
		insertAt = len(children)
	}
	newChildren := make([]ast.Node, 0, len(children)+1)
	newChildren = append(newChildren, children[:insertAt]...)
	newChildren = append(newChildren, p)
	newChildren = append(newChildren, children[insertAt:]...)
	docRoot.SetChildren(newChildren)
	return nil
}
