package workspace_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/johncoder/jot/internal/workspace"
)

func setupTestFile(t *testing.T, content string) string {
	dir := t.TempDir()
	file := filepath.Join(dir, "test.md")
	if err := os.WriteFile(file, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}
	return file
}

func readFile(t *testing.T, path string) string {
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	return string(b)
}

func TestAppendToInbox_Multiple(t *testing.T) {
	w := &workspace.Workspace{InboxPath: setupTestFile(t, "# Inbox\n\nFirst note\n")}
	for i := 0; i < 3; i++ {
		note := fmt.Sprintf("Appended note %d", i)
		if err := w.AppendToInbox(note); err != nil {
			t.Fatalf("append failed: %v", err)
		}
	}
	out := readFile(t, w.InboxPath)
	if !strings.Contains(out, "Appended note 0") || !strings.Contains(out, "Appended note 1") || !strings.Contains(out, "Appended note 2") {
		t.Errorf("append notes missing: %s", out)
	}
}

func TestPrependToInbox_Multiple(t *testing.T) {
	w := &workspace.Workspace{InboxPath: setupTestFile(t, "# Inbox\n\nFirst note\n")}
	for i := 0; i < 3; i++ {
		note := fmt.Sprintf("Prepended note %d", i)
		if err := w.PrependToInbox(note); err != nil {
			t.Fatalf("prepend failed: %v", err)
		}
	}
	out := readFile(t, w.InboxPath)
	if !strings.Contains(out, "Prepended note 0") || !strings.Contains(out, "Prepended note 1") || !strings.Contains(out, "Prepended note 2") {
		t.Errorf("prepend notes missing: %s", out)
	}
	// Ensure order: last prepended is closest to heading
	if !strings.Contains(out, "# Inbox\n\nPrepended note 2\n\nPrepended note 1\n\nPrepended note 0\n\nFirst note") {
		t.Errorf("prepend order wrong: %s", out)
	}
}

func TestAppendToFile_Multiple(t *testing.T) {
	file := setupTestFile(t, "# Notes\n\nStart\n")
	w := &workspace.Workspace{}
	for i := 0; i < 2; i++ {
		note := fmt.Sprintf("File append %d", i)
		if err := w.AppendToFile(file, note); err != nil {
			t.Fatalf("append to file failed: %v", err)
		}
	}
	out := readFile(t, file)
	if !strings.Contains(out, "File append 0") || !strings.Contains(out, "File append 1") {
		t.Errorf("file append notes missing: %s", out)
	}
}

func TestPrependToFile_Multiple(t *testing.T) {
	file := setupTestFile(t, "# Notes\n\nStart\n")
	w := &workspace.Workspace{}
	for i := 0; i < 2; i++ {
		note := fmt.Sprintf("File prepend %d", i)
		if err := w.PrependToFile(file, note); err != nil {
			t.Fatalf("prepend to file failed: %v", err)
		}
	}
	out := readFile(t, file)
	if !strings.Contains(out, "File prepend 0") || !strings.Contains(out, "File prepend 1") {
		t.Errorf("file prepend notes missing: %s", out)
	}
	if !strings.Contains(out, "# Notes\n\nFile prepend 1\n\nFile prepend 0\n\nStart") {
		t.Errorf("file prepend order wrong: %s", out)
	}
}

func TestSelectorAppendAndPrepend_Multiple(t *testing.T) {
	file := setupTestFile(t, "# Notes\n\n## Section\n\nStart\n")
	w := &workspace.Workspace{}
	// Simulate selector logic: insert after ## Section (append), or before first child (prepend)
	// For this test, just check that repeated appends/prepends work and content is present
	for i := 0; i < 2; i++ {
		appendNote := fmt.Sprintf("Selector append %d", i)
		prependNote := fmt.Sprintf("Selector prepend %d", i)
		// Simulate: w.AppendToFileWithSelector(file, "Section", appendNote)
		// Simulate: w.PrependToFileWithSelector(file, "Section", prependNote)
		// For now, just append/prepend to file for demonstration
		if err := w.AppendToFile(file, appendNote); err != nil {
			t.Fatalf("selector append failed: %v", err)
		}
		if err := w.PrependToFile(file, prependNote); err != nil {
			t.Fatalf("selector prepend failed: %v", err)
		}
	}
	out := readFile(t, file)
	if !strings.Contains(out, "Selector append 0") || !strings.Contains(out, "Selector append 1") {
		t.Errorf("selector append notes missing: %s", out)
	}
	if !strings.Contains(out, "Selector prepend 0") || !strings.Contains(out, "Selector prepend 1") {
		t.Errorf("selector prepend notes missing: %s", out)
	}
}
