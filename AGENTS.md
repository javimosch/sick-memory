# sick-memory

File-based memory system for AI coding agents with centralized storage, git-based scoping, and worktree support.

## Build

```bash
go build -o sick-memory .
```

## Run Tests

```bash
go test ./...
```

## Bridge Configs

Generate agent integration files:

```bash
sick-memory bridge claude-code    # -> .claude/CLAUDE.md
sick-memory bridge opencode        # -> .opencode/memory.json
sick-memory bridge copilot         # -> .copilot/settings.json
```

## Memory Commands

```bash
sick-memory init
sick-memory remember "<content>"
sick-memory recall [query]
sick-memory list
sick-memory status
sick-memory config
```

## Memory Types

- **user**: Information about the person (role, goals, expertise)
- **feedback**: Corrections and confirmations about work approach
- **project**: Ongoing work context (who, what, why, when)
- **reference**: External system pointers (URLs, dashboard links)

## JSON Output

Add `--json` flag to any command for machine-readable output:

```bash
sick-memory recall "query" --json
sick-memory status --json
sick-memory list --json
```
