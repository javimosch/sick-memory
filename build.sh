#!/bin/bash
set -euo pipefail

# Ensure the script runs from the repository root.
cd "$(dirname "$(readlink -f "$0")")"
REPO_ROOT=$(pwd)

# Run tests and static analysis before producing release binaries
echo "Running tests..."
go test -race -count=1 ./...

echo "Running vet..."
go vet ./...

SMOKE_DIR=$(mktemp -d)
OPT_SMOKE_DIR=$(mktemp -d)
BRIDGE_DIR=$(mktemp -d)
trap 'rm -rf "$SMOKE_DIR" "$OPT_SMOKE_DIR" "$BRIDGE_DIR"' EXIT

# Build sick-memory CLI - Default
echo "Building sick-memory default..."
CGO_ENABLED=0 go build -trimpath -o sick-memory .
ls -lh sick-memory

# Smoke test the freshly built binary
echo "Smoke testing sick-memory default binary..."
./sick-memory --version

echo "Smoke testing core commands with temporary memory dir..."
./sick-memory init --memory-dir "$SMOKE_DIR"
MEMORY_ID=$(./sick-memory remember "Smoke test memory" --memory-dir "$SMOKE_DIR" | awk '{print $NF}')
./sick-memory recall "Smoke test" --memory-dir "$SMOKE_DIR"
./sick-memory list --memory-dir "$SMOKE_DIR"
./sick-memory edit "$MEMORY_ID" "Edited smoke test memory" --memory-dir "$SMOKE_DIR"
./sick-memory search "Edited smoke test" --memory-dir "$SMOKE_DIR"
./sick-memory delete "$MEMORY_ID" --memory-dir "$SMOKE_DIR"
./sick-memory status --memory-dir "$SMOKE_DIR"

# Smoke test the config command without polluting the real HOME directory
mkdir -p "$SMOKE_DIR/home"
HOME="$SMOKE_DIR/home" ./sick-memory config --memory-dir "$SMOKE_DIR"

# Smoke test the bridge command in an isolated directory to avoid cluttering the repo
(
  cd "$BRIDGE_DIR"
  "$REPO_ROOT/sick-memory" bridge claude-code --memory-dir "$SMOKE_DIR"
  [ -f .claude/CLAUDE.md ]
  "$REPO_ROOT/sick-memory" bridge opencode --memory-dir "$SMOKE_DIR"
  [ -f .opencode/memory.json ]
  "$REPO_ROOT/sick-memory" bridge copilot --memory-dir "$SMOKE_DIR"
  [ -f .copilot/settings.json ]
)

# Build sick-memory CLI - Optimized (size + performance)
echo "Building sick-memory optimized..."
CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o sick-memory-optimized .
ls -lh sick-memory-optimized

# Smoke test the optimized binary too
echo "Smoke testing sick-memory optimized binary..."
./sick-memory-optimized --version

echo "Smoke testing core commands with optimized binary..."
./sick-memory-optimized init --memory-dir "$OPT_SMOKE_DIR"
OPT_MEMORY_ID=$(./sick-memory-optimized remember "Smoke test optimized" --memory-dir "$OPT_SMOKE_DIR" | awk '{print $NF}')
./sick-memory-optimized recall "Smoke test" --memory-dir "$OPT_SMOKE_DIR"
./sick-memory-optimized list --memory-dir "$OPT_SMOKE_DIR"
./sick-memory-optimized edit "$OPT_MEMORY_ID" "Edited smoke test optimized" --memory-dir "$OPT_SMOKE_DIR"
./sick-memory-optimized search "Edited smoke test" --memory-dir "$OPT_SMOKE_DIR"
./sick-memory-optimized delete "$OPT_MEMORY_ID" --memory-dir "$OPT_SMOKE_DIR"
./sick-memory-optimized status --memory-dir "$OPT_SMOKE_DIR"

# Smoke test the config command with the optimized binary
mkdir -p "$OPT_SMOKE_DIR/home"
HOME="$OPT_SMOKE_DIR/home" ./sick-memory-optimized config --memory-dir "$OPT_SMOKE_DIR"
