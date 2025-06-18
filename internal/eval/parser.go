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

// ParseEvalLink parses a markdown link of the form [](eval key1=val1 key2=val2 ...)
func ParseEvalLink(link string) (*EvalMetadata, error) {
	link = strings.TrimSpace(link)
	if !strings.HasPrefix(link, "[](") || !strings.HasSuffix(link, ")") {
		return nil, fmt.Errorf("not a valid markdown link")
	}
	content := link[3 : len(link)-1]
	if !strings.HasPrefix(content, "eval ") {
		return nil, fmt.Errorf("not an eval link")
	}
	paramsStr := strings.TrimSpace(content[5:])
	params, err := parseKeyValueParams(paramsStr)
	if err != nil {
		return nil, err
	}
	return &EvalMetadata{Params: params}, nil
}

// parseKeyValueParams parses key=value pairs, supporting quoted values
func parseKeyValueParams(s string) (map[string]string, error) {
	params := make(map[string]string)
	// Regex: key=value or key="value with spaces"
	re := regexp.MustCompile(`([a-zA-Z0-9_]+)=(("[^"]*")|([^\s]+))`)
	matches := re.FindAllStringSubmatch(s, -1)
	for _, m := range matches {
		key := m[1]
		val := m[3]
		if val == "" {
			val = m[4]
		}
		val = strings.Trim(val, `"`)
		params[key] = val
	}
	return params, nil
}

// IsEvalLink returns true if the given line is an eval link
func IsEvalLink(line string) bool {
	line = strings.TrimSpace(line)
	return strings.HasPrefix(line, "[](") && strings.Contains(line, "eval")
}
