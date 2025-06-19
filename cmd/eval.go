package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/johncoder/jot/internal/eval"
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

var evalCmd = &cobra.Command{
	Use:   "eval [file] [block_name]",
	Short: "Evaluate code blocks in markdown files",
	Long: `Evaluate code blocks in markdown files using standards-compliant metadata.

Examples:
  jot eval example.md                    # List blocks with approval status
  jot eval example.md hello_python       # Execute specific block (if approved)
  jot eval example.md hello_python --approve --mode hash  # Approve block (doesn't execute)
  jot eval example.md --all              # Execute all approved blocks
  jot eval example.md --approve-document --mode always    # Approve entire document
  jot eval --list-approved               # List all approved blocks`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Handle global operations
		if evalListApproved {
			return listApprovedBlocks()
		}
		
		if len(args) == 0 {
			return fmt.Errorf("please specify a markdown file")
		}
		
		filename := args[0]
		
		// Handle revoke operations
		if evalRevokeDocument {
			return revokeDocumentApproval(filename)
		}
		
		if evalRevoke {
			if len(args) < 2 {
				return fmt.Errorf("please specify a block name to revoke")
			}
			return revokeApproval(filename, args[1])
		}
		
		// Handle approve operations
		if evalApproveDocument {
			return approveDocument(filename, evalMode)
		}
		
		// If no block name specified, list blocks (unless --all is used)
		if len(args) == 1 && !evalAll {
			return listBlocks(filename)
		}
		
		var blockName string
		if len(args) > 1 {
			blockName = args[1]
		}
		
		// Handle approval workflow
		if evalApprove {
			if blockName == "" {
				return fmt.Errorf("please specify a block name to approve")
			}
			return approveBlock(filename, blockName, evalMode)
		}
		
		// Execute blocks
		var results []*eval.EvalResult
		var err error
		
		if blockName != "" {
			// Execute specific block by name
			results, err = eval.ExecuteEvaluableBlockByName(filename, blockName)
		} else if evalAll {
			// Execute all blocks
			results, err = eval.ExecuteEvaluableBlocks(filename)
		} else {
			return fmt.Errorf("please specify a block name or use --all to execute all blocks")
		}
		
		if err != nil {
			return fmt.Errorf("error executing blocks in %s: %w", filename, err)
		}
		
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
		err = eval.UpdateMarkdownWithResults(filename, results)
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
				fmt.Printf("✓ Executed block '%s' in %s\n", blockName, filename)
			}
		} else if evalAll {
			fmt.Printf("✓ Executed %d approved blocks in %s\n", executed, filename)
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
	
	fmt.Printf("✓ Revoked approval for block '%s' in %s\n", blockName, filename)
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
	fmt.Printf("Approve this block with %s mode? [y/N]: ", approvalMode)
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	
	response = strings.TrimSpace(strings.ToLower(response))
	if response != "y" && response != "yes" {
		fmt.Println("Approval cancelled.")
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
	
	fmt.Printf("✓ Block '%s' approved with %s mode.\n", blockName, approvalMode)
	fmt.Println("Use 'jot eval' to execute the approved block.")
	
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
	fmt.Printf("Approve entire document with %s mode? [y/N]: ", approvalMode)
	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	
	response = strings.TrimSpace(strings.ToLower(response))
	if response != "y" && response != "yes" {
		fmt.Println("Document approval cancelled.")
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

func init() {
	rootCmd.AddCommand(evalCmd)
	evalCmd.Flags().BoolVarP(&evalList, "list", "l", false, "List evaluable code blocks (deprecated, use without arguments)")
	evalCmd.Flags().BoolVarP(&evalAll, "all", "a", false, "Execute all approved evaluable code blocks")
	evalCmd.Flags().BoolVar(&evalApprove, "approve", false, "Approve and execute the specified block")
	evalCmd.Flags().StringVar(&evalMode, "mode", "hash", "Approval mode: hash, prompt, or always")
	evalCmd.Flags().BoolVar(&evalRevoke, "revoke", false, "Revoke approval for the specified block")
	evalCmd.Flags().BoolVar(&evalListApproved, "list-approved", false, "List all approved blocks")
	evalCmd.Flags().BoolVar(&evalApproveDocument, "approve-document", false, "Approve the entire document")
	evalCmd.Flags().BoolVar(&evalRevokeDocument, "revoke-document", false, "Revoke document approval")
}
