package markdown

import (
	"encoding/json"
)

// Utility functions for operations, serialization, etc.

// SerializeOp serializes a MarkdownOp to JSON.
func SerializeOp(op MarkdownOp) ([]byte, error) {
	return json.Marshal(op)
}

// DeserializeOp deserializes a JSON blob to a MarkdownOp.
// TODO: Implement type switching for each op type
func DeserializeOp(data []byte) (MarkdownOp, error) {
	return nil, nil // TODO
}

// Shared helpers for AST manipulation, selector resolution, etc. can go here.
