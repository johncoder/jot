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
	var pendingEval *EvalMetadata // Store eval element waiting for code block

	for scanner.Scan() {
		line := scanner.Text()
		lineNum++
		trim := strings.TrimSpace(line)

		// Check for eval element first (new pattern: eval before code)
		if !inCode && IsEvalElement(trim) {
			meta, err := ParseEvalElement(trim)
			if err == nil {
				pendingEval = meta
			}
			continue
		}

		if strings.HasPrefix(trim, "```") {
			if !inCode {
				// Start of code block
				inCode = true
				lang := strings.TrimSpace(trim[3:])
				codeBlock = &CodeBlock{
					StartLine: lineNum,
					Lang:      lang,
					Eval:      pendingEval, // Associate with preceding eval element
				}
				pendingEval = nil // Clear pending eval
			} else {
				// End of code block
				inCode = false
				codeBlock.EndLine = lineNum
				blocks = append(blocks, codeBlock)
				codeBlock = nil
			}
			continue
		}

		if inCode {
			codeBlock.Code = append(codeBlock.Code, line)
			continue
		}
	}
	return blocks, scanner.Err()
}
