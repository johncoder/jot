package markdown

import "github.com/gomarkdown/markdown/ast"

// OpRemove removes a subtree or content at a selector.
type OpRemove struct {
	Selector SelectorPath
}

func (op *OpRemove) Apply(doc ast.Node) error {
	node, level := op.Selector.FindHeadingNode(doc)
	if node == nil {
		return nil // node not found, nothing to remove
	}
	parent := node.GetParent()
	if parent != nil && parent.AsContainer() != nil {
		children := parent.AsContainer().GetChildren()
		headIdx := -1
		for i, c := range children {
			if c == node {
				headIdx = i
				break
			}
		}
		if headIdx == -1 {
			return nil // heading not found in parent's children
		}
		// Remove the heading and all content up to the next heading of same or higher level
		endIdx := headIdx + 1
		for i := headIdx + 1; i < len(children); i++ {
			if h, ok := children[i].(*ast.Heading); ok && h.Level <= level {
				break
			}
			endIdx = i + 1
		}
		newChildren := make([]ast.Node, 0, len(children)-(endIdx-headIdx))
		newChildren = append(newChildren, children[:headIdx]...)
		newChildren = append(newChildren, children[endIdx:]...)
		parent.AsContainer().SetChildren(newChildren)
	}
	return nil
}
