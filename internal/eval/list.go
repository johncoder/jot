package eval

import (
	"fmt"
)

// For demonstration: print all evaluable code blocks in a markdown file
func ListEvalBlocks(filename string) error {
	blocks, err := ParseMarkdownForEvalBlocks(filename)
	if err != nil {
		return err
	}
	for _, b := range blocks {
		if b.Eval != nil {
			fmt.Printf("%s (lines %d-%d): %s\n", b.Eval.Params["name"], b.StartLine, b.EndLine, b.Lang)
		}
	}
	return nil
}
