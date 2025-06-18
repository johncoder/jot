#!/bin/bash

# Test script for jot eval
set -euo pipefail

# Build the `jot` binary before running tests
BUILD_DIR="$(dirname "$0")/../build"
mkdir -p "$BUILD_DIR"
JOT_BINARY="$BUILD_DIR/jot"

# Correct the path to `main.go`
go build -o "$JOT_BINARY" main.go

# Use the built binary for testing
JOT_CMD="$JOT_BINARY"

# Create a temporary Markdown file
temp_file=$(mktemp /tmp/test_eval.XXXXXX.md)
expected_output_file="${temp_file%.md}.expected.txt"

# Write test content to the Markdown file using echo -e to interpret escaped characters
echo -e "# Test Markdown File\n\n\`\`\`bash\n echo \"Hello, jot!\"\n\`\`\`" > "$temp_file"

# Add a second code block that is not intended for evaluation
echo -e "\n\n\`\`\`plaintext\nThis is a plaintext block and should not be evaluated.\n\`\`\`" >> "$temp_file"

# Debugging: Print temp file content immediately after writing
cat "$temp_file" >&2

# Correctly write the expected output
cat <<EOF > "$expected_output_file"
Evaluating code blocks in file: $temp_file
Executing code block with language: bash
Output:
Hello, jot!
EOF

# Debugging: Print expected output for verification
echo "Expected Output:" >&2
cat "$expected_output_file" >&2

# Debugging: Print temp file path and content
echo "Temp File Path: $temp_file" >&2
echo "Temp File Content:" >&2
cat "$temp_file" >&2

# Debugging: Print file content before running jot eval
echo "File content before jot eval:" >&2
cat "$temp_file" >&2

# Run jot eval and capture output
actual_output=$($JOT_CMD eval "$temp_file" 2>&1 || true)

# Debugging: Print file content after jot eval
echo "File content after jot eval:" >&2
cat "$temp_file" >&2

# Simplify comparison logic
if [ "$actual_output" == "$(cat $expected_output_file)" ]; then
  echo "Test passed!"
else
  echo "Test failed!"
  echo "Actual Output:" >&2
  echo "$actual_output" >&2
  echo "Expected Output:" >&2
  cat "$expected_output_file" >&2
  exit 1
fi

# Verify updated Markdown file content
updated_content=$(cat "$temp_file")
expected_markdown_content="# Test Markdown File

\`\`\`bash
 echo \"Hello, jot!\"
\`\`\`
Output:
\`\`\`
Hello, jot!
\`\`\`"

if [ "$updated_content" == "$expected_markdown_content" ]; then
  echo "Markdown content test passed!"
else
  echo "Markdown content test failed!"
  echo "Updated Content:" >&2
  echo "$updated_content" >&2
  echo "Expected Content:" >&2
  echo "$expected_markdown_content" >&2
  exit 1
fi

# Clean up temporary files
rm -f "$temp_file" "$expected_output_file"

# Debugging: Print heredoc content before writing to the file
echo "Heredoc Content:" >&2
echo "# Test Markdown File\n\n\`\`\`bash\n echo \"Hello, jot!\"\n\`\`\`" >&2
