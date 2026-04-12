# Vectora Local Test Suite - Summary

## What Was Created

A **production-ready local testing framework** for Vectora CLI that allows developers to test all commands locally with real API keys (Gemini, Claude, OpenAI).

## Quick Start (5 Minutes)

```bash
# 1. Setup environment
cd /path/to/vectora
chmod +x test/*.sh
./test/setup-test-env.sh

# 2. Edit .env.local with your API keys
nano test/.env.local

# 3. Run tests
./test/local-test-suite.sh --interactive
# or
./test/local-test-suite.sh --batch
```

## Key Features

✅ **Interactive Mode** - Choose which tests to run
✅ **Batch Mode** - Run all tests automatically  
✅ **Specific Tests** - Test individual commands
✅ **Gemini Support** - Real API calls with Gemini
✅ **Real API Keys** - Use .env.local for Gemini, Claude, OpenAI
✅ **Verbose Output** - See detailed information
✅ **Debug Mode** - Full execution traces
✅ **Performance Timing** - Measure response times
✅ **Color Output** - Easy to read results
✅ **No Server Needed** - CLI-only testing

## File Structure

```
test/
├── local-test-suite.sh           # Main test script (400+ lines)
├── setup-test-env.sh             # Environment setup
├── README.md                      # Full documentation
├── QUICK_START.md                # 5-minute guide
├── DEVELOPMENT_INTEGRATION.md    # IDE/CI-CD integration
├── .env.local.example            # Template (copy this to .env.local)
└── fixtures/
    ├── prompts.txt              # 20 test prompts
    └── config-examples.json     # Configuration examples
```

## Usage Examples

### Setup First Time
```bash
./test/setup-test-env.sh
```
Creates `.env.local` and validates setup.

### Interactive Testing
```bash
./test/local-test-suite.sh --interactive
```
Menu-driven testing - great for learning.

### All Tests (Batch)
```bash
./test/local-test-suite.sh --batch
```
Run complete suite automatically.

### Specific Test
```bash
./test/local-test-suite.sh --test help
./test/local-test-suite.sh --test version
./test/local-test-suite.sh --test ask_simple
```

### With Gemini
```bash
./test/local-test-suite.sh --batch --with-gemini
```
Requires `GEMINI_API_KEY` in `.env.local`

### Verbose Mode
```bash
./test/local-test-suite.sh --batch --verbose
```
Show detailed output.

### Debug Mode
```bash
./test/local-test-suite.sh --batch --debug --trace
```
Full execution traces for debugging.

## What Gets Tested

### Core Commands
- `vectora --help` - Help text
- `vectora --version` - Version number
- `vectora ask "prompt"` - Simple ask command
- `vectora ask --model gemini-1.5-pro` - With Gemini
- `vectora configure --list` - Configuration listing
- `vectora list-models` - Available models
- `vectora completion bash` - Shell completion
- Invalid commands - Error handling

### Test Results
Each test shows:
- ✓ (Pass) or ✗ (Fail)
- Execution time
- Output preview
- Error details if failed

## Configuration (.env.local)

Copy template and add your keys:

```bash
cp test/.env.local.example test/.env.local
```

Edit with your API keys:

```
GEMINI_API_KEY=your_key_here
ANTHROPIC_API_KEY=optional_key
OPENAI_API_KEY=optional_key
DEFAULT_PROVIDER=gemini
DEFAULT_MODEL=gemini-1.5-pro
TEST_TIMEOUT=30
TEST_VERBOSE=false
```

**Important**: `.env.local` is in `.gitignore` - never committed!

## Common Workflows

### Daily Development
```bash
# Quick smoke test
./test/local-test-suite.sh --test help --test version

# Or full test
./test/local-test-suite.sh --batch
```

### Before Commit
```bash
# Full test with API calls
./test/local-test-suite.sh --batch --with-gemini

# Save output for review
./test/local-test-suite.sh --batch --save-outputs --verbose
```

### Debugging
```bash
# See what's happening
./test/local-test-suite.sh --test ask --debug --trace

# Save for analysis
./test/local-test-suite.sh --test ask --save-outputs
```

## Expected Output

```
ℹ Setting up test environment...
✓ Environment ready

→ Testing: vectora --help
✓ Help command works

→ Testing: vectora --version
✓ Version: 0.1.0

→ Testing: vectora ask (simple prompt)
✓ Ask command executed successfully

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
 Test Results Summary
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Total Tests: 7
Passed: 7
Failed: 0
Skipped: 0
Success Rate: 100%

✓ All tests passed!
```

## IDE Integration

### VS Code
Add to `.vscode/tasks.json`:

```json
{
  "label": "Vectora: Test Suite",
  "type": "shell",
  "command": "./test/local-test-suite.sh",
  "args": ["--batch", "--verbose"],
  "group": {
    "kind": "test",
    "isDefault": true
  }
}
```

Then run with `Ctrl+Shift+B`.

### GoLand / IntelliJ
Run → Edit Configurations → Add Shell Script → `test/local-test-suite.sh`

## Troubleshooting

### "Vectora binary not found"
Build it:
```bash
go build -o bin/vectora ./cmd/core
```

### "API key not set"
Add to `.env.local`:
```
GEMINI_API_KEY=your_key_here
```

### Timeout errors
Increase in `.env.local`:
```
TEST_TIMEOUT=60
```

### Tests hang
Check network connectivity and API key validity.

## Documentation

- **README.md** - Complete reference (400+ lines)
- **QUICK_START.md** - 5-minute guide
- **DEVELOPMENT_INTEGRATION.md** - IDE/CI-CD setup

## Key Stats

| Metric | Value |
|--------|-------|
| Main Script Lines | 400+ |
| Documentation Lines | 1000+ |
| Test Functions | 8 |
| Supported Providers | 3 (Gemini, Claude, OpenAI) |
| Setup Time | < 5 minutes |
| Test Execution Time | 20-30 seconds |

## Next Steps

1. ✓ Setup: `./test/setup-test-env.sh`
2. ✓ Configure: Edit `test/.env.local`
3. ✓ Test: `./test/local-test-suite.sh --interactive`
4. ✓ Integrate: Add to IDE or pre-commit hooks
5. ✓ Develop: Use as part of daily workflow

## Support

- Quick questions? See `test/QUICK_START.md`
- Full details? See `test/README.md`
- IDE setup? See `test/DEVELOPMENT_INTEGRATION.md`
- Specific issue? Run with `--verbose --debug --trace`

---

**Status**: ✅ **READY FOR LOCAL TESTING**

The test suite is complete and ready to use. Start with `./test/setup-test-env.sh` and follow the quick start guide.

**Commit**: `4046546`
**Files**: 8 created, 1700+ lines
**Date**: April 2026
