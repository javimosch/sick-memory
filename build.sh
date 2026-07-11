#!/bin/bash
set -euo pipefail

# Ensure the script runs from the repository root.
cd "$(dirname "$0")"

# Run tests before producing release binaries
echo "Running tests..."
go test ./...

# Build sick-memory CLI - Default
echo "Building sick-memory default..."
go build -o sick-memory .
ls -lh sick-memory

# Build sick-memory CLI - Optimized (size + performance)
echo "Building sick-memory optimized..."
go build -ldflags "-s -w" -o sick-memory-optimized .
ls -lh sick-memory-optimized
