# Troubleshooting

This guide helps you diagnose and fix common jot issues. Start with `jot doctor` for automated diagnostics.

## Quick Diagnostics

### Run jot doctor

The first step for any issue:

```bash
jot doctor
```

This checks:

- Workspace integrity
- File permissions
- Configuration validity
- Template security
- Common setup issues

### Get Detailed Status

```bash
jot status --verbose
jot doctor --show-config
```

## Common Issues

### Installation and Setup

#### "jot: command not found"

**Cause:** jot is not in your PATH or not installed.

**Solutions:**

```bash
# Check if jot is installed
which jot

# If using the installer, ensure ~/.local/bin is in PATH
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc

# Or install with Go
go install github.com/johncoder/jot@latest

# Or download manually from releases
```

#### "No workspace found"

**Cause:** Current directory doesn't contain a `.jot` directory.

**Solutions:**

```bash
# Initialize workspace in current directory
jot init

# Or specify workspace explicitly
jot --workspace ~/notes capture

# Or add to global config
echo '{"workspaces": {"default": "~/notes"}}' > ~/.jotrc
```

#### Permission errors

**Cause:** Incorrect file permissions on workspace or config.

**Solutions:**

```bash
# Fix workspace permissions
chmod 755 .jot/
chmod 644 .jot/config.json
chmod 644 inbox.md

# Fix config permissions
chmod 600 ~/.jotrc

# Fix template permissions
chmod 644 .jot/templates/*.md
```

### Configuration Issues

#### Config file not loading

**Symptoms:** Default behavior despite having config file.

**Debug:**

```bash
# Check config file exists
ls -la ~/.jotrc

# Validate JSON5 syntax
json5 validate ~/.jotrc

# Show effective config
jot doctor --show-config
```

**Solutions:**

```bash
# Fix JSON5 syntax errors
# Common issues: missing commas, unquoted keys, trailing commas

# Minimal valid config
echo '{"workspaces": {"default": "~/notes"}}' > ~/.jotrc
```

#### Workspace discovery not working

**Symptoms:** "No workspace found" despite having `.jot` directory.

**Debug:**

```bash
# Check current directory
pwd
ls -la .jot/

# Test discovery manually
find . -name ".jot" -type d
```

**Solutions:**

```bash
# Ensure .jot is a directory, not a file
file .jot

# Check parent directories for .jot
cd .. && ls -la .jot/

# Use explicit workspace
jot --workspace $(pwd) status
```

### Template Issues

#### "Template needs approval"

**Symptoms:** Template requires approval before use.

**Solutions:**

```bash
# List template status
jot template list

# Approve specific template
jot template approve meeting

# Review template before approval
jot template view meeting
```

#### Template commands failing

**Symptoms:** Template renders but shell commands don't work.

**Debug:**

```bash
# Test commands manually
date '+%Y-%m-%d'
git branch --show-current

# Check template content
jot template view problematic-template
```

**Solutions:**

```bash
# Use command fallbacks in templates
$(git branch --show-current 2>/dev/null || echo "no-branch")

# Check command availability
which git date curl

# Use absolute paths if needed
$(/usr/bin/date '+%Y-%m-%d')
```

#### Template permission denied

**Symptoms:** Cannot execute template commands.

**Solutions:**

```bash
# Check template file permissions
ls -la .jot/templates/

# Fix permissions
chmod 644 .jot/templates/*.md

# Re-approve template
jot template approve template-name
```

### Editor Integration

#### Editor not opening

**Symptoms:** `jot capture` doesn't open editor.

**Debug:**

```bash
# Check editor setting
echo $EDITOR
echo $VISUAL

# Test editor manually
$EDITOR test.md
```

**Solutions:**

```bash
# Set editor environment variable
export EDITOR=vim
export VISUAL=code

# Or set in config
echo '{"editor": "vim"}' > ~/.jotrc

# For VS Code, ensure it waits
echo '{"editor": "code --wait"}' > ~/.jotrc
```

#### Editor opens but doesn't save content

**Symptoms:** Editor opens but note isn't captured.

**Solutions:**

```bash
# For VS Code, use --wait flag
export EDITOR="code --wait"

# For GUI editors, ensure they wait for file close
export EDITOR="gedit --wait"

# Test with simple editor
export EDITOR=nano
jot capture
```

### File and Content Issues

#### Large inbox warning

**Symptoms:** Warning about too many notes in inbox.

**Solutions:**

```bash
# Check inbox size
wc -l inbox.md

# Refile notes to organize
jot refile

# Archive old notes
jot archive
```

#### Search not finding content

**Symptoms:** `jot find` doesn't return expected results.

**Debug:**

```bash
# Test with simple query
jot find "test"

# Check file permissions
ls -la lib/*.md

# Verify content exists
grep -r "search-term" lib/
```

**Solutions:**

```bash
# Rebuild search index (if supported)
jot doctor --rebuild-index

# Check file encoding
file lib/*.md

# Use absolute paths
jot find "term" $(pwd)/lib/work.md
```

#### Refile not working

**Symptoms:** `jot refile` doesn't move content.

**Debug:**

```bash
# Check inbox content
cat inbox.md

# Verify permissions
ls -la inbox.md lib/
```

**Solutions:**

```bash
# Ensure lib/ directory exists
mkdir -p lib/

# Check file permissions
chmod 644 inbox.md
chmod 755 lib/
chmod 644 lib/*.md

# Try manual refile
jot refile --verbose
```

### Git Integration

#### Auto-commit not working

**Symptoms:** Changes not automatically committed.

**Debug:**

```bash
# Check git status
git status

# Verify git config
git config user.name
git config user.email
```

**Solutions:**

```bash
# Initialize git repository
git init
git config user.name "Your Name"
git config user.email "you@example.com"

# Check jot git config
jot doctor --show-config | grep git

# Disable auto-commit if problematic
echo '{"git": {"auto_commit": false}}' > .jot/config.json
```

### Performance Issues

#### Slow search

**Symptoms:** `jot find` takes too long.

**Solutions:**

```bash
# Check number of files
find lib/ -name "*.md" | wc -l

# Exclude large files from search
echo '{"search": {"exclude_patterns": ["large-file.md"]}}' > .jot/config.json

# Archive old content
jot archive
```

#### Large workspace

**Symptoms:** All operations slow.

**Solutions:**

```bash
# Check workspace size
du -sh .

# Archive old content
jot archive

# Split into multiple workspaces
jot workspace add project2 ~/project2-notes
```

## Error Messages

### "workspace integrity check failed"

**Cause:** Corrupted or incomplete workspace structure.

**Solution:**

```bash
# Recreate workspace structure
mkdir -p lib .jot/logs .jot/templates

# Restore missing files
touch inbox.md
echo '{}' > .jot/config.json

# Run integrity check
jot doctor
```

### "template execution denied"

**Cause:** Template contains unapproved shell commands.

**Solution:**

```bash
# Approve template
jot template approve template-name

# Or review and modify template
jot template edit template-name
```

### "config parse error"

**Cause:** Invalid JSON5 in configuration file.

**Solution:**

```bash
# Validate config syntax
json5 validate ~/.jotrc

# Common fixes:
# - Add missing commas
# - Remove trailing commas
# - Quote object keys properly
# - Escape backslashes in paths
```

### "file permission denied"

**Cause:** Insufficient permissions on workspace files.

**Solution:**

```bash
# Fix common permission issues
chmod 755 .jot/
chmod 644 .jot/config.json
chmod 644 inbox.md
chmod 755 lib/
chmod 644 lib/*.md
```

## Recovery Procedures

### Corrupted Workspace

If your workspace becomes corrupted:

```bash
# 1. Backup existing data
cp -r .jot .jot.backup
cp inbox.md inbox.md.backup
cp -r lib lib.backup

# 2. Reinitialize workspace
rm -rf .jot
jot init

# 3. Restore user content
cp inbox.md.backup inbox.md
cp -r lib.backup/* lib/

# 4. Verify integrity
jot doctor
```

### Lost Configuration

If configuration is lost:

```bash
# 1. Recreate minimal config
echo '{
  "workspaces": {"default": "~/notes"},
  "editor": "vim"
}' > ~/.jotrc

# 2. Test basic functionality
jot status

# 3. Rebuild configuration gradually
jot doctor --show-config
```

### Template Issues

If templates stop working:

```bash
# 1. Backup templates
cp -r .jot/templates .jot/templates.backup

# 2. Clear approvals
rm .jot/template_permissions

# 3. Re-approve templates one by one
jot template list
jot template approve template-name
```

## Getting Help

### Information Gathering

When reporting issues, include:

```bash
# Version information
jot --version

# System information
uname -a
echo $SHELL

# Configuration
jot doctor --show-config

# Workspace status
jot status --verbose

# Error messages (exact text)
jot problematic-command 2>&1
```

### Debug Mode

Enable verbose output:

```bash
# Verbose output
jot --verbose command

# JSON output for parsing
jot --json command

# Environment debug
JOT_DEBUG=1 jot command
```

### Community Resources

- **GitHub Issues:** [Report bugs and feature requests](https://github.com/johncoder/jot/issues)
- **Discussions:** [Community Q&A](https://github.com/johncoder/jot/discussions)
- **Documentation:** [Complete user guide](../README.md)

### Before Reporting Issues

1. **Update jot** to the latest version
2. **Run `jot doctor`** and include output
3. **Check existing issues** on GitHub
4. **Provide minimal reproduction** steps
5. **Include version and system information**

## Prevention

### Regular Maintenance

```bash
# Weekly maintenance routine
jot doctor                    # Check for issues
jot status                   # Review workspace state
jot refile                   # Organize notes
jot archive                  # Archive old content
```

### Backup Strategy

```bash
# Simple backup script
#!/bin/bash
DATE=$(date +%Y%m%d)
tar -czf "notes-backup-$DATE.tar.gz" inbox.md lib/ .jot/

# Or use git for versioning
git add -A && git commit -m "Daily backup: $(date)"
```

### Configuration Validation

```bash
# Add to your shell profile for validation
alias jot-check='jot doctor && echo "✓ jot workspace healthy"'
```

## See Also

- **[Command Reference](commands.md)** - Complete command documentation
- **[Configuration](configuration.md)** - Configuration file options
- **[Getting Started](getting-started.md)** - Initial setup guide
- **[Templates](templates.md)** - Template troubleshooting
