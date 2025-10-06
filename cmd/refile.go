// jot/cmd/refile.go: CLI orchestration for the 'refile' command.
// All helpers, types, and logic are modularized in refile_pipeline.go, refile_output.go, refile_types.go, refile_interactive.go, and refile_selectors.go.
// This file should only contain CLI wiring, flag parsing, and orchestration logic.

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/johncoder/jot/internal/cmdutil"
	"github.com/johncoder/jot/internal/fzf"
	"github.com/johncoder/jot/internal/hooks"
	md "github.com/johncoder/jot/internal/markdown"
	"github.com/johncoder/jot/internal/workspace"
	"github.com/spf13/cobra"
)


var refileInteractive bool
var refileNoVerify bool

var refileCmd = &cobra.Command{
	Use:   "refile [SOURCE] --to DESTINATION",
	Short: "Move markdown subtrees between files using path-based selectors",
	Long: `Move entire markdown subtrees (headings with all nested content) between files.

Path-based selector syntax with contains matching:
- Each segment uses case-insensitive contains matching
- Must match exactly one subtree
- Leading slashes handle unusual document structures

Examples:
  jot refile "inbox.md#meeting" --to "work.md#projects"
  jot refile "notes.md#research/database" --to "archive.md#technical"  
  jot refile "inbox.md#/foo/bar" --to "work.md#tasks"  # Skip level 1
  jot refile --to "work.md#projects/frontend"          # Inspect destination`,

	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmdutil.StartCommand(cmd)

		ws, err := workspace.RequireWorkspace()
		if err != nil {
			return ctx.HandleError(err)
		}

		// Get flags
		to, _ := cmd.Flags().GetString("to")
		prepend, _ := cmd.Flags().GetBool("prepend")
		verbose, _ := cmd.Flags().GetBool("verbose")
	   // Check for interactive mode
	   if fzf.ShouldUseFZF(refileInteractive) {
		   return RunInteractiveRefile(ctx, args, ws)
	   }

		// No source and no destination: show usage help
		if len(args) == 0 && to == "" {
			err := fmt.Errorf("provide a source file or --to destination")
			return ctx.HandleError(err)
		}

		if to == "" {
			// Check if this is a request to show selectors for a specific file
			if len(args) == 1 && !strings.Contains(args[0], "#") {
				return ShowSelectorsForFile(ws, args[0])
			}
			err := fmt.Errorf("destination path required: use --to flag")
			if ctx.IsJSONOutput() {
				return ctx.HandleError(err)
			}
			return err
		}

		// Parse destination path
		destPath, err := md.ParsePath(to)
		if err != nil {
			err := cmdutil.NewValidationError("destination path", to, err)
			if ctx.IsJSONOutput() {
				return ctx.HandleError(err)
			}
			return err
		}

		// Source-less mode: inspect destination
		if len(args) == 0 {
			if ctx.IsJSONOutput() {
				return InspectDestinationJSON(ctx, ws, destPath)
			}
			return inspectDestination(ws, destPath)
		}

		// Parse source path
		sourcePath, err := md.ParsePath(args[0])
		if err != nil {
			err := cmdutil.NewValidationError("source path", args[0], err)
			if ctx.IsJSONOutput() {
				return ctx.HandleError(err)
			}
			return err
		}

		// Extract subtree from source (now in refile_pipeline.go)
		subtree, err := ExtractSubtree(ws, sourcePath)
		if err != nil {
			err := fmt.Errorf("failed to extract subtree: %w", err)
			if ctx.IsJSONOutput() {
				return ctx.HandleError(err)
			}
			return err
		}

		if verbose && !ctx.IsJSONOutput() {
			PrintVerboseSubtreeInfo(subtree, sourcePath.File)
		}

		// Resolve destination (now in refile_pipeline.go)
		dest, err := ResolveDestination(ws, destPath, prepend)
		if err != nil {
			err := fmt.Errorf("failed to resolve destination: %w", err)
			if ctx.IsJSONOutput() {
				return ctx.HandleError(err)
			}
			return err
		}

		if verbose && !ctx.IsJSONOutput() {
			PrintVerboseDestinationInfo(dest)
		}

		// Transform subtree level (now in refile_pipeline.go)
		transformedContent := TransformSubtreeLevel(subtree, dest.TargetLevel)

		// Run pre-refile hook
		hookManager := hooks.NewManager(ws)
		if !refileNoVerify {
			hookCtx := &hooks.HookContext{
				Type:        hooks.PreRefile,
				Workspace:   ws,
				SourceFile:  args[0],
				DestPath:    to,
				Timeout:     30 * time.Second,
				AllowBypass: refileNoVerify,
			}

			result, err := hookManager.Execute(hookCtx)
			if err != nil {
				err := cmdutil.NewExternalError("pre-refile hook", nil, err)
				if ctx.IsJSONOutput() {
					return ctx.HandleError(err)
				}
				return err
			}

			if result.Aborted {
				err := fmt.Errorf("pre-refile hook aborted operation")
				if ctx.IsJSONOutput() {
					return ctx.HandleError(err)
				}
				return err
			}
		}

		// Perform the refile operation (now in refile_pipeline.go)
		if err := PerformRefile(ws, sourcePath, subtree, dest, transformedContent); err != nil {
			err := fmt.Errorf("refile operation failed: %w", err)
			if ctx.IsJSONOutput() {
				return ctx.HandleError(err)
			}
			return err
		}

		// Run post-refile hook (informational only)
		if !refileNoVerify {
			hookCtx := &hooks.HookContext{
				Type:        hooks.PostRefile,
				Workspace:   ws,
				SourceFile:  args[0],
				DestPath:    to,
				Timeout:     30 * time.Second,
				AllowBypass: refileNoVerify,
			}

			_, hookErr := hookManager.Execute(hookCtx)
			if hookErr != nil && !ctx.IsJSONOutput() {
				fmt.Printf("Warning: post-refile hook failed: %s\n", hookErr.Error())
			}
		}

		// Handle JSON output (now in refile_output.go)
		if ctx.IsJSONOutput() {
			return OutputRefileJSON(ctx, sourcePath, destPath, subtree, dest, transformedContent)
		}

		// Human-readable output
		if verbose {
			fmt.Printf("Refile operation completed successfully!\n")
		}

		fmt.Printf("Successfully refiled '%s' to '%s'\n",
			subtree.Heading, destPath.File+"#"+strings.Join(destPath.Segments, "/"))

		return nil
	},
}

// inspectDestination analyzes destination path without performing refile

func init() {
   refileCmd.Flags().BoolVar(&refileInteractive, "interactive", false, "Launch interactive FZF-based refile UI")
   // Register other flags here as needed
}
func inspectDestination(ws *workspace.Workspace, destPath *md.HeadingPath) error {
	pathUtil := cmdutil.NewPathUtil(ws)
	fmt.Printf("Destination analysis for \"%s#%s\":\n",
		destPath.File, strings.Join(destPath.Segments, "/"))

	// Check if file exists
	var filePath string
	if destPath.File == "inbox.md" {
		filePath = ws.InboxPath
	} else if filepath.IsAbs(destPath.File) {
		filePath = destPath.File
	} else {
		// Use workspace root for relative paths, not lib/ directory
		filePath = pathUtil.WorkspaceJoin(destPath.File)
	}

	if _, err := os.Stat(filePath); err != nil {
		if cmdutil.IsFileNotFound(err) {
			cmdutil.ShowError("✗ File not found: %s", destPath.File)
			return nil
		}
		cmdutil.ShowError("✗ Error accessing file: %s", err.Error())
		return nil
	}
	cmdutil.ShowSuccess("✓ File exists: %s", destPath.File)

	// Read and parse the file to analyze the path
	content, err := os.ReadFile(filePath)
	if err != nil {
		// Use structured error inspection for better error handling
		if fileErr, ok := cmdutil.GetFileError(err); ok {
			cmdutil.ShowError("✗ Error reading file %s: %s", fileErr.Path, fileErr.Err.Error())
		} else {
			cmdutil.ShowError("✗ Error reading file: %s", err.Error())
		}
		return nil
	}

	doc := md.ParseMarkdownAST(content)
	// Update NavigateHeadingPath call to use only doc and path, handle both return values.
	res, err := md.NavigateHeadingPath(doc, destPath)
	if err != nil {
		return err
	}

	if res.TargetHeading != nil {
		// Complete path exists
		fmt.Printf("✓ Path exists: %s\n", strings.Join(destPath.Segments, " > "))
		targetLevel := res.TargetHeading.Level + 1
		fmt.Printf("Ready to receive content at level %d\n", targetLevel)
	} else if len(res.FoundSegments) > 0 {
		// Partial path exists
		cmdutil.ShowSuccess("✓ Partial path exists: %s", strings.Join(res.FoundSegments, " > "))
		cmdutil.ShowError("✗ Missing path: %s", strings.Join(res.MissingSegments, " > "))

		// Show what would be created
		baseLevel := res.ParentHeading.Level + 1
		for i, heading := range res.MissingSegments {
			level := baseLevel + i
			fmt.Printf("Would create: %s %s (level %d)\n",
				strings.Repeat("#", level), heading, level)
		}
		finalLevel := baseLevel + len(res.MissingSegments)
		fmt.Printf("Ready to receive content at level %d\n", finalLevel)
	} else {
		// No path exists
		fmt.Printf("✗ Missing path: %s\n", strings.Join(destPath.Segments, " > "))

		// Show what would be created
		baseLevel := destPath.SkipLevels + 1
		for i, heading := range destPath.Segments {
			level := baseLevel + i
			fmt.Printf("Would create: %s %s (level %d)\n",
				strings.Repeat("#", level), heading, level)
		}
		finalLevel := baseLevel + len(destPath.Segments)
		fmt.Printf("Ready to receive content at level %d\n", finalLevel)
	}

	return nil
}
