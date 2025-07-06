package cmd

import (
	"fmt"
	"os"

	"github.com/johncoder/jot/internal/cmdutil"
	"github.com/johncoder/jot/internal/eval"
	"github.com/johncoder/jot/internal/workspace"
	"github.com/spf13/cobra"
)

var evaluatorCmd = &cobra.Command{
	Use:   "evaluator [language]",
	Short: "Manage and run code evaluators",
	Long: `Manage and run code evaluators for the jot eval system.

This command serves two main purposes:
1. Run built-in evaluators for specific languages
2. List available evaluators in the system

The evaluator system uses a convention-based approach:
- PATH evaluators: jot-eval-<language> executables in PATH
- Built-in evaluators: jot evaluator <language> for supported languages

Built-in language support:
- python3     (python3)
- javascript  (node)
- bash        (bash)
- go          (go run)

Examples:
  jot evaluator                    # List all available evaluators
  jot evaluator python3            # Run built-in Python evaluator
  jot evaluator --json             # List evaluators in JSON format`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmdutil.StartCommand(cmd)

		// If no language specified, list available evaluators
		if len(args) == 0 {
			if ctx.IsJSONOutput() {
				return listEvaluatorsJSON(ctx)
			}
			return listEvaluators()
		}

		// Execute built-in evaluator for specified language
		language := args[0]
		return runBuiltinEvaluator(ctx, language)
	},
}

func init() {
	// No additional flags needed for now
}

func listEvaluators() error {
	manager := eval.NewEvaluatorManager()
	evaluators, err := manager.ListEvaluators()
	if err != nil {
		return fmt.Errorf("failed to list evaluators: %w", err)
	}

	// Separate built-in and PATH evaluators
	var builtins []*eval.EvaluatorInfo
	var pathEvaluators []*eval.EvaluatorInfo

	for _, evaluator := range evaluators {
		switch evaluator.Type {
		case "built-in":
			builtins = append(builtins, evaluator)
		case "path":
			pathEvaluators = append(pathEvaluators, evaluator)
		}
	}

	// Display built-in evaluators
	if len(builtins) > 0 {
		fmt.Println("Built-in evaluators:")
		for _, evaluator := range builtins {
			interpreter := getInterpreterName(evaluator.Language)
			fmt.Printf("  %-12s (%s)\n", evaluator.Language, interpreter)
		}
	}

	// Display PATH evaluators
	if len(pathEvaluators) > 0 {
		if len(builtins) > 0 {
			fmt.Println()
		}
		fmt.Println("PATH evaluators:")
		for _, evaluator := range pathEvaluators {
			fmt.Printf("  %-12s (%s)\n", evaluator.Language, evaluator.Command)
		}
	}

	if len(evaluators) == 0 {
		fmt.Println("No evaluators found.")
		fmt.Println("\nTo add custom evaluators, create executable scripts named 'jot-eval-<language>' in your PATH.")
	}

	return nil
}

func listEvaluatorsJSON(ctx *cmdutil.CommandContext) error {
	manager := eval.NewEvaluatorManager()
	evaluators, err := manager.ListEvaluators()
	if err != nil {
		return ctx.HandleOperationError("list evaluators", err)
	}

	// Build JSON response
	response := map[string]interface{}{
		"operation": "list_evaluators",
		"evaluators": map[string]interface{}{
			"built_in": []map[string]interface{}{},
			"path":     []map[string]interface{}{},
		},
		"metadata": cmdutil.CreateJSONMetadata(ctx.Cmd, true, ctx.StartTime),
	}

	// Separate evaluators by type
	for _, evaluator := range evaluators {
		evalInfo := map[string]interface{}{
			"language": evaluator.Language,
			"command":  evaluator.Command,
		}

		if evaluator.Type == "path" {
			evalInfo["path"] = evaluator.Path
		}

		switch evaluator.Type {
		case "built-in":
			evalInfo["interpreter"] = getInterpreterName(evaluator.Language)
			response["evaluators"].(map[string]interface{})["built_in"] = append(
				response["evaluators"].(map[string]interface{})["built_in"].([]map[string]interface{}),
				evalInfo,
			)
		case "path":
			response["evaluators"].(map[string]interface{})["path"] = append(
				response["evaluators"].(map[string]interface{})["path"].([]map[string]interface{}),
				evalInfo,
			)
		}
	}

	return cmdutil.OutputJSON(response)
}

func runBuiltinEvaluator(ctx *cmdutil.CommandContext, language string) error {
	// Try to get workspace context for enhanced features
	var manager *eval.EvaluatorManager
	if ws, err := workspace.GetWorkspaceContext(false); err == nil && ws != nil {
		manager = eval.NewEvaluatorManagerWithWorkspace(ws)
	} else {
		manager = eval.NewEvaluatorManager()
	}

	// Check if this is a built-in evaluator
	if !isBuiltinLanguage(language) {
		if ctx.IsJSONOutput() {
			return ctx.HandleOperationError("run evaluator", fmt.Errorf("'%s' is not a built-in evaluator", language))
		}
		return fmt.Errorf("'%s' is not a built-in evaluator.\n\nBuilt-in evaluators: python3, javascript, bash, go", language)
	}

	// Get code from environment variable
	code := os.Getenv("JOT_EVAL_CODE")
	if code == "" {
		if ctx.IsJSONOutput() {
			return ctx.HandleOperationError("run evaluator", fmt.Errorf("no code provided (JOT_EVAL_CODE environment variable not set)"))
		}
		return fmt.Errorf("no code provided, set JOT_EVAL_CODE environment variable")
	}

	// Get working directory
	workingDir := os.Getenv("JOT_EVAL_CWD")
	if workingDir == "" {
		// Try to get workspace context
		ws, err := workspace.GetWorkspaceContext(false)
		if err == nil && ws != nil {
			workingDir = ws.Root
		} else {
			// Fall back to current directory
			if wd, err := os.Getwd(); err == nil {
				workingDir = wd
			}
		}
	}

	// Build parameters from environment variables
	params := make(map[string]string)
	if timeout := os.Getenv("JOT_EVAL_TIMEOUT"); timeout != "" {
		params["timeout"] = timeout
	}
	if args := os.Getenv("JOT_EVAL_ARGS"); args != "" {
		params["args"] = args
	}
	if env := os.Getenv("JOT_EVAL_ENV"); env != "" {
		params["env"] = env
	}
	if name := os.Getenv("JOT_EVAL_BLOCK_NAME"); name != "" {
		params["name"] = name
	}

	// Execute the code
	output, err := manager.ExecuteWithEvaluator(language, code, params, workingDir)

	if ctx.IsJSONOutput() {
		response := map[string]interface{}{
			"operation": "run_evaluator",
			"language":  language,
			"success":   err == nil,
			"output":    output,
			"metadata":  cmdutil.CreateJSONMetadata(ctx.Cmd, err == nil, ctx.StartTime),
		}

		if err != nil {
			response["error"] = err.Error()
		}

		return cmdutil.OutputJSON(response)
	}

	// Human-readable output
	if err != nil {
		return fmt.Errorf("evaluator failed: %w", err)
	}

	// Print output directly (evaluator output should go to stdout)
	fmt.Print(output)
	return nil
}

func getInterpreterName(language string) string {
	switch language {
	case "python3":
		return "python3"
	case "javascript":
		return "node"
	case "bash":
		return "bash"
	case "go":
		return "go run"
	default:
		return language
	}
}

func isBuiltinLanguage(language string) bool {
	switch language {
	case "python3", "javascript", "bash", "go":
		return true
	default:
		return false
	}
}
