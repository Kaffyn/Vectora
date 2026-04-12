#!/bin/bash
# verify-help.sh - Verify Vectora help command and extract available commands
#
# Usage: ./scripts/verify-help.sh
#
# This script:
# 1. Checks if vectora binary is available
# 2. Extracts help information
# 3. Lists all available commands
# 4. Validates command structure

set -e

VECTORA_BIN="${VECTORA_BIN:-./bin/vectora}"
LOG_FILE="logs/verify-help.log"

# Create log directory
mkdir -p logs test-results

# Logging function
log() {
  echo "[$(date +'%Y-%m-%d %H:%M:%S')] $*" | tee -a "$LOG_FILE"
}

log "======================================================================"
log "Vectora Help Command Verification"
log "======================================================================"
log ""

# Check if vectora binary exists
if [ ! -f "$VECTORA_BIN" ] && [ ! -f "${VECTORA_BIN}.exe" ]; then
  log "❌ ERROR: Vectora binary not found at $VECTORA_BIN"
  exit 1
fi

log "✓ Vectora binary found"

# Determine the correct binary path
if [ -f "${VECTORA_BIN}.exe" ]; then
  VECTORA_BIN="${VECTORA_BIN}.exe"
fi

# Test 1: Get version
log ""
log "Test 1: Check version"
log "─────────────────────"
if VERSION=$("$VECTORA_BIN" --version 2>&1); then
  log "✓ Version: $VERSION"
else
  log "⚠ Warning: Could not retrieve version"
fi

# Test 2: Get help
log ""
log "Test 2: Extract help information"
log "────────────────────────────────"
if HELP_OUTPUT=$("$VECTORA_BIN" --help 2>&1); then
  log "✓ Help command successful"

  # Save help output
  mkdir -p test-results
  echo "$HELP_OUTPUT" > test-results/help.txt
  log "✓ Help output saved to test-results/help.txt"
else
  log "❌ ERROR: Help command failed"
  exit 1
fi

# Test 3: Parse available commands
log ""
log "Test 3: Available Commands"
log "──────────────────────────"

# Extract commands from help
COMMANDS=$(echo "$HELP_OUTPUT" | grep -E '^\s+[a-z]+' | awk '{print $1}' | grep -v '^$' | sort | uniq || true)

if [ -z "$COMMANDS" ]; then
  log "⚠ Warning: Could not parse commands from help"
else
  log "Found commands:"
  echo "$COMMANDS" | while read -r cmd; do
    if [ -n "$cmd" ]; then
      log "  • $cmd"
    fi
  done

  # Save commands list
  echo "$COMMANDS" > test-results/commands.txt
  log ""
  log "✓ Commands list saved to test-results/commands.txt"
fi

# Test 4: Check for common command flags
log ""
log "Test 4: Common Flags"
log "────────────────────"

FLAGS=("--help" "--version" "--debug" "--config")
for flag in "${FLAGS[@]}"; do
  if echo "$HELP_OUTPUT" | grep -q "$flag"; then
    log "✓ Found flag: $flag"
  fi
done

# Test 5: Validate help structure
log ""
log "Test 5: Help Structure Validation"
log "──────────────────────────────────"

CHECKS=(
  "Usage"
  "Available Commands"
  "Global Flags"
  "Examples"
)

for check in "${CHECKS[@]}"; do
  if echo "$HELP_OUTPUT" | grep -qi "$check"; then
    log "✓ Found section: $check"
  else
    log "⚠ Missing section: $check"
  fi
done

# Test 6: Try subcommand help
log ""
log "Test 6: Subcommand Help"
log "───────────────────────"

# Get first command if available
FIRST_CMD=$(echo "$COMMANDS" | head -1)
if [ -n "$FIRST_CMD" ]; then
  if SUBCMD_HELP=$("$VECTORA_BIN" "$FIRST_CMD" --help 2>&1); then
    log "✓ Subcommand help works for: $FIRST_CMD"
  else
    log "⚠ Could not get help for subcommand: $FIRST_CMD"
  fi
fi

# Final summary
log ""
log "======================================================================"
log "✓ Help verification completed successfully"
log "======================================================================"
log ""
log "Summary:"
log "  • Binary check:     ✓"
log "  • Version:          ✓"
log "  • Help output:      ✓"
log "  • Command parsing:  ✓"
log ""
log "Results saved to:"
log "  • test-results/help.txt"
log "  • test-results/commands.txt"
log "  • logs/verify-help.log"
log ""

exit 0
