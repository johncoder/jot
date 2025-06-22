package eval

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
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

		output, err := executeBlock(b)
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

		output, err := executeBlock(b)
		results = append(results, &EvalResult{Block: b, Output: output, Err: err})
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("no evaluable block found with name '%s'", name)
	}
	return results, nil
}

// executeBlock runs the code block using the shell/interpreter specified in EvalMetadata or inferred from language
func executeBlock(b *CodeBlock) (string, error) {
	lang := b.Lang
	if shell, ok := b.Eval.Params["shell"]; ok && shell != "" {
		lang = shell
	}

	cmd, args := getInterpreter(lang)
	if cmd == "" {
		return "", fmt.Errorf("unsupported language/shell: %s", lang)
	}

	// Add additional args if specified
	if extraArgs, ok := b.Eval.Params["args"]; ok && extraArgs != "" {
		// Parse quoted arguments
		args = append(args, parseArgs(extraArgs)...)
	}

	// Create command with context for timeout support
	ctx := context.Background()
	var cancel context.CancelFunc

	// Set timeout if specified
	if timeoutStr, ok := b.Eval.Params["timeout"]; ok && timeoutStr != "" {
		timeout, err := time.ParseDuration(timeoutStr)
		if err != nil {
			return "", fmt.Errorf("invalid timeout format: %s", timeoutStr)
		}
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	c := exec.CommandContext(ctx, cmd, args...)
	c.Stdin = strings.NewReader(strings.Join(b.Code, "\n"))

	// Set working directory if specified
	if cwd, ok := b.Eval.Params["cwd"]; ok && cwd != "" {
		c.Dir = cwd
	}

	// Set environment variables if specified
	if envStr, ok := b.Eval.Params["env"]; ok && envStr != "" {
		c.Env = os.Environ() // Start with current environment
		envVars := parseEnvVars(envStr)
		for key, value := range envVars {
			c.Env = append(c.Env, fmt.Sprintf("%s=%s", key, value))
		}
	}

	out, err := c.CombinedOutput()

	// Handle timeout errors more gracefully
	if ctx.Err() == context.DeadlineExceeded {
		return string(out), fmt.Errorf("command timed out")
	}

	return string(out), err
}

// getInterpreter returns the command and args for a given language/shell
func getInterpreter(lang string) (string, []string) {
	switch lang {
	case "python", "python3":
		return "python3", nil
	case "bash", "sh":
		return "bash", nil
	case "node", "javascript":
		return "node", nil
	case "go":
		return "go", []string{"run", "-"}
	default:
		return "", nil
	}
}

// parseArgs parses a space-separated argument string, handling quoted arguments
func parseArgs(argsStr string) []string {
	var args []string
	var current strings.Builder
	inQuote := false
	var quote rune

	for _, r := range argsStr {
		switch {
		case !inQuote && (r == '"' || r == '\''):
			inQuote = true
			quote = r
		case inQuote && r == quote:
			inQuote = false
			quote = 0
		case !inQuote && r == ' ':
			if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(r)
		}
	}

	if current.Len() > 0 {
		args = append(args, current.String())
	}

	return args
}

// parseEnvVars parses comma-separated environment variables of the form KEY=VALUE
func parseEnvVars(envStr string) map[string]string {
	envVars := make(map[string]string)
	pairs := strings.Split(envStr, ",")

	for _, pair := range pairs {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}

		parts := strings.SplitN(pair, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			envVars[key] = value
		}
	}

	return envVars
}
