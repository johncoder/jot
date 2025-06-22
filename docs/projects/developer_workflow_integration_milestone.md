# Developer Workflow Integration Milestone - June 22, 2025

## Executive Summary

This milestone focuses on seamless integration of jot into developer workflows, emphasizing zero-friction note-taking that works naturally alongside VS Code, Neovim, and zsh. The goal is to make jot feel like a native extension of your development environment rather than a separate tool.

**Priority**: High - Core productivity enhancement
**Approach**: Iterative implementation for a fun side project

## 🎯 Feature Overview

### 1. Configuration System (JSON5)

**Objective**: Flexible, hierarchical configuration supporting global and project-specific settings

#### Configuration Structure

```json5
{
  // Global defaults
  editor: "${EDITOR:-code}",
  workspace: {
    default_location: "~/Documents/notes",
    auto_create: true,
  },

  // Template defaults
  templates: {
    default_shell: "zsh",
    default_timeout: "30s",
  },

  // Capture behavior
  capture: {
    default_template: null,
    open_editor: true,
    timestamp_format: "2006-01-02 15:04:05",
  },

  // External commands
  external_commands: {
    enabled: true,
    search_paths: ["$HOME/.jot/bin", "/usr/local/bin"],
  },

  // Editor integrations
  integrations: {
    vscode: {
      enabled: true,
      workspace_detection: true,
    },
    nvim: {
      enabled: true,
      socket_path: "/tmp/jot-nvim.sock",
    },
  },
}
```

#### Configuration Hierarchy

1. **Environment Variables**: `JOT_EDITOR`, `JOT_WORKSPACE`, etc.
2. **Project Config**: `.jotrc` in workspace root
3. **Global Config**: `~/.jotrc`
4. **Built-in Defaults**: Sensible fallbacks

#### Implementation Plan

- Use `spf13/viper` for configuration management
- JSON5 parsing with fallback to regular JSON
- Environment variable substitution (`${VAR}` syntax)
- Configuration validation and helpful error messages

### 2. Editor Integration Enhancements

**Objective**: Seamless jot integration within VS Code and Neovim workflows

#### VS Code Extension Features

```typescript
// Primary commands
"jot.capture"; // Quick capture from selection or cursor
"jot.captureTemplate"; // Template-based capture
"jot.peek"; // View note content in sidebar
"jot.find"; // Search notes with results panel
"jot.refile"; // Refile current selection

// Workflow integrations
"jot.captureComment"; // Capture TODO/FIXME comments as notes
"jot.linkToLine"; // Create note linked to current file:line
"jot.insertReference"; // Insert reference to existing note
```

#### Extension Configuration

```json
{
  "jot.workspace": null, // Auto-detect or specify
  "jot.defaultTemplate": null, // Default template for captures
  "jot.openAfterCapture": false, // Open captured note
  "jot.autoRefile": true, // Auto-refile using templates
  "jot.keyBindings": {
    "capture": "ctrl+shift+j c",
    "peek": "ctrl+shift+j p",
    "find": "ctrl+shift+j f"
  }
}
```

#### Neovim Integration Features

```lua
-- Core commands
:JotCapture [template]      -- Capture with optional template
:JotPeek {selector}         -- View note in floating window
:JotFind {query}            -- Search with quickfix results
:JotRefile {destination}    -- Refile current buffer/selection

-- Advanced integrations
:JotCaptureVisual          -- Capture visual selection
:JotLinkToLine             -- Create note linked to current line
:JotInsertReference        -- Fuzzy-find and insert note reference
```

#### Neovim Configuration

```lua
require('jot').setup({
  workspace = nil,                    -- Auto-detect
  default_template = nil,
  keymaps = {
    capture = '<leader>jc',
    peek = '<leader>jp',
    find = '<leader>jf',
    refile = '<leader>jr'
  },
  ui = {
    float_opts = {
      border = 'rounded',
      width = 0.8,
      height = 0.6
    }
  }
})
```

### 3. External Command Support (Git-style)

**Objective**: Enable extensible command system for custom workflows and integrations

#### Implementation Scope

- **Command Discovery**: Automatic detection of `jot-<command>` executables in PATH
- **Execution Flow**: `jot mycommand` → searches for `jot-mycommand` → executes with remaining args
- **Fallback Handling**: Clear error messages when external commands aren't found
- **Documentation**: Guidelines for creating external commands

#### Technical Details

```go
// In cmd/root.go - modify unknown command handling
func handleUnknownCommand(cmd *cobra.Command, args []string) error {
    if len(args) > 0 {
        externalCmd := "jot-" + args[0]
        if path, err := exec.LookPath(externalCmd); err == nil {
            return exec.Command(path, args[1:]...).Run()
        }
    }
    return cmd.Help()
}
```

#### Use Cases

- `jot-daily` - Custom daily note creation
- `jot-project` - Project-specific note management
- `jot-sync` - Custom synchronization logic
- `jot-review` - Code review note workflows

## 📋 Implementation Phases

### Phase 1: Configuration System (Foundation)

**Solid foundation for all other features**

#### Core Configuration Implementation

- [ ] Implement JSON5 configuration parsing with viper
- [ ] Create default configuration structure
- [ ] Add environment variable support with substitution
- [ ] Implement configuration hierarchy (env → project → global → defaults)

#### Configuration Integration

- [ ] Update all commands to use configuration system
- [ ] Add `jot config` command for configuration management
- [ ] Implement configuration validation and error reporting
- [ ] Create example configuration files

#### Testing & Documentation

- [ ] Add comprehensive configuration tests
- [ ] Document all configuration options
- [ ] Test environment variable substitution
- [ ] Verify cross-platform compatibility

### Phase 2: Editor Integration (High Impact)

**Native integration for daily workflow**

#### VS Code Extension Development

- [ ] Create VS Code extension project structure
- [ ] Implement basic jot command execution
- [ ] Add configuration schema and settings
- [ ] Create basic capture command
- [ ] Implement capture with template selection
- [ ] Add peek command with document preview
- [ ] Create find command with results panel
- [ ] Add refile functionality
- [ ] Implement selection-based capture
- [ ] Add workspace detection and integration
- [ ] Create note reference insertion
- [ ] Add status bar integration

#### Neovim Plugin Development

- [ ] Create Neovim plugin structure (Lua)
- [ ] Implement jot command execution wrapper
- [ ] Add basic configuration handling
- [ ] Create core capture functionality
- [ ] Implement floating window for peek
- [ ] Add quickfix integration for find results
- [ ] Create fuzzy finder integration (telescope.nvim)
- [ ] Add template selection interface
- [ ] Implement visual selection capture
- [ ] Add line linking functionality
- [ ] Create note reference insertion with completion
- [ ] Add workspace detection

#### Polish & Documentation

- [ ] Add comprehensive error handling for both editors
- [ ] Implement proper logging and diagnostics
- [ ] Create extension/plugin documentation
- [ ] Test across different editor versions and configurations

### Phase 3: External Command Support (Extensibility)

**Enable custom workflows and community extensions**

#### External Command Discovery

- [ ] Implement external command discovery in PATH
- [ ] Add fallback handling for unknown commands
- [ ] Create command execution wrapper with proper error handling
- [ ] Test with sample external commands

#### Integration & Polish

- [ ] Add external command help integration
- [ ] Implement command completion for external commands
- [ ] Add configuration options for external command behavior
- [ ] Create documentation for external command development

## 🚀 Expected Benefits

### Developer Productivity

- **Zero-context switching**: Notes directly from your editor
- **Custom workflows**: External commands for project-specific needs
- **Consistent experience**: Same jot functionality across all tools
- **Reduced friction**: Quick capture without leaving development flow

### Configuration Flexibility

- **Project-specific settings**: Different templates and behaviors per project
- **Environment adaptation**: Automatic configuration based on environment
- **Extensibility**: Easy to add new integrations and workflows
- **Maintainability**: Clear, documented configuration options

### Editor Integration

- **Native experience**: Feels like built-in editor functionality
- **Workflow optimization**: Shortcuts and commands tailored to each editor
- **Visual feedback**: Proper UI integration with each editor's paradigms
- **Error handling**: Clear feedback when operations fail

## 📊 Success Criteria

### Technical Validation

- [ ] Configuration system supports all specified hierarchy levels
- [ ] External commands execute properly with argument passing
- [ ] VS Code extension passes marketplace requirements
- [ ] Neovim plugin works across different Neovim versions
- [ ] All integrations respect existing editor configurations

### User Experience Validation

- [ ] Configuration feels intuitive and well-documented
- [ ] External commands enable custom workflows without friction
- [ ] Editor integrations feel native to each environment
- [ ] Error messages provide clear guidance for resolution
- [ ] Setup process is simple and well-documented

### Integration Validation

- [ ] VS Code extension integrates with existing workspace settings
- [ ] Neovim plugin works with popular plugin managers
- [ ] External commands can access jot functionality programmatically
- [ ] Configuration changes apply immediately without restart
- [ ] All features work consistently across Linux, macOS, and Windows

## 🔮 Future Extension Opportunities

### Immediate Extensions (Next Milestone)

- **Shell completion enhancements**: Advanced zsh completion for all commands
- **Additional editor support**: Emacs, IntelliJ IDEA plugins
- **Workflow templates**: Pre-built external commands for common use cases
- **Integration testing**: Automated testing of editor integrations

### Long-term Possibilities

- **Language Server Protocol**: jot-lsp for advanced editor integrations
- **Workspace synchronization**: Multi-device note synchronization
- **Plugin ecosystem**: Community-driven external commands and integrations
- **Advanced automation**: Workflow automation based on development events

## 💡 Implementation Notes

### Technical Considerations

- **Configuration parsing**: Use go-json5 for robust JSON5 support
- **External command security**: Validate external command paths
- **Editor APIs**: Leverage each editor's extension APIs effectively
- **Cross-platform**: Ensure all features work on Linux, macOS, Windows

### Risk Mitigation

- **Incremental rollout**: Each phase builds on previous functionality
- **Backward compatibility**: Maintain existing command-line interface
- **Comprehensive testing**: Unit, integration, and user acceptance testing
- **Documentation**: Clear setup guides for each integration

This milestone transforms jot from a standalone tool into a natural extension of your development environment, enabling seamless note-taking without disrupting your coding flow.
