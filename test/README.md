# Vectora Local Test Suite

Complete testing framework for Vectora CLI that can be run locally during development.

## Quick Start

```bash
# 1. Setup environment
chmod +x test/*.sh
./test/setup-test-env.sh

# 2. Run interactive tests
./test/local-test-suite.sh --interactive

# 3. Or run all tests automatically
./test/local-test-suite.sh --batch

# 4. Test specific command
./test/local-test-suite.sh --test ask
```

## Configuration

### Setup .env.local

Copy the template and add your API keys:

```bash
cp test/.env.local.example test/.env.local
```

Then edit `.env.local`:

```bash
GEMINI_API_KEY=your_key_here
ANTHROPIC_API_KEY=your_key_here
OPENAI_API_KEY=your_key_here
DEFAULT_PROVIDER=gemini
DEFAULT_MODEL=gemini-1.5-pro
```

**Important**: `.env.local` is in `.gitignore` and never committed.

## Usage

### Interactive Mode

Run tests with a menu:

```bash
./test/local-test-suite.sh --interactive
```

Menu options:
- Run all tests
- Test individual commands
- Show help
- Exit

### Batch Mode

Run all tests automatically:

```bash
./test/local-test-suite.sh --batch
```

### Specific Test

Run a single test:

```bash
./test/local-test-suite.sh --test help
./test/local-test-suite.sh --test version
./test/local-test-suite.sh --test ask_simple
```

### With Gemini

Enable tests that use Gemini API:

```bash
./test/local-test-suite.sh --batch --with-gemini
```

### Verbose Output

Show detailed output:

```bash
./test/local-test-suite.sh --batch --verbose
```

### Debug Mode

Show debug information:

```bash
./test/local-test-suite.sh --batch --debug
```

### Trace Mode

Show command execution traces:

```bash
./test/local-test-suite.sh --batch --trace
```

### Save Outputs

Save test outputs to files:

```bash
./test/local-test-suite.sh --batch --save-outputs
```

Outputs saved to: `test/outputs/`

## Available Tests

### Core Tests

- **help**: Test `vectora --help` command
- **version**: Test `vectora --version` command
- **ask_simple**: Test `vectora ask` with simple prompt
- **ask_with_gemini**: Test with Gemini model (requires API key)
- **configure_list**: Test configuration commands
- **list_providers**: Test listing available models/providers
- **completion**: Test shell completion generation
- **invalid_command**: Test error handling for invalid commands

### Run Specific Tests

```bash
./test/local-test-suite.sh --test help
./test/local-test-suite.sh --test version
./test/local-test-suite.sh --test ask_simple
./test/local-test-suite.sh --test ask_with_gemini
./test/local-test-suite.sh --test configure_list
./test/local-test-suite.sh --test completion
./test/local-test-suite.sh --test invalid_command
```

## Test Output

### Success Output

```
✓ Help command works
✓ Version: 0.1.0
✓ Ask command executed successfully
✓ Configure list executed successfully
✓ List models/providers executed successfully
✓ Completion script generated successfully
✓ Invalid command properly rejected
```

### Summary Report

```
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
 Test Results Summary
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Total Tests: 7
Passed: 7
Failed: 0
Skipped: 0
Success Rate: 100%

Details:
  ✓ help
  ✓ version
  ✓ ask_simple
  ✓ ask_with_gemini
  ✓ configure_list
  ✓ list_providers
  ✓ completion
  ✓ invalid_command

✓ All tests passed!
```

## Troubleshooting

### Vectora binary not found

Build Vectora first:

```bash
cd /path/to/vectora
go build -o bin/vectora ./cmd/core
```

### API key errors

Check that your `.env.local` has valid API keys:

```bash
grep -E "GEMINI_API_KEY|ANTHROPIC_API_KEY|OPENAI_API_KEY" test/.env.local
```

### Timeout errors

Increase timeout in `.env.local`:

```bash
TEST_TIMEOUT=60  # Default is 30 seconds
```

### Commands not found

Make sure Vectora was built correctly:

```bash
./bin/vectora --help
./bin/vectora --version
```

## File Structure

```
test/
├── README.md                      # This file
├── local-test-suite.sh           # Main test script
├── setup-test-env.sh             # Environment setup
├── .env.local.example            # Template
├── fixtures/                     # Test data
│   ├── prompts.txt              # Test prompts
│   └── queries.txt              # Test queries
├── outputs/                      # Test output files
│   └── last-output.log
└── reports/                      # Generated reports
    ├── test-report.html
    ├── test-report.json
    └── test-report.txt
```

## Environment Variables

All variables in `.env.local`:

| Variable | Description | Default |
|----------|-------------|---------|
| `GEMINI_API_KEY` | Google Gemini API key | - |
| `ANTHROPIC_API_KEY` | Anthropic Claude API key | - |
| `OPENAI_API_KEY` | OpenAI API key | - |
| `VECTORA_LOG_LEVEL` | Logging level | info |
| `VECTORA_MODE` | Test mode | test |
| `TEST_TIMEOUT` | Timeout for tests (seconds) | 30 |
| `TEST_VERBOSE` | Verbose output | false |
| `TEST_SAVE_OUTPUTS` | Save outputs to file | true |
| `DEFAULT_PROVIDER` | Default provider | gemini |
| `DEFAULT_MODEL` | Default model | gemini-1.5-pro |

## Common Commands

```bash
# Setup
./test/setup-test-env.sh

# Interactive testing
./test/local-test-suite.sh --interactive

# All tests
./test/local-test-suite.sh --batch

# Specific test
./test/local-test-suite.sh --test ask_simple

# With Gemini
./test/local-test-suite.sh --batch --with-gemini

# Verbose mode
./test/local-test-suite.sh --batch --verbose

# Debug mode
./test/local-test-suite.sh --batch --debug --trace

# Save outputs
./test/local-test-suite.sh --batch --save-outputs
```

## Performance Benchmarks

Typical execution times:

- `--help`: ~100ms
- `--version`: ~50ms
- `ask` (simple): ~2-5 seconds
- `ask` (with Gemini): ~3-8 seconds
- `configure --list`: ~500ms
- Full test suite: ~20-30 seconds

## Best Practices

1. **Always setup first**: Run `./test/setup-test-env.sh` before first use
2. **Fill .env.local**: Add at least one API key for full testing
3. **Use specific tests during development**: Faster than running all tests
4. **Check verbose output**: Use `--verbose` to understand failures
5. **Save outputs**: Use `--save-outputs` for debugging
6. **Clean outputs**: Regularly clean `outputs/` directory

## Notes

- Tests run commands against the actual Vectora binary
- All API calls are real (not mocked)
- Timeout is 30 seconds by default (configurable)
- Exit codes are validated (0 for success, non-0 for errors)
- Output is captured and validated for expected patterns

## Support

For issues or questions:

1. Check `.env.local` is properly configured
2. Verify Vectora binary exists: `./bin/vectora --help`
3. Run with `--verbose --debug --trace` for detailed output
4. Check test outputs in `test/outputs/`

## License

See root LICENSE file.
