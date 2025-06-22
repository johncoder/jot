package eval

import (
	"fmt"
	"regexp"
	"strings"
)

// EvalMetadata holds parsed parameters from an eval link
// Example: [](eval name=foo shell=python3 timeout=5s)
type EvalMetadata struct {
	Params map[string]string
}

// ParseEvalElement parses an HTML eval element of the form <eval name="foo" shell="python3" timeout="5s" />
func ParseEvalElement(element string) (*EvalMetadata, error) {
	element = strings.TrimSpace(element)
	if !strings.HasPrefix(element, "<eval") || !strings.HasSuffix(element, "/>") {
		return nil, fmt.Errorf("not a valid eval element")
	}

	// Extract the attributes portion between <eval and />
	content := element[5 : len(element)-2] // Remove "<eval" and "/>"
	content = strings.TrimSpace(content)

	params, err := parseHTMLAttributes(content)
	if err != nil {
		return nil, err
	}

	return &EvalMetadata{Params: params}, nil
}

// parseHTMLAttributes parses HTML-style attributes: name="value" key="value with spaces"
func parseHTMLAttributes(s string) (map[string]string, error) {
	params := make(map[string]string)
	// Regex: key="value" or key='value'
	re := regexp.MustCompile(`([a-zA-Z0-9_-]+)=["']([^"']*)["']`)
	matches := re.FindAllStringSubmatch(s, -1)

	for _, m := range matches {
		key := m[1]
		value := m[2]
		params[key] = value
	}

	return params, nil
}

// IsEvalElement returns true if the given line is an eval element
func IsEvalElement(line string) bool {
	line = strings.TrimSpace(line)
	return strings.HasPrefix(line, "<eval") && strings.Contains(line, "/>")
}
