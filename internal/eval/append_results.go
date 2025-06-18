package eval

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// UpdateMarkdownWithResults updates the markdown file by inserting result blocks after eval links
func UpdateMarkdownWithResults(filename string, results []*EvalResult) error {
	input, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	lines := strings.Split(string(input), "\n")

	// Find eval links and insert results after them
	for _, r := range results {
		if r.Block == nil || r.Block.Eval == nil {
			continue
		}

		// Check if results should be processed
		resultsParam := getResultsParam(r.Block.Eval.Params)
		if resultsParam == "none" || resultsParam == "silent" {
			continue // Don't insert results for these modes
		}

		// Find the eval element for this specific block using line numbers
		// The eval element should be right after the code block ends
		evalElementLineIndex := -1
		for i := r.Block.EndLine - 1; i < len(lines) && i < r.Block.EndLine + 4; i++ { // search within 5 lines after code block (0-indexed)
			if i < 0 || i >= len(lines) {
				continue
			}
			if IsEvalElement(strings.TrimSpace(lines[i])) {
				// Verify this is the correct eval element by checking if it matches our block's metadata
				if meta, err := ParseEvalElement(strings.TrimSpace(lines[i])); err == nil {
					if blockName, ok := r.Block.Eval.Params["name"]; ok {
						if elemName, ok := meta.Params["name"]; ok && elemName == blockName {
							evalElementLineIndex = i
							break
						}
					} else {
						// If no name, match by line proximity (first eval element found)
						evalElementLineIndex = i
						break
					}
				}
			}
		}

		if evalElementLineIndex == -1 {
			continue // Couldn't find corresponding eval element
		}

		// Format the output based on results parameters
		formattedResult, err := formatResult(r, r.Block.Eval.Params, filename)
		if err != nil {
			return err
		}

		if formattedResult == "" {
			continue // No result to insert
		}

		// Handle different result insertion modes
		handling := getResultsHandling(r.Block.Eval.Params)
		switch handling {
		case "replace":
			lines = replaceResultBlock(lines, evalElementLineIndex, formattedResult)
		case "append":
			lines = appendResultBlock(lines, evalElementLineIndex, formattedResult)
		case "prepend":
			lines = prependResultBlock(lines, evalElementLineIndex, formattedResult)
		default:
			lines = replaceResultBlock(lines, evalElementLineIndex, formattedResult) // default to replace
		}
	}

	return os.WriteFile(filename, []byte(strings.Join(lines, "\n")), 0644)
}

// getResultsParam extracts the results parameter, defaulting to "code"
func getResultsParam(params map[string]string) string {
	if result, ok := params["results"]; ok {
		return result
	}
	return "code" // default
}

// getResultsHandling extracts the results handling mode, defaulting to "replace"
func getResultsHandling(params map[string]string) string {
	resultsParam := getResultsParam(params)
	
	// Check if results param contains handling mode
	if strings.Contains(resultsParam, "append") {
		return "append"
	}
	if strings.Contains(resultsParam, "prepend") {
		return "prepend"
	}
	if strings.Contains(resultsParam, "silent") {
		return "silent"
	}
	if strings.Contains(resultsParam, "none") {
		return "none"
	}
	
	return "replace" // default
}

// formatResult formats the execution result based on the results parameter
func formatResult(result *EvalResult, params map[string]string, filename string) (string, error) {
	if result.Output == "" && result.Err == nil {
		return "", nil
	}

	output := result.Output
	if result.Err != nil {
		output = fmt.Sprintf("Error: %v\n%s", result.Err, output)
	}

	resultsParam := getResultsParam(params)
	
	// Parse compound results parameters (e.g., "table", "output", "code")
	resultType := parseResultType(resultsParam)
	
	switch resultType {
	case "code":
		return formatAsCodeBlock(output), nil
	case "table":
		return formatAsTable(output), nil
	case "list":
		return formatAsList(output), nil
	case "raw":
		return output, nil
	case "file":
		return formatAsFile(output, params, filename)
	case "html":
		return formatAsHTML(output), nil
	case "verbatim":
		return output, nil
	default:
		return formatAsCodeBlock(output), nil // default to code block
	}
}

// parseResultType extracts the result type from compound results parameter
func parseResultType(resultsParam string) string {
	// Handle compound parameters like "output table" or "value code"
	parts := strings.Fields(resultsParam)
	for _, part := range parts {
		switch part {
		case "code", "table", "list", "raw", "file", "html", "verbatim":
			return part
		}
	}
	return "code" // default
}

// formatAsCodeBlock wraps output in a code block
func formatAsCodeBlock(output string) string {
	return fmt.Sprintf("```\n%s\n```", strings.TrimRight(output, "\n"))
}

// formatAsTable attempts to format output as a markdown table
func formatAsTable(output string) string {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) == 0 {
		return ""
	}

	// Try to detect if this is already table-like data (CSV, TSV, etc.)
	var table strings.Builder
	
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		
		// Split by common delimiters
		var cells []string
		if strings.Contains(line, "\t") {
			cells = strings.Split(line, "\t")
		} else if strings.Contains(line, ",") {
			cells = strings.Split(line, ",")
		} else if strings.Contains(line, "|") {
			cells = strings.Split(line, "|")
		} else {
			// Single column
			cells = []string{line}
		}
		
		// Clean up cells
		for j, cell := range cells {
			cells[j] = strings.TrimSpace(cell)
		}
		
		// Format as markdown table row
		table.WriteString("| ")
		table.WriteString(strings.Join(cells, " | "))
		table.WriteString(" |\n")
		
		// Add header separator for first row
		if i == 0 {
			table.WriteString("|")
			for range cells {
				table.WriteString(" --- |")
			}
			table.WriteString("\n")
		}
	}
	
	return table.String()
}

// formatAsList formats output as a markdown list
func formatAsList(output string) string {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	var list strings.Builder
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		list.WriteString("- ")
		list.WriteString(line)
		list.WriteString("\n")
	}
	
	return list.String()
}

// formatAsHTML wraps output in an HTML code block
func formatAsHTML(output string) string {
	return fmt.Sprintf("```html\n%s\n```", strings.TrimRight(output, "\n"))
}

// formatAsFile saves output to a file and returns a markdown link
func formatAsFile(output string, params map[string]string, baseFilename string) (string, error) {
	// Determine output file path
	var outputPath string
	if filePath, ok := params["file"]; ok {
		outputPath = filePath
	} else {
		// Generate a filename based on the markdown file and block name
		baseName := strings.TrimSuffix(filepath.Base(baseFilename), filepath.Ext(baseFilename))
		blockName := params["name"]
		if blockName == "" {
			blockName = "output"
		}
		outputPath = fmt.Sprintf("%s_%s.txt", baseName, blockName)
	}
	
	// Make path absolute if relative
	if !filepath.IsAbs(outputPath) {
		dir := filepath.Dir(baseFilename)
		outputPath = filepath.Join(dir, outputPath)
	}
	
	// Write output to file
	err := os.WriteFile(outputPath, []byte(output), 0644)
	if err != nil {
		return "", fmt.Errorf("failed to write output file: %v", err)
	}
	
	// Return markdown link
	relPath, err := filepath.Rel(filepath.Dir(baseFilename), outputPath)
	if err != nil {
		relPath = outputPath
	}
	
	// Determine if it's an image or regular file
	ext := strings.ToLower(filepath.Ext(outputPath))
	if ext == ".png" || ext == ".jpg" || ext == ".jpeg" || ext == ".gif" || ext == ".svg" {
		return fmt.Sprintf("![Output](%s)", relPath), nil
	} else {
		return fmt.Sprintf("[Output File](%s)", relPath), nil
	}
}

// replaceResultBlock replaces any existing result block after the eval element
func replaceResultBlock(lines []string, evalElementIndex int, result string) []string {
	// Remove any existing result blocks after this eval element
	j := evalElementIndex + 1
	for j < len(lines) {
		line := strings.TrimSpace(lines[j])
		if strings.HasPrefix(line, "```") {
			// Found start of a code block, find its end
			k := j + 1
			for k < len(lines) && !strings.HasPrefix(strings.TrimSpace(lines[k]), "```") {
				k++
			}
			if k < len(lines) {
				k++ // include closing ```
			}
			// Remove this code block
			lines = append(lines[:j], lines[k:]...)
		} else if isTableLine(line) {
			// Found start of a markdown table, find its end
			k := j
			for k < len(lines) {
				if isTableLine(strings.TrimSpace(lines[k])) {
					k++
				} else if strings.TrimSpace(lines[k]) == "" {
					// Skip empty lines after table
					k++
				} else {
					// Hit non-table, non-empty content
					break
				}
			}
			// Remove the table
			lines = append(lines[:j], lines[k:]...)
		} else if line == "" {
			// Skip empty lines
			j++
		} else {
			// Hit non-empty, non-result content, stop removing
			break
		}
	}
	
	// Insert new result with proper blank line separation
	resultLines := strings.Split(result, "\n")
	
	// After removing existing results, j points to where we should insert
	// Check if there's already a blank line at the insertion point or just before it
	needsBlankLine := true
	if j > 0 && j <= len(lines) {
		// Check if there's a blank line just before our insertion point
		if j > evalElementIndex+1 && strings.TrimSpace(lines[j-1]) == "" {
			needsBlankLine = false
		}
		// Also check if the line right after eval element is already blank
		if j == evalElementIndex+1 && j < len(lines) && strings.TrimSpace(lines[j]) == "" {
			needsBlankLine = false
		}
	}
	
	var newContent []string
	if needsBlankLine {
		newContent = append([]string{""}, resultLines...)
	} else {
		newContent = resultLines
	}
	
	lines = append(lines[:j], append(newContent, lines[j:]...)...)
	
	return lines
}

// isTableLine returns true if the line looks like a markdown table row
func isTableLine(line string) bool {
	if line == "" {
		return false
	}
	// A table line contains pipes and typically starts/ends with |
	return strings.Contains(line, "|") && 
		   (strings.HasPrefix(line, "|") || strings.HasSuffix(line, "|") || strings.Count(line, "|") >= 2)
}

// appendResultBlock adds result after existing results
func appendResultBlock(lines []string, evalElementIndex int, result string) []string {
	// Find end of existing results
	j := evalElementIndex + 1
	for j < len(lines) {
		line := strings.TrimSpace(lines[j])
		if strings.HasPrefix(line, "```") {
			// Skip over code blocks
			k := j + 1
			for k < len(lines) && !strings.HasPrefix(strings.TrimSpace(lines[k]), "```") {
				k++
			}
			if k < len(lines) {
				k++ // include closing ```
			}
			j = k
		} else if isTableLine(line) {
			// Skip over markdown tables
			k := j
			for k < len(lines) && (isTableLine(strings.TrimSpace(lines[k])) || strings.TrimSpace(lines[k]) == "") {
				if isTableLine(strings.TrimSpace(lines[k])) {
					k++
				} else {
					// Empty line after table, include it
					k++
				}
			}
			j = k
		} else if line == "" {
			j++
		} else {
			break
		}
	}
	
	// Insert result at position j with proper blank line separation
	resultLines := strings.Split(result, "\n")
	
	// Add a blank line before the result for proper Markdown parsing
	newContent := append([]string{""}, resultLines...)
	lines = append(lines[:j], append(newContent, lines[j:]...)...)
	
	return lines
}

// prependResultBlock adds result before existing results
func prependResultBlock(lines []string, evalElementIndex int, result string) []string {
	// Insert right after eval element with proper blank line separation
	resultLines := strings.Split(result, "\n")
	
	// Always add a blank line before the result for proper Markdown parsing
	newContent := append([]string{""}, resultLines...)
	lines = append(lines[:evalElementIndex+1], append(newContent, lines[evalElementIndex+1:]...)...)
	
	return lines
}
