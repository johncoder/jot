package markdown

import (
	"github.com/gomarkdown/markdown/ast"
)

// OpAppend appends content as a new paragraph after the heading node.
type OpAppend struct {
	Selector SelectorPath
	Content  string
}

func (op *OpAppend) Apply(doc ast.Node) error {
	head, level := op.Selector.FindHeadingNode(doc)
	if head == nil {
		return nil // heading not found, nothing to append to
	}
	parent := head.GetParent()
	docRoot, isDoc := parent.(*ast.Document)
	if !isDoc {
		return nil // only insert as sibling at the document root
	}
	children := docRoot.GetChildren()
	headIdx := -1
	for i, c := range children {
		if c == head {
			headIdx = i
			break
		}
	}
	if headIdx == -1 {
		return nil // heading not found in document root's children
	}
	// Find the last node in this heading's section
	insertIdx := headIdx + 1
	for i := headIdx + 1; i < len(children); i++ {
		if h, ok := children[i].(*ast.Heading); ok && h.Level <= level {
			break
		}
		insertIdx = i + 1
	}
	p := &ast.Paragraph{}
	p.SetParent(docRoot)
	textNode := &ast.Text{}
	textNode.Literal = []byte(op.Content)
	ast.AppendChild(p, textNode)
	newChildren := make([]ast.Node, 0, len(children)+1)
	newChildren = append(newChildren, children[:insertIdx]...)
	newChildren = append(newChildren, p)
	newChildren = append(newChildren, children[insertIdx:]...)
	docRoot.SetChildren(newChildren)
	return nil
}
