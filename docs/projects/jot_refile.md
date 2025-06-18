# Project: jot refile

## Objective
Implement the `jot refile` command to move notes from `inbox.md` to files in the `lib/` directory using an AST-based approach for robust note parsing and flexible targeting.

## Scope
- `jot refile` command with AST-based note parsing
- Multiple targeting mechanisms (index, exact match, pattern, interactive)
- Move notes from `inbox.md` to files in `lib/`
- Natural sort order preservation
- Robust markdown handling with proper AST parsing
- Update both `inbox.md` and destination file
- Handle errors and conflicts

## Technical Approach

### AST-Based Note Parsing
Replace the current line-by-line scanning with markdown AST parsing using goldmark:

- **Current**: Simple `## ` header detection with string concatenation
- **Enhanced**: Full markdown AST traversal with proper heading detection
- **Benefits**: Natural document order, robust parsing, extensibility

### Enhanced Note Structure
```go
type Note struct {
    Title       string        // Heading text (timestamp)
    Content     []byte        // Raw markdown content 
    AST         ast.Node      // Full AST subtree
    SourcePos   text.Segment  // Position in source document
    LineStart   int           // For backwards compatibility
    LineEnd     int           // For backwards compatibility
}
```

### Flexible Targeting Mechanisms

1. **Index Targeting**: `jot refile 1,3,5 target.md`
   - Target specific notes by their numbered position
   - Support ranges: `1-3` and individual indices: `1,5,7`

2. **Exact Match**: `jot refile --exact "2024-01-15 14:30:00" target.md`
   - Target notes by exact timestamp match
   - Precise targeting for specific notes

3. **Pattern Match**: `jot refile --pattern "2024-01-15.*" target.md`
   - Target notes matching regex patterns
   - Flexible batch operations

4. **Offset Targeting**: `jot refile --offset 150 target.md`
   - Target notes by cursor byte position (for editor integration)
   - Enables seamless editor workflow integration

5. **Interactive Mode**: `jot refile --interactive target.md`
   - Present numbered list with multi-select interface
   - User-friendly note selection

### Implementation Phases

#### Phase 1: AST Foundation
- Add goldmark dependency
- Implement AST-based note parsing
- Maintain backwards compatibility with existing Note structure
- Update parseInboxNotes() function

#### Phase 2: Enhanced Targeting
- Implement IndexTargeter for numeric targeting
- Add command-line parsing for index syntax
- Test with existing refile command structure

#### Phase 3: Advanced Targeting
- Implement ExactTargeter and PatternTargeter
- Add --exact and --pattern command-line flags
- Enhance error handling and validation

#### Phase 4: Editor Integration  
- Implement OffsetTargeter for cursor-based targeting
- Add --offset flag for editor integration
- Enable byte-level precise note extraction
- Preserve exact formatting through byte-range extraction

#### Phase 5: Interactive Mode
- Implement InteractiveTargeter
- Add --interactive flag and user interface
- Polish user experience

## Dependencies
- **goldmark**: Modern, CommonMark-compliant markdown parser
  - Chosen over blackfriday/v2 for better performance and compliance
  - Provides robust AST with source position tracking

## Deliverables
- Working `jot refile` command with AST-based parsing
- Multiple targeting mechanisms (index, exact, pattern, interactive)
- Documentation for usage and workflows
- Updated test suite covering new functionality
- Backwards compatibility with existing workflows

## Implementation Status
- ✅ **Analysis Complete**: Current implementation analyzed
- ✅ **Technical Design**: AST approach designed with code samples
- ✅ **Phase 1**: AST foundation implementation complete
  - ✅ Added goldmark dependency
  - ✅ Implemented AST-based note parsing with `parseInboxNotesAST()`
  - ✅ Enhanced Note structure with AST fields
  - ✅ Maintains backwards compatibility with legacy parser fallback
- ✅ **Phase 2**: Enhanced Index Targeting complete
  - ✅ Implemented IndexTargeter for numeric targeting
  - ✅ Support for individual indices: `1,3,5`
  - ✅ Support for ranges: `1-3`
  - ✅ Support for mixed syntax: `1,3-5,7`
  - ✅ Comprehensive error handling and validation
  - ✅ Updated refile command interface
- ✅ **Phase 3**: Advanced Targeting features (COMPLETED)
  - ✅ ExactTargeter for timestamp matching  
  - ✅ PatternTargeter for regex patterns
  - ✅ --exact flag for precise timestamp targeting
  - ✅ --pattern flag for flexible pattern-based targeting
  - ✅ Content matching (not just titles)
  - ✅ Comprehensive error handling for invalid patterns
- ✅ **Phase 4**: Editor Integration (COMPLETED)
  - ✅ OffsetTargeter for cursor-based note targeting
  - ✅ --offset flag for editor integration workflows
  - ✅ Byte-level precise note extraction
  - ✅ Perfect formatting preservation through exact byte ranges
  - ✅ Enhanced Note structure with ByteStart/ByteEnd fields
  - ✅ Manual byte position calculation for accuracy
- ⏳ **Phase 5**: Interactive Mode pending
  - ⏳ InteractiveTargeter with multi-select interface

## Current Functionality (COMPLETED)

### Index-Based Targeting ✅
```bash
# Move specific notes by index
jot refile 1,3,5 --dest target.md

# Move a range of notes  
jot refile 1-3 --dest target.md

# Move all notes
jot refile --all --dest target.md

# Mixed targeting
jot refile 1,3-5,7 --dest target.md
```

### Advanced Targeting ✅
```bash
# Exact timestamp matching
jot refile --exact '2025-06-06 10:30' --dest meetings.md

# Pattern matching (regex) - title and content
jot refile --pattern 'Meeting|Task' --dest work.md
jot refile --pattern 'technical debt|database' --dest dev.md

# Case-insensitive pattern matching
jot refile --pattern '(?i)important' --dest priority.md
```

### Editor Integration ✅
```bash
# Offset-based targeting for editor integration
jot refile --offset 150 --dest daily-notes.md

# Editor workflow: user positions cursor, editor sends byte offset
# Example: cursor at byte 1024 targets note containing that position
jot refile --offset 1024 --dest research.md
```

### Error Handling ✅
- Out-of-range index validation
- Invalid format detection  
- Missing destination file validation
- Comprehensive error messages

### File Operations ✅
- AST-based note parsing and extraction
- Clean removal from inbox.md
- Proper formatting in destination files
- Natural document order preservation

## Benefits of AST Approach
1. **Natural Sort Order**: Notes processed in document order, not scan order
2. **Robust Parsing**: Proper markdown structure handling
3. **Flexible Targeting**: Multiple ways to select notes
4. **Future Extensibility**: AST enables advanced features (metadata, nested content)
5. **Better Error Handling**: Precise source position tracking
6. **Performance**: Efficient single-pass parsing

## Editor Integration Benefits
The offset targeting feature enables seamless integration with text editors:

1. **Precise Targeting**: Cursor position directly identifies the containing note
2. **Format Preservation**: Exact byte extraction preserves all original formatting
3. **Workflow Integration**: No manual note selection required
4. **Universal Compatibility**: Works with any editor that can provide byte positions
5. **Zero Configuration**: No editor-specific setup required

### Technical Implementation
- **Byte Position Tracking**: Each note records `ByteStart` and `ByteEnd` positions
- **Manual Calculation**: Line-by-line parsing ensures accurate byte offsets
- **Exact Extraction**: Direct byte range copying preserves formatting perfectly
- **Robust Validation**: Comprehensive error handling for invalid offsets

## What Was Implemented

### Phase 1: AST Foundation ✅
The foundation has been successfully implemented with:

- **Goldmark Integration**: Added goldmark dependency for robust markdown parsing
- **Enhanced Note Structure**: Extended with AST fields while maintaining backwards compatibility
- **AST-based Parser**: New `parseInboxNotesAST()` function with fallback to legacy parser
- **Natural Order Processing**: Notes processed in document order, not scan order
- **Robust Content Extraction**: Proper handling of paragraphs, lists, and other markdown elements

### Phase 2: Index Targeting ✅  
Full index-based targeting system implemented:

- **Individual Indices**: `jot refile 1,3,5 --dest target.md`
- **Range Support**: `jot refile 1-3 --dest target.md`  
- **Mixed Syntax**: `jot refile 1,3-5,7 --dest target.md`
- **Comprehensive Validation**: Out-of-range detection, format validation, error messages
- **File Operations**: Clean note removal from inbox, proper destination formatting

### Phase 3: Advanced Targeting ✅  
Advanced targeting mechanisms implemented:

- **Exact Timestamp Matching**: `jot refile --exact '2025-06-06 10:30' --dest target.md`
- **Pattern Matching**: `jot refile --pattern 'Meeting|Task' --dest target.md`  
- **Content Search**: Patterns match both note titles and content
- **Regex Support**: Full regular expression functionality with proper error handling
- **Enhanced User Experience**: Clear error messages for invalid patterns and non-matching searches

### Testing Results ✅
- All existing tests pass (25 assertions across 8 test categories)
- Index targeting validated with various scenarios
- Range targeting tested and working
- Exact timestamp targeting verified
- Pattern matching (content and title) tested
- Regex error handling validated
- **Offset targeting tested and working**
- **Editor integration workflow validated**
- **Format preservation verified with complex content**
- Error handling verified for edge cases
- Backwards compatibility maintained

The refile command has evolved from a placeholder ("coming soon") into a fully functional note organization tool with robust AST-based parsing, flexible targeting mechanisms, and seamless editor integration. The addition of offset targeting in Phase 4 enables powerful editor workflows where users can position their cursor anywhere within a note and automatically refile it to the desired location while preserving all original formatting.
