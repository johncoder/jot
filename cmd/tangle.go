package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

var tangleCmd = &cobra.Command{
	Use:   "tangle <file>",
	Short: "Extract code blocks into standalone source files",
	Long: `Extract code blocks from Markdown files into standalone source files.

The tangle command looks for code blocks with :tangle or :file header arguments 
and extracts them to the specified file paths. Directories are created as needed.

Examples:
  jot tangle notes.md              # Extract code blocks from notes.md
  jot tangle docs/tutorial.md      # Extract from tutorial file`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[0]
		fmt.Printf("Tangling code blocks in file: %s\n", filePath)
		return tangleMarkdown(filePath)
	},
}

func init() {
}

func tangleMarkdown(filePath string) error {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	md := goldmark.New()
	node := md.Parser().Parse(text.NewReader(content))

	// Traverse the AST to find code blocks with :tangle
	walker := func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if n.Kind() == ast.KindFencedCodeBlock {
			codeBlock := n.(*ast.FencedCodeBlock)
			info := string(codeBlock.Info.Text(content))
			args := parseHeaderArguments(info)

			if _, ok := args["tangle"]; ok {
				filePath, hasFile := args["file"]
				if !hasFile {
					fmt.Printf("Skipping code block without :file argument\n")
					return ast.WalkContinue, nil
				}

				code := string(codeBlock.Text(content))
				if err := writeToFile(filePath, code); err != nil {
					return ast.WalkStop, err
				}
			}
		}
		return ast.WalkContinue, nil
	}

	if err := ast.Walk(node, walker); err != nil {
		return fmt.Errorf("failed to walk AST: %w", err)
	}

	return nil
}

func writeToFile(filePath, content string) error {
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directories for %s: %w", filePath, err)
	}

	if err := ioutil.WriteFile(filePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file %s: %w", filePath, err)
	}

	fmt.Printf("Wrote code block to %s\n", filePath)
	return nil
}

func parseHeaderArguments(info string) map[string]string {
	args := make(map[string]string)
	for _, part := range strings.Fields(info) {
		if strings.HasPrefix(part, ":") {
			// Handle arguments like :name
			args[part] = "true"
		} else if strings.Contains(part, "=") {
			kv := strings.SplitN(part, "=", 2)
			args[kv[0]] = kv[1]
		} else {
			args[part] = ""
		}
	}
	return args
}
