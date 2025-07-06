package eval

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/johncoder/jot/internal/cmdutil"
	"github.com/johncoder/jot/internal/workspace"
)

// EvaluatorInfo represents information about an available evaluator
type EvaluatorInfo struct {
	Language string
	Type     string // "built-in" or "path"
	Command  string
	Path     string // Full path for PATH evaluators
}

// EvaluatorManager handles evaluator discovery and execution
type EvaluatorManager struct {
	// Cache of discovered evaluators
	cache map[string]*EvaluatorInfo
	// Command executor for running evaluators
	executor *cmdutil.CommandExecutor
	// Workspace context
	workspace *workspace.Workspace
}

// NewEvaluatorManager creates a new evaluator manager
func NewEvaluatorManager() *EvaluatorManager {
	return &EvaluatorManager{
		cache: make(map[string]*EvaluatorInfo),
	}
}

// NewEvaluatorManagerWithWorkspace creates a new evaluator manager with workspace context
func NewEvaluatorManagerWithWorkspace(ws *workspace.Workspace) *EvaluatorManager {
	executor := cmdutil.NewCommandExecutor(ws, 30*time.Second) // Default timeout
	return &EvaluatorManager{
		cache:     make(map[string]*EvaluatorInfo),
		executor:  executor,
		workspace: ws,
	}
}

// DiscoverEvaluator finds an evaluator for the given language
func (m *EvaluatorManager) DiscoverEvaluator(lang string) (*EvaluatorInfo, error) {
	// Check cache first
	if info, ok := m.cache[lang]; ok {
		return info, nil
	}

	// Try PATH evaluator first
	pathEvaluator := fmt.Sprintf("jot-eval-%s", lang)
	if path, err := exec.LookPath(pathEvaluator); err == nil {
		info := &EvaluatorInfo{
			Language: lang,
			Type:     "path",
			Command:  pathEvaluator,
			Path:     path,
		}
		m.cache[lang] = info
		return info, nil
	}

	// Try built-in evaluator
	if m.isBuiltinEvaluator(lang) {
		info := &EvaluatorInfo{
			Language: lang,
			Type:     "built-in",
			Command:  fmt.Sprintf("jot evaluator %s", lang),
			Path:     "", // Built-in evaluators don't have a path
		}
		m.cache[lang] = info
		return info, nil
	}

	// No evaluator found
	return nil, m.GetEvaluatorError(lang)
}

// isBuiltinEvaluator checks if a language has a built-in evaluator
func (m *EvaluatorManager) isBuiltinEvaluator(lang string) bool {
	switch lang {
	case "python", "python3":
		return true
	case "bash", "sh":
		return true
	case "javascript", "node":
		return true
	case "go":
		return true
	default:
		return false
	}
}

// ListEvaluators returns all available evaluators
func (m *EvaluatorManager) ListEvaluators() ([]*EvaluatorInfo, error) {
	var evaluators []*EvaluatorInfo

	// Add built-in evaluators
	builtins := []string{"python3", "javascript", "bash", "go"}
	for _, lang := range builtins {
		evaluators = append(evaluators, &EvaluatorInfo{
			Language: lang,
			Type:     "built-in",
			Command:  fmt.Sprintf("jot evaluator %s", lang),
			Path:     "",
		})
	}

	// Discover PATH evaluators
	pathEvaluators, err := m.discoverPathEvaluators()
	if err != nil {
		return evaluators, err // Return built-ins even if PATH discovery fails
	}

	evaluators = append(evaluators, pathEvaluators...)
	return evaluators, nil
}

// discoverPathEvaluators finds all jot-eval-* commands in PATH
func (m *EvaluatorManager) discoverPathEvaluators() ([]*EvaluatorInfo, error) {
	var evaluators []*EvaluatorInfo

	// Get all directories in PATH
	pathEnv := os.Getenv("PATH")
	if pathEnv == "" {
		return evaluators, nil
	}

	pathSeparator := ":"
	if os.PathSeparator == '\\' { // Windows
		pathSeparator = ";"
	}

	dirs := strings.Split(pathEnv, pathSeparator)

	// Track seen evaluators to avoid duplicates
	seen := make(map[string]bool)

	for _, dir := range dirs {
		if dir == "" {
			continue
		}

		// Read directory contents
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue // Skip directories we can't read
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			name := entry.Name()
			if !strings.HasPrefix(name, "jot-eval-") {
				continue
			}

			// Extract language from filename
			lang := strings.TrimPrefix(name, "jot-eval-")
			if lang == "" {
				continue
			}

			// Skip if we've already seen this language
			if seen[lang] {
				continue
			}

			// Check if file is executable
			fullPath := filepath.Join(dir, name)
			if info, err := os.Stat(fullPath); err == nil {
				// On Unix-like systems, check if executable
				if info.Mode()&0111 != 0 {
					evaluators = append(evaluators, &EvaluatorInfo{
						Language: lang,
						Type:     "path",
						Command:  name,
						Path:     fullPath,
					})
					seen[lang] = true
				}
			}
		}
	}

	return evaluators, nil
}

// ExecuteWithEvaluator executes code using the discovered evaluator
func (m *EvaluatorManager) ExecuteWithEvaluator(lang string, code string, params map[string]string, workingDir string) (string, error) {
	evaluator, err := m.DiscoverEvaluator(lang)
	if err != nil {
		return "", err
	}

	switch evaluator.Type {
	case "path":
		return m.executePathEvaluator(evaluator, code, params, workingDir)
	case "built-in":
		return m.executeBuiltinEvaluator(lang, code, params, workingDir)
	default:
		return "", fmt.Errorf("unknown evaluator type: %s", evaluator.Type)
	}
}

// executePathEvaluator executes a PATH-based evaluator using CommandExecutor
func (m *EvaluatorManager) executePathEvaluator(evaluator *EvaluatorInfo, code string, params map[string]string, workingDir string) (string, error) {
	// If no executor available, fall back to direct execution
	if m.executor == nil {
		return m.executePathEvaluatorDirect(evaluator, code, params, workingDir)
	}

	// Create external command
	cmd := &cmdutil.ExternalCommand{
		Name:          evaluator.Path,
		Args:          []string{},
		WorkingDir:    workingDir,
		Environment:   m.buildEnvironmentMap(evaluator.Language, code, params, workingDir),
		CaptureOutput: true,
		Interactive:   false,
	}

	// Set timeout if specified
	if timeoutStr, ok := params["timeout"]; ok && timeoutStr != "" {
		timeout, err := time.ParseDuration(timeoutStr)
		if err != nil {
			return "", cmdutil.NewValidationError("timeout", timeoutStr, err)
		}
		cmd.Timeout = timeout
	}

	// Execute command
	result, err := m.executor.Execute(cmd)
	if err != nil {
		return "", cmdutil.NewExternalError(evaluator.Command, []string{}, err)
	}

	// Check for execution errors
	if result.ExitCode != 0 {
		errorMsg := fmt.Sprintf("evaluator exited with code %d", result.ExitCode)
		if result.Stderr != "" {
			errorMsg += ": " + result.Stderr
		}
		return result.Stdout, fmt.Errorf("%s", errorMsg)
	}

	return result.Stdout, nil
}

// executeBuiltinEvaluator executes a built-in evaluator
func (m *EvaluatorManager) executeBuiltinEvaluator(lang string, code string, params map[string]string, workingDir string) (string, error) {
	// Get the interpreter command and args
	cmd, args := m.getBuiltinInterpreter(lang)
	if cmd == "" {
		return "", fmt.Errorf("unsupported built-in language: %s", lang)
	}

	// Add additional args if specified
	if extraArgs, ok := params["args"]; ok && extraArgs != "" {
		args = append(args, parseArgs(extraArgs)...)
	}

	// Create context for timeout
	ctx := context.Background()
	var cancel context.CancelFunc

	// Set timeout if specified
	if timeoutStr, ok := params["timeout"]; ok && timeoutStr != "" {
		timeout, err := time.ParseDuration(timeoutStr)
		if err != nil {
			return "", cmdutil.NewValidationError("timeout", timeoutStr, err)
		}
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	// Create command
	c := exec.CommandContext(ctx, cmd, args...)
	c.Stdin = strings.NewReader(code)

	// Set working directory
	if workingDir != "" {
		c.Dir = workingDir
	}

	// Set environment variables if specified
	if envStr, ok := params["env"]; ok && envStr != "" {
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

	if err != nil {
		return string(out), cmdutil.NewExternalError(cmd, args, err)
	}

	return string(out), nil
}

// getBuiltinInterpreter returns the command and args for built-in evaluators
func (m *EvaluatorManager) getBuiltinInterpreter(lang string) (string, []string) {
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

// buildEnvironment builds the environment variables for PATH evaluators
func (m *EvaluatorManager) buildEnvironment(lang string, code string, params map[string]string, workingDir string) []string {
	env := os.Environ()

	// Add standard jot context (these would be set by the calling code)
	// JOT_WORKSPACE_PATH, JOT_WORKSPACE_NAME, JOT_CONFIG_FILE are set by caller

	// Add eval-specific context
	env = append(env, fmt.Sprintf("JOT_EVAL_CODE=%s", code))
	env = append(env, fmt.Sprintf("JOT_EVAL_LANG=%s", lang))

	if workingDir != "" {
		env = append(env, fmt.Sprintf("JOT_EVAL_CWD=%s", workingDir))
	}

	// Add parameters as environment variables
	for key, value := range params {
		switch key {
		case "timeout":
			env = append(env, fmt.Sprintf("JOT_EVAL_TIMEOUT=%s", value))
		case "args":
			env = append(env, fmt.Sprintf("JOT_EVAL_ARGS=%s", value))
		case "name":
			env = append(env, fmt.Sprintf("JOT_EVAL_BLOCK_NAME=%s", value))
		case "env":
			// Parse and add custom environment variables
			envVars := parseEnvVars(value)
			for k, v := range envVars {
				env = append(env, fmt.Sprintf("JOT_EVAL_ENV_%s=%s", k, v))
			}
		}
	}

	return env
}

// GetEvaluatorError returns a helpful error message when no evaluator is found
func (m *EvaluatorManager) GetEvaluatorError(lang string) error {
	pathEvaluator := fmt.Sprintf("jot-eval-%s", lang)

	// Capitalize first letter for display
	displayLang := lang
	if len(lang) > 0 {
		displayLang = strings.ToUpper(lang[:1]) + lang[1:]
	}

	err := fmt.Errorf("no evaluator found for '%s'", lang)
	helpMsg := fmt.Sprintf(`
Tried:
  1. PATH evaluator: %s (not found)
  2. Built-in evaluator: jot evaluator %s (not available)

To add %s support:
  - Create %s script in PATH
  - Check available evaluators: jot evaluator`,
		pathEvaluator, lang, displayLang, pathEvaluator)

	return fmt.Errorf("%w%s", err, helpMsg)
}

// buildEnvironmentMap builds environment variables as a map for CommandExecutor
func (m *EvaluatorManager) buildEnvironmentMap(lang string, code string, params map[string]string, workingDir string) map[string]string {
	env := make(map[string]string)

	// Add eval-specific context
	env["JOT_EVAL_CODE"] = code
	env["JOT_EVAL_LANG"] = lang

	if workingDir != "" {
		env["JOT_EVAL_CWD"] = workingDir
	}

	// Add parameters as environment variables
	for key, value := range params {
		switch key {
		case "timeout":
			env["JOT_EVAL_TIMEOUT"] = value
		case "args":
			env["JOT_EVAL_ARGS"] = value
		case "name":
			env["JOT_EVAL_BLOCK_NAME"] = value
		case "env":
			// Parse and add custom environment variables
			envVars := parseEnvVars(value)
			for k, v := range envVars {
				env[fmt.Sprintf("JOT_EVAL_ENV_%s", k)] = v
			}
		}
	}

	return env
}

// executePathEvaluatorDirect executes a PATH-based evaluator directly (fallback)
func (m *EvaluatorManager) executePathEvaluatorDirect(evaluator *EvaluatorInfo, code string, params map[string]string, workingDir string) (string, error) {
	// Create context for timeout
	ctx := context.Background()
	var cancel context.CancelFunc

	// Set timeout if specified
	if timeoutStr, ok := params["timeout"]; ok && timeoutStr != "" {
		timeout, err := time.ParseDuration(timeoutStr)
		if err != nil {
			return "", cmdutil.NewValidationError("timeout", timeoutStr, err)
		}
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	// Create command
	cmd := exec.CommandContext(ctx, evaluator.Path)

	// Set working directory
	if workingDir != "" {
		cmd.Dir = workingDir
	}

	// Set environment variables
	cmd.Env = m.buildEnvironment(evaluator.Language, code, params, workingDir)

	// Execute
	out, err := cmd.CombinedOutput()

	// Handle timeout errors more gracefully
	if ctx.Err() == context.DeadlineExceeded {
		return string(out), fmt.Errorf("evaluator timed out")
	}

	if err != nil {
		return string(out), cmdutil.NewExternalError(evaluator.Command, []string{}, err)
	}

	return string(out), nil
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
