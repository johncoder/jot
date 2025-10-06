package markdown

import (
	"strings"

	"github.com/gomarkdown/markdown/ast"
)

// ExtractSubtreeByPath finds the subtree for a path of heading segments (case-insensitive contains matching).
func ExtractSubtreeByPath(doc ast.Node, segments []string) (ast.Node, error) {
	var match ast.Node
	var found bool
	ast.WalkFunc(doc, func(node ast.Node, entering bool) ast.WalkStatus {
		h, ok := node.(*ast.Heading)
		if !ok || !entering {
			return ast.GoToNext
		}
		text := strings.ToLower(string(h.Literal))
		for _, seg := range segments {
			if !strings.Contains(text, strings.ToLower(seg)) {
				return ast.GoToNext
			}
		}
		match = node
		found = true
		return ast.Terminate
	})
	if !found {
		return nil, ErrSubtreeNotFound
	}
	return match, nil
}

// RenderSubtreeContent renders the subtree as markdown text.
func RenderSubtreeContent(subtree ast.Node) string {
	if subtree == nil {
		return ""
	}
	var sb strings.Builder
	ast.WalkFunc(subtree, func(node ast.Node, entering bool) ast.WalkStatus {
		if !entering {
			return ast.GoToNext
		}
		switch n := node.(type) {
		case *ast.Heading:
			sb.WriteString(strings.Repeat("#", n.Level))
			sb.WriteString(" ")
			sb.Write(n.Literal)
			sb.WriteString("\n")
		case *ast.Text:
			sb.Write(n.Literal)
			sb.WriteString("\n")
		}
		return ast.GoToNext
	})
	return sb.String()
}

// GetHeadingText returns the heading text for a subtree.
func GetHeadingText(subtree ast.Node) string {
	h, ok := subtree.(*ast.Heading)
	if ok {
		return string(h.Literal)
	}
	return ""
}

// GetHeadingLevel returns the heading level for a subtree.
func GetHeadingLevel(subtree ast.Node) int {
	h, ok := subtree.(*ast.Heading)
	if ok {
		return h.Level
	}
	return 0
}

// CountChildHeadings counts child headings in a subtree.
func CountChildHeadings(subtree ast.Node) int {
	count := 0
	ast.WalkFunc(subtree, func(node ast.Node, entering bool) ast.WalkStatus {
		if entering {
			if _, ok := node.(*ast.Heading); ok {
				count++
			}
		}
		return ast.GoToNext
	})
	if count > 0 {
		count-- // exclude root
	}
	return count
}

var ErrSubtreeNotFound = &SubtreeError{"subtree not found"}

type SubtreeError struct{ msg string }
func (e *SubtreeError) Error() string { return e.msg }
