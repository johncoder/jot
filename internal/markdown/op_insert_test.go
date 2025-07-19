package markdown

import (
	"testing"

	"github.com/gomarkdown/markdown/parser"
)

func TestOpInsert_Apply(t *testing.T) {
	input := "# Meetings\n\nLine 1\n\nLine 2\n"
	p := parser.New()
	doc := p.Parse([]byte(input))

	op := &OpInsert{
		Selector: SelectorPath{Segments: []string{"Meetings"}},
		Offset:   1,
		Content:  "Inserted line.",
	}
	err := op.Apply(doc)
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	output := RenderMarkdown(doc)
	expected := "# Meetings\nLine 1\nInserted line.\nLine 2\n"
	if output != expected {
		t.Errorf("Output mismatch.\nExpected:\n%q\nGot:\n%q", expected, output)
	}
}
