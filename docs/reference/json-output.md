[Documentation](../README.md) > JSON Output Reference

# JSON Output Reference

All jot commands support JSON output via the `--json` flag for automation and scripting. This reference documents the standardized JSON format, common patterns, and integration examples.

## Standard Response Structure

All JSON responses follow a consistent structure:

```json
{
  "operation": "command_name",
  "data": {
    // Command-specific data
  },
  "metadata": {
    "success": true,
    "command": "jot command",
    "execution_time_ms": 125,
    "timestamp": "2025-01-01T12:00:00Z"
  }
}
```

## Error Response Structure

When commands fail, they return a standardized error format:

```json
{
  "error": {
    "message": "Error description",
    "code": "error_code",
    "details": {
      // Additional error context
    }
  },
  "metadata": {
    "success": false,
    "command": "jot command",
    "execution_time_ms": 125,
    "timestamp": "2025-01-01T12:00:00Z"
  }
}
```

## Operations Response Structure

For commands that perform multiple operations (like batch operations):

```json
{
  "operations": [
    {
      "operation": "action_name",
      "result": "success",
      "details": {
        // Operation-specific details
      }
    }
  ],
  "summary": {
    "total_operations": 1,
    "successful": 1,
    "failed": 0
  },
  "metadata": {
    "success": true,
    "command": "jot command",
    "execution_time_ms": 125,
    "timestamp": "2025-01-01T12:00:00Z"
  }
}
```

## Metadata Fields

All JSON responses include standard metadata:

| Field | Type | Description |
|-------|------|-------------|
| `success` | boolean | Whether the command completed successfully |
| `command` | string | Full command that was executed |
| `execution_time_ms` | number | Command execution time in milliseconds |
| `timestamp` | string | ISO 8601 timestamp of command completion |

## Command-Specific Examples

### Status Command

```json
{
  "operation": "status",
  "workspace": {
    "name": "my-workspace",
    "root": "/path/to/workspace",
    "active": true
  },
  "files": {
    "inbox": {
      "path": "inbox.md",
      "exists": true,
      "size": 1024,
      "modified": "2025-01-01T12:00:00Z"
    }
  },
  "stats": {
    "total_files": 5,
    "total_size": 15360,
    "last_activity": "2025-01-01T11:30:00Z"
  },
  "metadata": {
    "success": true,
    "command": "jot status",
    "execution_time_ms": 25,
    "timestamp": "2025-01-01T12:00:00Z"
  }
}
```

### Capture Command

```json
{
  "operation": "capture",
  "file": "inbox.md",
  "content": "# Meeting Notes\n\nImportant discussion points",
  "template": "meeting",
  "bytes_written": 256,
  "metadata": {
    "success": true,
    "command": "jot capture --template meeting",
    "execution_time_ms": 45,
    "timestamp": "2025-01-01T12:00:00Z"
  }
}
```

### Find Command

```json
{
  "operation": "find",
  "query": "meeting",
  "results": [
    {
      "file": "inbox.md",
      "line": 15,
      "content": "# Meeting Notes",
      "context": "# Meeting Notes\n\nImportant discussion points"
    }
  ],
  "summary": {
    "total_matches": 1,
    "files_searched": 5,
    "search_time_ms": 12
  },
  "metadata": {
    "success": true,
    "command": "jot find meeting",
    "execution_time_ms": 35,
    "timestamp": "2025-01-01T12:00:00Z"
  }
}
```

### Hooks Command

```json
{
  "operation": "list",
  "hooks_dir": "/path/to/workspace/.jot/hooks",
  "hooks": [
    {
      "name": "pre-capture",
      "path": "/path/to/workspace/.jot/hooks/pre-capture",
      "executable": true,
      "sample": false,
      "size": 256,
      "mod_time": "2025-01-01T12:00:00Z"
    }
  ],
  "summary": {
    "total_hooks": 1,
    "executable_hooks": 1,
    "sample_hooks": 0
  },
  "metadata": {
    "success": true,
    "command": "jot hooks list",
    "execution_time_ms": 15,
    "timestamp": "2025-01-01T12:00:00Z"
  }
}
```

## Common Error Codes

| Code | Description | Commands |
|------|-------------|----------|
| `workspace_not_found` | No workspace found | All commands |
| `file_not_found` | File doesn't exist | peek, refile, archive |
| `permission_denied` | Access denied | All file operations |
| `invalid_selector` | Path selector invalid | refile, archive |
| `template_not_found` | Template doesn't exist | capture, template |
| `hook_failure` | Hook script failed | capture, refile, archive |
| `validation_error` | Input validation failed | All commands |

## Integration Examples

### Command Line Processing with jq

```bash
# Process jot JSON output with jq
jot status --json | jq '.workspace.name'

# Extract specific fields
jot find "meeting" --json | jq '.results[].file'

# Check for errors
jot capture --template nonexistent --json | jq '.error.message'

# Count files in workspace
jot status --json | jq '.stats.total_files'

# Get all hook names
jot hooks list --json | jq '.hooks[].name'

# Extract file paths from search results
jot find "todo" --json | jq '.results[].file' | sort | uniq
```

### Shell Scripting Integration

```bash
#!/bin/bash
# Script to process jot operations

result=$(jot capture --content "Script note" --json)
success=$(echo "$result" | jq -r '.metadata.success')

if [ "$success" = "true" ]; then
    echo "Capture successful"
    file=$(echo "$result" | jq -r '.file')
    echo "Content written to: $file"
else
    echo "Capture failed"
    error=$(echo "$result" | jq -r '.error.message')
    echo "Error: $error"
fi
```

### Error Handling

```bash
# Check command success
if jot status --json | jq -e '.metadata.success' > /dev/null; then
    echo "Command succeeded"
else
    echo "Command failed"
fi

# Extract error details
jot capture --content "" --json | jq '.error.details // empty'
```

### Python API Integration

```python
import subprocess
import json

def jot_status():
    """Get workspace status as JSON"""
    result = subprocess.run(
        ['jot', 'status', '--json'],
        capture_output=True,
        text=True
    )
    
    return json.loads(result.stdout)

def jot_capture(content, template=None):
    """Capture content with optional template"""
    cmd = ['jot', 'capture', '--content', content, '--json']
    if template:
        cmd.extend(['--template', template])
    
    result = subprocess.run(cmd, capture_output=True, text=True)
    return json.loads(result.stdout)

# Usage
status = jot_status()
if status['metadata']['success']:
    print(f"Workspace: {status['workspace']['name']}")
    
    # Capture new content
    capture_result = jot_capture("API test note", "default")
    if capture_result['metadata']['success']:
        print(f"Captured to: {capture_result['file']}")
```

## Performance Considerations

### JSON Processing Overhead

- **Minimal impact**: JSON formatting adds ~1-5ms to command execution
- **Memory usage**: Slightly higher memory usage for large datasets
- **Streaming**: Large results are not streamed, entire response is buffered

### Optimization Tips

- **Filter early**: Use command-specific filters rather than JSON post-processing
- **Batch operations**: Process multiple operations in single commands when possible
- **Use jq efficiently**: Stream processing for large datasets

## Best Practices

### Error Handling

1. **Always check metadata.success** before processing data
2. **Use appropriate error codes** for different failure modes
3. **Provide meaningful error messages** with context in details

### Automation

1. **Use jq for field extraction** rather than complex parsing
2. **Cache results** for expensive operations
3. **Handle network timeouts** and retries appropriately

### Integration

1. **Validate JSON structure** before processing
2. **Use consistent field names** across different commands
3. **Handle version compatibility** in automation scripts

## See Also

- [Global Options](../commands/README.md#global-options) - The `--json` flag
- [External Commands](../commands/jot-external.md) - JSON output in external commands
- [Command Reference](../commands/README.md) - All commands support JSON output
