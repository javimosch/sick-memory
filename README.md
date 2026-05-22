# sick-memory

File-based memory system for AI coding agents with centralized storage, git-based scoping, and worktree support. Inspired by Claude Code's memory system, sick-memory provides persistent knowledge storage across sessions using markdown files with YAML frontmatter.

## Features

- **Centralized Storage**: All memories stored in `~/.sick-memory/` with git-based project scoping
- **Git-based Scoping**: Memory automatically scoped to git repository root
- **Worktree Support**: All git worktrees of the same repository share memory directory
- **Global Configuration**: Centralized config file for user preferences
- **File-based Storage**: Markdown files with YAML frontmatter for human readability
- **Agent-agnostic**: Works with Claude Code, OpenCode, Copilot, and other AI agents
- **JSON Output**: Machine-readable output for agent integration
- **Bridge Commands**: Generate agent-specific configurations automatically
- **Zero Infrastructure**: No database or server required, works offline

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/javimosch/sick-memory.git
cd sick-memory

# Build the binary
go build -o sick-memory .

# Install to local bin
cp sick-memory ~/.local/bin/
```

### SuperCLI Plugin

```bash
# Install the SuperCLI plugin
cd ~/ai/supercli
sc plugins install ./plugins/sick-memory --on-conflict replace --json

# Sync skills to make the skill available
sc skills sync
```

## Usage

### Basic Commands

```bash
# Initialize memory system for current project
sick-memory init

# Add a memory
sick-memory remember "Use real database instances in tests, not mocks"

# Retrieve memories
sick-memory recall "database"

# List all memories
sick-memory list

# Show system status
sick-memory status

# Show configuration and storage location
sick-memory config
```

### Memory Types

The system supports four memory types (following Claude Code's taxonomy):

- **user**: Information about the person (role, goals, expertise)
- **feedback**: Corrections and confirmations about work approach
- **project**: Ongoing work context (who, what, why, when)
- **reference**: External system pointers (URLs, dashboard links)

```bash
# Add memory with specific type
sick-memory remember "API uses JWT tokens with 24h expiration" --type project
```

### JSON Output

Add `--json` flag to any command for structured output:

```bash
sick-memory recall "query" --json
sick-memory status --json
sick-memory list --json
sick-memory config --json
```

## Storage Architecture

### Centralized Storage

- **Location**: `~/.sick-memory/projects/<sanitized-git-root>/memory/`
- **Git-based Scoping**: Memory automatically scoped to git repository root
- **Worktree Support**: All git worktrees share the same memory directory
- **Fallback**: Uses local `.sick-memory/` directory if not in a git repository

### Global Configuration

Global configuration is stored in `~/.sick-memory/config.json`:

- **default_memory_type**: Default memory type (user, feedback, project, reference)
- **max_memory_size**: Maximum memory file size in bytes
- **auto_index**: Whether to automatically update memory index

## Agent Integration

### Claude Code

```bash
sick-memory bridge claude-code
```

Creates `.claude/CLAUDE.md` with memory loading instructions and centralized storage information.

### OpenCode

```bash
sick-memory bridge opencode
```

Creates `.opencode/memory.json` with OpenCode configuration.

### Copilot

```bash
sick-memory bridge copilot
```

Creates `.copilot/settings.json` with Copilot configuration.

## SuperCLI Plugin

When installed as a SuperCLI plugin, sick-memory provides enhanced commands:

```bash
# SuperCLI plugin commands
sc sick-memory init
sc sick-memory remember "<content>"
sc sick-memory recall [query]
sc sick-memory list
sc sick-memory status
sc sick-memory config show
sc sick-memory bridge generate <agent>

# Quick access via /sm shorthand
/sm init
/sm remember "<content>"
/sm recall [query]
```

### Learning the Skill

```bash
# Get comprehensive sick-memory documentation
sc skills get sick-memory:quickstart

# Search for the skill
sc skills search --query sick-memory
```

## Development

### Building

```bash
# Build default binary
go build -o sick-memory .

# Build optimized binary (smaller size)
go build -ldflags "-s -w" -o sick-memory .

# Using the build script
./build.sh
```

### Testing

```bash
# Run tests
go test ./...

# Test with coverage
go test -cover ./...
```

## Architecture

Sick-memory follows a clean room implementation inspired by Claude Code's memory system:

- **Storage**: Markdown files with YAML frontmatter
- **Scoping**: Git-based project isolation
- **Configuration**: JSON-based global config
- **Output**: Human-readable text with optional JSON
- **Exit Codes**: Semantic exit codes for programmatic use

## License

MIT License - see LICENSE file for details

## Author

Javier Leandro Arancibia

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.