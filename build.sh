#!/bin/bash
set -euo pipefail

# Build sick-memory CLI - Default
go build -o sick-memory-default .
ls -lh sick-memory-default

# Build sick-memory CLI - Optimized (size + performance)
go build -ldflags "-s -w" -o sick-memory-optimized .
ls -lh sick-memory-optimized