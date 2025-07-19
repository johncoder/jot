package markdown

import (
	"io"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/html"
)

// CommandBuffer accumulates a set of operations to apply to a Markdown document.
type CommandBuffer struct {
	Operations []MarkdownOp
}

type ApplyOption func(*applyOptions)

type applyOptions struct {
	Debug    bool
	Original []byte // If Debug, provide original source for diff
}

// WithDebug enables diff output in Apply.
func WithDebug(original []byte) ApplyOption {
	return func(opts *applyOptions) {
		opts.Debug = true
		opts.Original = original
	}
}

// Apply applies all operations in the buffer to the AST and writes the result to w.
func (buf *CommandBuffer) Apply(doc ast.Node, w io.Writer, opts ...ApplyOption) error {
	var options applyOptions
	for _, o := range opts {
		o(&options)
	}

	for _, op := range buf.Operations {
		if err := op.Apply(doc); err != nil {
			return err
		}
	}

	// Render to HTML using gomarkdown/markdown
	out := markdown.Render(doc, html.NewRenderer(html.RendererOptions{}))

	if options.Debug && options.Original != nil {
		// Output unified diff between options.Original and out
		diff := UnifiedDiff(options.Original, out)
		w.Write([]byte(diff))
	}

	_, err := w.Write(out)
	return err
}

// MarkdownOp is the interface for all buffer operations.
type MarkdownOp interface {
	Apply(doc ast.Node) error
	// Optionally: String() for human-readable diff/debug output
}

// UnifiedDiff returns a unified diff string between a and b.
func UnifiedDiff(a, b []byte) string {
	// TODO: Implement a real unified diff, or use an external package.
	return "[diff output not implemented]"
}
