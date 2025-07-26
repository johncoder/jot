package markdown

import "github.com/gomarkdown/markdown/ast"

// RemoveSubtree removes a subtree (heading and all its children) from the AST document.
func RemoveSubtree(doc ast.Node, heading *ast.Heading) {
	parent := heading.GetParent()
	if parent == nil || parent.AsContainer() == nil {
		return
	}
	children := parent.AsContainer().GetChildren()
	idx := -1
	for i, n := range children {
		if n == heading {
			idx = i
			break
		}
	}
	if idx == -1 {
		return
	}
	// Find the end of the subtree
	endIdx := idx + 1
	level := heading.Level
	for endIdx < len(children) {
		if h, ok := children[endIdx].(*ast.Heading); ok && h.Level <= level {
			break
		}
		endIdx++
	}
	newChildren := append(children[:idx], children[endIdx:]...)
	parent.AsContainer().SetChildren(newChildren)
}
