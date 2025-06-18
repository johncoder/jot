package utils

import "strings"

// ParseHeaderArguments parses a string of header arguments into a map of key-value pairs.
func ParseHeaderArguments(info string) map[string]string {
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
