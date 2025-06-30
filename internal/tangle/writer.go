package tangle

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Writer handles writing tangled code blocks to files
type Writer struct {
	createDirs bool
	verbose    bool
}

// NewWriter creates a new file writer
func NewWriter() *Writer {
	return &Writer{
		createDirs: true,
		verbose:    false,
	}
}

// SetCreateDirs sets whether to create parent directories automatically
func (w *Writer) SetCreateDirs(create bool) {
	w.createDirs = create
}

// SetVerbose sets whether to output verbose logging
func (w *Writer) SetVerbose(verbose bool) {
	w.verbose = verbose
}

// WriteBlocks writes grouped tangle blocks to their respective files
func (w *Writer) WriteBlocks(groups map[string][]TangleBlock) error {
	for filePath, blocks := range groups {
		if err := w.writeFile(filePath, blocks); err != nil {
			return fmt.Errorf("failed to write %s: %w", filePath, err)
		}
	}
	return nil
}

// writeFile writes a collection of blocks to a single file
func (w *Writer) writeFile(filePath string, blocks []TangleBlock) error {
	// Create parent directories if needed
	if w.createDirs {
		dir := filepath.Dir(filePath)
		if dir != "." && dir != "" {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", dir, err)
			}
		}
	}
	
	// Combine all block contents
	var content strings.Builder
	for i, block := range blocks {
		if i > 0 {
			content.WriteString("\n") // Add newline between blocks
		}
		content.WriteString(block.Content)
	}
	
	// Write to file
	if err := os.WriteFile(filePath, []byte(content.String()), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	
	if w.verbose {
		fmt.Printf("Wrote %d block(s) to %s\n", len(blocks), filePath)
	} else {
		fmt.Printf("Tangled %s\n", filePath)
	}
	
	return nil
}

// DryRun shows what would be written without actually writing files
func (w *Writer) DryRun(groups map[string][]TangleBlock) {
	fmt.Println("Dry run - would tangle the following files:")
	
	for filePath, blocks := range groups {
		fmt.Printf("  %s (%d block(s))\n", filePath, len(blocks))
		
		if w.verbose {
			for i, block := range blocks {
				blockInfo := fmt.Sprintf("Block %d", i+1)
				if name := block.Metadata.GetName(); name != "" {
					blockInfo = fmt.Sprintf("Block %d (%s)", i+1, name)
				}
				
				lines := strings.Split(strings.TrimSpace(block.Content), "\n")
				preview := ""
				if len(lines) > 0 {
					preview = lines[0]
					if len(preview) > 50 {
						preview = preview[:47] + "..."
					}
				}
				
				fmt.Printf("    - %s: %s\n", blockInfo, preview)
			}
		}
	}
}
