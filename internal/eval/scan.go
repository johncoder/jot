package eval

import (
	"bufio"
	"os"
	"strings"
)

// CodeBlock represents a fenced code block in markdown
// It stores the code, language, and line numbers
// Optionally, it can be associated with an EvalMetadata
// and a result block

type CodeBlock struct {
	StartLine   int
	EndLine     int
	Lang        string
	Code        []string
	Eval        *EvalMetadata
	ResultBlock *ResultBlock
}

type ResultBlock struct {
	StartLine int
	EndLine   int
	Content   []string
}

// ParseMarkdownForEvalBlocks scans a markdown file and returns all evaluable code blocks
func ParseMarkdownForEvalBlocks(filename string) ([]*CodeBlock, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	var blocks []*CodeBlock
	var inCode bool
	var codeBlock *CodeBlock
	var lineNum int
	var lastCodeBlock *CodeBlock
	for scanner.Scan() {
		line := scanner.Text()
		lineNum++
		trim := strings.TrimSpace(line)
		if strings.HasPrefix(trim, "```") {
			if !inCode {
				// Start of code block
				inCode = true
				lang := strings.TrimSpace(trim[3:])
				codeBlock = &CodeBlock{StartLine: lineNum, Lang: lang}
			} else {
				// End of code block
				inCode = false
				codeBlock.EndLine = lineNum
				blocks = append(blocks, codeBlock)
				lastCodeBlock = codeBlock
				codeBlock = nil
			}
			continue
		}
		if inCode {
			codeBlock.Code = append(codeBlock.Code, line)
			continue
		}
		// If not in code, check for eval link
		if IsEvalLink(trim) && lastCodeBlock != nil && lastCodeBlock.Eval == nil {
			meta, err := ParseEvalLink(trim)
			if err == nil {
				lastCodeBlock.Eval = meta
			}
		}
	}
	return blocks, scanner.Err()
}
