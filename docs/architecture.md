# Architecture

## Storage Layout

```
~/.sick-memory/
  config.json              # Global configuration
  projects/
    <sanitized-git-root>/  # One directory per repository
      memory/
        MEMORY.md          # Index file (created by init)
        memory_<ts>.md     # Individual memory files
        search_index.json  # Cached TF-IDF search index
```

### Git-Based Scoping

When `sick-memory` runs inside a git repository, it extracts the repository
root path via `git rev-parse --show-toplevel`, sanitizes it (replacing `/`,
`\`, `:` with `-`, spaces with `_`), and uses it as the project subdirectory
under `~/.sick-memory/projects/`. All git worktrees of the same repository
share the same memory directory.

When not in a git repository, it falls back to a local `.sick-memory/`
directory in the current working directory.

## Memory File Format

Each memory is stored as a markdown file with YAML-style frontmatter:

```markdown
---
name: Memory <timestamp>
description: <first 50 chars of content>
type: user|feedback|project|reference
created: <unix timestamp>
---

<full memory content>
```

The filename is `memory_<unix-timestamp>.md`. Memory ID is derived by
stripping the `.md` extension (e.g., `memory_1718000000`).

## Search Algorithm

The search system uses a multi-layered relevance scoring approach:

### TF-IDF Core

Term Frequency-Inverse Document Frequency (TF-IDF) forms the base scoring
layer. Keywords are extracted by lowercasing, removing punctuation, and
filtering common English stop words. Words shorter than 3 characters are
excluded.

$$TF\text{-}IDF(t, d) = TF(t, d) \times \log\left(\frac{N+1}{DF(t)+1}\right)$$

### Scoring Layers

1. **TF-IDF score** — Sum of TF-IDF scores for matching query keywords
2. **Exact phrase boost** — +2.0 if the query appears verbatim in content
3. **Word-overlap fallback** — Ratio of matched keywords when TF-IDF is 0
4. **Recency multiplier** — 1.2 for <24h, 1.1 for <7d, 1.0 otherwise
5. **Type boost** — 1.15 for `project` type memories

### Caching

The search index is cached in `search_index.json` and rebuilt automatically
when a new memory is added (if `auto_index` is enabled).

## Configuration

Global configuration is stored at `~/.sick-memory/config.json`:

```json
{
  "default_memory_type": "user",
  "max_memory_size": 1048576,
  "auto_index": true
}
```

Created automatically on first run with sensible defaults.

## Agent Bridges

The `bridge` command generates agent-specific integration files:

| Agent | File | Format |
|-------|------|--------|
| Claude Code | `.claude/CLAUDE.md` | Markdown with `%bash` blocks |
| OpenCode | `.opencode/memory.json` | JSON configuration |
| Copilot | `.copilot/settings.json` | JSON configuration |

Each bridge file includes the memory directory path, usage examples, and
instructions for the target agent's configuration format.

## Exit Code Convention

- **0**: Success
- **1**: Generic failure
- **80–89**: Input/validation errors
- **90–99**: Resource/state errors
- **100–109**: Integration/external errors
- **110–119**: Internal software errors

## Implementation

The project was initially implemented in Zig but migrated to Go because Zig
0.16.0 had API compatibility issues with `std.process.argsAlloc` and
`std.os.argv`. Go was chosen following SuperCLI plugin recommendations for
better ecosystem support and API stability.
