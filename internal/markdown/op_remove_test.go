package markdown

import (
	"testing"

	"github.com/gomarkdown/markdown/parser"
)

func TestOpRemove_Apply(t *testing.T) {
	input := "# Meetings\nRemove this line.\nKeep this line.\n"
	p := parser.New()
	doc := p.Parse([]byte(input))

	op := &OpRemove{
		Selector: SelectorPath{Segments: []string{"Meetings"}},
	}
	err := op.Apply(doc)
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	output := RenderMarkdown(doc)
	expected := ""
	if output != expected {
		t.Errorf("Output mismatch.\nExpected:\n%q\nGot:\n%q", expected, output)
	}
}
