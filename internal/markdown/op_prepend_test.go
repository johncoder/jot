package markdown

import (
	"testing"

	"github.com/gomarkdown/markdown/parser"
)

func TestOpPrepend_Apply(t *testing.T) {
	input := "# Meetings\n\nSome notes here.\n"
	p := parser.New()
	doc := p.Parse([]byte(input))

	op := &OpPrepend{
		Selector: SelectorPath{Segments: []string{"Meetings"}},
		Content:  "Prepended note.",
	}
	err := op.Apply(doc)
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	output := RenderMarkdown(doc)
	expected := "# Meetings\nPrepended note.\nSome notes here.\n"
	if output != expected {
		t.Errorf("Output mismatch.\nExpected:\n%q\nGot:\n%q", expected, output)
	}
}
