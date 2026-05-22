# May 2026 — Product Summary

## Semantic Search Engine
The biggest update this month: sick-memory now has a proper search engine with TF-IDF scoring, exact phrase boosting, and word-overlap fallback. Search across all your memories with natural language queries and get ranked results. The search index is cached for fast subsequent lookups.

## Edit & Delete Support
You can now edit existing memories and delete outdated ones directly from the CLI. No more manual file editing — `edit` and `delete` commands work non-interactively, making them agent-friendly.

## Agent-First Design
Every command now supports JSON output and requires no interactive prompts. sick-memory integrates naturally with Claude Code, OpenCode, and Copilot via bridge commands that auto-generate configuration files.

## Bug Fixes
- TF-IDF now correctly counts document frequency (was counting total occurrences, causing negative scores for common terms)
- `--json` flag no longer leaks into search queries
- Search index is now cached on recall, not just on remember
- Word-overlap fallback handles queries like "UI design" matching "UI/Design" or "disk cleanup" matching "disk-cleanup"
