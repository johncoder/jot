package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/johncoder/jot/internal/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/text"
)

var evalCmd = &cobra.Command{
	Use:   "eval",
	Short: "Evaluate code blocks in a Markdown file",
	Long: `The eval command allows you to execute code blocks embedded in Markdown files.
You can target specific blocks by name or evaluate all approved blocks in the file.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			fmt.Fprintln(os.Stderr, "Error: You must specify a Markdown file to evaluate.")
			os.Exit(1)
		}

		filePath := args[0]
		parseMarkdownAndExecuteWithResults(filePath)
	},
}

func parseMarkdownAndExecuteWithResults(filePath string) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to read file:", err)
		os.Exit(1)
	}

	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
		),
	)

	node := md.Parser().Parse(text.NewReader(content))
	results := make(map[int]string) // Map byte offsets to execution results

	// Declare supportedLanguages at the beginning of the function
	var supportedLanguages map[string]string

	// Load supported languages from jotrc
	viper.SetConfigName("jotrc")
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME")
	// Handle absence of jotrc gracefully
	if err := viper.ReadInConfig(); err != nil {
		supportedLanguages = map[string]string{
			"bash":       "true",
			"python":     "true",
			"javascript": "true",
			"java":       "true",
			"c":          "true",
			"cpp":        "true",
			"csharp":     "true",
			"ruby":       "true",
			"go":         "true",
			"swift":      "true",
			"php":        "true",
			"typescript": "true",
			"kotlin":     "true",
			"r":          "true",
			"perl":       "true",
			"scala":      "true",
			"rust":       "true",
			"dart":       "true",
			"shell":      "true",
			"lua":        "true",
		}
	} else {
		supportedLanguages = viper.GetStringMapString("supported_languages")
		if len(supportedLanguages) == 0 {
			supportedLanguages = map[string]string{
				"bash":       "true",
				"python":     "true",
				"javascript": "true",
				"java":       "true",
				"c":          "true",
				"cpp":        "true",
				"csharp":     "true",
				"ruby":       "true",
				"go":         "true",
				"swift":      "true",
				"php":        "true",
				"typescript": "true",
				"kotlin":     "true",
				"r":          "true",
				"perl":       "true",
				"scala":      "true",
				"rust":       "true",
				"dart":       "true",
				"shell":      "true",
				"lua":        "true",
			}
		}
	}

	// Traverse the AST to find and execute code blocks
	walker := func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {		if n.Kind() == ast.KindCodeSpan {
			codeSpan := n.(*ast.CodeSpan)
			// Use the same offset calculation as ParseMarkdownAST
			endOffset := getNodeEnd(n, content)

			// Use Lines().At(i).Value(content) to extract content from CodeSpan nodes
			rawCode := ""
			for i := 0; i < codeSpan.Lines().Len(); i++ {
				segment := codeSpan.Lines().At(i)
				rawCode += string(segment.Value(content))
			}

			// Interpret escaped characters (e.g., \n as newline)
			code := strings.ReplaceAll(rawCode, "\\n", "\n")

			// Check if the CodeSpan contains newlines and conforms to the expected format
			if strings.Contains(code, "\n") {
				lines := strings.SplitN(code, "\n", 2)
				firstLine := lines[0]
				remainingCode := lines[1]

				// Handle cases where the first line specifies only the language
				language := firstLine
				if strings.HasPrefix(firstLine, ":language") {
					args := utils.ParseHeaderArguments(firstLine)
					if lang, ok := args[":language"]; ok {
						language = lang
					}
				}

				// Check if the language is supported
				if _, ok := supportedLanguages[language]; ok {
					output, err := executeCodeBlock(language, remainingCode)
					if err == nil {
						results[endOffset] = output
					}
				}
			}
		}
		if n.Kind() == ast.KindFencedCodeBlock {
			fencedCodeBlock := n.(*ast.FencedCodeBlock)
			// Use the same offset calculation as ParseMarkdownAST
			endOffset := getNodeEnd(n, content)
			language := string(fencedCodeBlock.Language(content))
			code := ""
			for i := 0; i < fencedCodeBlock.Lines().Len(); i++ {
				segment := fencedCodeBlock.Lines().At(i)
				code += string(segment.Value(content))
			}

			// Check if the language is supported
			if _, ok := supportedLanguages[language]; ok {
				output, err := executeCodeBlock(language, code)
				if err == nil {
					results[endOffset] = output // Use byte offset as key
				}
			}
		}
		}

		return ast.WalkContinue, nil
	}

	if err := ast.Walk(node, walker); err != nil {
		fmt.Fprintln(os.Stderr, "Failed to walk AST:", err)
		os.Exit(1)
	}

	appendResultsToMarkdown(filePath, results)
}

func executeCodeBlock(language, code string) (string, error) {
	// Map language to interpreter command
	interpreter := map[string]string{
		"python":     "python3",
		"bash":       "bash",
		"javascript": "node",
	}

	cmd, ok := interpreter[language]
	if !ok {
		return "", fmt.Errorf("unsupported language: %s", language)
	}

	// Execute the code block
	execCmd := exec.Command(cmd)
	execCmd.Stdin = strings.NewReader(code)
	output, err := execCmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("execution failed: %s", err)
	}

	return string(output), nil
}

func AppendResultsToMarkdownStream(input io.Reader, output io.Writer, results map[int]string) error {
	content, err := io.ReadAll(input)
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}

	var updatedContent bytes.Buffer
	lastOffset := 0

	for offset, result := range results {
		if offset > len(content) {
			return fmt.Errorf("offset %d is beyond content length %d", offset, len(content))
		}
		updatedContent.Write(content[lastOffset:offset])
		updatedContent.WriteString("\nOutput:\n```\n")
		updatedContent.WriteString(result)
		updatedContent.WriteString("\n```\n")
		lastOffset = offset
	}

	if lastOffset > len(content) {
		return fmt.Errorf("lastOffset %d is beyond content length %d", lastOffset, len(content))
	}
	updatedContent.Write(content[lastOffset:])

	if _, err := output.Write(updatedContent.Bytes()); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	return nil
}

// Update the original function to use the refactored stream-based version
func appendResultsToMarkdown(filePath string, results map[int]string) {
	// Read the original file content
	inputFile, err := os.Open(filePath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to open file:", err)
		os.Exit(1)
	}
	defer inputFile.Close()

	// Create a buffer to hold the updated content
	var outputBuffer bytes.Buffer
	
	// Process the content
	if err := AppendResultsToMarkdownStream(inputFile, &outputBuffer, results); err != nil {
		fmt.Fprintln(os.Stderr, "Failed to append results:", err)
		os.Exit(1)
	}

	// Close the input file before writing
	inputFile.Close()

	// Write the updated content back to the file
	if err := os.WriteFile(filePath, outputBuffer.Bytes(), 0644); err != nil {
		fmt.Fprintln(os.Stderr, "Failed to write file:", err)
		os.Exit(1)
	}
}

// ParseMarkdownAST parses the Markdown content and returns a list of code block nodes with their offsets.
// Updated ParseMarkdownAST to use getNodeStart and getNodeEnd for accurate offsets
func ParseMarkdownAST(content []byte) ([]struct {
	Type  string
	Start int
	End   int
}, error) {
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
		),
	)

	node := md.Parser().Parse(text.NewReader(content))
	var nodes []struct {
		Type  string
		Start int
		End   int
	}

	err := ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			if n.Kind() == ast.KindFencedCodeBlock {
				codeBlock := n.(*ast.FencedCodeBlock)
				start := getNodeStart(n, content)
				end := getNodeEnd(n, content)
				nodes = append(nodes, struct {
					Type  string
					Start int
					End   int
				}{
					Type:  string(codeBlock.Language(content)),
					Start: start,
					End:   end,
				})
			}
		}
		return ast.WalkContinue, nil
	})

	if err != nil {
		return nil, err
	}

	return nodes, nil
}

// isFencedCodeBlock checks if a CodeSpan node is actually a fenced code block
func isFencedCodeBlock(spanContent string, node ast.Node, content []byte) bool {
	// Check if the span content looks like a fenced code block opening
	// (starts with language name and possibly options)
	lines := strings.Split(spanContent, "\n")
	if len(lines) < 2 {
		return false
	}

	// First line should be just the language (and possibly options)
	firstLine := strings.TrimSpace(lines[0])
	if firstLine == "" {
		return false
	}

	// Check if this CodeSpan is at the beginning of a line (after ```language)
	if hasSegment, ok := node.(interface{ Segment() *text.Segment }); ok {
		seg := hasSegment.Segment()
		if seg != nil {
			start := seg.Start
			// Look backwards to see if we have ``` before this span
			if start >= 3 && string(content[start-3:start]) == "```" {
				return true
			}
		}
	}

	return false
}

// getCodeSpanFencedStart gets the start position of a CodeSpan that's actually a fenced code block
func getCodeSpanFencedStart(node ast.Node, content []byte) int {
	if hasSegment, ok := node.(interface{ Segment() *text.Segment }); ok {
		seg := hasSegment.Segment()
		if seg != nil {
			start := seg.Start
			// Go back to include the ``` opening
			if start >= 3 && string(content[start-3:start]) == "```" {
				return start - 3
			}
		}
	}
	return 0
}

// getCodeSpanFencedEnd gets the end position of a CodeSpan that's actually a fenced code block
func getCodeSpanFencedEnd(node ast.Node, content []byte) int {
	if hasSegment, ok := node.(interface{ Segment() *text.Segment }); ok {
		seg := hasSegment.Segment()
		if seg != nil {
			end := seg.Stop
			// Look forward to find the closing ```
			remaining := content[end:]
			closingIndex := bytes.Index(remaining, []byte("```"))
			if closingIndex >= 0 {
				return end + closingIndex + 3
			}
		}
	}
	return 0
}

// extractLanguageFromFencedBlock extracts the language from a fenced code block content
func extractLanguageFromFencedBlock(spanContent string) string {
	lines := strings.Split(spanContent, "\n")
	if len(lines) > 0 {
		firstLine := strings.TrimSpace(lines[0])
		// The first line should be the language (and possibly options)
		parts := strings.Fields(firstLine)
		if len(parts) > 0 {
			return parts[0]
		}
	}
	return ""
}

func init() {
	rootCmd.AddCommand(evalCmd)

	// Add flags for targeting specific blocks or evaluating all blocks
	evalCmd.Flags().String("name", "", "Name of the code block to evaluate")
	evalCmd.Flags().Bool("all", false, "Evaluate all approved code blocks in the file")
}
