# Command Reference

## Global Options

| Flag | Shorthand | Description |
|------|-----------|-------------|
| `--json` | `-j` | Output in JSON format |
| `--no-interactive` | `-y` | Disable all prompts |
| `--memory-dir <dir>` | | Override memory directory |

## Commands

### init

Initialize the memory system for the current project.

```bash
sick-memory init
```

Creates the memory directory and a `MEMORY.md` index file. Uses git-based
scoping to isolate memories by repository root. Falls back to local
`.sick-memory/` if not in a git repository.

### remember

Store a new memory. Alias: `keep`.

```bash
sick-memory remember "Use real database instances in tests, not mocks"
sick-memory remember --type project "API uses JWT tokens with 24h expiration"
```

The `--type` flag accepts: `user`, `feedback`, `project`, `reference`.

Memories are stored as markdown files with YAML frontmatter at
`~/.sick-memory/projects/<sanitized-git-root>/memory/memory_<timestamp>.md`.

### recall

Search memories by relevance. Alias: `search`.

```bash
sick-memory recall "database"
sick-memory recall "authentication" --json
```

Uses TF-IDF scoring with exact phrase boost (+2.0), word-overlap fallback,
recency multipliers (1.2 for <24h, 1.1 for <7d), and project-type boost
(1.15). Results are sorted by score descending.

Without a query, returns all memories sorted by recency.

### list

List all memory IDs. Alias: `ls`.

```bash
sick-memory list
sick-memory list --json
```

### edit

Update a memory by ID.

```bash
sick-memory edit <memory-id> "Updated content here"
```

The ID can be the memory filename, the `memory_` prefixed version, or
just the numeric timestamp portion. Updates the frontmatter description
to match the new content.

### delete

Remove a memory by ID.

```bash
sick-memory delete <memory-id>
```

Same ID matching logic as `edit`.

### status

Show the current memory system status.

```bash
sick-memory status
sick-memory status --json
```

Displays whether the system is initialized or uninitialized, the memory
directory path, and total memory count.

### config

Show configuration and storage location.

```bash
sick-memory config
sick-memory config --json
```

Displays the global directory, per-project memory directory, project root,
storage mode (centralized or local), and global configuration values
(default memory type, max size, auto-index toggle).

### bridge

Generate agent-specific integration files.

```bash
sick-memory bridge claude-code   # Creates .claude/CLAUDE.md
sick-memory bridge opencode      # Creates .opencode/memory.json
sick-memory bridge copilot       # Creates .copilot/settings.json
```

Available agents: `claude-code`, `opencode`, `copilot`.

## Help and Version

```bash
sick-memory --help        # Show full help
sick-memory --version     # Show version string
sick-memory help          # Same as --help
sick-memory version       # Same as --version
```

## Exit Codes

| Code | Category | Description |
|------|----------|-------------|
| 0 | Success | Command completed successfully |
| 1 | Generic | Generic failure |
| 80 | Input | Missing argument |
| 85 | Input | Invalid argument or unknown command |
| 92 | Resource | Resource not found, memory directory missing |
| 110 | Internal | Read/write error or not implemented |
