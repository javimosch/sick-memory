#!/bin/bash
set -euo pipefail

# Ensure the script runs from the repository root.
cd "$(dirname "$(readlink -f "$0")")"

# Run tests and static analysis before producing release binaries
echo "Running tests..."
go test -count=1 ./...

echo "Running vet..."
go vet ./...

# Build sick-memory CLI - Default
echo "Building sick-memory default..."
CGO_ENABLED=0 go build -trimpath -o sick-memory .
ls -lh sick-memory

# Smoke test the freshly built binary
echo "Smoke testing sick-memory default binary..."
./sick-memory --version

# Build sick-memory CLI - Optimized (size + performance)
echo "Building sick-memory optimized..."
CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o sick-memory-optimized .
ls -lh sick-memory-optimized
