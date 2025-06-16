# Editor Integration Guide

## Overview
The `jot refile --offset` feature enables seamless integration with text editors, allowing users to position their cursor anywhere within a note and automatically refile it without manual selection.

## How It Works
1. **User Workflow**: User positions cursor anywhere within a note in their editor
2. **Editor Action**: Editor calculates the byte offset of the cursor position
3. **Command Execution**: Editor executes `jot refile --offset <byte_position> --dest <target_file>`
4. **Automatic Targeting**: jot identifies which note contains that byte offset
5. **Perfect Preservation**: Note is moved with exact formatting preservation

## Technical Implementation
- **Byte Position Tracking**: Each note records precise `ByteStart` and `ByteEnd` positions
- **Accurate Calculation**: Manual line-by-line parsing ensures correct byte offsets
- **Exact Extraction**: Direct byte range copying preserves all formatting
- **Robust Validation**: Comprehensive error handling for invalid offsets

## Command Syntax
```bash
jot refile --offset <cursor_byte_position> --dest <destination_file>
```

## Examples
```bash
# Editor integration: cursor at byte 150 targets containing note
jot refile --offset 150 --dest daily-notes.md

# Example workflow: cursor at byte 1024
jot refile --offset 1024 --dest research.md
```

## Editor-Specific Integration

### VS Code
```typescript
// Get cursor byte position
const editor = vscode.window.activeTextEditor;
const position = editor.selection.active;
const document = editor.document;
const byteOffset = Buffer.from(document.getText(new vscode.Range(new vscode.Position(0, 0), position))).length;

// Execute jot refile
const terminal = vscode.window.createTerminal();
terminal.sendText(`jot refile --offset ${byteOffset} --dest target.md`);
```

### Vim/Neovim
```vim
" Get byte offset of cursor position
function! GetByteOffset()
    return line2byte(line('.')) + col('.') - 1
endfunction

" Refile note containing cursor
command! -nargs=1 JotRefile execute '!jot refile --offset ' . GetByteOffset() . ' --dest ' . <q-args>
```

### Emacs
```elisp
;; Get byte position of point
(defun jot-get-byte-position ()
  "Get byte position of point in current buffer."
  (1- (position-bytes (point))))

;; Refile note at point
(defun jot-refile (destination)
  "Refile note containing point to DESTINATION."
  (interactive "sDestination file: ")
  (shell-command 
    (format "jot refile --offset %d --dest %s" 
            (jot-get-byte-position) 
            destination)))
```

## Benefits
1. **Zero Configuration**: No editor-specific setup required
2. **Universal Compatibility**: Works with any editor that can provide byte positions
3. **Precise Targeting**: Cursor position directly identifies the containing note
4. **Format Preservation**: All original formatting (markdown, code blocks, tables) preserved
5. **Seamless Workflow**: No interruption to user's editing flow

## Error Handling
- **Invalid Offsets**: Clear error messages for out-of-range positions
- **No Matching Note**: Helpful feedback when offset doesn't target any note
- **File Validation**: Ensures destination file path is valid

## Format Preservation
The implementation preserves **all** original formatting:
- ✅ Markdown formatting (headers, bold, italic, links)
- ✅ Code blocks with syntax highlighting
- ✅ Tables with proper alignment  
- ✅ Lists (ordered and unordered)
- ✅ Line breaks and spacing
- ✅ Special characters and Unicode

## Implementation Status
✅ **Complete** - The offset targeting feature is fully implemented and production-ready.

This feature enables powerful editor workflows where users can seamlessly move between editing and organizing their notes without leaving their preferred text editor environment.
