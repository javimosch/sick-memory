# Sick-Memory Integration for Claude Code

This file provides project-specific memory for Claude Code via sick-memory CLI.

## Memory Loading

To load memories at session start, add this to your Claude Code workflow:

```bash
# Load relevant memories
sick-memory recall --json
```

## Adding Memories

```bash
# Add a memory
sick-memory remember "Use real database instances in tests, not mocks"

# Add with type
sick-memory remember --type feedback "Integration tests must hit real DB"
```

## Editing and Deleting

```bash
# Edit a memory by ID
sick-memory edit <id> "Updated content"

# Delete a memory by ID
sick-memory delete <id>
```

## Listing and Searching

```bash
# List all memories
sick-memory list

# Search with natural language
sick-memory recall "database configuration"
```

## Aliases

```bash
# Shorthand equivalents
sick-memory keep <content>     # alias for remember
sick-memory search <query>     # alias for recall
sick-memory ls                 # alias for list
```

## JSON Output

Add `--json` to any command for machine-readable output:

```bash
sick-memory recall "query" --json
sick-memory status --json
sick-memory list --json
sick-memory config --json
```

## Memory Location

Memories are stored in: `~/.sick-memory/projects/<sanitized-git-root>/memory/`

## Storage Mode

This project uses centralized storage with git-based scoping.
All git worktrees of this repository share the same memory directory.

## Bridge Commands

The following bridge commands are available:
- `/sm` - Access sick-memory functionality
- `/sm remember <content>` - Add a memory
- `/sm recall [query]` - Retrieve memories
- `/sm list` - List all memories
- `/sm edit <id>` - Edit a memory by ID
- `/sm delete <id>` - Delete a memory by ID
- `/sm status` - Check memory system status
- `/sm config` - Show configuration and storage location
