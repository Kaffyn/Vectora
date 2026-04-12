# Vectora Local Test Suite - Complete Guide

## Quick Start

Run the complete test suite:

```bash
# Go to project root
cd /path/to/Vectora

# Run all tests (132 tests, ~17 seconds)
go run ./tests

# Or use Makefile (if available)
make test-local
```

## Test Results

```
========================================
Results: 132/132 PASSED
Status: ✅ ALL TESTS PASSED
Total Time: 17.05s
========================================
```

## Test Categories

### 1. CLI Tests (22 tests)
Tests all command-line interface functionality:
- Command parsing and execution
- Help text validation
- Flag processing
- Exit code validation
- Error handling

**Examples:**
```bash
vectora --help          # Test: Help Command
vectora --version       # Test: Version Command
vectora ask "query"     # Test: Ask Without Core
vectora embed .         # Test: Embed Without Core
```

### 2. Integration Tests (35 tests)
End-to-end workflows with real file operations:
- Workspace creation and management
- File system operations
- Project structure handling
- Process execution
- Multi-step workflows
- Cleanup verification

**Coverage:**
- Core binary integration
- Command integration
- Workspace operations
- Process-level operations
- Error recovery

### 3. ACP Protocol Tests (34 tests)
JSON-RPC 2.0 compliance validation:
- Request/response formatting
- Error handling
- Streaming format
- Tool definitions
- Message validation
- Protocol compliance

**Standards Tested:**
- JSON-RPC 2.0 specification
- Error codes and messages
- Request/response pairing
- Batch requests
- Notifications

### 4. Feature Tests (41 tests)
Functional testing of Vectora features:
- Token counting
- Cost tracking
- Model switching
- Chat sessions
- RAG indexing and search
- Streaming operations
- Advanced features

## Configuration

### .testenv File

Located in project root:

```env
# Core Binary Path
VECTORA_CORE_BIN=./bin/core

# API Keys
GEMINI_API_KEY=                    # Uses environment if not set

# Test Environment
TEST_WORKSPACE=/tmp/vectora-test-workspace
TEST_LOG_LEVEL=info                # debug, info, warn, error
TEST_CORE_TIMEOUT=30               # seconds
TEST_PORT=42780
```

## Running Specific Tests

### By Category

```go
// Edit tests/main.go to run only specific test suites

// Run only CLI tests
TestCLI(config, runner)

// Run only integration tests
TestIntegration(config, runner)

// Run only ACP tests
TestACP(config, runner)

// Run only feature tests
TestFeatures(config, runner)
```

### By Name

Tests are organized by functionality. Edit `tests/main.go` to comment/uncomment specific test suites.

## Test Reports

### timing.txt
Execution times sorted by duration (longest first):

```
Test Execution Times
====================

 1. Recovery: Missing Core [PASS] 6.48s
 2. Ask Without Core [PASS] 1.13s
 3. Ask Command Too Many Args [PASS] 1.08s
...

Total: 17.05s
```

### results.json
Structured test results in JSON format:

```json
{
  "total": 132,
  "passed": 132,
  "failed": 0,
  "duration": "17.05s",
  "tests": [
    {
      "name": "Test Name",
      "duration": "100ms",
      "passed": true,
      "errorMsg": ""
    }
  ]
}
```

## Test Files

```
tests/
├── main.go              # Test runner & orchestration
├── helpers.go           # Utility functions (300+ lines)
├── fixtures.go          # Test data (200+ lines)
├── test_cli.go          # CLI tests (250+ lines)
├── test_integration.go  # Integration tests (500+ lines)
├── test_acp.go          # ACP protocol tests (400+ lines)
├── test_features.go     # Feature tests (300+ lines)
├── go.mod               # Module definition
├── README.md            # Detailed documentation
└── reports/
    ├── results.json     # Structured results
    └── timing.txt       # Execution times
```

## Helper Functions

### Command Execution

```go
import "context"

ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

result := ExecCommand(ctx, "vectora", "ask", "query")
// result.Stdout, result.Stderr, result.ExitCode, result.Duration

// With input
result := ExecCommandWithInput(ctx, "input\n", "vectora", "chat")
```

### File Operations

```go
CreateTestFile(dir, "file.go", "content")
ReadFile(path)
FileExists(path)
TempDir("prefix-")
CleanupDir(path)

WriteFile(path, content)
```

### Test Fixtures

```go
fixture, _ := NewTestFixture()
fixture.CreateProjectStructure()       // Create files
fixture.CreateSubdirectories()         // Create dirs
fixture.CreateLargeFiles()             // Large test files
fixture.CreateConfigFile(content)      // Config file
defer fixture.Cleanup()                // Clean up
```

### JSON-RPC Testing

```go
req := JSONRPCRequest(1, "method", params)
resp := JSONRPCResult(1, result)
err := ValidateJSONRPCResponse(resp)

errResp := JSONRPCError(1, -32600, "Invalid Request")
```

### Assertions

```go
AssertEqual(expected, actual)
AssertNotEqual(expected, actual)
AssertContains(s, substring)
AssertNotContains(s, substring)
AssertError(err)
AssertNoError(err)
AssertNil(v)
AssertNotNil(v)
AssertExitCode(expected, actual)
```

## Adding New Tests

### Example: Adding a CLI Test

1. Create test function in `test_cli.go`:

```go
func TestNewFeature(config *EnvironmentConfig, runner *TestRunner) {
    runner.RunTest("Feature: New Command", func() error {
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()

        result := ExecCommand(ctx, "vectora", "new-cmd")

        if result.ExitCode != 0 {
            return fmt.Errorf("command failed: %s", result.Stderr)
        }

        return nil
    })
}
```

2. Add to `main.go`:

```go
TestNewFeature(config, runner)
```

3. Run tests:

```bash
cd tests && go run .
```

## Troubleshooting

### Tests Fail: "vectora not found"

1. Build Core:
   ```bash
   go build -o bin/core ./cmd/core
   ```

2. Or set in `.testenv`:
   ```env
   VECTORA_CORE_BIN=/path/to/vectora
   ```

### Port Already in Use

Change `TEST_PORT` in `.testenv`:
```env
TEST_PORT=42781
```

### Tests Hang

Increase timeout in `.testenv`:
```env
TEST_CORE_TIMEOUT=60
```

### Permission Errors

```bash
chmod -R 755 tests/
```

## Performance

### Slowest Tests

1. `Recovery: Missing Core` - 6.48s (intentional)
2. `Ask Without Core` - 1.13s (waiting for timeout)
3. `Ask Command Too Many Args` - 1.08s

### Fast Tests

- ACP Protocol tests - mostly < 1ms
- Unit tests - < 10ms
- File operations - < 50ms

### Total Runtime

- Full suite: ~17 seconds
- CLI tests only: ~3 seconds
- Integration tests only: ~5 seconds

## CI/CD Integration

### GitHub Actions

```yaml
- name: Run tests
  run: |
    cd tests
    go run .

- name: Upload reports
  uses: actions/upload-artifact@v2
  with:
    name: test-reports
    path: tests/reports/
```

### Example Workflow

```yaml
name: Tests
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-go@v2
      - run: cd tests && go run .
```

## Best Practices

1. **Test Isolation**: Each test is independent
2. **Cleanup**: Always use `defer` for cleanup
3. **Timeouts**: Set appropriate context timeouts
4. **Error Handling**: Check both error and exit codes
5. **Assertions**: Use helper assertion functions
6. **Documentation**: Document complex test logic

## Contributing Tests

When adding new tests:

1. Follow existing naming conventions
2. Use helper functions from `helpers.go`
3. Clean up resources with `defer`
4. Document complex scenarios
5. Run `go run ./tests` before committing
6. Check `tests/reports/timing.txt` for performance

## Requirements

- Go 1.26.1 or later
- Vectora CLI or Core binary
- ~500MB free disk space for test artifacts
- bash/sh shell for Makefile

## Advanced Usage

### Environment Variables

```bash
# Override log level
TEST_LOG_LEVEL=debug go run ./tests

# Custom workspace
TEST_WORKSPACE=/tmp/custom go run ./tests

# Custom Core binary
VECTORA_CORE_BIN=/custom/path go run ./tests
```

### Code Coverage

```bash
go test -cover -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Benchmarking

```bash
go test -bench=. -benchmem ./...
```

## References

- Test Suite: `/tests/README.md`
- Main Runner: `/tests/main.go`
- Configuration: `/.testenv`
- Makefile: `/Makefile`

## Support

For issues or questions:

1. Check test output in `tests/reports/`
2. Review specific test in source code
3. Check `.testenv` configuration
4. Verify Core binary path

## License

Same as Vectora project
