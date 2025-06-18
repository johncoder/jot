package tests

import (
	"bytes"
	"testing"

	"github.com/johncoder/jot/cmd"
)

func TestAppendResultsToMarkdownStream(t *testing.T) {
	input := "# Test Markdown File\n\n```bash eval=true\necho \"Hello, jot!\"\n```"

	// Assert input length for debugging
	t.Logf("Input length: %d", len(input))
	t.Logf("Input content: %q", input)

	results := map[int]string{
		62: "Hello, jot!",
	}

	var output bytes.Buffer

	err := cmd.AppendResultsToMarkdownStream(bytes.NewReader([]byte(input)), &output, results)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := "# Test Markdown File\n\n```bash eval=true\necho \"Hello, jot!\"\n```\nOutput:\n```\nHello, jot!\n```\n"

	if output.String() != expected {
		t.Errorf("output mismatch\nexpected len=%d:\n%q\nactual len=%d:\n%q", len(expected), expected, len(output.String()), output.String())
	}
}

func TestMarkdownASTOffsets(t *testing.T) {
	input := "# Test Markdown File\n\n```bash eval=true\necho \"Hello, jot!\"\n```"

	// Assert input length for debugging
	t.Logf("Input length: %d", len(input))

	// Parse the Markdown input into an AST
	nodes, err := cmd.ParseMarkdownAST([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error parsing Markdown AST: %v", err)
	}

	// Expected offsets for code block
	expectedOffsets := []struct {
		Type  string
		Start int
		End   int
	}{
		{"bash", 22, 62},
	}

	if len(nodes) != len(expectedOffsets) {
		t.Fatalf("expected %d code blocks, got %d", len(expectedOffsets), len(nodes))
	}

	for i, node := range nodes {
		if node.Type != expectedOffsets[i].Type {
			t.Errorf("expected type %s, got %s", expectedOffsets[i].Type, node.Type)
		}
		if node.Start != expectedOffsets[i].Start {
			t.Errorf("expected start offset %d, got %d", expectedOffsets[i].Start, node.Start)
		}
		if node.End != expectedOffsets[i].End {
			t.Errorf("expected end offset %d, got %d", expectedOffsets[i].End, node.End)
		}
	}
}
