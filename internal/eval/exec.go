package eval

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/johncoder/jot/internal/workspace"
)

// ExecuteEvaluableBlocks executes all evaluable code blocks in a file and returns a slice of results
// Each result contains the block, its output, and any error

type EvalResult struct {
	Block  *CodeBlock
	Output string
	Err    error
}

func ExecuteEvaluableBlocks(filename string) ([]*EvalResult, error) {
	blocks, err := ParseMarkdownForEvalBlocks(filename)
	if err != nil {
		return nil, err
	}

	// Initialize security manager
	sm, err := NewSecurityManager()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize security manager: %w", err)
	}

	// Get absolute path for consistent checking
	absPath, err := filepath.Abs(filename)
	if err != nil {
		return nil, err
	}

	var results []*EvalResult
	for _, b := range blocks {
		if b.Eval == nil {
			continue
		}

		// Check security approval
		approved, err := sm.CheckApproval(absPath, b)
		if err != nil {
			results = append(results, &EvalResult{
				Block:  b,
				Output: "",
				Err:    fmt.Errorf("security check failed: %w", err),
			})
			continue
		}

		if !approved {
			blockName := "unnamed"
			if b.Eval.Params["name"] != "" {
				blockName = b.Eval.Params["name"]
			}
			results = append(results, &EvalResult{
				Block:  b,
				Output: "",
				Err:    fmt.Errorf("code block '%s' requires approval", blockName),
			})
			continue
		}

		output, err := executeBlock(b, filename)
		results = append(results, &EvalResult{Block: b, Output: output, Err: err})
	}
	return results, nil
}

// ExecuteEvaluableBlockByName executes a specific evaluable code block by name
func ExecuteEvaluableBlockByName(filename, name string) ([]*EvalResult, error) {
	blocks, err := ParseMarkdownForEvalBlocks(filename)
	if err != nil {
		return nil, err
	}

	// Initialize security manager
	sm, err := NewSecurityManager()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize security manager: %w", err)
	}

	// Get absolute path for consistent checking
	absPath, err := filepath.Abs(filename)
	if err != nil {
		return nil, err
	}

	var results []*EvalResult
	for _, b := range blocks {
		if b.Eval == nil {
			continue
		}
		blockName, ok := b.Eval.Params["name"]
		if !ok || blockName != name {
			continue
		}

		// Check security approval
		approved, err := sm.CheckApproval(absPath, b)
		if err != nil {
			results = append(results, &EvalResult{
				Block:  b,
				Output: "",
				Err:    fmt.Errorf("security check failed: %w", err),
			})
			break
		}

		if !approved {
			results = append(results, &EvalResult{
				Block:  b,
				Output: "",
				Err:    fmt.Errorf("code block '%s' requires approval", name),
			})
			break
		}

		output, err := executeBlock(b, filename)
		results = append(results, &EvalResult{Block: b, Output: output, Err: err})
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("no evaluable block found with name '%s'", name)
	}
	return results, nil
}

// executeBlock runs the code block using the new evaluator system
func executeBlock(b *CodeBlock, filename string) (string, error) {
	lang := b.Lang
	if shell, ok := b.Eval.Params["shell"]; ok && shell != "" {
		lang = shell
	}

	// Try to get workspace context for enhanced features
	var manager *EvaluatorManager
	if ws, err := workspace.GetWorkspaceContext(false); err == nil && ws != nil {
		manager = NewEvaluatorManagerWithWorkspace(ws)
	} else {
		manager = NewEvaluatorManager()
	}

	// Set working directory - default to file's directory (org-mode behavior)
	workingDir := filepath.Dir(filename)
	if cwd, ok := b.Eval.Params["cwd"]; ok && cwd != "" {
		workingDir = cwd
	}

	// Build code string
	code := strings.Join(b.Code, "\n")

	// Execute using the evaluator system
	output, err := manager.ExecuteWithEvaluator(lang, code, b.Eval.Params, workingDir)
	if err != nil {
		// If no evaluator found, return the helpful error message
		if evalErr, ok := err.(*EvaluatorError); ok {
			return "", evalErr
		}
		return output, err
	}

	return output, nil
}

// EvaluatorError represents an error from the evaluator system
type EvaluatorError struct {
	Language string
	Message  string
}

func (e *EvaluatorError) Error() string {
	return e.Message
}
