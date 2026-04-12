# CI/CD Testing Guide

This document describes the comprehensive testing pipeline for Vectora VS Code extension and CLI.

## Overview

The CI/CD pipeline includes:

1. **Command Tests** - Test all Vectora CLI commands
2. **Extension Tests** - Test VS Code extension
3. **CLI Extension Tests** - Test Gemini CLI extension
4. **Integration Tests** - Test core integration
5. **Security Scanning** - Dependency and vulnerability scanning
6. **Performance Monitoring** - Bundle size and benchmark tracking

## Workflows

### 1. Test Workflow (`tests.yml`)

Runs on every push and pull request. Tests all commands with multi-platform support.

**Triggers:**
- Push to `master` or `develop`
- Pull requests to `master` or `develop`
- Manual workflow dispatch with optional full test

**Jobs:**
- `command-tests` - Matrix: OS x Go version x Node version
- `extension-tests` - VS Code extension testing
- `cli-extension-tests` - Gemini CLI testing
- `integration-tests` - Core integration tests
- `test-summary` - Final summary report

**Matrix Testing:**
- OS: Ubuntu, macOS, Windows
- Go: 1.22, 1.23
- Node: 20, 22

### 2. Security Workflow (`security.yml`)

Runs on push, pull requests, and daily schedule (2 AM UTC).

**Jobs:**
- `dependency-scan` - npm and Go dependencies
- `trivy-scan` - Container/filesystem scanning
- `secret-scan` - Secret detection (TruffleHog)
- `code-quality` - golangci-lint, ESLint
- `license-check` - License compliance

### 3. Performance Workflow (`performance.yml`)

Monitors bundle size and build performance.

**Jobs:**
- `bundle-size` - Analyzes extension bundle size
- `benchmarks` - Go benchmarks
- `build-time` - Tracks build time
- `performance-summary` - Summary and PR comments

## Test Scripts

### `verify-help.sh`

Verifies the `vectora --help` command and extracts available commands.

```bash
./scripts/verify-help.sh
```

**Output:**
- `test-results/help.txt` - Full help output
- `test-results/commands.txt` - List of available commands
- `logs/verify-help.log` - Execution log

### `setup-test-env.sh`

Sets up test environment with secure .env handling.

```bash
./scripts/setup-test-env.sh
```

**Features:**
- Creates `.env.test` from `.env.example`
- Injects GitHub Secrets (if in CI)
- Creates test directories
- Validates configuration
- Git-ignores sensitive files

**Environment variables supported:**
- `GEMINI_API_KEY`
- `ANTHROPIC_API_KEY`
- `OPENAI_API_KEY`
- `VOYAGE_API_KEY`

### `run-command-suite.sh`

Comprehensive test suite for all Vectora commands.

```bash
./scripts/run-command-suite.sh
```

**Tests:**
- Core commands: `--version`, `--help`, `version`, `help`
- Optional commands: `ask`, `chat`, `rag`, `configure`, `completion`, `list-models`
- Error handling: Invalid commands, missing arguments
- Output validation: Help format, section checks

**Output:**
- `test-results/command-tests.json` - Test results in JSON format
- `logs/command-tests.log` - Detailed execution log

### `test-vectora-commands.sh`

Individual command testing with detailed validation.

```bash
./scripts/test-vectora-commands.sh [command] [args...]
```

**Examples:**
```bash
./scripts/test-vectora-commands.sh help
./scripts/test-vectora-commands.sh ask "What is AI?"
./scripts/test-vectora-commands.sh chat
```

## GitHub Secrets Configuration

Required secrets for CI/CD:

```bash
GEMINI_API_KEY          # Gemini API key for tests
ANTHROPIC_API_KEY       # (Optional) Anthropic API key
OPENAI_API_KEY          # (Optional) OpenAI API key
VOYAGE_API_KEY          # (Optional) Voyage AI key
```

**Setup:**
1. Go to GitHub repository settings
2. Secrets and variables → Actions
3. New repository secret
4. Add each secret with its value

**Security Notes:**
- Secrets are encrypted and never logged
- Only injected into test environments in CI
- `.env.test` is automatically git-ignored
- Secrets are never committed to the repository

## Test Environment

### .env.test

The test environment file (`.env.test`) contains:

- Test API keys (fake/mock values)
- Vectora core configuration
- RAG settings
- Cache configuration
- Test flags
- Feature toggles
- Security settings

**Important:** This file should NOT be committed. It's automatically created during CI and ignored by git.

### Test Directories

Created during setup:
- `test-data/` - Test data and artifacts
- `test-data/rag-index/` - RAG index cache
- `test-data/cache/` - Cache directory
- `test-data/logs/` - Test logs
- `test-results/` - Test results
- `logs/` - Execution logs
- `coverage/` - Coverage reports

## Running Tests Locally

### Prerequisites

```bash
# Go (1.22+)
go version

# Node.js (20+)
node --version
npm --version

# Build Vectora
go build -o bin/vectora ./cmd/core
```

### Full Test Suite

```bash
# Setup environment
./scripts/setup-test-env.sh

# Run all tests
./scripts/run-command-suite.sh

# Run with GEMINI_API_KEY
GEMINI_API_KEY=your-key ./scripts/run-command-suite.sh
```

### Individual Tests

```bash
# Help verification
./scripts/verify-help.sh

# Test specific command
./scripts/test-vectora-commands.sh ask "Hello, Vectora!"

# VS Code extension
cd extensions/vscode
npm test
npm run compile
```

### Extension Tests

```bash
# VS Code Extension
cd extensions/vscode
npm ci
npm run lint
npm run compile
npm test -- --coverage

# Gemini CLI
cd extensions/geminicli
npm ci
npm run build
```

## Test Results

### Test Report Format

Commands tests produce a JSON report:

```json
{
  "timestamp": "2024-04-12T10:30:00Z",
  "tests_run": 25,
  "tests_passed": 24,
  "tests_failed": 1,
  "tests_skipped": 0,
  "pass_rate": 96.0,
  "vectora_bin": "./bin/vectora",
  "platform": "Linux",
  "environment": "test"
}
```

### Viewing Results

After a workflow run:

1. Go to GitHub Actions tab
2. Click the workflow run
3. Expand the job (e.g., "Test Vectora Commands")
4. View logs or download artifacts

### Coverage Reports

Coverage reports are uploaded to:
- `coverage/` directory
- Codecov integration (if enabled)

View with:
```bash
# After running tests locally
go tool cover -html=coverage.out
```

## Troubleshooting

### Tests Failing on CI but Pass Locally

**Common causes:**
- Different environment variables
- Platform differences (Windows vs Linux)
- Missing dependencies
- Network issues (if tests hit real APIs)

**Solutions:**
1. Check GitHub Secrets are configured
2. Review test logs in workflow artifacts
3. Run same tests locally with CI environment

### Timeout Issues

Tests have default timeouts:
- Command tests: 10s per command
- Integration tests: 15 minutes
- Build time: depends on task

**To increase timeout:**
Edit `tests.yml` and adjust `timeout-minutes` in the job.

### Network Errors

If tests fail with network errors:
1. Check if API keys are valid (not expired)
2. Verify network connectivity in runner
3. Consider mocking external API calls

### Build Artifacts Not Found

If binary is not found:
1. Check build step completed successfully
2. Verify binary path: `./bin/vectora` (or `.exe` on Windows)
3. Review build logs for errors

## Performance Budgets

The performance workflow monitors:

**Bundle Size Limits:**
- Extension: 2MB
- Webview: 3MB
- Total: 5MB

**Build Time Limits:**
- Go build: <60 seconds
- Node build: <120 seconds

**Benchmark Thresholds:**
- Depends on specific operations
- See benchmark results in artifacts

## Continuous Improvement

### Adding New Tests

1. Create test script in `scripts/`
2. Add job to workflow (e.g., `tests.yml`)
3. Document in this file
4. Test locally before pushing

### Updating Dependencies

```bash
# Go
go get -u ./...
go mod tidy

# Node (vscode extension)
cd extensions/vscode
npm update
npm audit fix
```

### Security Updates

Run security workflows:
```bash
# Manual trigger
gh workflow run security.yml
```

## References

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Vectora Documentation](../../README.md)
- [Contributing Guide](../../CONTRIBUTING.md)
