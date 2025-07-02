package cmd

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/johncoder/jot/internal/cmdutil"
	"github.com/johncoder/jot/internal/eval"
	"github.com/johncoder/jot/internal/hooks"
	"github.com/johncoder/jot/internal/workspace"
	"github.com/spf13/cobra"
)

var evalList bool
var evalAll bool
var evalApprove bool
var evalMode string
var evalRevoke bool
var evalListApproved bool
var evalApproveDocument bool
var evalRevokeDocument bool
var evalNoVerify bool

var evalCmd = &cobra.Command{
	Use:   "eval [file] [block_name]",
	Short: "Evaluate code blocks in markdown files",
	Long: `Evaluate code blocks in markdown files using standards-compliant metadata.

EVAL ELEMENT PLACEMENT AND CONFIGURATION:

Eval elements are HTML-style self-closing tags that precede code blocks:

  <eval name="hello" />
  ` + "```" + `python
  print("Hello, world!")
  ` + "```" + `

Core Parameters:
  name="block_name"     Unique identifier for the code block
  shell="python3"       Shell/interpreter (default: inferred from language)
  timeout="30s"         Execution timeout (default: 30s)
  cwd="/tmp"            Working directory for execution
  env="VAR=value"       Environment variables (comma-separated)
  args="--verbose"      Additional arguments to interpreter

Result Parameters:
  results="output"      Capture stdout/stderr (default)
  results="value"       Return function/expression value
  results="code"        Wrap in code block (default)
  results="table"       Format as markdown table
  results="raw"         Insert directly as markdown
  results="replace"     Replace previous results (default)
  results="append"      Add after previous results
  results="silent"      Execute but don't show results

Security:
All eval blocks require explicit approval before execution. Approval is tied
to the block's content hash - changes require re-approval.

Examples:
  jot eval example.md                    # List blocks with approval status
  jot eval example.md hello_python       # Execute specific block (if approved)
  jot eval example.md hello_python --approve --mode hash  # Approve block (doesn't execute)
  jot eval example.md --all              # Execute all approved blocks
  jot eval example.md --approve-document --mode always    # Approve entire document
  jot eval --list-approved               # List all approved blocks`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmdutil.StartCommand(cmd)

		// Handle global operations
		if evalListApproved {
			if ctx.IsJSONOutput() {
				return listApprovedBlocksJSON(ctx)
			}
			return listApprovedBlocks()
		}

		if len(args) == 0 {
			return ctx.HandleError(fmt.Errorf("please specify a markdown file"))
		}

		// Get workspace for file path resolution
		noWorkspace, _ := cmd.Flags().GetBool("no-workspace")
		ws, err := workspace.GetWorkspaceContext(noWorkspace)
		if err != nil {
			return ctx.HandleError(err)
		}

		filename := args[0]
		// Resolve file path relative to workspace or current directory
		resolvedFilename := cmdutil.ResolvePath(ws, filename, noWorkspace)

		// Handle revoke operations
		if evalRevokeDocument {
			if ctx.IsJSONOutput() {
				return revokeDocumentApprovalJSON(ctx, resolvedFilename)
			}
			return revokeDocumentApproval(resolvedFilename)
		}

		if evalRevoke {
			if len(args) < 2 {
				return ctx.HandleError(fmt.Errorf("please specify a block name to revoke"))
			}
			if ctx.IsJSONOutput() {
				return revokeApprovalJSON(ctx, resolvedFilename, args[1])
			}
			return revokeApproval(resolvedFilename, args[1])
		}

		// Handle approve operations
		if evalApproveDocument {
			if ctx.IsJSONOutput() {
				return approveDocumentJSON(ctx, resolvedFilename, evalMode)
			}
			return approveDocument(resolvedFilename, evalMode)
		}

		// If no block name specified, list blocks (unless --all is used)
		if len(args) == 1 && !evalAll {
			if ctx.IsJSONOutput() {
				return listBlocksJSON(ctx, resolvedFilename)
			}
			return listBlocks(resolvedFilename)
		}

		var blockName string
		if len(args) > 1 {
			blockName = args[1]
		}

		// Handle approval workflow
		if evalApprove {
			if blockName == "" {
				return ctx.HandleError(fmt.Errorf("please specify a block name to approve"))
			}
			if ctx.IsJSONOutput() {
				return approveBlockJSON(ctx, resolvedFilename, blockName, evalMode)
			}
			return approveBlock(resolvedFilename, blockName, evalMode)
		}

		// Execute blocks
		var results []*eval.EvalResult

		// Initialize hook manager and run pre-eval hook
		if ws != nil && !evalNoVerify {
			hookManager := hooks.NewManager(ws)
			hookCtx := &hooks.HookContext{
				Type:        hooks.PreEval,
				Workspace:   ws,
				SourceFile:  resolvedFilename,
				Timeout:     30 * time.Second,
				AllowBypass: evalNoVerify,
			}

			result, err := hookManager.Execute(hookCtx)
			if err != nil {
				return ctx.HandleExternalCommand("pre-eval hook", nil, err)
			}

			if result.Aborted {
				return ctx.HandleOperationError("pre-eval hook", fmt.Errorf("pre-eval hook aborted operation"))
			}
		}

		if blockName != "" {
			// Execute specific block by name
			results, err = eval.ExecuteEvaluableBlockByName(resolvedFilename, blockName)
		} else if evalAll {
			// Execute all blocks
			results, err = eval.ExecuteEvaluableBlocks(resolvedFilename)
		} else {
			return ctx.HandleError(fmt.Errorf("please specify a block name or use --all to execute all blocks"))
		}

		if err != nil {
			return ctx.HandleOperationError("execute blocks", fmt.Errorf("error executing blocks in %s: %w", filename, err))
		}

		// Run post-eval hook (informational only)
		if ws != nil && !evalNoVerify {
			hookManager := hooks.NewManager(ws)
			hookCtx := &hooks.HookContext{
				Type:        hooks.PostEval,
				Workspace:   ws,
				SourceFile:  resolvedFilename,
				Timeout:     30 * time.Second,
				AllowBypass: evalNoVerify,
			}

			_, hookErr := hookManager.Execute(hookCtx)
			if hookErr != nil && !ctx.IsJSONOutput() {
				cmdutil.ShowWarning("Warning: post-eval hook failed: %s", hookErr.Error())
			}
		}

		// Handle JSON output for execution results
		if ctx.IsJSONOutput() {
			return outputExecutionResultsJSON(ctx, filename, blockName, results)
		}

		// Human-readable output for execution results
		// Check for approval errors
		hasApprovalErrors := false
		for _, result := range results {
			if result.Err != nil && strings.Contains(result.Err.Error(), "requires approval") {
				hasApprovalErrors = true
				fmt.Printf("⚠ %s\n", result.Err.Error())
			}
		}

		if hasApprovalErrors {
			fmt.Printf("\nTo approve blocks, use: jot eval %s <block_name> --approve --mode <hash|prompt|always>\n", filename)
		}

		// Update results in markdown
		err = eval.UpdateMarkdownWithResults(resolvedFilename, results)
		if err != nil {
			return fmt.Errorf("error updating results in %s: %w", filename, err)
		}

		// Report success
		executed := 0
		for _, result := range results {
			if result.Err == nil {
				executed++
			}
		}

		if blockName != "" {
			if executed > 0 {
				cmdutil.ShowSuccess("✓ Executed block '%s' in %s", blockName, filename)
			}
		} else if evalAll {
			cmdutil.ShowSuccess("✓ Executed %d approved blocks in %s", executed, filename)
		}

		return nil
	},
}

func listBlocks(filename string) error {
	fmt.Printf("Blocks in %s:\n", filename)
	return eval.ListEvalBlocks(filename)
}

func listApprovedBlocks() error {
	sm, err := eval.NewSecurityManager()
	if err != nil {
		return fmt.Errorf("failed to initialize security manager: %w", err)
	}

	approvals := sm.ListApprovals()
	docApprovals := sm.ListDocumentApprovals()

	if len(approvals) == 0 && len(docApprovals) == 0 {
		fmt.Println("No approved blocks or documents found.")
		return nil
	}

	if len(docApprovals) > 0 {
		fmt.Println("Approved documents:")
		for _, approval := range docApprovals {
			fmt.Printf("  ✓ %s (%s mode)\n",
				approval.FilePath, approval.Mode)
		}
		fmt.Println()
	}

	if len(approvals) > 0 {
		fmt.Println("Approved individual blocks:")
		for _, approval := range approvals {
			fmt.Printf("  ✓ %s:%s (%s mode)\n",
				approval.FilePath, approval.BlockName, approval.Mode)
		}
	}

	return nil
}

func revokeApproval(filename, blockName string) error {
	sm, err := eval.NewSecurityManager()
	if err != nil {
		return fmt.Errorf("failed to initialize security manager: %w", err)
	}

	absPath, err := filepath.Abs(filename)
	if err != nil {
		return err
	}

	err = sm.RevokeApproval(absPath, blockName)
	if err != nil {
		return fmt.Errorf("failed to revoke approval: %w", err)
	}

	cmdutil.ShowSuccess("✓ Revoked approval for block '%s' in %s", blockName, filename)
	return nil
}

func approveBlock(filename, blockName, mode string) error {
	// Parse and validate mode
	var approvalMode eval.ApprovalMode
	switch mode {
	case "hash", "":
		approvalMode = eval.ApprovalModeHash
	case "prompt":
		approvalMode = eval.ApprovalModePrompt
	case "always":
		approvalMode = eval.ApprovalModeAlways
	default:
		return fmt.Errorf("invalid approval mode: %s (must be hash, prompt, or always)", mode)
	}

	// Find the block
	blocks, err := eval.ParseMarkdownForEvalBlocks(filename)
	if err != nil {
		return err
	}

	var targetBlock *eval.CodeBlock
	for _, block := range blocks {
		if block.Eval != nil && block.Eval.Params["name"] == blockName {
			targetBlock = block
			break
		}
	}

	if targetBlock == nil {
		return fmt.Errorf("no block named '%s' found in %s", blockName, filename)
	}

	// Show the code block for approval
	fmt.Printf("Approving code block '%s':\n", blockName)
	fmt.Println("────────────────────────────────────────")
	for _, line := range targetBlock.Code {
		fmt.Println(line)
	}
	fmt.Println("────────────────────────────────────────")

	// Confirm approval
	confirmed, err := cmdutil.ConfirmOperation(fmt.Sprintf("Approve this block with %s mode?", approvalMode))
	if err != nil {
		return err
	}

	if !confirmed {
		cmdutil.ShowInfo("Approval cancelled.")
		return nil
	}

	// Approve the block
	sm, err := eval.NewSecurityManager()
	if err != nil {
		return fmt.Errorf("failed to initialize security manager: %w", err)
	}

	absPath, err := filepath.Abs(filename)
	if err != nil {
		return err
	}

	err = sm.ApproveBlock(absPath, targetBlock, approvalMode)
	if err != nil {
		return fmt.Errorf("failed to approve block: %w", err)
	}

	cmdutil.ShowSuccess("✓ Block '%s' approved with %s mode.", blockName, approvalMode)
	cmdutil.ShowInfo("Use 'jot eval' to execute the approved block.")

	return nil
}

func approveDocument(filename, mode string) error {
	// Parse and validate mode
	var approvalMode eval.ApprovalMode
	switch mode {
	case "hash", "":
		approvalMode = eval.ApprovalModeHash
	case "prompt":
		approvalMode = eval.ApprovalModePrompt
	case "always":
		approvalMode = eval.ApprovalModeAlways
	default:
		return fmt.Errorf("invalid approval mode: %s (must be hash, prompt, or always)", mode)
	}

	// Get all blocks in the document
	blocks, err := eval.ParseMarkdownForEvalBlocks(filename)
	if err != nil {
		return err
	}

	// Count evaluable blocks
	evalBlocks := 0
	for _, block := range blocks {
		if block.Eval != nil && block.Eval.Params["name"] != "" {
			evalBlocks++
		}
	}

	if evalBlocks == 0 {
		return fmt.Errorf("no evaluable blocks found in %s", filename)
	}

	// Show the blocks for approval
	fmt.Printf("Approving entire document '%s' (%d blocks):\n", filename, evalBlocks)
	fmt.Println("────────────────────────────────────────")
	for _, block := range blocks {
		if block.Eval != nil && block.Eval.Params["name"] != "" {
			blockName := block.Eval.Params["name"]
			fmt.Printf("Block: %s (lines %d-%d) %s\n", blockName, block.StartLine, block.EndLine, block.Lang)
		}
	}
	fmt.Println("────────────────────────────────────────")

	// Confirm approval
	confirmed, err := cmdutil.ConfirmOperation(fmt.Sprintf("Approve entire document with %s mode?", approvalMode))
	if err != nil {
		return err
	}

	if !confirmed {
		cmdutil.ShowInfo("Document approval cancelled.")
		return nil
	}

	// Approve the document
	sm, err := eval.NewSecurityManager()
	if err != nil {
		return fmt.Errorf("failed to initialize security manager: %w", err)
	}

	absPath, err := filepath.Abs(filename)
	if err != nil {
		return err
	}

	err = sm.ApproveDocument(absPath, approvalMode)
	if err != nil {
		return fmt.Errorf("failed to approve document: %w", err)
	}

	fmt.Printf("✓ Document '%s' approved with %s mode (%d blocks).\n", filename, approvalMode, evalBlocks)
	return nil
}

func revokeDocumentApproval(filename string) error {
	sm, err := eval.NewSecurityManager()
	if err != nil {
		return fmt.Errorf("failed to initialize security manager: %w", err)
	}

	absPath, err := filepath.Abs(filename)
	if err != nil {
		return err
	}

	err = sm.RevokeDocumentApproval(absPath)
	if err != nil {
		return fmt.Errorf("failed to revoke document approval: %w", err)
	}

	fmt.Printf("✓ Revoked document approval for %s\n", filename)
	return nil
}

// JSON response structures for eval command
type EvalResponse struct {
	Operation string               `json:"operation"`
	Results   []EvalResult         `json:"results,omitempty"`
	Blocks    []EvalBlock          `json:"blocks,omitempty"`
	Approvals []EvalApproval       `json:"approvals,omitempty"`
	Summary   EvalSummary          `json:"summary"`
	Metadata  cmdutil.JSONMetadata `json:"metadata"`
}

type EvalResult struct {
	BlockName string `json:"block_name"`
	Language  string `json:"language"`
	Code      string `json:"code"`
	Output    string `json:"output,omitempty"`
	Error     string `json:"error,omitempty"`
	Success   bool   `json:"success"`
	StartLine int    `json:"start_line"`
	EndLine   int    `json:"end_line"`
}

type EvalBlock struct {
	Name         string `json:"name"`
	Language     string `json:"language"`
	StartLine    int    `json:"start_line"`
	EndLine      int    `json:"end_line"`
	IsApproved   bool   `json:"is_approved"`
	ApprovalMode string `json:"approval_mode,omitempty"`
}

type EvalApproval struct {
	Type      string `json:"type"` // "block" or "document"
	FilePath  string `json:"file_path"`
	BlockName string `json:"block_name,omitempty"`
	Mode      string `json:"mode"`
}

type EvalSummary struct {
	TotalBlocks    int `json:"total_blocks"`
	ExecutedBlocks int `json:"executed_blocks"`
	FailedBlocks   int `json:"failed_blocks"`
	ApprovedBlocks int `json:"approved_blocks,omitempty"`
}

func init() {
	evalCmd.Flags().BoolVarP(&evalList, "list", "l", false, "List evaluable code blocks (deprecated, use without arguments)")
	evalCmd.Flags().BoolVarP(&evalAll, "all", "a", false, "Execute all approved evaluable code blocks")
	evalCmd.Flags().BoolVar(&evalApprove, "approve", false, "Approve and execute the specified block")
	evalCmd.Flags().StringVar(&evalMode, "mode", "hash", "Approval mode: hash, prompt, or always")
	evalCmd.Flags().BoolVar(&evalRevoke, "revoke", false, "Revoke approval for the specified block")
	evalCmd.Flags().BoolVar(&evalListApproved, "list-approved", false, "List all approved blocks")
	evalCmd.Flags().BoolVar(&evalApproveDocument, "approve-document", false, "Approve the entire document")
	evalCmd.Flags().BoolVar(&evalRevokeDocument, "revoke-document", false, "Revoke document approval")
	evalCmd.Flags().Bool("no-workspace", false, "Resolve file paths relative to current directory instead of workspace")
	evalCmd.Flags().BoolVar(&evalNoVerify, "no-verify", false, "Skip hooks verification")
}

// JSON output functions for eval command

// listBlocksJSON outputs JSON response for listing blocks
func listBlocksJSON(ctx *cmdutil.CommandContext, filename string) error {
	blocks, err := eval.ParseMarkdownForEvalBlocks(filename)
	if err != nil {
		return ctx.HandleError(fmt.Errorf("error parsing %s: %w", filename, err))
	}

	// Get security manager to check approvals
	sm, err := eval.NewSecurityManager()
	if err != nil {
		return ctx.HandleError(fmt.Errorf("failed to initialize security manager: %w", err))
	}

	absPath, err := filepath.Abs(filename)
	if err != nil {
		return ctx.HandleError(err)
	}

	var evalBlocks []EvalBlock
	approvedCount := 0

	for _, block := range blocks {
		if block.Eval != nil && block.Eval.Params["name"] != "" {
			blockName := block.Eval.Params["name"]

			// Check if block is approved
			isApproved, err := sm.CheckApproval(absPath, block)
			if err != nil {
				// If error checking approval, assume not approved
				isApproved = false
			}

			var approvalMode string
			if isApproved {
				approvedCount++
				// Try to find the approval record to get the mode
				approvals := sm.ListApprovals()
				key := filepath.Base(absPath) + ":" + blockName
				for _, approval := range approvals {
					approvalKey := filepath.Base(approval.FilePath) + ":" + approval.BlockName
					if approvalKey == key {
						approvalMode = string(approval.Mode)
						break
					}
				}
			}

			evalBlocks = append(evalBlocks, EvalBlock{
				Name:         blockName,
				Language:     block.Lang,
				StartLine:    block.StartLine,
				EndLine:      block.EndLine,
				IsApproved:   isApproved,
				ApprovalMode: approvalMode,
			})
		}
	}

	response := EvalResponse{
		Operation: "list_blocks",
		Blocks:    evalBlocks,
		Summary: EvalSummary{
			TotalBlocks:    len(evalBlocks),
			ApprovedBlocks: approvedCount,
		},
		Metadata: cmdutil.CreateJSONMetadata(ctx.Cmd, true, ctx.StartTime),
	}

	return outputJSON(response)
}

// listApprovedBlocksJSON outputs JSON response for listing approved blocks
func listApprovedBlocksJSON(ctx *cmdutil.CommandContext) error {
	sm, err := eval.NewSecurityManager()
	if err != nil {
		return ctx.HandleError(fmt.Errorf("failed to initialize security manager: %w", err))
	}

	approvals := sm.ListApprovals()
	docApprovals := sm.ListDocumentApprovals()

	var evalApprovals []EvalApproval

	// Add document approvals
	for _, approval := range docApprovals {
		evalApprovals = append(evalApprovals, EvalApproval{
			Type:     "document",
			FilePath: approval.FilePath,
			Mode:     string(approval.Mode),
		})
	}

	// Add individual block approvals
	for _, approval := range approvals {
		evalApprovals = append(evalApprovals, EvalApproval{
			Type:      "block",
			FilePath:  approval.FilePath,
			BlockName: approval.BlockName,
			Mode:      string(approval.Mode),
		})
	}

	response := EvalResponse{
		Operation: "list_approved",
		Approvals: evalApprovals,
		Summary: EvalSummary{
			ApprovedBlocks: len(evalApprovals),
		},
		Metadata: cmdutil.CreateJSONMetadata(ctx.Cmd, true, ctx.StartTime),
	}

	return outputJSON(response)
}

// outputExecutionResultsJSON outputs JSON response for execution results
func outputExecutionResultsJSON(ctx *cmdutil.CommandContext, filename, blockName string, results []*eval.EvalResult) error {
	var evalResults []EvalResult
	executed := 0
	failed := 0

	for _, result := range results {
		var output, errorMsg string
		var blockName, language, code string
		var startLine, endLine int
		success := result.Err == nil

		if result.Block != nil {
			if result.Block.Eval != nil && result.Block.Eval.Params["name"] != "" {
				blockName = result.Block.Eval.Params["name"]
			}
			language = result.Block.Lang
			code = strings.Join(result.Block.Code, "\n")
			startLine = result.Block.StartLine
			endLine = result.Block.EndLine
		}

		if result.Err != nil {
			errorMsg = result.Err.Error()
			failed++
		} else {
			executed++
		}

		if result.Output != "" {
			output = result.Output
		}

		evalResults = append(evalResults, EvalResult{
			BlockName: blockName,
			Language:  language,
			Code:      code,
			Output:    output,
			Error:     errorMsg,
			Success:   success,
			StartLine: startLine,
			EndLine:   endLine,
		})
	}

	operation := "execute_all"
	if blockName != "" {
		operation = "execute_block"
	}

	response := EvalResponse{
		Operation: operation,
		Results:   evalResults,
		Summary: EvalSummary{
			TotalBlocks:    len(evalResults),
			ExecutedBlocks: executed,
			FailedBlocks:   failed,
		},
		Metadata: cmdutil.CreateJSONMetadata(ctx.Cmd, true, ctx.StartTime),
	}

	return outputJSON(response)
}

// approveBlockJSON outputs JSON response for block approval (non-interactive for JSON)
func approveBlockJSON(ctx *cmdutil.CommandContext, filename, blockName, mode string) error {
	// For JSON output, we cannot do interactive approval, so we return an error
	err := fmt.Errorf("interactive approval not supported in JSON mode - use non-JSON mode for approval operations")
	return ctx.HandleError(err)
}

// approveDocumentJSON outputs JSON response for document approval (non-interactive for JSON)
func approveDocumentJSON(ctx *cmdutil.CommandContext, filename, mode string) error {
	// For JSON output, we cannot do interactive approval, so we return an error
	err := fmt.Errorf("interactive approval not supported in JSON mode - use non-JSON mode for approval operations")
	return ctx.HandleError(err)
}

// revokeApprovalJSON outputs JSON response for revoking block approval
func revokeApprovalJSON(ctx *cmdutil.CommandContext, filename, blockName string) error {
	sm, err := eval.NewSecurityManager()
	if err != nil {
		return ctx.HandleError(fmt.Errorf("failed to initialize security manager: %w", err))
	}

	absPath, err := filepath.Abs(filename)
	if err != nil {
		return ctx.HandleError(err)
	}

	err = sm.RevokeApproval(absPath, blockName)
	if err != nil {
		return ctx.HandleError(fmt.Errorf("failed to revoke approval: %w", err))
	}

	response := EvalResponse{
		Operation: "revoke_approval",
		Summary: EvalSummary{
			TotalBlocks: 1,
		},
		Metadata: cmdutil.CreateJSONMetadata(ctx.Cmd, true, ctx.StartTime),
	}

	return outputJSON(response)
}

// revokeDocumentApprovalJSON outputs JSON response for revoking document approval
func revokeDocumentApprovalJSON(ctx *cmdutil.CommandContext, filename string) error {
	sm, err := eval.NewSecurityManager()
	if err != nil {
		return ctx.HandleError(fmt.Errorf("failed to initialize security manager: %w", err))
	}

	absPath, err := filepath.Abs(filename)
	if err != nil {
		return ctx.HandleError(err)
	}

	err = sm.RevokeDocumentApproval(absPath)
	if err != nil {
		return ctx.HandleError(fmt.Errorf("failed to revoke document approval: %w", err))
	}

	response := EvalResponse{
		Operation: "revoke_document_approval",
		Summary: EvalSummary{
			TotalBlocks: 1,
		},
		Metadata: cmdutil.CreateJSONMetadata(ctx.Cmd, true, ctx.StartTime),
	}

	return outputJSON(response)
}
