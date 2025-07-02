package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/johncoder/jot/internal/cmdutil"
	"github.com/johncoder/jot/internal/hooks"
	"github.com/johncoder/jot/internal/workspace"
	"github.com/spf13/cobra"
)

var hooksCmd = &cobra.Command{
	Use:   "hooks",
	Short: "Manage workspace hooks",
	Long: `Manage git-style hooks for workspace automation.

Hooks are executable scripts in .jot/hooks/ that run at specific points
in jot operations. Similar to git hooks, they can modify content, trigger
notifications, or integrate with external tools.

Available hook types:
  pre-capture    - Called before capturing content (can modify content)
  post-capture   - Called after capturing content (notifications, logging)
  pre-refile     - Called before refiling content (can modify content)
  post-refile    - Called after refiling content (cleanup, notifications)
  pre-archive    - Called before archiving content
  post-archive   - Called after archiving content
  workspace-change - Called when switching workspaces

Examples:
  jot hooks list                    # List all hooks in workspace
  jot hooks install-samples         # Install sample hook scripts
  jot hooks test pre-capture        # Test a specific hook type`,
}

var hooksListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all hooks in the workspace",
	Long:  `List all executable hooks found in the workspace .jot/hooks/ directory.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmdutil.StartCommand(cmd)

		ws, err := workspace.RequireWorkspace()
		if err != nil {
			return ctx.HandleError(err)
		}

		return listHooks(ctx, ws)
	},
}

var hooksInstallSamplesCmd = &cobra.Command{
	Use:   "install-samples",
	Short: "Install sample hook scripts",
	Long: `Install sample hook scripts in the workspace .jot/hooks/ directory.

Sample hooks demonstrate common use cases and can be customized for your workflow.
They are created with .sample extension and must be renamed to be active.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmdutil.StartCommand(cmd)

		ws, err := workspace.RequireWorkspace()
		if err != nil {
			return ctx.HandleError(err)
		}

		return installSampleHooks(ctx, ws)
	},
}

var hooksTestCmd = &cobra.Command{
	Use:   "test <hook-type>",
	Short: "Test a specific hook type",
	Long: `Test a specific hook type with sample data to verify it works correctly.

This runs the hook with test data and shows the output, allowing you to
debug and verify your hook scripts work as expected.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmdutil.StartCommand(cmd)

		ws, err := workspace.RequireWorkspace()
		if err != nil {
			return ctx.HandleError(err)
		}

		return testHook(ctx, ws, args[0])
	},
}

// listHooks lists all hooks in the workspace
func listHooks(ctx *cmdutil.CommandContext, ws *workspace.Workspace) error {
	hooksDir := filepath.Join(ws.JotDir, "hooks")

	// Check if hooks directory exists
	if _, err := os.Stat(hooksDir); os.IsNotExist(err) {
		if ctx.IsJSONOutput() {
			response := HooksResponse{
				Operation: "list",
				HooksDir:  hooksDir,
				Hooks:     []HookInfo{},
				Summary: HooksSummary{
					TotalHooks:      0,
					ExecutableHooks: 0,
					SampleHooks:     0,
				},
				Metadata: cmdutil.CreateJSONMetadata(ctx.Cmd, true, ctx.StartTime),
			}
			return outputJSON(response)
		}

		fmt.Printf("No hooks directory found at %s\n", hooksDir)
		fmt.Println("Run 'jot hooks install-samples' to get started with sample hooks.")
		return nil
	}

	// Read hooks directory
	entries, err := os.ReadDir(hooksDir)
	if err != nil {
		if ctx.IsJSONOutput() {
			return ctx.HandleError(err)
		}
		return fmt.Errorf("failed to read hooks directory: %w", err)
	}

	var hooksList []HookInfo
	var executableCount, sampleCount int

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		path := filepath.Join(hooksDir, name)

		info, err := entry.Info()
		if err != nil {
			continue
		}

		isSample := strings.HasSuffix(name, ".sample")
		isExecutable := info.Mode()&0111 != 0

		hookInfo := HookInfo{
			Name:       name,
			Path:       path,
			Executable: isExecutable,
			Sample:     isSample,
			Size:       info.Size(),
			ModTime:    info.ModTime(),
		}

		if isSample {
			sampleCount++
		} else if isExecutable {
			executableCount++
		}

		hooksList = append(hooksList, hookInfo)
	}

	if ctx.IsJSONOutput() {
		response := HooksResponse{
			Operation: "list",
			HooksDir:  hooksDir,
			Hooks:     hooksList,
			Summary: HooksSummary{
				TotalHooks:      len(hooksList),
				ExecutableHooks: executableCount,
				SampleHooks:     sampleCount,
			},
			Metadata: cmdutil.CreateJSONMetadata(ctx.Cmd, true, ctx.StartTime),
		}
		return outputJSON(response)
	}

	// Human-readable output
	if len(hooksList) == 0 {
		fmt.Printf("No hooks found in %s\n", hooksDir)
		fmt.Println("Run 'jot hooks install-samples' to get started with sample hooks.")
		return nil
	}

	fmt.Printf("Hooks in %s:\n\n", hooksDir)

	for _, hook := range hooksList {
		status := "inactive"
		if hook.Executable && !hook.Sample {
			status = "active"
		} else if hook.Sample {
			status = "sample"
		}

		fmt.Printf("  %-20s %s\n", hook.Name, status)
	}

	fmt.Printf("\nSummary: %d total, %d active, %d samples\n",
		len(hooksList), executableCount, sampleCount)

	if sampleCount > 0 {
		fmt.Println("\nTo activate a sample hook:")
		fmt.Println("  1. Copy the .sample file: cp pre-capture.sample pre-capture")
		fmt.Println("  2. Make it executable: chmod +x pre-capture")
		fmt.Println("  3. Customize as needed")
	}

	return nil
}

// installSampleHooks installs sample hook scripts
func installSampleHooks(ctx *cmdutil.CommandContext, ws *workspace.Workspace) error {
	manager := hooks.NewManager(ws)

	if err := manager.CreateSampleHooks(); err != nil {
		if ctx.IsJSONOutput() {
			return ctx.HandleError(err)
		}
		return fmt.Errorf("failed to install sample hooks: %w", err)
	}

	hooksDir := filepath.Join(ws.JotDir, "hooks")

	if ctx.IsJSONOutput() {
		response := HooksResponse{
			Operation: "install-samples",
			HooksDir:  hooksDir,
			Summary: HooksSummary{
				SampleHooks: 4, // Number of sample hooks we create
			},
			Metadata: cmdutil.CreateJSONMetadata(ctx.Cmd, true, ctx.StartTime),
		}
		return outputJSON(response)
	}

	fmt.Printf("✓ Sample hooks installed in %s\n", hooksDir)
	fmt.Println("\nAvailable samples:")
	fmt.Println("  pre-capture.sample  - Modify content before capture")
	fmt.Println("  post-capture.sample - Notifications after capture")
	fmt.Println("  pre-refile.sample   - Modify content before refile")
	fmt.Println("  post-refile.sample  - Logging after refile")
	fmt.Println()
	fmt.Println("To activate a hook:")
	fmt.Println("  1. Copy: cp pre-capture.sample pre-capture")
	fmt.Println("  2. Make executable: chmod +x pre-capture")
	fmt.Println("  3. Customize as needed")

	return nil
}

// testHook tests a specific hook type with sample data
func testHook(ctx *cmdutil.CommandContext, ws *workspace.Workspace, hookType string) error {
	// Validate hook type
	validTypes := []string{
		"pre-capture", "post-capture", "pre-refile", "post-refile",
		"pre-archive", "post-archive", "workspace-change",
	}

	valid := false
	for _, validType := range validTypes {
		if hookType == validType {
			valid = true
			break
		}
	}

	if !valid {
		err := fmt.Errorf("invalid hook type '%s'. Valid types: %s",
			hookType, strings.Join(validTypes, ", "))
		if ctx.IsJSONOutput() {
			return ctx.HandleError(err)
		}
		return err
	}

	// Check if hook exists
	hooksDir := filepath.Join(ws.JotDir, "hooks")
	hookPath := filepath.Join(hooksDir, hookType)

	if _, err := os.Stat(hookPath); os.IsNotExist(err) {
		err := fmt.Errorf("hook '%s' not found at %s", hookType, hookPath)
		if ctx.IsJSONOutput() {
			return ctx.HandleError(err)
		}
		return err
	}

	// Create sample context based on hook type
	var testContent string
	var extraEnv map[string]string

	switch hookType {
	case "pre-capture", "post-capture":
		testContent = "# Test Capture\n\nThis is test content for the capture hook."
		extraEnv = map[string]string{
			"JOT_TEMPLATE_NAME": "test-template",
		}
	case "pre-refile", "post-refile":
		testContent = "# Test Refile\n\nThis content is being refiled for testing."
		extraEnv = map[string]string{
			"JOT_REFILE_SOURCE": "inbox.md#test-section",
			"JOT_REFILE_DEST":   "work.md#test-project",
		}
	case "pre-archive", "post-archive":
		testContent = "# Test Archive\n\nThis content is being archived for testing."
		extraEnv = map[string]string{
			"JOT_ARCHIVE_SOURCE":   "work.md#old-project",
			"JOT_ARCHIVE_LOCATION": "archive/archive.md#Archive",
		}
	case "workspace-change":
		extraEnv = map[string]string{
			"JOT_OLD_WORKSPACE": "/old/workspace/path",
			"JOT_NEW_WORKSPACE": ws.Root,
		}
	}

	manager := hooks.NewManager(ws)
	hookCtx := &hooks.HookContext{
		Type:        hooks.HookType(hookType),
		Workspace:   ws,
		Content:     testContent,
		Timeout:     30 * time.Second,
		AllowBypass: false,
		ExtraEnv:    extraEnv,
	}

	if ctx.IsJSONOutput() {
		result, err := manager.Execute(hookCtx)
		response := HookTestResponse{
			Operation: "test",
			HookType:  hookType,
			HookPath:  hookPath,
			TestData: HookTestData{
				Content:  testContent,
				ExtraEnv: extraEnv,
			},
			Metadata: cmdutil.CreateJSONMetadata(ctx.Cmd, true, ctx.StartTime),
		}

		if err != nil {
			response.Success = false
			response.Error = err.Error()
		} else {
			response.Success = true
			response.Result = HookTestResult{
				ExitCode: result.ExitCode,
				Output:   result.Output,
				Content:  result.Content,
				Aborted:  result.Aborted,
			}
		}

		return outputJSON(response)
	}

	// Human-readable output
	fmt.Printf("Testing hook: %s\n", hookType)
	fmt.Printf("Hook path: %s\n", hookPath)
	fmt.Printf("Test content: %s\n", testContent)
	fmt.Println()

	result, err := manager.Execute(hookCtx)
	if err != nil {
		fmt.Printf("❌ Hook failed to execute: %v\n", err)
		return nil
	}

	if result.ExitCode == 0 {
		fmt.Printf("✅ Hook executed successfully\n")
	} else {
		fmt.Printf("❌ Hook failed with exit code %d\n", result.ExitCode)
	}

	if result.Output != "" {
		fmt.Printf("Hook output:\n%s\n", result.Output)
	}

	if result.Content != testContent && result.Content != "" {
		fmt.Printf("Modified content:\n%s\n", result.Content)
	}

	return nil
}

// JSON response structures
type HooksResponse struct {
	Operation string               `json:"operation"`
	HooksDir  string               `json:"hooks_dir"`
	Hooks     []HookInfo           `json:"hooks"`
	Summary   HooksSummary         `json:"summary"`
	Metadata  cmdutil.JSONMetadata `json:"metadata"`
}

type HookInfo struct {
	Name       string    `json:"name"`
	Path       string    `json:"path"`
	Executable bool      `json:"executable"`
	Sample     bool      `json:"sample"`
	Size       int64     `json:"size"`
	ModTime    time.Time `json:"mod_time"`
}

type HooksSummary struct {
	TotalHooks      int `json:"total_hooks"`
	ExecutableHooks int `json:"executable_hooks"`
	SampleHooks     int `json:"sample_hooks"`
}

type HookTestResponse struct {
	Operation string               `json:"operation"`
	HookType  string               `json:"hook_type"`
	HookPath  string               `json:"hook_path"`
	Success   bool                 `json:"success"`
	TestData  HookTestData         `json:"test_data"`
	Result    HookTestResult       `json:"result,omitempty"`
	Error     string               `json:"error,omitempty"`
	Metadata  cmdutil.JSONMetadata `json:"metadata"`
}

type HookTestData struct {
	Content  string            `json:"content"`
	ExtraEnv map[string]string `json:"extra_env"`
}

type HookTestResult struct {
	ExitCode int    `json:"exit_code"`
	Output   string `json:"output"`
	Content  string `json:"content"`
	Aborted  bool   `json:"aborted"`
}

func init() {
	// Add subcommands
	hooksCmd.AddCommand(hooksListCmd)
	hooksCmd.AddCommand(hooksInstallSamplesCmd)
	hooksCmd.AddCommand(hooksTestCmd)
}
