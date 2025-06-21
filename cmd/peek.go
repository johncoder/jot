package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/johncoder/jot/internal/markdown"
	"github.com/johncoder/jot/internal/workspace"
	"github.com/spf13/cobra"
	"github.com/yuin/goldmark/ast"
)

var peekCmd = &cobra.Command{
	Use:   "peek SELECTOR",
	Short: "View a specific markdown subtree without opening the entire file",
	Long: `View a specific markdown subtree (heading with all nested content) without opening the entire file.

The peek command uses the same path-based selector syntax as refile:
- Each segment uses case-insensitive contains matching
- Must match exactly one subtree
- Leading slashes handle unusual document structures

Examples:
  jot peek "inbox.md#meeting"                    # View meeting notes subtree
  jot peek "work.md#projects/frontend"          # View frontend project section
  jot peek "notes.md#research/database"         # View database research
  jot peek "inbox.md#/foo/bar"                  # Skip level 1, find foo/bar
  jot peek "inbox.md" --toc                     # Show table of contents for entire file
  jot peek "work.md#projects" --toc             # Show TOC for projects subtree

This is useful for quickly reviewing specific sections without opening full files in an editor.`,

	Args: cobra.RangeArgs(0, 1), // Allow 0 or 1 arguments for --toc mode
	RunE: func(cmd *cobra.Command, args []string) error {
		ws, err := workspace.RequireWorkspace()
		if err != nil {
			return err
		}

		// Get flags
		raw, _ := cmd.Flags().GetBool("raw")
		info, _ := cmd.Flags().GetBool("info")
		toc, _ := cmd.Flags().GetBool("toc")

		// Handle TOC mode
		if toc {
			if len(args) == 0 {
				return fmt.Errorf("table of contents requires a file or selector (e.g., 'inbox.md' or 'work.md#projects')")
			}
			return showTableOfContents(ws, args[0])
		}

		// Regular peek mode requires exactly one argument
		if len(args) != 1 {
			return fmt.Errorf("peek requires a selector argument (e.g., 'inbox.md#meeting')")
		}

		// Parse the source path selector
		sourcePath, err := markdown.ParsePath(args[0])
		if err != nil {
			return fmt.Errorf("invalid selector: %w", err)
		}

		// Extract the subtree
		subtree, err := ExtractSubtree(ws, sourcePath)
		if err != nil {
			return fmt.Errorf("failed to extract subtree: %w", err)
		}

		// Display subtree information if requested
		if info {
			printSubtreeInfo(subtree, sourcePath.File)
			fmt.Println()
		}

		// Display the subtree content
		if raw {
			// Raw mode: output just the content without any formatting
			os.Stdout.Write(subtree.Content)
		} else {
			// Formatted mode: show with nice header
			fmt.Printf("# Subtree: %s#%s\n\n", sourcePath.File, 
				strings.Join(sourcePath.Segments, "/"))
			
			// Remove trailing newlines for cleaner output
			content := subtree.Content
			for len(content) > 0 && content[len(content)-1] == '\n' {
				content = content[:len(content)-1]
			}
			
			fmt.Println(string(content))
		}

		return nil
	},
}

// printSubtreeInfo displays metadata about the subtree
func printSubtreeInfo(subtree *markdown.Subtree, filename string) {
	fmt.Printf("Subtree Information:\n")
	fmt.Printf("  File: %s\n", filename)
	fmt.Printf("  Heading: %q\n", subtree.Heading)
	fmt.Printf("  Level: %d\n", subtree.Level)
	fmt.Printf("  Content length: %d bytes\n", len(subtree.Content))
	fmt.Printf("  Byte range: %d-%d\n", subtree.StartOffset, subtree.EndOffset)
	
	// Count nested headings
	nestedCount := countNestedHeadings(subtree.Content, subtree.Level)
	if nestedCount > 0 {
		fmt.Printf("  Nested headings: %d\n", nestedCount)
	}
}

// countNestedHeadings counts how many headings are nested within this subtree
func countNestedHeadings(content []byte, baseLevel int) int {
	lines := splitLines(content)
	count := 0
	
	for _, line := range lines {
		if len(line) > 0 && line[0] == '#' {
			// Count the heading level
			level := 0
			for i := 0; i < len(line) && line[i] == '#'; i++ {
				level++
			}
			
			// Only count headings deeper than the base level
			if level > baseLevel {
				count++
			}
		}
	}
	
	return count
}

// splitLines splits content into lines
func splitLines(content []byte) []string {
	if len(content) == 0 {
		return []string{}
	}
	
	var lines []string
	start := 0
	
	for i := 0; i < len(content); i++ {
		if content[i] == '\n' {
			lines = append(lines, string(content[start:i]))
			start = i + 1
		}
	}
	
	// Add the last line - even if it's empty due to trailing newline
	if start <= len(content) {
		lines = append(lines, string(content[start:]))
	}
	
	return lines
}

// showTableOfContents displays a table of contents for a file or subtree
func showTableOfContents(ws *workspace.Workspace, selector string) error {
	// Check if this is a simple file name or a path selector
	var content []byte
	var filename string
	var baseFilename string
	var subtreePath string
	var err error
	
	if strings.Contains(selector, "#") {
		// This is a path selector - extract the subtree first
		sourcePath, parseErr := markdown.ParsePath(selector)
		if parseErr != nil {
			return fmt.Errorf("invalid selector: %w", parseErr)
		}
		
		subtree, extractErr := ExtractSubtree(ws, sourcePath)
		if extractErr != nil {
			return fmt.Errorf("failed to extract subtree: %w", extractErr)
		}
		
		content = subtree.Content
		baseFilename = sourcePath.File
		subtreePath = strings.Join(sourcePath.Segments, "/")
		filename = fmt.Sprintf("%s#%s", baseFilename, subtreePath)
	} else {
		// This is just a file name
		baseFilename = selector
		filename = selector
		var filePath string
		
		if selector == "inbox.md" {
			filePath = ws.InboxPath
		} else {
			// Handle files in lib/ directory
			if !strings.HasSuffix(selector, ".md") {
				selector += ".md"
				baseFilename = selector
				filename = selector
			}
			filePath = filepath.Join(ws.LibDir, selector)
		}
		
		// Check if file exists
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			return fmt.Errorf("file not found: %s", selector)
		}
		
		// Read file content
		content, err = os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", selector, err)
		}
	}
	
	if len(content) == 0 {
		fmt.Printf("File %s is empty (no table of contents available)\n", filename)
		return nil
	}
	
	// Parse document and extract headings
	doc := markdown.ParseDocument(content)
	headings := extractHeadingsFromContent(doc, content)
	
	if len(headings) == 0 {
		fmt.Printf("No headings found in %s\n", filename)
		return nil
	}
	
	// Display table of contents
	fmt.Printf("Table of Contents: %s\n", filename)
	fmt.Printf("%s\n\n", strings.Repeat("=", len("Table of Contents: ")+len(filename)))
	
	for i, heading := range headings {
		// Create indentation based on heading level
		indent := strings.Repeat("  ", heading.Level-1)
		
		// Format the heading line
		fmt.Printf("%s%s %s\n", indent, strings.Repeat("#", heading.Level), heading.Text)
		
		// Create a helpful selector hint for navigation
		if subtreePath == "" { // Full file TOC
			var selectorPath string
			if heading.Level == 1 {
				// For level 1 headings, simple path
				selectorPath = strings.ToLower(heading.Text)
			} else {
				// For deeper levels, we need to build a hierarchical path
				// For simplicity, show the skip-level syntax
				skipPrefix := strings.Repeat("/", heading.Level-1)
				selectorPath = fmt.Sprintf("%s%s", skipPrefix, strings.ToLower(heading.Text))
			}
			
			selectorHint := fmt.Sprintf("jot peek \"%s#%s\"", baseFilename, selectorPath)
			fmt.Printf("%s%s\n", indent, fmt.Sprintf("  → %s", selectorHint))
		}
		
		// Add spacing between entries for readability
		if i < len(headings)-1 {
			fmt.Println()
		}
	}
	
	// Add helpful usage notes at the end
	fmt.Println()
	if subtreePath == "" {
		fmt.Printf("Use 'jot peek \"%s#<heading>\"' to view specific sections.\n", baseFilename)
		fmt.Printf("Tip: Heading names are matched case-insensitively using 'contains' logic.\n")
	} else {
		fmt.Printf("This is a table of contents for the subtree '%s'.\n", subtreePath)
		fmt.Printf("Use 'jot peek \"%s#%s/<subheading>\"' to view nested sections.\n", baseFilename, subtreePath)
	}
	
	return nil
}

// HeadingInfo represents a heading with its metadata
type HeadingInfo struct {
	Text  string
	Level int
	Line  int
}

// extractHeadingsFromContent extracts all headings from markdown content
func extractHeadingsFromContent(doc ast.Node, content []byte) []HeadingInfo {
	var headings []HeadingInfo
	
	// Walk the AST to find all headings
	ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		
		if heading, ok := n.(*ast.Heading); ok {
			headingText := markdown.ExtractHeadingText(heading, content)
			if strings.TrimSpace(headingText) != "" {
				offset := markdown.GetNodeOffset(heading, content)
				lineNum := markdown.CalculateLineNumber(content, offset)
				
				headings = append(headings, HeadingInfo{
					Text:  headingText,
					Level: heading.Level,
					Line:  lineNum,
				})
			}
		}
		
		return ast.WalkContinue, nil
	})
	
	return headings
}

func init() {
	// Add flags
	peekCmd.Flags().BoolP("raw", "r", false, "Output raw content without formatting")
	peekCmd.Flags().BoolP("info", "i", false, "Show subtree metadata information")
	peekCmd.Flags().BoolP("toc", "t", false, "Show table of contents for file or subtree")
	
	// Add to root command
	rootCmd.AddCommand(peekCmd)
}
