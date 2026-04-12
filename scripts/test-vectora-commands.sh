#!/bin/bash
# test-vectora-commands.sh - Individual command testing with detailed validation
#
# Usage: ./scripts/test-vectora-commands.sh [command] [args...]
#
# Examples:
#   ./scripts/test-vectora-commands.sh vectora --help
#   ./scripts/test-vectora-commands.sh vectora ask "What is Rust?"
#   ./scripts/test-vectora-commands.sh vectora chat

set -e

VECTORA_BIN="${VECTORA_BIN:-./bin/vectora}"
COMMAND="${1:-help}"
ARGS=("${@:2}")
TIMEOUT=30

# Create test directory
mkdir -p test-results logs

# Get platform
PLATFORM=$(uname -s)

# Logging
log() {
  echo "[$(date +'%Y-%m-%d %H:%M:%S')] $*" | tee -a "logs/test-${COMMAND}.log"
}

# Initialize
log "╔════════════════════════════════════════════════════════════╗"
log "║ Vectora Command Tester                                     ║"
log "╚════════════════════════════════════════════════════════════╝"
log ""
log "Configuration:"
log "  Platform:     $PLATFORM"
log "  Vectora Bin:  $VECTORA_BIN"
log "  Command:      $COMMAND"
log "  Args:         ${ARGS[*]:-none}"
log "  Timeout:      ${TIMEOUT}s"
log ""

# Check binary
if [ ! -f "$VECTORA_BIN" ] && [ ! -f "${VECTORA_BIN}.exe" ]; then
  log "❌ ERROR: Vectora binary not found at $VECTORA_BIN"
  exit 1
fi

if [ -f "${VECTORA_BIN}.exe" ]; then
  VECTORA_BIN="${VECTORA_BIN}.exe"
fi

# Run command
log "Running command: $VECTORA_BIN $COMMAND ${ARGS[*]}"
log "─────────────────────────────────────────────────"

# Capture output and exit code
OUTPUT=$(timeout "$TIMEOUT" "$VECTORA_BIN" "$COMMAND" "${ARGS[@]}" 2>&1 || true)
EXIT_CODE=$?

# Check timeout
if [ $EXIT_CODE -eq 124 ]; then
  log "⏱ Command timed out after ${TIMEOUT}s"
elif [ $EXIT_CODE -eq 0 ]; then
  log "✓ Command executed successfully (exit code: 0)"
else
  log "❌ Command failed with exit code: $EXIT_CODE"
fi

# Output
log ""
log "Output:"
log "───────"
if [ -n "$OUTPUT" ]; then
  echo "$OUTPUT" | tee -a "logs/test-${COMMAND}.log"
else
  log "(empty output)"
fi

# Statistics
log ""
log "Statistics:"
log "  Output size:  $(echo -n "$OUTPUT" | wc -c) bytes"
log "  Lines:        $(echo -n "$OUTPUT" | wc -l)"
log "  Exit code:    $EXIT_CODE"

# Save to file
OUTPUT_FILE="test-results/${COMMAND}-output.txt"
echo "$OUTPUT" > "$OUTPUT_FILE"
log "  Saved to:    $OUTPUT_FILE"

log ""
log "✓ Test completed"

exit $EXIT_CODE
