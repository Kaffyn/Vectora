#!/bin/bash

################################################################################
# Vectora Local Test Suite
# Complete testing framework for Vectora CLI with validation and reporting
#
# Usage:
#   ./test/local-test-suite.sh [OPTIONS]
#
# Options:
#   --help              Show this help message
#   --interactive       Run in interactive mode (choose tests)
#   --batch             Run all tests automatically
#   --test COMMAND      Run specific test (ask, chat, rag, configure, etc)
#   --with-gemini       Use Gemini for tests (requires GEMINI_API_KEY)
#   --verbose           Show detailed output
#   --debug             Show debug information
#   --trace             Show command execution trace
#   --save-outputs      Save test outputs to files
#   --generate-report   Generate HTML report
#
# Examples:
#   ./test/local-test-suite.sh --interactive
#   ./test/local-test-suite.sh --batch
#   ./test/local-test-suite.sh --test ask --with-gemini
#   ./test/local-test-suite.sh --batch --verbose --generate-report
#
################################################################################

set -o pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
MAGENTA='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
ENV_FILE="${SCRIPT_DIR}/.env.local"
VECTORA_BIN="${PROJECT_ROOT}/bin/vectora"
FIXTURES_DIR="${SCRIPT_DIR}/fixtures"
REPORTS_DIR="${SCRIPT_DIR}/reports"
OUTPUTS_DIR="${SCRIPT_DIR}/outputs"

# Test configuration
TEST_TIMEOUT=30
VERBOSE=false
DEBUG=false
TRACE=false
SAVE_OUTPUTS=false
GENERATE_REPORT=false
INTERACTIVE=false
BATCH=false
SPECIFIC_TEST=""
WITH_GEMINI=false

# Test results
TESTS_TOTAL=0
TESTS_PASSED=0
TESTS_FAILED=0
TESTS_SKIPPED=0
TEST_START_TIME=0
TEST_RESULTS=()

################################################################################
# Utilities
################################################################################

log_info() {
    echo -e "${BLUE}ℹ ${NC}$*"
}

log_success() {
    echo -e "${GREEN}✓ ${NC}$*"
}

log_error() {
    echo -e "${RED}✗ ${NC}$*"
}

log_warning() {
    echo -e "${YELLOW}⚠ ${NC}$*"
}

log_debug() {
    [[ $DEBUG == true ]] && echo -e "${MAGENTA}🐛 ${NC}$*"
}

log_trace() {
    [[ $TRACE == true ]] && echo -e "${CYAN}→ ${NC}$*"
}

print_header() {
    echo ""
    echo -e "${MAGENTA}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${MAGENTA}${NC} $*"
    echo -e "${MAGENTA}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
}

print_section() {
    echo ""
    echo -e "${CYAN}→ $*${NC}"
}

show_help() {
    head -33 "$0" | tail -30
}

setup_environment() {
    log_info "Setting up test environment..."

    # Create directories
    mkdir -p "$REPORTS_DIR" "$OUTPUTS_DIR"

    # Load .env.local
    if [[ -f "$ENV_FILE" ]]; then
        log_debug "Loading environment from $ENV_FILE"
        set -a
        source "$ENV_FILE"
        set +a
    else
        log_warning ".env.local not found. Copy from .env.local.example and fill API keys."
        log_warning "Continuing with limited functionality..."
    fi

    # Check Vectora binary
    if [[ ! -f "$VECTORA_BIN" ]]; then
        log_error "Vectora binary not found at $VECTORA_BIN"
        log_info "Please build Vectora first: cd $PROJECT_ROOT && go build -o bin/vectora ./cmd/core"
        exit 1
    fi

    log_success "Environment ready"
}

run_command() {
    local cmd="$*"
    log_trace "$cmd"

    local output
    local exit_code
    local start_time
    local duration

    start_time=$(date +%s%N)
    output=$(eval "$cmd" 2>&1)
    exit_code=$?
    duration=$(( ($(date +%s%N) - start_time) / 1000000 ))

    if [[ $SAVE_OUTPUTS == true ]]; then
        echo "$output" >> "${OUTPUTS_DIR}/last-output.log"
    fi

    [[ $TRACE == true ]] && log_trace "Exit code: $exit_code, Duration: ${duration}ms"

    echo "$output"
    return $exit_code
}

################################################################################
# Test Functions
################################################################################

test_help() {
    print_section "Testing: vectora --help"

    local output
    output=$(run_command "$VECTORA_BIN --help")
    local exit_code=$?

    if [[ $exit_code -eq 0 ]] && [[ -n "$output" ]] && [[ "$output" == *"usage"* || "$output" == *"Usage"* ]]; then
        log_success "Help command works"
        [[ $VERBOSE == true ]] && echo "$output"
        return 0
    else
        log_error "Help command failed"
        [[ $VERBOSE == true ]] && echo "$output"
        return 1
    fi
}

test_version() {
    print_section "Testing: vectora --version"

    local output
    output=$(run_command "$VECTORA_BIN --version")
    local exit_code=$?

    if [[ $exit_code -eq 0 ]] && [[ "$output" =~ ^[0-9]+\.[0-9]+\.[0-9]+ ]]; then
        log_success "Version: $output"
        return 0
    else
        log_error "Version command failed or returned invalid format"
        [[ $VERBOSE == true ]] && echo "$output"
        return 1
    fi
}

test_ask_simple() {
    print_section "Testing: vectora ask (simple prompt)"

    local prompt="Hello, what is Vectora?"
    local output
    output=$(timeout $TEST_TIMEOUT run_command "$VECTORA_BIN ask \"$prompt\"" 2>&1)
    local exit_code=$?

    if [[ $exit_code -eq 0 ]] && [[ -n "$output" ]]; then
        log_success "Ask command executed successfully"
        [[ $VERBOSE == true ]] && echo "Output: ${output:0:200}..."
        return 0
    elif [[ $exit_code -eq 124 ]]; then
        log_warning "Ask command timed out (${TEST_TIMEOUT}s)"
        return 1
    else
        log_error "Ask command failed (exit code: $exit_code)"
        [[ $VERBOSE == true ]] && echo "$output"
        return 1
    fi
}

test_ask_with_gemini() {
    print_section "Testing: vectora ask with Gemini model"

    if [[ -z "$GEMINI_API_KEY" ]]; then
        log_warning "GEMINI_API_KEY not set, skipping Gemini test"
        return 2 # Skip
    fi

    local prompt="What is machine learning?"
    local output
    output=$(timeout $TEST_TIMEOUT run_command "$VECTORA_BIN ask --model gemini-1.5-pro \"$prompt\"" 2>&1)
    local exit_code=$?

    if [[ $exit_code -eq 0 ]] && [[ -n "$output" ]]; then
        log_success "Ask with Gemini executed successfully"
        [[ $VERBOSE == true ]] && echo "Output: ${output:0:200}..."
        return 0
    elif [[ $exit_code -eq 124 ]]; then
        log_warning "Ask with Gemini timed out (${TEST_TIMEOUT}s)"
        return 1
    else
        log_error "Ask with Gemini failed (exit code: $exit_code)"
        [[ $VERBOSE == true ]] && echo "$output"
        return 1
    fi
}

test_configure_list() {
    print_section "Testing: vectora configure --list"

    local output
    output=$(run_command "$VECTORA_BIN configure --list" 2>&1)
    local exit_code=$?

    if [[ $exit_code -eq 0 ]]; then
        log_success "Configure list executed successfully"
        [[ $VERBOSE == true ]] && echo "$output"
        return 0
    else
        log_warning "Configure list command not available or failed"
        [[ $VERBOSE == true ]] && echo "$output"
        return 1
    fi
}

test_list_providers() {
    print_section "Testing: vectora list-models"

    local output
    output=$(run_command "$VECTORA_BIN list-models 2>&1 || $VECTORA_BIN list-providers 2>&1")
    local exit_code=$?

    if [[ $exit_code -eq 0 ]] && [[ -n "$output" ]]; then
        log_success "List models/providers executed successfully"
        [[ $VERBOSE == true ]] && echo "$output"
        return 0
    else
        log_warning "List models/providers command not available"
        return 1
    fi
}

test_completion() {
    print_section "Testing: vectora completion bash"

    local output
    output=$(run_command "$VECTORA_BIN completion bash 2>&1")
    local exit_code=$?

    if [[ $exit_code -eq 0 ]] && [[ -n "$output" ]]; then
        log_success "Completion script generated successfully"
        [[ $VERBOSE == true ]] && echo "Script length: ${#output} characters"
        return 0
    else
        log_warning "Completion command not available"
        return 1
    fi
}

test_invalid_command() {
    print_section "Testing: invalid command (error handling)"

    local output
    output=$(run_command "$VECTORA_BIN invalid-command-xyz 2>&1")
    local exit_code=$?

    if [[ $exit_code -ne 0 ]]; then
        log_success "Invalid command properly rejected (exit code: $exit_code)"
        return 0
    else
        log_error "Invalid command should have failed but didn't"
        return 1
    fi
}

################################################################################
# Test Management
################################################################################

run_test() {
    local test_name="$1"
    local test_func="test_$test_name"

    if ! declare -f "$test_func" > /dev/null; then
        log_error "Test function $test_func not found"
        return 1
    fi

    TESTS_TOTAL=$((TESTS_TOTAL + 1))
    TEST_START_TIME=$(date +%s%N)

    if $test_func; then
        TESTS_PASSED=$((TESTS_PASSED + 1))
        TEST_RESULTS+=("✓ $test_name")
        return 0
    else
        local result=$?
        if [[ $result -eq 2 ]]; then
            TESTS_SKIPPED=$((TESTS_SKIPPED + 1))
            TEST_RESULTS+=("⊘ $test_name (skipped)")
        else
            TESTS_FAILED=$((TESTS_FAILED + 1))
            TEST_RESULTS+=("✗ $test_name")
        fi
        return 1
    fi
}

run_all_tests() {
    print_header "Running All Tests"

    run_test "help"
    run_test "version"
    run_test "ask_simple"

    if [[ $WITH_GEMINI == true ]]; then
        run_test "ask_with_gemini"
    fi

    run_test "configure_list"
    run_test "list_providers"
    run_test "completion"
    run_test "invalid_command"

    print_results
}

run_specific_test() {
    local test_name="$1"
    print_header "Running Specific Test: $test_name"

    run_test "$test_name"

    print_results
}

print_results() {
    echo ""
    print_header "Test Results Summary"

    echo -e "${CYAN}Total Tests:${NC} $TESTS_TOTAL"
    echo -e "${GREEN}Passed:${NC} $TESTS_PASSED"
    echo -e "${RED}Failed:${NC} $TESTS_FAILED"
    echo -e "${YELLOW}Skipped:${NC} $TESTS_SKIPPED"

    if [[ $TESTS_TOTAL -gt 0 ]]; then
        local percentage=$(( (TESTS_PASSED * 100) / TESTS_TOTAL ))
        echo -e "${CYAN}Success Rate:${NC} $percentage%"
    fi

    echo ""
    echo "Details:"
    for result in "${TEST_RESULTS[@]}"; do
        echo "  $result"
    done

    echo ""
    if [[ $TESTS_FAILED -eq 0 ]]; then
        log_success "All tests passed!"
        return 0
    else
        log_error "$TESTS_FAILED test(s) failed"
        return 1
    fi
}

################################################################################
# Interactive Menu
################################################################################

show_interactive_menu() {
    while true; do
        echo ""
        echo -e "${CYAN}Vectora Test Suite - Interactive Menu${NC}"
        echo "1) Run all tests"
        echo "2) Test help & version"
        echo "3) Test ask command"
        echo "4) Test ask with Gemini"
        echo "5) Test configuration"
        echo "6) Test completion"
        echo "7) Test error handling"
        echo "8) Show help"
        echo "9) Exit"
        echo ""
        read -p "Choose option: " choice

        case $choice in
            1) run_all_tests ;;
            2) run_test "help"; run_test "version" ;;
            3) run_test "ask_simple" ;;
            4) run_test "ask_with_gemini" ;;
            5) run_test "configure_list" ;;
            6) run_test "completion" ;;
            7) run_test "invalid_command" ;;
            8) show_help ;;
            9) echo "Exiting..."; exit 0 ;;
            *) echo "Invalid option" ;;
        esac
    done
}

################################################################################
# Main Entry Point
################################################################################

main() {
    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            --help)
                show_help
                exit 0
                ;;
            --interactive)
                INTERACTIVE=true
                shift
                ;;
            --batch)
                BATCH=true
                shift
                ;;
            --test)
                SPECIFIC_TEST="$2"
                shift 2
                ;;
            --with-gemini)
                WITH_GEMINI=true
                shift
                ;;
            --verbose)
                VERBOSE=true
                shift
                ;;
            --debug)
                DEBUG=true
                VERBOSE=true
                shift
                ;;
            --trace)
                TRACE=true
                DEBUG=true
                VERBOSE=true
                shift
                ;;
            --save-outputs)
                SAVE_OUTPUTS=true
                shift
                ;;
            --generate-report)
                GENERATE_REPORT=true
                shift
                ;;
            *)
                echo "Unknown option: $1"
                show_help
                exit 1
                ;;
        esac
    done

    # Setup
    setup_environment

    # Run tests
    if [[ $INTERACTIVE == true ]]; then
        show_interactive_menu
    elif [[ -n "$SPECIFIC_TEST" ]]; then
        run_specific_test "$SPECIFIC_TEST"
        exit $?
    elif [[ $BATCH == true ]]; then
        run_all_tests
        exit $?
    else
        # Default: run all tests
        run_all_tests
        exit $?
    fi
}

# Run main function
main "$@"
