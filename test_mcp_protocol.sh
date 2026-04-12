#!/bin/bash

# Vectora MCP Protocol Testing Script
# Tests JSON-RPC 2.0 compliance and tool functionality

set -e

COLOR_GREEN='\033[0;32m'
COLOR_RED='\033[0;31m'
COLOR_YELLOW='\033[1;33m'
COLOR_BLUE='\033[0;34m'
COLOR_NC='\033[0m' # No Color

# Test counter
TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0

# Create test workspace
TEST_WORKSPACE="/tmp/vectora-test-workspace"
mkdir -p "$TEST_WORKSPACE"

# Helper function to run test
run_test() {
    local test_name="$1"
    local input="$2"
    local expected_contains="$3"

    TESTS_RUN=$((TESTS_RUN + 1))

    echo -e "${COLOR_BLUE}[TEST $TESTS_RUN]${COLOR_NC} $test_name"

    # Run the test with timeout
    output=$(echo "$input" | timeout 10s ./core.exe mcp "$TEST_WORKSPACE" 2>&1 || true)

    if echo "$output" | grep -q "$expected_contains"; then
        echo -e "${COLOR_GREEN}✅ PASSED${COLOR_NC}"
        TESTS_PASSED=$((TESTS_PASSED + 1))
        return 0
    else
        echo -e "${COLOR_RED}❌ FAILED${COLOR_NC}"
        echo "Expected to contain: $expected_contains"
        echo "Got: $output"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
    fi
}

echo -e "${COLOR_BLUE}======================================${COLOR_NC}"
echo -e "${COLOR_BLUE}Vectora MCP Protocol Testing${COLOR_NC}"
echo -e "${COLOR_BLUE}======================================${COLOR_NC}"
echo ""

# Build the project first
echo -e "${COLOR_YELLOW}Building Vectora core...${COLOR_NC}"
if go build -o core.exe ./cmd/core; then
    echo -e "${COLOR_GREEN}✅ Build successful${COLOR_NC}"
else
    echo -e "${COLOR_RED}❌ Build failed${COLOR_NC}"
    exit 1
fi
echo ""

# Category 1: Protocol Tests
echo -e "${COLOR_BLUE}Category 1: JSON-RPC 2.0 Protocol Tests${COLOR_NC}"
echo "==========================================="

# Test 1.1: Initialize request
run_test "1.1: Initialize request" \
    '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}' \
    '"protocolVersion":"2024-11-05"'

# Test 1.2: Tools list
run_test "1.2: Tools list" \
    '{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}' \
    '"name":"embed"'

# Test 1.3: Invalid JSON
run_test "1.3: Invalid JSON error handling" \
    'invalid json' \
    'Parse error'

# Test 1.4: Missing method
run_test "1.4: Missing method error" \
    '{"jsonrpc":"2.0","id":3}' \
    'method is required'

# Test 1.5: Wrong JSON-RPC version
run_test "1.5: Wrong JSON-RPC version" \
    '{"jsonrpc":"1.0","id":4,"method":"initialize"}' \
    'jsonrpc must be'

# Test 1.6: Unknown method
run_test "1.6: Unknown method error" \
    '{"jsonrpc":"2.0","id":5,"method":"unknown/method"}' \
    'Method not found'

echo ""

# Category 2: Tool Tests
echo -e "${COLOR_BLUE}Category 2: Tool Discovery Tests${COLOR_NC}"
echo "==========================================="

# Test 2.1: All tools discoverable
run_test "2.1: All tools discoverable" \
    '{"jsonrpc":"2.0","id":6,"method":"tools/list","params":{}}' \
    '"search_database"'

# Additional tool check
run_test "2.1b: analyze_code_patterns discoverable" \
    '{"jsonrpc":"2.0","id":7,"method":"tools/list","params":{}}' \
    '"analyze_code_patterns"'

# Test 2.2: Non-existent tool error
run_test "2.2: Non-existent tool error" \
    '{"jsonrpc":"2.0","id":8,"method":"tools/call","params":{"name":"fake_tool","input":{}}}' \
    'tool not found'

echo ""

# Category 3: Error Handling Tests
echo -e "${COLOR_BLUE}Category 3: Error Handling Tests${COLOR_NC}"
echo "==========================================="

# Test 3.1: Tool call with missing parameters
run_test "3.1: Tool call validation" \
    '{"jsonrpc":"2.0","id":9,"method":"tools/call","params":{"name":"embed","input":{}}}' \
    'error\|Error\|required'

echo ""

# Summary
echo -e "${COLOR_BLUE}======================================${COLOR_NC}"
echo -e "${COLOR_BLUE}Test Summary${COLOR_NC}"
echo -e "${COLOR_BLUE}======================================${COLOR_NC}"
echo "Total Tests Run: $TESTS_RUN"
echo -e "Passed: ${COLOR_GREEN}$TESTS_PASSED${COLOR_NC}"
echo -e "Failed: ${COLOR_RED}$TESTS_FAILED${COLOR_NC}"

if [ $TESTS_FAILED -eq 0 ]; then
    echo ""
    echo -e "${COLOR_GREEN}✅ ALL TESTS PASSED${COLOR_NC}"
    exit 0
else
    echo ""
    echo -e "${COLOR_RED}❌ SOME TESTS FAILED${COLOR_NC}"
    exit 1
fi
