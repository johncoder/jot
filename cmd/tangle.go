package cmd

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

var tangleCmd = &cobra.Command{
	Use:   "tangle",
	Short: "Extract code blocks into standalone source files",
	Long: `The tangle command extracts code blocks from Markdown files into standalone source files.
It uses the :tangle and :file header arguments to determine the output file paths.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			fmt.Println("Error: You must specify a Markdown file to tangle.")
			os.Exit(1)
		}

		filePath := args[0]
		fmt.Printf("Tangling code blocks in file: %s\n", filePath)
		tangleMarkdown(filePath)
	},
}

func init() {
}

func tangleMarkdown(filePath string) {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Failed to read file: %s", err)
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
					log.Printf("Skipping code block without :file argument\n")
					return ast.WalkContinue, nil
				}

				code := string(codeBlock.Text(content))
				writeToFile(filePath, code)
			}
		}
		return ast.WalkContinue, nil
	}

	if err := ast.Walk(node, walker); err != nil {
		log.Fatalf("Failed to walk AST: %s", err)
	}
}

func writeToFile(filePath, content string) {
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Fatalf("Failed to create directories for %s: %s", filePath, err)
	}

	if err := ioutil.WriteFile(filePath, []byte(content), 0644); err != nil {
		log.Fatalf("Failed to write file %s: %s", filePath, err)
	}

	fmt.Printf("Wrote code block to %s\n", filePath)
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
