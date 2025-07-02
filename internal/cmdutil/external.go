package cmdutil

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/johncoder/jot/internal/config"
	"github.com/johncoder/jot/internal/workspace"
)

// ExternalCommand represents a command to be executed
type ExternalCommand struct {
	Name         string            `json:"name"`
	Args         []string          `json:"args"`
	WorkingDir   string            `json:"working_dir,omitempty"`
	Environment  map[string]string `json:"environment,omitempty"`
	Timeout      time.Duration     `json:"timeout,omitempty"`
	Interactive  bool              `json:"interactive"`
	CaptureOutput bool             `json:"capture_output"`
}

// CommandResult represents the result of command execution
type CommandResult struct {
	ExitCode int    `json:"exit_code"`
	Stdout   string `json:"stdout,omitempty"`
	Stderr   string `json:"stderr,omitempty"`
	Error    string `json:"error,omitempty"`
	Duration time.Duration `json:"duration"`
}

// CommandExecutor handles external command execution with unified patterns
type CommandExecutor struct {
	workspace      *workspace.Workspace
	defaultTimeout time.Duration
	baseEnv        map[string]string
}

// NewCommandExecutor creates a new command executor
func NewCommandExecutor(ws *workspace.Workspace, timeout time.Duration) *CommandExecutor {
	return &CommandExecutor{
		workspace:      ws,
		defaultTimeout: timeout,
		baseEnv:        make(map[string]string),
	}
}

// Execute runs a command and returns the result
func (ce *CommandExecutor) Execute(cmd *ExternalCommand) (*CommandResult, error) {
	start := time.Now()
	
	// Build the exec.Cmd
	execCmd := exec.Command(cmd.Name, cmd.Args...)
	
	// Set working directory
	if cmd.WorkingDir != "" {
		execCmd.Dir = cmd.WorkingDir
	} else if ce.workspace != nil {
		execCmd.Dir = ce.workspace.Root
	}
	
	// Build environment
	env := ce.buildEnvironment(cmd.Environment)
	execCmd.Env = env
	
	// Configure I/O based on command type
	var result *CommandResult
	var err error
	
	if cmd.Interactive {
		result, err = ce.executeInteractive(execCmd)
	} else if cmd.CaptureOutput {
		result, err = ce.executeWithCapture(execCmd)
	} else {
		result, err = ce.executeWithInheritedIO(execCmd)
	}
	
	result.Duration = time.Since(start)
	return result, err
}

// executeInteractive runs a command with inherited stdin/stdout/stderr
func (ce *CommandExecutor) executeInteractive(cmd *exec.Cmd) (*CommandResult, error) {
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	err := cmd.Run()
	result := &CommandResult{
		ExitCode: 0,
	}
	
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitError.ExitCode()
		} else {
			result.Error = err.Error()
			result.ExitCode = -1
		}
	}
	
	return result, nil
}

// executeWithCapture runs a command and captures output
func (ce *CommandExecutor) executeWithCapture(cmd *exec.Cmd) (*CommandResult, error) {
	stdout, err := cmd.Output()
	result := &CommandResult{
		ExitCode: 0,
		Stdout:   string(stdout),
	}
	
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitError.ExitCode()
			result.Stderr = string(exitError.Stderr)
		} else {
			result.Error = err.Error()
			result.ExitCode = -1
		}
	}
	
	return result, nil
}

// executeWithInheritedIO runs a command with inherited I/O but no capture
func (ce *CommandExecutor) executeWithInheritedIO(cmd *exec.Cmd) (*CommandResult, error) {
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	err := cmd.Run()
	result := &CommandResult{
		ExitCode: 0,
	}
	
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitError.ExitCode()
		} else {
			result.Error = err.Error()
			result.ExitCode = -1
		}
	}
	
	return result, nil
}

// buildEnvironment constructs the environment for command execution
func (ce *CommandExecutor) buildEnvironment(additional map[string]string) []string {
	env := os.Environ()
	
	// Add base environment
	for key, value := range ce.baseEnv {
		env = append(env, fmt.Sprintf("%s=%s", key, value))
	}
	
	// Add additional environment
	for key, value := range additional {
		env = append(env, fmt.Sprintf("%s=%s", key, value))
	}
	
	return env
}

// SetBaseEnvironment sets the base environment for all commands
func (ce *CommandExecutor) SetBaseEnvironment(env map[string]string) {
	ce.baseEnv = env
}

// EnvironmentBuilder helps build environment variables for different contexts
type EnvironmentBuilder struct {
	workspace  *workspace.Workspace
	jsonOutput bool
	configFile string
}

// NewEnvironmentBuilder creates a new environment builder
func NewEnvironmentBuilder(ws *workspace.Workspace, jsonOutput bool, configFile string) *EnvironmentBuilder {
	return &EnvironmentBuilder{
		workspace:  ws,
		jsonOutput: jsonOutput,
		configFile: configFile,
	}
}

// BuildJotEnvironment builds environment variables for jot external commands
func (eb *EnvironmentBuilder) BuildJotEnvironment(subcommand string) map[string]string {
	env := make(map[string]string)
	
	// Always set command context
	env["JOT_SUBCOMMAND"] = subcommand
	env["JOT_JSON_OUTPUT"] = fmt.Sprintf("%t", eb.jsonOutput)
	
	// Set config file if specified
	if eb.configFile != "" {
		env["JOT_CONFIG_FILE"] = eb.configFile
	}
	
	// Set workspace information if available
	if eb.workspace != nil {
		env["JOT_WORKSPACE_ROOT"] = eb.workspace.Root
		env["JOT_WORKSPACE_NAME"] = workspace.GetNameFromPath(eb.workspace.Root)
		env["JOT_WORKSPACE_INBOX"] = eb.workspace.InboxPath
		env["JOT_WORKSPACE_LIB"] = eb.workspace.LibDir
		env["JOT_WORKSPACE_JOTDIR"] = eb.workspace.JotDir
		env["JOT_DISCOVERY_METHOD"] = workspace.GetDiscoveryMethod(eb.workspace)
	} else {
		env["JOT_DISCOVERY_METHOD"] = "none"
	}
	
	// Add resolved configuration values
	env["JOT_EDITOR"] = config.GetEditor()
	env["JOT_PAGER"] = config.GetPager()
	
	return env
}

// BuildEditorEnvironment builds environment variables for editor commands
func (eb *EnvironmentBuilder) BuildEditorEnvironment() map[string]string {
	env := make(map[string]string)
	
	// Set editor configuration
	env["EDITOR"] = config.GetEditor()
	env["VISUAL"] = config.GetEditor()
	
	// Set workspace context if available
	if eb.workspace != nil {
		env["JOT_WORKSPACE_ROOT"] = eb.workspace.Root
	}
	
	return env
}

// BuildToolEnvironment builds environment variables for specific tools
func (eb *EnvironmentBuilder) BuildToolEnvironment(tool string, options map[string]string) map[string]string {
	env := make(map[string]string)
	
	// Set tool-specific environment based on tool type
	switch strings.ToLower(tool) {
	case "fzf":
		// Set basic FZF options
		env["FZF_DEFAULT_OPTS"] = "--height=40% --layout=reverse --border"
		if eb.workspace != nil {
			env["FZF_DEFAULT_COMMAND"] = fmt.Sprintf("find %s -type f -name '*.md'", eb.workspace.Root)
		}
	case "git":
		if eb.workspace != nil {
			env["GIT_DIR"] = eb.workspace.Root + "/.git"
			env["GIT_WORK_TREE"] = eb.workspace.Root
		}
	}
	
	// Add any additional options
	for key, value := range options {
		env[key] = value
	}
	
	return env
}

// Command builder functions for common external commands

// NewEditorCommand creates a command to open a file in the configured editor
func NewEditorCommand(filePath string, ws *workspace.Workspace) *ExternalCommand {
	editorCmd := config.GetEditor()
	editorParts := strings.Fields(editorCmd)
	
	cmd := &ExternalCommand{
		Name:        editorParts[0],
		Args:        append(editorParts[1:], filePath),
		Interactive: true,
		Timeout:     0, // No timeout for editor
	}
	
	if ws != nil {
		cmd.WorkingDir = ws.Root
		env := NewEnvironmentBuilder(ws, false, "").BuildEditorEnvironment()
		cmd.Environment = env
	}
	
	return cmd
}

// NewFZFCommand creates a command for FZF selection
func NewFZFCommand(input []string, options map[string]string) *ExternalCommand {
	args := []string{}
	
	// Add common FZF options
	for key, value := range options {
		args = append(args, fmt.Sprintf("--%s=%s", key, value))
	}
	
	return &ExternalCommand{
		Name:         "fzf",
		Args:         args,
		Interactive:  true,
		CaptureOutput: false, // FZF handles its own I/O
		Environment:  map[string]string{"FZF_DEFAULT_OPTS": "--height=40% --layout=reverse --border"},
	}
}

// NewExternalJotCommand creates a command for external jot subcommands
func NewExternalJotCommand(subcommand string, args []string, ws *workspace.Workspace, jsonOutput bool, configFile string) *ExternalCommand {
	cmdName := "jot-" + subcommand
	
	env := NewEnvironmentBuilder(ws, jsonOutput, configFile).BuildJotEnvironment(subcommand)
	
	return &ExternalCommand{
		Name:        cmdName,
		Args:        args,
		Interactive: true,
		Environment: env,
		Timeout:     30 * time.Second, // Default timeout for external commands
	}
}

// NewShellCommand creates a command to execute shell scripts
func NewShellCommand(script string, ws *workspace.Workspace) *ExternalCommand {
	cmd := &ExternalCommand{
		Name:        "sh",
		Args:        []string{"-c", script},
		Interactive: false,
		CaptureOutput: true,
		Timeout:     10 * time.Second, // Default timeout for shell commands
	}
	
	if ws != nil {
		cmd.WorkingDir = ws.Root
	}
	
	return cmd
}

// ExecuteWithTimeout executes a command with a timeout context
func (ce *CommandExecutor) ExecuteWithTimeout(cmd *ExternalCommand, timeout time.Duration) (*CommandResult, error) {
	if timeout == 0 {
		timeout = ce.defaultTimeout
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	
	// Build the exec.Cmd
	execCmd := exec.CommandContext(ctx, cmd.Name, cmd.Args...)
	
	// Set working directory
	if cmd.WorkingDir != "" {
		execCmd.Dir = cmd.WorkingDir
	} else if ce.workspace != nil {
		execCmd.Dir = ce.workspace.Root
	}
	
	// Build environment
	env := ce.buildEnvironment(cmd.Environment)
	execCmd.Env = env
	
	start := time.Now()
	
	// Execute based on command type
	var result *CommandResult
	var err error
	
	if cmd.CaptureOutput {
		result, err = ce.executeWithCapture(execCmd)
	} else {
		result, err = ce.executeWithInheritedIO(execCmd)
	}
	
	result.Duration = time.Since(start)
	
	// Check for timeout
	if ctx.Err() == context.DeadlineExceeded {
		result.Error = fmt.Sprintf("command timed out after %v", timeout)
		result.ExitCode = -1
	}
	
	return result, err
}

// LookupCommand checks if a command exists in PATH
func LookupCommand(name string) (string, error) {
	return exec.LookPath(name)
}

// IsCommandAvailable checks if a command is available in PATH
func IsCommandAvailable(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}
