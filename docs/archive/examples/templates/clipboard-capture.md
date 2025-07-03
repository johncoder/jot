# Clipboard Capture Template

Capture content directly from your system clipboard with automatic formatting and context.

## Template File

Save as `.jot/templates/clipboard.md`:

```markdown
---
destination: inbox.md
refile_mode: append
tags: [clipboard, $(date '+%Y-%m-%d')]
---

## Clipboard Capture - $(date '+%Y-%m-%d %H:%M')

**Source:** $(if command -v pbpaste >/dev/null; then echo "macOS Clipboard"; elif command -v xclip >/dev/null; then echo "Linux Clipboard (xclip)"; elif command -v wl-paste >/dev/null; then echo "Wayland Clipboard"; else echo "Manual"; fi)
**Context:** $(pwd | basename)

### Content

$(if command -v pbpaste >/dev/null; then pbpaste; elif command -v xclip >/dev/null; then xclip -o -selection clipboard; elif command -v wl-paste >/dev/null; then wl-paste; else echo "[Paste clipboard content here]"; fi)

### Notes

### Tags

<!-- Add relevant tags for organization -->

#clipboard

---
```

## Usage

```bash
# Approve template first
jot template approve clipboard

# Copy something to clipboard, then:
jot capture clipboard

# Quick capture with additional context
jot capture clipboard --content "From Stack Overflow discussion about React hooks"
```

## Platform-Specific Versions

### macOS Version

Save as `.jot/templates/clipboard-mac.md`:

```markdown
---
destination: lib/clips/$(date '+%Y-%m').md
refile_mode: append
tags: [clipboard, mac]
---

## Clipboard - $(date '+%Y-%m-%d %H:%M')

**App:** $(osascript -e 'tell application "System Events" to get name of first process whose frontmost is true' 2>/dev/null || echo "Unknown")
**URL:** $(if echo "$(pbpaste)" | grep -E '^https?://'; then pbpaste | head -1; else echo "N/A"; fi)

### Content
```

$(pbpaste)

```

### Context
- **Directory:** $(pwd)
- **Git Branch:** $(git branch --show-current 2>/dev/null || echo "N/A")
- **Clipboard Size:** $(pbpaste | wc -c) characters

### Processing Notes


---
```

### Linux Version

Save as `.jot/templates/clipboard-linux.md`:

```markdown
---
destination: lib/clips/$(date '+%Y-%m').md
refile_mode: append
tags: [clipboard, linux]
---

## Clipboard - $(date '+%Y-%m-%d %H:%M')

**Window:** $(xdotool getactivewindow getwindowname 2>/dev/null || echo "Unknown")
**Type:** $(xclip -o -selection clipboard | file - | cut -d: -f2- || echo "text")

### Content
```

$(xclip -o -selection clipboard 2>/dev/null || echo "[No clipboard content available]")

```

### Metadata
- **Size:** $(xclip -o -selection clipboard 2>/dev/null | wc -c || echo "0") characters
- **Lines:** $(xclip -o -selection clipboard 2>/dev/null | wc -l || echo "0") lines
- **Format:** $(xclip -o -selection clipboard 2>/dev/null | head -1 | file - | grep -o 'text\|data' || echo "unknown")

### Notes


---
```

### Windows Version (WSL/Git Bash)

Save as `.jot/templates/clipboard-win.md`:

```markdown
---
destination: lib/clips/$(date '+%Y-%m').md
refile_mode: append
tags: [clipboard, windows]
---

## Clipboard - $(date '+%Y-%m-%d %H:%M')

**System:** $(uname -s)
**Shell:** $(basename $SHELL)

### Content
```

$(if command -v clip.exe >/dev/null; then powershell.exe Get-Clipboard; elif command -v xclip >/dev/null; then xclip -o; else echo "[Paste clipboard content manually]"; fi)

```

### Context
- **Working Directory:** $(pwd)
- **Timestamp:** $(date --iso-8601=seconds)

### Notes


---
```

## Smart Clipboard Template

A more advanced version that detects content type and formats accordingly:

Save as `.jot/templates/smart-clipboard.md`:

````markdown
---
destination: lib/clips/$(date '+%Y-%m').md
refile_mode: append
tags: [clipboard, smart, $(date '+%Y-%m-%d')]
---

## Smart Clipboard - $(date '+%Y-%m-%d %H:%M')

$(

# Get clipboard content

if command -v pbpaste >/dev/null; then
CLIP_CONTENT=$(pbpaste)
elif command -v xclip >/dev/null; then
    CLIP_CONTENT=$(xclip -o -selection clipboard)
elif command -v wl-paste >/dev/null; then
CLIP_CONTENT=$(wl-paste)
else
CLIP_CONTENT="[Manual paste required]"
fi

# Detect content type and format accordingly

if echo "$CLIP_CONTENT" | grep -qE '^https?://'; then
    echo "**Type:** URL"
    echo "**Link:** $CLIP_CONTENT"
    echo ""
    echo "### URL Analysis"
    echo "- **Domain:** $(echo "$CLIP*CONTENT" | sed 's|https\?://||' | cut -d'/' -f1)"
echo "- **Protocol:** $(echo "$CLIP_CONTENT" | grep -oE '^https?')"
echo ""
elif echo "$CLIP_CONTENT" | grep -qE '^\{.*\}$|^\[.\*\]$'; then
    echo "**Type:** JSON"
    echo ""
    echo "### JSON Content"
    echo '```json'
    echo "$CLIP_CONTENT"
echo '```'
echo ""
elif echo "$CLIP_CONTENT" | grep -qE '^[a-zA-Z0-9.*%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$'; then
    echo "**Type:** Email Address"
    echo "**Email:** $CLIP_CONTENT"
    echo ""
elif echo "$CLIP_CONTENT" | grep -qE '^[0-9]{3}-?[0-9]{3}-?[0-9]{4}$'; then
    echo "**Type:** Phone Number"
    echo "**Phone:** $CLIP_CONTENT"
    echo ""
elif echo "$CLIP_CONTENT" | grep -qE '^[a-f0-9]{40}$|^[a-f0-9]{64}$'; then
echo "**Type:** Hash/Commit ID"
echo "**Hash:** $CLIP_CONTENT"
    echo ""
elif echo "$CLIP_CONTENT" | wc -w | grep -qE '^[1-5]$'; then
    echo "**Type:** Short Text/Command"
    echo "**Content:** \`$CLIP_CONTENT\`"
echo ""
else
echo "**Type:** Text Content"
echo "**Length:** $(echo "$CLIP_CONTENT" | wc -c) characters"
echo "**Lines:** $(echo "$CLIP_CONTENT" | wc -l) lines"
echo ""
echo "### Content"
echo '`'
    echo "$CLIP_CONTENT"
    echo '`'
echo ""
fi
)

### Context

- **Source Application:** $(if command -v osascript >/dev/null; then osascript -e 'tell application "System Events" to get name of first process whose frontmost is true' 2>/dev/null; elif command -v xdotool >/dev/null; then xdotool getactivewindow getwindowname 2>/dev/null; else echo "Unknown"; fi)
- **Working Directory:** $(basename $(pwd))
- **Git Context:** $(git branch --show-current 2>/dev/null || echo "Not a git repository")

### Notes

### Follow-up Actions

- [ ] Review and categorize
- [ ] Add to relevant project
- [ ] Share with team
- [ ] Archive after use

---
````

## Usage Patterns

### Quick Clipboard Capture

```bash
# Basic clipboard capture
jot capture clipboard

# With immediate context
jot capture clipboard --content "Research for the authentication bug"

# Smart formatting
jot capture smart-clipboard
```

### Integration with Browser/Apps

```bash
# Create shell alias for quick access
alias jclip='jot capture clipboard'
alias jurl='jot capture clipboard --content "Reference URL for current research"'

# Keyboard shortcut integration (macOS)
# In Automator, create service that runs:
# /usr/local/bin/jot capture clipboard --content "Quick capture via hotkey"
```

### Advanced Workflows

```bash
# Capture clipboard and immediately refile
jot capture clipboard && jot refile --last --dest research.md

# Capture with automatic tagging based on content
if pbpaste | grep -q "github.com"; then
    jot capture clipboard --content "GitHub reference"
elif pbpaste | grep -q "stackoverflow.com"; then
    jot capture clipboard --content "Stack Overflow solution"
else
    jot capture clipboard
fi
```

## Shell Integration Examples

### Bash/Zsh Functions

Add to your `.bashrc` or `.zshrc`:

```bash
# Quick clipboard capture
jclip() {
    local context="${1:-Quick clipboard capture}"
    jot capture clipboard --content "$context"
}

# URL-specific capture
jurl() {
    if command -v pbpaste >/dev/null; then
        local url=$(pbpaste)
    elif command -v xclip >/dev/null; then
        local url=$(xclip -o -selection clipboard)
    fi

    if echo "$url" | grep -qE '^https?://'; then
        jot capture --content "## Reference URL
**URL:** $url
**Context:** ${1:-Web research}
**Date:** $(date)"
    else
        echo "Clipboard doesn't contain a URL"
    fi
}

# Code snippet capture
jcode() {
    local lang="${1:-text}"
    local desc="${2:-Code snippet}"
    jot capture --content "## Code Snippet: $desc
**Language:** $lang
**Source:** Clipboard
**Date:** $(date)

\`\`\`$lang
$(if command -v pbpaste >/dev/null; then pbpaste; else xclip -o -selection clipboard; fi)
\`\`\`

### Notes
"
}
```

### Alfred/Raycast Integration (macOS)

Alfred workflow script:

```bash
#!/bin/bash
# Alfred workflow for clipboard capture
query="$1"
content=$(pbpaste)

if [ -z "$query" ]; then
    query="Quick capture via Alfred"
fi

/usr/local/bin/jot capture clipboard --content "$query"
echo "Captured clipboard content: $query"
```

## Troubleshooting

### Clipboard Access Issues

```bash
# Linux: Install clipboard tools
sudo apt-get install xclip  # X11
sudo apt-get install wl-clipboard  # Wayland

# macOS: Should work out of the box with pbpaste
which pbpaste

# Windows (WSL): Enable clipboard integration
echo "alias pbcopy='clip.exe'" >> ~/.bashrc
echo "alias pbpaste='powershell.exe Get-Clipboard'" >> ~/.bashrc
```

### Template Approval

```bash
# If template fails to execute
jot template list
jot template approve clipboard

# Check clipboard access manually
pbpaste  # macOS
xclip -o -selection clipboard  # Linux
```

### Content Formatting Issues

```bash
# For binary content, use file detection
file <(pbpaste)

# For large content, limit capture
pbpaste | head -100 | jot capture --content "Large clipboard content (truncated)"
```

## Security Considerations

- **Review clipboard content** before approving templates
- **Avoid capturing sensitive data** like passwords or API keys
- **Use template approval** to prevent automatic execution of clipboard commands
- **Consider clipboard history** - some tools save clipboard history

## See Also

- **[Templates Guide](../../user-guide/templates.md)** - Template system basics
- **[Basic Workflows](../../user-guide/basic-workflows.md)** - Integration patterns
- **[Daily Notes](../daily-notes.md)** - Using clipboard capture in daily workflows
