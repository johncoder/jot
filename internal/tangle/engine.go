package tangle

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/johncoder/jot/internal/eval"
	"github.com/johncoder/jot/internal/workspace"
)

// TangleBlock represents a code block that can be tangled
type TangleBlock struct {
	Metadata *eval.EvalMetadata
	Content  string
	FilePath string
	Language string
}

// Engine handles the extraction of code blocks to files
type Engine struct {
	blocks []TangleBlock
}

// NewEngine creates a new tangle engine
func NewEngine() *Engine {
	return &Engine{
		blocks: make([]TangleBlock, 0),
	}
}

// FindTangleBlocks scans a markdown file for code blocks with tangle attributes
func (e *Engine) FindTangleBlocks(ws *workspace.Workspace, filePath string, noWorkspace bool) error {
	// Parse the markdown file for eval blocks using the existing eval system
	codeBlocks, err := eval.ParseMarkdownForEvalBlocks(filePath)
	if err != nil {
		return fmt.Errorf("failed to parse markdown: %w", err)
	}

	// Filter for tangle blocks and convert to TangleBlock format
	for _, block := range codeBlocks {
		if block.Eval != nil && block.Eval.IsTangleElement() {
			tangleFilePath := block.Eval.GetTangleFile()
			if tangleFilePath == "" {
				// Skip blocks without file specification
				continue
			}

			// Resolve tangle file path relative to workspace or current directory
			absoluteTangleFilePath := resolveTangleFilePath(ws, tangleFilePath, noWorkspace)

			content := ""
			if len(block.Code) > 0 {
				// Join the code lines back together
				for i, line := range block.Code {
					if i > 0 {
						content += "\n"
					}
					content += line
				}
			}

			tangleBlock := TangleBlock{
				Metadata: block.Eval,
				Content:  content,
				FilePath: absoluteTangleFilePath,
				Language: block.Lang,
			}

			e.blocks = append(e.blocks, tangleBlock)
		}
	}

	return nil
}

// GetTangleBlocks returns all found tangle blocks
func (e *Engine) GetTangleBlocks() []TangleBlock {
	return e.blocks
}

// GroupBlocksByFile groups tangle blocks by their target file path
func (e *Engine) GroupBlocksByFile() map[string][]TangleBlock {
	groups := make(map[string][]TangleBlock)

	for _, block := range e.blocks {
		groups[block.FilePath] = append(groups[block.FilePath], block)
	}

	return groups
}

// resolveTangleFilePath consolidates file path resolution logic for tangle operations
func resolveTangleFilePath(ws *workspace.Workspace, filename string, noWorkspace bool) string {
	if noWorkspace {
		// Non-workspace mode: resolve relative to current directory
		if filepath.IsAbs(filename) {
			return filename
		}
		// Get current working directory and resolve relative to it
		cwd, _ := os.Getwd()
		return filepath.Join(cwd, filename)
	}

	// Workspace mode: existing logic
	if filepath.IsAbs(filename) {
		return filename
	}
	if ws != nil {
		return filepath.Join(ws.Root, filename)
	}
	return filename // Fallback
}
