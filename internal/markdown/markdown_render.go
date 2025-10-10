package markdown

import (
	"bytes"

	"github.com/gomarkdown/markdown/ast"
)

// RenderMarkdown renders the AST back to Markdown source.
func RenderMarkdown(node ast.Node) string {
	var buf bytes.Buffer
	renderNode(&buf, node, 0)
	return buf.String()
}

func renderNode(buf *bytes.Buffer, node ast.Node, depth int) {
	switch n := node.(type) {
	case *ast.Document:
		for _, c := range n.GetChildren() {
			renderNode(buf, c, depth)
		}
	case *ast.Heading:
		for i := 0; i < n.Level; i++ {
			buf.WriteByte('#')
		}
		buf.WriteByte(' ')
		for _, c := range n.GetChildren() {
			renderNode(buf, c, depth+1)
		}
		buf.WriteByte('\n')
	case *ast.Paragraph:
		for _, c := range n.GetChildren() {
			renderNode(buf, c, depth+1)
		}
		buf.WriteByte('\n')
	case *ast.Text:
		buf.Write(n.Literal)
		// Add more node types as needed
	}
}
