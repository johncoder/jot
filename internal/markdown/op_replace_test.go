package markdown

import (
	"testing"

	"github.com/gomarkdown/markdown/parser"
)

func TestOpReplace_Apply(t *testing.T) {
	input := "# Meetings\n\nOld content.\n"
	p := parser.New()
	doc := p.Parse([]byte(input))

	op := &OpReplace{
		Selector: SelectorPath{Segments: []string{"Meetings"}},
		Content:  "New content.",
	}
	err := op.Apply(doc)
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	output := RenderMarkdown(doc)
	expected := "# Meetings\nNew content.\n"
	if output != expected {
		t.Errorf("Output mismatch.\nExpected:\n%q\nGot:\n%q", expected, output)
	}
}
