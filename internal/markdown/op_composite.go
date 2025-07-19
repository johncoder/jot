package markdown

import "github.com/gomarkdown/markdown/ast"

// OpComposite batches multiple operations as a single logical command.
type OpComposite struct {
	Ops []MarkdownOp
}

func (op *OpComposite) Apply(doc ast.Node) error {
	for _, subop := range op.Ops {
		if err := subop.Apply(doc); err != nil {
			return err
		}
	}
	return nil
}
