package markdown

import (
	"github.com/gomarkdown/markdown/ast"
)

// OpPrepend prepends content as a new paragraph after the heading node.
type OpPrepend struct {
	Selector SelectorPath
	Content  string
}

func (op *OpPrepend) Apply(doc ast.Node) error {
	head, _ := op.Selector.FindHeadingNode(doc)
	if head == nil {
		return nil // heading not found, nothing to prepend to
	}
	parent := head.GetParent()
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
	newChildren := make([]ast.Node, 0, len(children)+1)
	for _, c := range children {
		newChildren = append(newChildren, c)
		if c == head {
			// Insert after the heading
			newChildren = append(newChildren, p)
		}
	}
	docRoot.SetChildren(newChildren)
	return nil
}
