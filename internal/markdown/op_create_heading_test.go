package markdown

import (
	"testing"

	"github.com/gomarkdown/markdown/parser"
)

func TestOpCreateHeading_Apply(t *testing.T) {
	input := ""
	p := parser.New()
	doc := p.Parse([]byte(input))

	op := &OpCreateHeading{
		Selector: SelectorPath{Segments: []string{"New Heading"}},
		Level:    1,
	}
	err := op.Apply(doc)
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	output := RenderMarkdown(doc)
	expected := "# New Heading\n"
	if output != expected {
		t.Errorf("Output mismatch.\nExpected:\n%q\nGot:\n%q", expected, output)
	}
}
