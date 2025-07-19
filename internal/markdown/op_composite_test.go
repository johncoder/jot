package markdown

import (
	"testing"

	"github.com/gomarkdown/markdown/parser"
)

func TestOpComposite_Apply(t *testing.T) {
	input := "# Meetings\nOld content.\n"
	p := parser.New()
	doc := p.Parse([]byte(input))

	composite := &OpComposite{
		Ops: []MarkdownOp{
			&OpReplace{
				Selector: SelectorPath{Segments: []string{"Meetings"}},
				Content:  "Composite new content.",
			},
			&OpAppend{
				Selector: SelectorPath{Segments: []string{"Meetings"}},
				Content:  "Appended by composite.",
			},
		},
	}
	err := composite.Apply(doc)
	if err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	output := RenderMarkdown(doc)
	expected := "# Meetings\nComposite new content.\nAppended by composite.\n"
	if output != expected {
		t.Errorf("Output mismatch.\nExpected:\n%q\nGot:\n%q", expected, output)
	}
}
