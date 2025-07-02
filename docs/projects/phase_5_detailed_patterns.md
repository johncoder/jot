# Phase 5: Detailed Pattern Analysis for Consolidation

## Overview

This document identifies the exact patterns that will be consolidated across the three main areas: User Interaction, Path Operations, and Error Management. Each pattern includes code examples and the proposed unified approach.

---

## 🎯 AREA 1: User Interaction Framework

### Pattern 1: Confirmation Dialogs

**Current Scattered Implementation (4+ instances):**

```go
// Pattern A: refile.go (most complete)
fmt.Printf("\n🚀 Execute refile operation? [y/N]: ")
var response string
fmt.Scanln(&response)
response = strings.ToLower(strings.TrimSpace(response))
return response == "y" || response == "yes", nil

// Pattern B: template.go (basic)
fmt.Print("Approve this template? [y/N]: ")
var response string
fmt.Scanln(&response)
if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
    fmt.Println("Template not approved.")
    return nil
}

// Pattern C: eval.go (with reader)
fmt.Printf("Approve this block with %s mode? [y/N]: ", approvalMode)
reader := bufio.NewReader(os.Stdin)
response, err := reader.ReadString('\n')
if err != nil {
    return err
}
response = strings.TrimSpace(strings.ToLower(response))
if response != "y" && response != "yes" {
    fmt.Println("Approval cancelled.")
    return nil
}
```

**Proposed Unified Pattern:**
```go
// In internal/cmdutil/interact.go
func ConfirmOperation(prompt string) (bool, error)
func ConfirmWithIcon(icon, operation string) (bool, error)

// Usage examples:
confirmed, err := interact.ConfirmWithIcon("🚀", "Execute refile operation")
confirmed, err := interact.ConfirmOperation("Approve this template")
```

### Pattern 2: Status Messages with Icons

**Current Scattered Implementation (36+ instances):**

```go
// Success patterns
fmt.Printf("✓ Note captured (%d characters)\n", len(finalContent))
fmt.Printf("✓ Found workspace at: %s\n", ws.Root)
fmt.Printf("✓ File exists: %s\n", destPath.File)

// Error patterns  
fmt.Printf("✗ File not found: %s\n", destPath.File)
fmt.Println("✗ Not in a jot workspace")
fmt.Printf("✗ Failed to create inbox.md: %v\n", err)

// Warning patterns
fmt.Printf("Warning: post-capture hook failed: %s\n", err.Error())
fmt.Printf("⚠️  Warning: Found duplicate headings in %s: %s\n", sourceFile, duplicates)
```

**Proposed Unified Pattern:**
```go
// In internal/cmdutil/interact.go
func ShowSuccess(message string, args ...interface{})
func ShowError(message string, args ...interface{})
func ShowWarning(message string, args ...interface{})
func ShowInfo(message string, args ...interface{})
func ShowProgress(message string, args ...interface{})

// Usage examples:
interact.ShowSuccess("Note captured (%d characters)", len(finalContent))
interact.ShowError("File not found: %s", destPath.File)
interact.ShowWarning("Post-capture hook failed: %s", err.Error())
```

### Pattern 3: Input Normalization

**Current Scattered Implementation (3+ instances):**

```go
// Various normalization approaches
response = strings.ToLower(strings.TrimSpace(response))
strings.TrimSpace(strings.ToLower(response))
strings.ToLower(response) != "y" && strings.ToLower(response) != "yes"
```

**Proposed Unified Pattern:**
```go
// In internal/cmdutil/interact.go
func NormalizeUserInput(input string) string
func IsConfirmationYes(input string) bool

// Usage:
normalized := interact.NormalizeUserInput(rawInput)
confirmed := interact.IsConfirmationYes(userResponse)
```

---

## 📁 AREA 2: Path/File Operations Utilities

### Pattern 1: Workspace-Relative Path Construction

**Current Scattered Implementation (16+ instances):**

```go
// Direct pattern repetition
destinationPath = filepath.Join(ws.Root, destination)
filePath = filepath.Join(ws.Root, destPath.File)
filePath = filepath.Join(ws.Root, filename)
absolutePath = filepath.Join(ws.Root, file)
testFile := filepath.Join(libDir, "test_peek.md")
```

**Proposed Unified Pattern:**
```go
// In internal/cmdutil/paths.go (extend existing)
func (p *PathUtil) WorkspaceJoin(relativePath string) string
func (p *PathUtil) LibJoin(relativePath string) string  
func (p *PathUtil) JotDirJoin(relativePath string) string

// Usage:
paths := cmdutil.NewPathUtil(ws)
filePath := paths.WorkspaceJoin(destPath.File)
libFile := paths.LibJoin("test_peek.md")
```

### Pattern 2: Directory Creation Before File Operations

**Current Scattered Implementation (17+ instances):**

```go
// Pattern A: In content.go (already some consolidation)
dir := filepath.Dir(resolvedPath)
if err := os.MkdirAll(dir, 0755); err != nil {
    return fmt.Errorf("failed to create directory %s: %w", dir, err)
}

// Pattern B: In various commands
if err := os.MkdirAll(libDir, 0755); err != nil {
    return fmt.Errorf("failed to create lib directory: %w", err)
}

// Pattern C: In tests
if err := os.MkdirAll(jotDir, 0755); err != nil {
    t.Fatalf("failed to create jot directory: %v", err)
}
```

**Proposed Unified Pattern:**
```go
// In internal/cmdutil/paths.go
func (p *PathUtil) EnsureDir(path string) error
func (p *PathUtil) EnsureDirForFile(filePath string) error
func (p *PathUtil) SafeWriteFile(path string, content []byte) error
func (p *PathUtil) SafeAppendFile(path string, content []byte) error

// Usage:
err := paths.EnsureDirForFile(destinationPath)
err := paths.SafeWriteFile(filePath, content)  // auto-creates directories
```

### Pattern 3: Path Resolution and Validation

**Current Scattered Implementation (51+ instances):**

```go
// Absolute path resolution
absPath, err := filepath.Abs(filename)
absPath, _ := filepath.Abs(workspacePath)

// Relative path calculation  
relPath, err := filepath.Rel(ws.Root, absolutePath)
relPath, _ := filepath.Rel(ws.Root, path)

// Directory/basename extraction
dir := filepath.Dir(baseFilename)
baseName := filepath.Base(file)
```

**Proposed Unified Pattern:**
```go
// In internal/cmdutil/paths.go
func (p *PathUtil) ToAbsolute(path string) (string, error)
func (p *PathUtil) ToWorkspaceRelative(absolutePath string) (string, error)
func (p *PathUtil) IsWithinWorkspace(path string) bool
func (p *PathUtil) SplitPath(path string) (dir, base, ext string)

// Usage:
absPath, err := paths.ToAbsolute(filename)
relPath, err := paths.ToWorkspaceRelative(absPath)
isInside := paths.IsWithinWorkspace(somePath)
```

---

## ⚠️ AREA 3: Enhanced Error Management

### Pattern 1: Standard Error Message Formats

**Current Scattered Implementation (32+ instances):**

```go
// "failed to X: %w" pattern (most common)
return fmt.Errorf("failed to read file %s: %w", sourcePath.File, err)
return fmt.Errorf("failed to create temp file: %w", err)
return fmt.Errorf("failed to resolve destination: %w", err)

// "error X: %w" pattern
return fmt.Errorf("error reading file: %w", err)
return fmt.Errorf("error analyzing path: %w", err)

// "invalid X: %w" pattern  
return fmt.Errorf("invalid destination '%s': %w", destination, err)
return fmt.Errorf("invalid source path '%s': %w", args[0], err)
```

**Proposed Unified Pattern:**
```go
// In internal/cmdutil/errors.go (extend existing)
func NewOperationError(operation, target string, err error) error
func NewValidationError(item, value string, err error) error
func NewFileError(operation, filepath string, err error) error

// Usage:
return cmdutil.NewFileError("read", sourcePath.File, err)
return cmdutil.NewOperationError("create temp file", "", err)  
return cmdutil.NewValidationError("destination", destination, err)
```

### Pattern 2: Contextual Error Wrapping

**Current Implementation:**
```go
// Already partially consolidated through HandleError/HandleOperationError
return ctx.HandleOperationError("pre-capture hook", fmt.Errorf("pre-capture hook failed: %s", err.Error()))
return ctx.HandleError(fmt.Errorf("please specify a markdown file"))
```

**Proposed Enhancement:**
```go
// Enhanced context preservation
func (ctx *Context) WrapFileError(operation, filepath string, err error) error
func (ctx *Context) WrapValidationError(field, value string, err error) error
func (ctx *Context) WrapExternalError(command string, err error) error

// Usage:
return ctx.WrapFileError("read", filename, err)
return ctx.WrapExternalError("fzf", err)
```

---

## 🚀 Implementation Priority & Impact

### High Impact Patterns (Implement First)

1. **Confirmation Dialogs** (4 identical implementations → 1 utility)
2. **Status Messages** (36+ scattered implementations → 5 utilities)  
3. **Workspace Path Construction** (16+ identical patterns → 3 utilities)

### Medium Impact Patterns (Implement Second)

4. **Directory Creation** (17+ similar patterns → 4 utilities)
5. **Input Normalization** (3+ variations → 2 utilities)
6. **Path Resolution** (51+ scattered operations → 6 utilities)

### Lower Impact Patterns (Polish Phase)

7. **Error Message Standardization** (32+ variations → 3 utilities)
8. **Contextual Error Wrapping** (enhancement of existing system)

---

## 📊 Expected Reduction Metrics

| Pattern Category | Current Instances | Proposed Utilities | Reduction Ratio |
|------------------|-------------------|-------------------|-----------------|
| Confirmation Dialogs | 4 implementations | 2 functions | 2:1 |
| Status Messages | 36+ instances | 5 functions | 7:1 |
| Workspace Paths | 16+ instances | 3 functions | 5:1 |
| Directory Creation | 17+ instances | 4 functions | 4:1 |
| Path Operations | 51+ instances | 6 functions | 8:1 |
| Error Patterns | 32+ instances | 3 functions | 10:1 |

**Total Estimated Reduction: ~156 duplicate implementations → ~23 utility functions**

---

## 🎯 Success Criteria

### Code Quality
- [ ] Eliminate 150+ duplicate implementations
- [ ] Standardize user interaction patterns
- [ ] Reduce path operation complexity
- [ ] Improve error message consistency

### User Experience  
- [ ] Consistent confirmation dialog behavior
- [ ] Uniform status message formatting
- [ ] Predictable success/error/warning indicators
- [ ] Clear, contextual error messages

### Maintainability
- [ ] Centralized interaction logic
- [ ] Reusable path operation utilities
- [ ] Consistent error handling patterns
- [ ] Improved test coverage for utilities

This pattern analysis provides the detailed roadmap for implementing the compression-oriented refactoring with specific code examples and clear consolidation targets.
