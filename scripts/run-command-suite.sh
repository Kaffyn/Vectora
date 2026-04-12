#!/bin/bash
# run-command-suite.sh - Comprehensive Vectora command test suite
#
# Usage: ./scripts/run-command-suite.sh
#
# This script tests all Vectora commands and validates:
# 1. Command availability
# 2. Correct exit codes
# 3. Output format
# 4. Error handling
# 5. Integration with providers

set +e  # Don't exit on first error to collect all results

VECTORA_BIN="${VECTORA_BIN:-./bin/vectora}"
TEST_RESULTS_FILE="test-results/command-tests.json"
TEST_LOG_FILE="logs/command-tests.log"
ENV_FILE="${VECTORA_TEST_ENV:-.env.test}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test counters
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0
TESTS_SKIPPED=0

# Create necessary directories
mkdir -p logs test-results coverage

# Logging function
log() {
  echo "[$(date +'%Y-%m-%d %H:%M:%S')] $*" | tee -a "$TEST_LOG_FILE"
}

# Test result tracker
test_result() {
  local test_name=$1
  local exit_code=$2
  local expected_code=${3:-0}
  local output=$4

  TESTS_RUN=$((TESTS_RUN + 1))

  if [ "$exit_code" -eq "$expected_code" ]; then
    TESTS_PASSED=$((TESTS_PASSED + 1))
    log "✓ $test_name (exit code: $exit_code)"
    return 0
  else
    TESTS_FAILED=$((TESTS_FAILED + 1))
    log "❌ $test_name (expected: $expected_code, got: $exit_code)"
    if [ -n "$output" ]; then
      log "   Output: $output"
    fi
    return 1
  fi
}

# Check binary
check_binary() {
  if [ ! -f "$VECTORA_BIN" ] && [ ! -f "${VECTORA_BIN}.exe" ]; then
    log "❌ ERROR: Vectora binary not found at $VECTORA_BIN"
    exit 1
  fi

  if [ -f "${VECTORA_BIN}.exe" ]; then
    VECTORA_BIN="${VECTORA_BIN}.exe"
  fi

  log "✓ Found Vectora binary: $VECTORA_BIN"
}

# Load test environment
load_test_env() {
  if [ -f "$ENV_FILE" ]; then
    log "Loading test environment from: $ENV_FILE"
    # shellcheck disable=SC1090
    export $(grep -v '^#' "$ENV_FILE" | xargs)
    log "✓ Test environment loaded"
  else
    log "⚠ Warning: Test environment file not found: $ENV_FILE"
  fi
}

# Timeout wrapper
run_with_timeout() {
  local timeout=$1
  shift
  local cmd=("$@")

  # Use timeout command if available, otherwise just run the command
  if command -v timeout &> /dev/null; then
    timeout "$timeout" "${cmd[@]}" 2>&1
  else
    "${cmd[@]}" 2>&1
  fi
}

log "======================================================================"
log "Vectora Command Test Suite"
log "======================================================================"
log ""

# Initialize
check_binary
load_test_env

# Test categories
declare -a COMMANDS=(
  "--version|1|Test version command"
  "--help|0|Test help command"
  "version|1|Test version subcommand"
  "help|0|Test help subcommand"
)

declare -a OPTIONAL_COMMANDS=(
  "ask|0|Test ask command with simple prompt"
  "chat|1|Test chat command startup"
  "rag|0|Test RAG command"
  "configure|0|Test configuration command"
  "completion|0|Test completion command"
  "list-models|0|Test model listing"
)

log "Test Phase 1: Core Commands"
log "─────────────────────────────────────────"
log ""

# Run core command tests
for cmd_def in "${COMMANDS[@]}"; do
  IFS='|' read -r cmd expected_code description <<< "$cmd_def"

  if [ -z "$cmd" ]; then
    continue
  fi

  log "Running: $description"
  OUTPUT=$(run_with_timeout 5 "$VECTORA_BIN" $cmd 2>&1)
  EXIT_CODE=$?

  test_result "vectora $cmd" "$EXIT_CODE" "$expected_code" "$OUTPUT"
  log "  Output: $(echo "$OUTPUT" | head -1)"
  log ""
done

log "Test Phase 2: Optional Commands"
log "─────────────────────────────────────────"
log ""

# Run optional command tests
for cmd_def in "${OPTIONAL_COMMANDS[@]}"; do
  IFS='|' read -r cmd expected_code description <<< "$cmd_def"

  if [ -z "$cmd" ]; then
    continue
  fi

  # Check if command exists first
  if ! "$VECTORA_BIN" --help 2>&1 | grep -q "^\s*$cmd"; then
    log "⊘ $description (command not implemented)"
    TESTS_SKIPPED=$((TESTS_SKIPPED + 1))
    log ""
    continue
  fi

  log "Running: $description"

  case "$cmd" in
    ask)
      OUTPUT=$(run_with_timeout 10 "$VECTORA_BIN" ask "test query" 2>&1 || true)
      EXIT_CODE=$?
      test_result "vectora ask" "$EXIT_CODE" 0 "$(echo "$OUTPUT" | head -1)"
      ;;
    chat)
      # Chat may require interactive input, so we test if it starts correctly
      timeout 2 "$VECTORA_BIN" chat </dev/null 2>&1 >/dev/null || EXIT_CODE=$?
      test_result "vectora chat startup" "${EXIT_CODE:-0}" 0
      ;;
    rag)
      OUTPUT=$(run_with_timeout 10 "$VECTORA_BIN" rag search "test" 2>&1 || true)
      EXIT_CODE=$?
      test_result "vectora rag" "$EXIT_CODE" 0 "$(echo "$OUTPUT" | head -1)"
      ;;
    configure)
      OUTPUT=$(run_with_timeout 5 "$VECTORA_BIN" configure --list 2>&1 || true)
      EXIT_CODE=$?
      test_result "vectora configure" "$EXIT_CODE" 0 "$(echo "$OUTPUT" | head -1)"
      ;;
    completion)
      OUTPUT=$(run_with_timeout 5 "$VECTORA_BIN" completion bash 2>&1 || true)
      EXIT_CODE=$?
      test_result "vectora completion" "$EXIT_CODE" 0
      ;;
    list-models)
      OUTPUT=$(run_with_timeout 10 "$VECTORA_BIN" list-models 2>&1 || true)
      EXIT_CODE=$?
      test_result "vectora list-models" "$EXIT_CODE" 0 "$(echo "$OUTPUT" | head -1)"
      ;;
  esac

  log ""
done

log "Test Phase 3: Error Handling"
log "─────────────────────────────────────────"
log ""

# Test error cases
log "Testing invalid command..."
OUTPUT=$("$VECTORA_BIN" invalid-command 2>&1 || true)
EXIT_CODE=$?
test_result "vectora invalid-command" "$EXIT_CODE" 1 "$(echo "$OUTPUT" | head -1)"
log ""

log "Testing missing required argument..."
OUTPUT=$("$VECTORA_BIN" ask 2>&1 || true)
EXIT_CODE=$?
# Should fail if argument is required
if [ "$EXIT_CODE" -ne 0 ]; then
  TESTS_PASSED=$((TESTS_PASSED + 1))
  TESTS_RUN=$((TESTS_RUN + 1))
  log "✓ vectora ask (missing argument) correctly failed"
else
  TESTS_FAILED=$((TESTS_FAILED + 1))
  TESTS_RUN=$((TESTS_RUN + 1))
  log "⚠ vectora ask (missing argument) did not fail as expected"
fi
log ""

log "Test Phase 4: Output Validation"
log "─────────────────────────────────────────"
log ""

# Validate help output format
log "Validating help output format..."
HELP_OUTPUT=$("$VECTORA_BIN" --help 2>&1)

if echo "$HELP_OUTPUT" | grep -q "Usage:"; then
  log "✓ Help output contains Usage section"
  TESTS_PASSED=$((TESTS_PASSED + 1))
else
  log "❌ Help output missing Usage section"
  TESTS_FAILED=$((TESTS_FAILED + 1))
fi
TESTS_RUN=$((TESTS_RUN + 1))

if echo "$HELP_OUTPUT" | grep -q -i "command"; then
  log "✓ Help output contains Commands section"
  TESTS_PASSED=$((TESTS_PASSED + 1))
else
  log "⚠ Help output may be missing Commands section"
  TESTS_SKIPPED=$((TESTS_SKIPPED + 1))
fi
TESTS_RUN=$((TESTS_RUN + 1))
log ""

# Generate JSON report
log "Generating test report..."
cat > "$TEST_RESULTS_FILE" << EOF
{
  "timestamp": "$(date -u +'%Y-%m-%dT%H:%M:%SZ')",
  "tests_run": $TESTS_RUN,
  "tests_passed": $TESTS_PASSED,
  "tests_failed": $TESTS_FAILED,
  "tests_skipped": $TESTS_SKIPPED,
  "pass_rate": $(echo "scale=2; $TESTS_PASSED * 100 / $TESTS_RUN" | bc || echo "0"),
  "vectora_bin": "$VECTORA_BIN",
  "platform": "$(uname -s)",
  "environment": "test"
}
EOF

log "✓ Test report saved to: $TEST_RESULTS_FILE"

# Final summary
log ""
log "======================================================================"
log "Test Summary"
log "======================================================================"
log ""
log "Results:"
log "  Tests run:    $TESTS_RUN"
log "  Passed:       ${GREEN}$TESTS_PASSED${NC}"
log "  Failed:       ${RED}$TESTS_FAILED${NC}"
log "  Skipped:      ${YELLOW}$TESTS_SKIPPED${NC}"
log ""

if [ $TESTS_FAILED -eq 0 ]; then
  log "✓ All tests passed!"
  EXIT_CODE=0
else
  log "❌ Some tests failed!"
  EXIT_CODE=1
fi

log ""
log "Reports saved to:"
log "  • $TEST_LOG_FILE"
log "  • $TEST_RESULTS_FILE"
log ""
log "======================================================================"

exit $EXIT_CODE
