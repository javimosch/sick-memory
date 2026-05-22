# Sick-Memory Integration for Claude Code

This file provides project-specific memory for Claude Code via sick-memory CLI.

## Memory Loading

To load memories at session start, add this to your Claude Code workflow:

%bash
# Load relevant memories
sick-memory recall --json
%

## Adding Memories

%bash
# Add a memory
sick-memory remember "Use real database instances in tests, not mocks"

# Add with type
sick-memory remember --type feedback "Integration tests must hit real DB"
%

## Memory Location

Memories are stored in: .sick-memory

## Bridge Commands

The following bridge commands are available:
- /sm - Access sick-memory functionality
- /sm remember <content> - Add a memory
- /sm recall [query] - Retrieve memories
- /sm status - Check memory system status
