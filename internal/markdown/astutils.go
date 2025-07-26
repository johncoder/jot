package markdown

import (
	"fmt"
	"strings"

	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/parser"
)

// ExtractSubtree extracts a heading and all its nested content as a slice of AST nodes.
func ExtractSubtree(doc ast.Node, heading *ast.Heading) []ast.Node {
	parent := heading.GetParent()
	if parent == nil || parent.AsContainer() == nil {
		return nil
	}
	children := parent.AsContainer().GetChildren()
	headIdx := -1
	for i, n := range children {
		if n == heading {
			headIdx = i
			break
		}
	}
	if headIdx == -1 {
		return nil
	}
	level := heading.Level
	endIdx := headIdx + 1
	for endIdx < len(children) {
		if h, ok := children[endIdx].(*ast.Heading); ok && h.Level <= level {
			break
		}
		endIdx++
	}
	return children[headIdx:endIdx]
}

// InsertSubtree inserts a subtree at the specified index in the destination AST.
func InsertSubtree(doc ast.Node, index int, subtree []ast.Node) {
	container := doc.AsContainer()
	if container == nil {
		return
	}
	children := container.GetChildren()
	if index < 0 {
		index = 0
	}
	if index > len(children) {
		index = len(children)
	}
	newChildren := make([]ast.Node, 0, len(children)+len(subtree))
	newChildren = append(newChildren, children[:index]...)
	newChildren = append(newChildren, subtree...)
	newChildren = append(newChildren, children[index:]...)
	container.SetChildren(newChildren)
	for _, n := range subtree {
		n.SetParent(doc)
	}
}

// RelevelSubtree re-levels all headings in a subtree by delta, clamping to H6.
func RelevelSubtree(subtree []ast.Node, delta int) {
	for _, n := range subtree {
		walkAndRelevel(n, delta)
	}
}

func walkAndRelevel(n ast.Node, delta int) {
	if h, ok := n.(*ast.Heading); ok {
		h.Level = ClampHeadingLevel(h.Level + delta)
	}
	if c := n.AsContainer(); c != nil {
		for _, child := range c.GetChildren() {
			walkAndRelevel(child, delta)
		}
	}
}

// ClampHeadingLevel clamps heading levels to a maximum of 6 (Markdown spec).
func ClampHeadingLevel(level int) int {
	if level < 1 {
		return 1
	}
	if level > 6 {
		return 6
	}
	return level
}

// ParseMarkdownAST parses Markdown content into an *ast.Document.
func ParseMarkdownAST(content []byte) *ast.Document {
	p := parser.New()
	doc := p.Parse(content)
	if d, ok := doc.(*ast.Document); ok {
		return d
	}
	panic("failed to parse markdown as *ast.Document")
}

// RenderMarkdownAST renders an ast.Node back to Markdown string.
func RenderMarkdownAST(doc ast.Node) string {
	return RenderMarkdown(doc)
}

// ParseContentAsSubtree parses Markdown content and returns its top-level children as a subtree.
func ParseContentAsSubtree(content string) []ast.Node {
	doc := ParseMarkdownAST([]byte(content))
	return doc.GetChildren()
}

// FindSubtree locates a subtree in the AST and returns a Subtree struct for the matching heading path.
func FindSubtree(doc ast.Node, content []byte, path *HeadingPath) (*Subtree, error) {
	headings := FindAllHeadings(doc)
	for _, h := range headings {
		if len(h.Path) == len(path.Segments) {
			match := true
			for i := range h.Path {
				if strings.ToLower(h.Path[i]) != strings.ToLower(path.Segments[i]) {
					match = false
					break
				}
			}
			if match {
				subtreeNodes := ExtractSubtree(doc, h.Node)
				var subtreeContent strings.Builder
				for _, n := range subtreeNodes {
					subtreeContent.WriteString(RenderMarkdownAST(n))
				}
				return &Subtree{
					Heading:     h.Text,
					Level:       h.Level,
					Content:     []byte(subtreeContent.String()),
					StartOffset: 0, // Offset logic removed
					EndOffset:   0, // Offset logic removed
				}, nil
			}
		}
	}
	return nil, fmt.Errorf("subtree not found for path: %v", path.Segments)
}
