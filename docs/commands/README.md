[Documentation](../README.md) > Commands

# Command Reference

Complete reference documentation for all jot commands and their options.

## Global Options

These options are available for all jot commands:

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--config FILE` | | Use custom configuration file | `~/.jotrc` |
| `--workspace NAME` | `-w` | Use specific workspace | auto-detect |
| `--json` | | Output in JSON format for automation ([reference](../reference/json-output.md)) | false |
| `--help` | `-h` | Show help information | |
| `--version` | | Show version information | |

## Core Commands

| Command | Description |
|---------|-------------|
| [jot init](jot-init.md) | Initialize a new workspace |
| [jot capture](jot-capture.md) | Capture notes with templates |
| [jot refile](jot-refile.md) | Move and organize notes |
| [jot find](jot-find.md) | Search workspace content |
| [jot archive](jot-archive.md) | Archive old notes |
| [jot status](jot-status.md) | Show workspace information |
| [jot doctor](jot-doctor.md) | Diagnose workspace issues |
| [jot template](jot-template.md) | Manage note templates |
| [jot workspace](jot-workspace.md) | Manage workspace registry |

## Advanced Commands

| Command | Description |
|---------|-------------|
| [jot eval](jot-eval.md) | Execute code blocks in notes |
| [jot tangle](jot-tangle.md) | Extract code from markdown |
| [jot peek](jot-peek.md) | Preview content and navigation |
| [jot files](jot-files.md) | Browse workspace files |
| [jot hooks](jot-hooks.md) | Manage hooks system |

## Utility Commands

All commands support JSON output with the `--json` flag. See the [JSON Output Reference](../reference/json-output.md) for format details and integration examples.

## External Commands

jot supports external commands that can be added to your PATH with the `jot-` prefix. See [External Commands](jot-external.md) for details on creating and using external commands, including environment variables and integration patterns.
