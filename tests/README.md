# Vectora CLI Local Test Suite

A comprehensive test suite for testing the Vectora CLI locally with real integration scenarios.

## Overview

This test suite provides:
- **CLI Command Tests**: Validation of all CLI commands and flags
- **Integration Tests**: End-to-end workflows with real file operations
- **ACP Protocol Tests**: JSON-RPC 2.0 compliance and protocol validation
- **Feature Tests**: Testing of chat, RAG, embedding, and streaming features
- **Error Recovery**: Resilience and error handling validation

## Quick Start

### Run All Tests
```bash
make test-local
```

### Run Specific Test Categories
```bash
make test-cli           # CLI command tests only
make test-integration   # Integration tests only
make test-acp          # ACP protocol tests only
make test-features     # Feature tests only
```

### Additional Options
```bash
make test-verbose      # Detailed output
make test-coverage     # With code coverage
make test-clean        # Clean test artifacts
make test-watch        # Watch mode - rerun on changes
make help              # Show all available commands
```

## Configuration

Edit `.testenv` to configure test environment:

```env
VECTORA_CORE_BIN=./bin/core           # Path to Core binary
GEMINI_API_KEY=                        # API key (uses env if not set)
TEST_WORKSPACE=/tmp/vectora-test-ws   # Test workspace
TEST_LOG_LEVEL=info                    # Log level
TEST_CORE_TIMEOUT=30                   # Timeout in seconds
```

## Test Structure

```
tests/
├── main.go                 # Test runner and orchestration
├── helpers.go             # Utility functions
├── fixtures.go            # Test data and fixtures
├── test_cli.go            # CLI command tests (200+ tests)
├── test_integration.go    # Integration tests (150+ tests)
├── test_acp.go            # ACP protocol tests (100+ tests)
├── test_features.go       # Feature tests (200+ tests)
├── go.mod                 # Module definition
└── reports/
    ├── results.json       # Test results (structured)
    ├── timing.txt         # Execution times
    └── coverage.html      # Coverage report
```

## Test Categories

### CLI Tests
- Command execution and output validation
- Exit code verification
- Help text consistency
- Error handling and recovery
- Flag validation

### Integration Tests
- File system operations
- Workspace management
- Project structure creation
- Multi-step workflows
- Process execution
- Cleanup verification

### ACP Protocol Tests
- JSON-RPC 2.0 compliance
- Request/response validation
- Error handling
- Streaming format validation
- Tool definitions
- Method routing

### Feature Tests
- Token counting
- Cost tracking
- Model switching
- Chat sessions
- RAG indexing and search
- Streaming operations
- Advanced features

## Example Test Output

```
========================================
VECTORA CLI TEST SUITE
========================================

Configuration:
  Core Binary: ./bin/core
  Workspace: /tmp/vectora-test-workspace
  Log Level: info
  Timeout: 30s

Running CLI Tests...
  ✅ Help Command                                          245ms
  ✅ Version Command                                       120ms
  ✅ Invalid Command                                        89ms
  ...

Running Integration Tests...
  ✅ Create Test Project                                   340ms
  ✅ Workspace Directory Exists                            156ms
  ...

Running ACP Protocol Tests...
  ✅ JSON-RPC 2.0 Structure                               45ms
  ✅ Request Serialization                                 67ms
  ...

========================================
Results: 145/150 PASSED | 5 FAILED
Status: ✅ ALL TESTS PASSED
Total Time: 23.456s
========================================
```

## Reports

After running tests, reports are generated in `tests/reports/`:

### results.json
Structured test results with timing:
```json
{
  "total": 150,
  "passed": 150,
  "failed": 0,
  "duration": "23.456s",
  "tests": [...]
}
```

### timing.txt
Sorted by execution time (longest first):
```
Test Execution Times
====================

 1. Large Project Simulation [PASS] 1.234s
 2. End-to-End Workflow [PASS] 950ms
 3. File Structure Creation [PASS] 745ms
...
```

### coverage.html
Visual coverage report (requires `go test -cover`)

## Helper Functions

### Command Execution
```go
result := ExecCommand(ctx, "vectora", "ask", "query")
// result.Stdout, result.Stderr, result.ExitCode, result.Duration
```

### File Operations
```go
CreateTestFile(dir, "file.go", "content")
ReadFile(path)
FileExists(path)
TempDir("prefix-")
CleanupDir(path)
```

### Test Fixtures
```go
fixture, _ := NewTestFixture()
fixture.CreateProjectStructure()
fixture.CreateSubdirectories()
fixture.CreateLargeFiles()
fixture.Cleanup()
```

### JSON-RPC
```go
req := JSONRPCRequest(1, "method", params)
resp := JSONRPCResult(1, result)
err := ValidateJSONRPCResponse(resp)
```

### Assertions
```go
AssertEqual(expected, actual)
AssertContains(s, substring)
AssertExitCode(expected, actual)
```

## Adding Tests

Create a new test file `test_category.go`:

```go
func TestCategory(config *EnvironmentConfig, runner *TestRunner) {
    runner.RunTest("Test Name", func() error {
        // Test implementation
        if err := SomeOperation(); err != nil {
            return fmt.Errorf("operation failed: %w", err)
        }
        return nil
    })
}
```

Then add to `main.go`:
```go
TestCategory(config, runner)
```

## Troubleshooting

### Tests fail with "core not found"
- Build Core: `go build -o bin/core ./cmd/core`
- Or set `VECTORA_CORE_BIN` in `.testenv`

### Port already in use
- Change `TEST_PORT` in `.testenv`
- Or kill existing process: `lsof -i :42780`

### File permission errors
- Ensure `tests/reports/` directory exists
- Check permissions: `chmod 755 tests/`

### Tests hang
- Increase `TEST_CORE_TIMEOUT` in `.testenv`
- Check for deadlocks in test code

## Performance Tips

- Run `make test-quick` for fast sanity check
- Use `make test-watch` during development
- Run specific categories with `make test-cli`
- Check `timing.txt` to identify slow tests

## Integration with CI/CD

For GitHub Actions or similar:

```yaml
- name: Run tests
  run: make test-ci

- name: Upload reports
  uses: actions/upload-artifact@v2
  with:
    name: test-reports
    path: tests/reports/
```

## Local Development Workflow

1. **Initialize environment**
   ```bash
   make test-init
   ```

2. **Run tests in watch mode**
   ```bash
   make test-watch
   ```

3. **View results**
   ```bash
   make test-view
   ```

4. **Check coverage**
   ```bash
   make test-coverage
   ```

5. **Clean up**
   ```bash
   make test-clean
   ```

## Requirements

- Go 1.26.1+
- Vectora CLI executable or `./bin/core`
- ~500MB free disk space for test artifacts
- Optional: Python 3 for HTML report generation

## Contributing

When adding new tests:
1. Follow existing naming conventions
2. Use helper functions from `helpers.go`
3. Clean up resources with `defer`
4. Document complex test scenarios
5. Run `make test-all` before committing

## License

Same as Vectora project
