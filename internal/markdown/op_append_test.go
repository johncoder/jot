package markdown

import (
	"testing"

	"github.com/gomarkdown/markdown/parser"
)

func TestOpAppend_Apply(t *testing.T) {
	input := "# Meetings\n\nSome notes here.\n"
	p := parser.New()
	doc := p.Parse([]byte(input))

	op := &OpAppend{
		Selector: SelectorPath{Segments: []string{"Meetings"}},
		Content:  "Discussed project X.",
	}
	err := op.Apply(doc)
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	output := RenderMarkdown(doc)
	expected := "# Meetings\nSome notes here.\nDiscussed project X.\n"
	if output != expected {
		t.Errorf("Output mismatch.\nExpected:\n%q\nGot:\n%q", expected, output)
	}
}
