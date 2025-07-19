package markdown

import (
	"strings"

	"github.com/gomarkdown/markdown/ast"
)

// SelectorPath represents a parsed selector (supports fuzzy matching).
type SelectorPath struct {
	Segments []string
	Fuzzy    bool
}

// FindHeadingNode traverses the AST and returns the heading node matching the selector path.
// Returns the node and its level, or nil if not found.
func (sp SelectorPath) FindHeadingNode(doc ast.Node) (ast.Node, int) {
	current := doc
	level := 0
	for _, seg := range sp.Segments {
		found := false
		children := current.GetChildren()
		for _, n := range children {
			h, ok := n.(*ast.Heading)
			if !ok {
				continue
			}
			txt := headingText(h)
			if sp.Fuzzy {
				if fuzzyMatch(txt, seg) {
					current = h // return the heading node itself
					level = h.Level
					found = true
					break
				}
			} else {
				if txt == seg {
					current = h
					level = h.Level
					found = true
					break
				}
			}
		}
		if !found {
			return nil, 0
		}
	}
	return current, level
}

// headingText extracts the text from a heading node.
func headingText(h *ast.Heading) string {
	var sb strings.Builder
	children := h.GetChildren()
	for _, n := range children {
		if t, ok := n.(*ast.Text); ok {
			sb.WriteString(string(t.Literal))
		}
	}
	return sb.String()
}

// fuzzyMatch does a case-insensitive comparison, ignoring punctuation and whitespace.
func fuzzyMatch(a, b string) bool {
	a = strings.ToLower(strings.TrimSpace(a))
	b = strings.ToLower(strings.TrimSpace(b))
	a = strings.Map(func(r rune) rune {
		if strings.ContainsRune("abcdefghijklmnopqrstuvwxyz0123456789", r) {
			return r
		}
		return -1
	}, a)
	b = strings.Map(func(r rune) rune {
		if strings.ContainsRune("abcdefghijklmnopqrstuvwxyz0123456789", r) {
			return r
		}
		return -1
	}, b)
	return a == b
}

// TODO: Add helpers for creating headings if missing, fuzzy/partial matching, etc.
