package markdown

import (
	"fmt"

	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/parser"
)

func PrintAST(n ast.Node, depth int) {
	if n == nil {
		return
	}
	for i := 0; i < depth; i++ {
		fmt.Print("  ")
	}
	fmt.Printf("%T", n)
	if h, ok := n.(*ast.Heading); ok {
		fmt.Printf(" (Heading level %d)", h.Level)
	}
	if t, ok := n.(*ast.Text); ok {
		fmt.Printf(" (Text: %q)", t.Literal)
	}
	fmt.Println()
	for _, child := range n.GetChildren() {
		PrintAST(child, depth+1)
	}
}

func PrintTestOpAppendAST() {
	input := "# Meetings\n\nSome notes here.\n"
	p := parser.New()
	doc := p.Parse([]byte(input))
	PrintAST(doc, 0)
}
