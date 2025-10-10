package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/johncoder/jot/internal/cmdutil"
	"github.com/johncoder/jot/internal/markdown"
	"github.com/johncoder/jot/internal/workspace"
	"github.com/spf13/cobra"
)

var peekCmd = &cobra.Command{
	Use:   "peek [SELECTOR]",
	Short: "View markdown file or subtree without opening an editor",
	Long: `View markdown content or subtree using path-based selectors.

Supports:
- Whole file: "filename.md"
- Subtree: "file.md#path/to/heading"
- Table of contents: --toc
- Info: --info
- Raw output: --raw
- Short selectors: --short (with --toc)

Selector syntax:
- Case-insensitive contains matching
- Must match exactly one subtree
- Leading slashes skip level 1

Examples:
  jot peek "inbox.md"
  jot peek "work.md#projects/frontend"
  jot peek "work.md" --toc
  jot peek "work.md#projects" --info
  jot peek "work.md" --toc --short
  jot peek "work.md:42"
`,
	Args: cobra.MaximumNArgs(1),
	RunE: runPeek,
}

func init() {
	peekCmd.Flags().BoolP("raw", "r", false, "Output raw content without formatting")
	peekCmd.Flags().BoolP("info", "i", false, "Show subtree metadata information")
	peekCmd.Flags().BoolP("toc", "t", false, "Show table of contents for file or subtree")
	peekCmd.Flags().BoolP("short", "s", false, "Generate shortest possible selectors (use with --toc)")
	peekCmd.Flags().Bool("no-workspace", false, "Resolve file paths relative to current directory")
	peekCmd.Flags().Bool("json", false, "Output result as JSON")
}

func runPeek(cmd *cobra.Command, args []string) error {
	ctx := cmdutil.StartCommand(cmd)
	noWorkspace, _ := cmd.Flags().GetBool("no-workspace")
	jsonOut, _ := cmd.Flags().GetBool("json")
	raw, _ := cmd.Flags().GetBool("raw")
	info, _ := cmd.Flags().GetBool("info")
	toc, _ := cmd.Flags().GetBool("toc")
	short, _ := cmd.Flags().GetBool("short")

	ws, err := workspace.GetWorkspaceContext(noWorkspace)
	if err != nil {
		return ctx.HandleError(err)
	}

	if toc {
		if len(args) == 0 {
			return ctx.HandleError(fmt.Errorf("table of contents requires a file or selector (e.g., 'inbox.md' or 'work.md#projects')"))
		}
		selector := args[0]
		return peekTOC(ctx, ws, selector, short, jsonOut, noWorkspace)
	}

	if len(args) == 0 {
		return ctx.HandleError(fmt.Errorf("peek requires a selector argument (e.g., 'inbox.md' or 'work.md#heading')"))
	}
	selector := args[0]

	// Enhanced selector: file:42 or file:42#heading
	selector = resolveEnhancedSelector(ws, selector)

	if !strings.Contains(selector, "#") {
		return peekWholeFile(ctx, ws, selector, raw, info, jsonOut, noWorkspace)
	}
	return peekSubtree(ctx, ws, selector, raw, info, jsonOut, noWorkspace)
}

func peekWholeFile(ctx *cmdutil.CommandContext, ws *workspace.Workspace, filename string, raw, info, jsonOut, noWorkspace bool) error {
	filePath := resolvePeekFilePath(ws, filename, noWorkspace)
	content, err := os.ReadFile(filePath)
	if err != nil {
		return ctx.HandleError(fmt.Errorf("file not found: %s", filename))
	}
	if jsonOut {
		return cmdutil.OutputJSON(map[string]interface{}{
			"selector":  filename,
			"file_path": filePath,
			"content":   string(content),
			"metadata":  cmdutil.CreateJSONMetadata(ctx.Cmd, true, ctx.StartTime),
		})
	}
	if info {
		fmt.Printf("File: %s\nPath: %s\nSize: %d bytes\nLines: %d\n\n", filename, filePath, len(content), strings.Count(string(content), "\n")+1)
	}
	if raw {
		os.Stdout.Write(content)
	} else {
		fmt.Printf("# File: %s\n\n%s\n", filename, string(content))
	}
	return nil
}

func peekSubtree(ctx *cmdutil.CommandContext, ws *workspace.Workspace, selector string, raw, info, jsonOut, noWorkspace bool) error {
	sourcePath, err := markdown.ParsePath(selector)
	if err != nil {
		return ctx.HandleError(fmt.Errorf("invalid selector: %w", err))
	}
	filePath := resolvePeekFilePath(ws, sourcePath.File, noWorkspace)
	content, err := os.ReadFile(filePath)
	if err != nil {
		return ctx.HandleError(fmt.Errorf("file not found: %s", sourcePath.File))
	}
	doc := markdown.ParseMarkdownAST(content)
	subtree, matchErr := markdown.ExtractSubtreeByPath(doc, sourcePath.Segments)
	if matchErr != nil {
		return ctx.HandleError(matchErr)
	}
	subtreeContent := markdown.RenderSubtreeContent(subtree)
	heading := markdown.GetHeadingText(subtree)
	level := markdown.GetHeadingLevel(subtree)
	childCount := markdown.CountChildHeadings(subtree)
	if jsonOut {
		return cmdutil.OutputJSON(map[string]interface{}{
			"selector":  selector,
			"file_path": filePath,
			"subtree": map[string]interface{}{
				"heading":        heading,
				"level":          level,
				"content_size":   len(subtreeContent),
				"child_headings": childCount,
				"content":        subtreeContent,
			},
			"metadata": cmdutil.CreateJSONMetadata(ctx.Cmd, true, ctx.StartTime),
		})
	}
	if info {
		fmt.Printf("Subtree Information:\n  File: %s\n  Heading: %s\n  Level: %d\n  Content size: %d bytes\n  Child headings: %d\n\n", sourcePath.File, heading, level, len(subtreeContent), childCount)
	}
	if raw {
		os.Stdout.Write([]byte(subtreeContent))
	} else {
		fmt.Print(subtreeContent)
	}
	return nil
}

func peekTOC(ctx *cmdutil.CommandContext, ws *workspace.Workspace, selector string, short, jsonOut, noWorkspace bool) error {
	file, _ := parseSelector(selector)
	filePath := resolvePeekFilePath(ws, file, noWorkspace)
	content, err := os.ReadFile(filePath)
	if err != nil {
		return ctx.HandleError(fmt.Errorf("file not found: %s", file))
	}
	doc := markdown.ParseMarkdownAST(content)
	headings := markdown.FindAllHeadings(doc)
	selectors := markdown.GenerateSelectors(file, headings, short)
	if jsonOut {
		toc := []map[string]interface{}{}
		for i, h := range headings {
			toc = append(toc, map[string]interface{}{
				"selector":       selectors[i],
				"heading":        h.Text,
				"level":          h.Level,
				"short_selector": markdown.GenerateShortSelector(file, h, headings),
			})
		}
		return cmdutil.OutputJSON(map[string]interface{}{
			"file_path":         filePath,
			"table_of_contents": toc,
			"metadata":          cmdutil.CreateJSONMetadata(ctx.Cmd, true, ctx.StartTime),
		})
	}
	fmt.Printf("Table of Contents for %s:\n\n", file)
	for _, sel := range selectors {
		fmt.Println(sel)
	}
	return nil
}

// --- Helpers ---

func resolvePeekFilePath(ws *workspace.Workspace, filename string, noWorkspace bool) string {
	if noWorkspace {
		if filepath.IsAbs(filename) {
			return filename
		}
		cwd, _ := os.Getwd()
		return filepath.Join(cwd, filename)
	}
	if ws != nil {
		if filename == "inbox.md" {
			return ws.InboxPath
		}
		return filepath.Join(ws.Root, filename)
	}
	return filename
}

func parseSelector(selector string) (string, []string) {
	if strings.Contains(selector, "#") {
		parts := strings.SplitN(selector, "#", 2)
		return parts[0], strings.Split(parts[1], "/")
	}
	return selector, nil
}

// Enhanced selector: file:42 or file:42#heading
func resolveEnhancedSelector(ws *workspace.Workspace, selector string) string {
	if !strings.Contains(selector, ":") {
		return selector
	}
	parts := strings.SplitN(selector, ":", 2)
	filename := parts[0]
	lineStr := parts[1]
	lineNum := 0
	fmt.Sscanf(lineStr, "%d", &lineNum)
	filePath := resolvePeekFilePath(ws, filename, false)
	content, err := os.ReadFile(filePath)
	if err != nil {
		return selector
	}
	doc := markdown.ParseMarkdownAST(content)
	heading := markdown.FindHeadingByLine(doc, lineNum)
	if heading != "" {
		return fmt.Sprintf("%s#%s", filename, heading)
	}
	return filename
}
