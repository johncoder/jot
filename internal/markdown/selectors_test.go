package markdown

import (
	"testing"

	"github.com/gomarkdown/markdown/ast"
)

func TestSelectorPathFromString(t *testing.T) {
	sp := SelectorPathFromString("Foo/Bar/Baz", false)
	if len(sp.Segments) != 3 || sp.Segments[0] != "Foo" || sp.Segments[2] != "Baz" {
		t.Errorf("unexpected segments: %+v", sp.Segments)
	}
	if sp.Fuzzy {
		t.Errorf("expected Fuzzy=false")
	}
}

func TestParseSelector(t *testing.T) {
	sp := ParseSelector("A/B/C", true)
	if len(sp.Segments) != 3 || !sp.Fuzzy {
		t.Errorf("unexpected parse: %+v", sp)
	}
}

func TestCreateHeadingNode(t *testing.T) {
	h := CreateHeadingNode("Hello", 2)
	if h.Level != 2 {
		t.Errorf("expected level 2, got %d", h.Level)
	}
	if got := headingText(h); got != "Hello" {
		t.Errorf("expected heading text 'Hello', got '%s'", got)
	}
}

func TestFindOrCreateHeadingNode(t *testing.T) {
	doc := &ast.Document{}
	sp := ParseSelector("A/B/C", false)
	h, lvl := sp.FindOrCreateHeadingNode(doc, 1)
	if h == nil {
		t.Fatal("expected heading node, got nil")
	}
	if headingText(h) != "C" {
		t.Errorf("expected heading text 'C', got '%s'", headingText(h))
	}
	if lvl != 3 {
		t.Errorf("expected level 3, got %d", lvl)
	}
	// Should not duplicate if called again
	h2, lvl2 := sp.FindOrCreateHeadingNode(doc, 1)
	if h2 != h || lvl2 != 3 {
		t.Errorf("expected to find same heading node, got different")
	}
}
