package markdown

import (
	"github.com/gomarkdown/markdown/ast"
)

// OpCreateHeading creates missing headings along a selector path.
type OpCreateHeading struct {
	Selector SelectorPath
	Level    int // Markdown heading level
}

func (op *OpCreateHeading) Apply(doc ast.Node) error {
	parent := doc
	for i, seg := range op.Selector.Segments {
		head, _ := SelectorPath{Segments: op.Selector.Segments[:i+1], Fuzzy: op.Selector.Fuzzy}.FindHeadingNode(parent)
		if head == nil {
			h := &ast.Heading{Level: op.Level}
			t := &ast.Text{}
			t.Literal = []byte(seg)
			ast.AppendChild(h, t)
			ast.AppendChild(parent, h)
			parent = h
		} else {
			parent = head
		}
	}
	return nil
}
