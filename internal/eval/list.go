package eval

import (
	"fmt"
	"path/filepath"
)

// ListEvalBlocks lists all evaluable code blocks in a markdown file with approval status
func ListEvalBlocks(filename string) error {
	blocks, err := ParseMarkdownForEvalBlocks(filename)
	if err != nil {
		return err
	}

	// Initialize security manager
	sm, err := NewSecurityManager()
	if err != nil {
		return fmt.Errorf("failed to initialize security manager: %w", err)
	}

	// Get absolute path for consistent checking
	absPath, err := filepath.Abs(filename)
	if err != nil {
		return err
	}

	fmt.Printf("Blocks in %s:\n", filename)

	for _, b := range blocks {
		if b.Eval != nil && b.Eval.Params["name"] != "" {
			blockName := b.Eval.Params["name"]

			// Check approval status
			approved, err := sm.CheckApproval(absPath, b)
			var status string
			if err != nil {
				status = fmt.Sprintf("ERROR: %v", err)
			} else if approved {
				status = "✓ APPROVED"
			} else {
				status = "⚠ NEEDS APPROVAL"
			}

			fmt.Printf("  %s (lines %d-%d) %s - %s\n",
				blockName, b.StartLine, b.EndLine, b.Lang, status)
		}
	}

	return nil
}
