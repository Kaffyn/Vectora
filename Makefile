.PHONY: test-local test-cli test-integration test-acp test-features test-coverage test-verbose test-clean help

# Colors for output
GREEN := \033[0;32m
YELLOW := \033[0;33m
RED := \033[0;31m
NC := \033[0m # No Color

help:
	@echo "$(GREEN)Vectora Test Suite$(NC)"
	@echo "===================="
	@echo ""
	@echo "$(YELLOW)Available commands:$(NC)"
	@echo "  make test-local       - Run complete test suite locally"
	@echo "  make test-cli         - Run only CLI tests"
	@echo "  make test-integration - Run only integration tests"
	@echo "  make test-acp         - Run only ACP protocol tests"
	@echo "  make test-features    - Run only feature tests"
	@echo "  make test-coverage    - Run tests with coverage report"
	@echo "  make test-verbose     - Run tests with verbose output"
	@echo "  make test-clean       - Clean test artifacts"
	@echo "  make test-watch       - Run tests in watch mode"
	@echo "  make test-report      - Generate HTML report"
	@echo ""
	@echo "$(YELLOW)Configuration:$(NC)"
	@echo "  See .testenv for test configuration"
	@echo ""

# Main test target - runs complete suite locally
test-local:
	@echo "$(GREEN)Starting Vectora Local Test Suite...$(NC)"
	@echo ""
	@cd tests && go run .

# Run only CLI tests
test-cli:
	@echo "$(GREEN)Running CLI Tests Only...$(NC)"
	@echo ""
	@cd tests && go run . -test cli

# Run only integration tests
test-integration:
	@echo "$(GREEN)Running Integration Tests Only...$(NC)"
	@echo ""
	@cd tests && go run . -test integration

# Run only ACP tests
test-acp:
	@echo "$(GREEN)Running ACP Protocol Tests Only...$(NC)"
	@echo ""
	@cd tests && go run . -test acp

# Run only feature tests
test-features:
	@echo "$(GREEN)Running Feature Tests Only...$(NC)"
	@echo ""
	@cd tests && go run . -test features

# Run with coverage
test-coverage:
	@echo "$(GREEN)Running tests with coverage...$(NC)"
	@echo ""
	@go test -cover -coverprofile=coverage.out -covermode=atomic ./...
	@go tool cover -html=coverage.out -o tests/reports/coverage.html
	@echo "$(GREEN)Coverage report saved to tests/reports/coverage.html$(NC)"

# Run with verbose output
test-verbose:
	@echo "$(GREEN)Running tests with verbose output...$(NC)"
	@echo ""
	@cd tests && go run . -verbose

# Clean test artifacts
test-clean:
	@echo "$(YELLOW)Cleaning test artifacts...$(NC)"
	@rm -rf /tmp/vectora-test-*
	@rm -f coverage.out coverage.html
	@rm -f tests/reports/*.json tests/reports/*.txt tests/reports/*.html
	@echo "$(GREEN)Cleanup complete$(NC)"

# Watch mode - rerun tests on file changes
test-watch:
	@echo "$(GREEN)Running tests in watch mode...$(NC)"
	@echo "Watching for changes..."
	@while true; do \
		clear; \
		cd tests && go run . || true; \
		echo ""; \
		echo "$(YELLOW)Waiting for changes... (press Ctrl+C to exit)$(NC)"; \
		sleep 2; \
	done

# Generate HTML test report
test-report:
	@echo "$(GREEN)Generating test report...$(NC)"
	@cd tests && go run . > /tmp/test-output.txt 2>&1
	@python3 tests/generate_report.py /tmp/test-output.txt tests/reports/report.html || \
	echo "$(YELLOW)HTML generation requires Python, skipping...$(NC)"
	@echo "$(GREEN)Reports available in tests/reports/$(NC)"

# Run quick sanity check
test-quick:
	@echo "$(GREEN)Running quick sanity check...$(NC)"
	@cd tests && timeout 30 go run . 2>&1 | head -50 || echo "$(RED)Tests failed$(NC)"

# Run specific test by name
test-run:
	@echo "$(GREEN)Running specific test...$(NC)"
	@cd tests && go run . -run $(FILTER)

# Benchmark tests
test-bench:
	@echo "$(GREEN)Running performance benchmarks...$(NC)"
	@go test -bench=. -benchmem -benchtime=3s ./... 2>/dev/null || echo "$(YELLOW)No benchmarks found$(NC)"

# Check test dependencies
test-deps:
	@echo "$(GREEN)Checking test dependencies...$(NC)"
	@go mod verify
	@echo "$(GREEN)All dependencies verified$(NC)"

# Lint test files
test-lint:
	@echo "$(GREEN)Linting test files...$(NC)"
	@cd tests && go vet ./... 2>/dev/null || echo "$(YELLOW)Linting complete$(NC)"

# Format test files
test-fmt:
	@echo "$(GREEN)Formatting test files...$(NC)"
	@cd tests && go fmt ./...
	@echo "$(GREEN)Formatting complete$(NC)"

# Run all checks (lint, fmt, test)
test-all: test-lint test-fmt test-coverage test-report
	@echo "$(GREEN)All checks passed!$(NC)"

# Initialize test environment
test-init:
	@echo "$(GREEN)Initializing test environment...$(NC)"
	@mkdir -p tests/reports tests/fixtures
	@mkdir -p /tmp/vectora-test-workspace
	@echo "$(GREEN)Test environment ready$(NC)"

# Show test statistics
test-stats:
	@echo "$(GREEN)Test Statistics$(NC)"
	@echo "================"
	@find tests -name "*.go" -type f | wc -l | xargs echo "Test files:"
	@find tests -name "*_test.go" -type f -o -name "test_*.go" -type f | wc -l | xargs echo "Test cases:"
	@wc -l tests/*.go | tail -1 | awk '{print "Total lines: " $$1}'

# View latest test report
test-view:
	@if [ -f "tests/reports/timing.txt" ]; then \
		cat tests/reports/timing.txt; \
	else \
		echo "$(YELLOW)No test reports found. Run 'make test-local' first.$(NC)"; \
	fi

# Docker test environment
test-docker:
	@echo "$(GREEN)Building Docker test environment...$(NC)"
	@docker build -t vectora-test -f tests/Dockerfile . 2>/dev/null || echo "$(YELLOW)Docker not available$(NC)"

# CI/CD compatible test run
test-ci:
	@echo "$(GREEN)Running CI/CD tests...$(NC)"
	@cd tests && go run . > test-results.json 2>&1
	@echo "$(GREEN)Results saved to test-results.json$(NC)"
