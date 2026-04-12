# Vectora Test Suite - Quick Start Guide

Get started testing Vectora CLI in 5 minutes.

## Step 1: Setup Environment

```bash
cd /path/to/vectora
chmod +x test/*.sh
./test/setup-test-env.sh
```

This will:
- Check if Go is installed
- Build Vectora binary if needed
- Create `.env.local` from template
- Validate configuration

## Step 2: Add Your API Keys

Edit `test/.env.local`:

```bash
nano test/.env.local
# or
code test/.env.local
```

Add at least one API key:

```
GEMINI_API_KEY=your_google_gemini_key
ANTHROPIC_API_KEY=your_anthropic_key
OPENAI_API_KEY=your_openai_key
DEFAULT_PROVIDER=gemini
DEFAULT_MODEL=gemini-1.5-pro
```

Save and close.

## Step 3: Run Tests

### Interactive Mode (Recommended for first time)

```bash
./test/local-test-suite.sh --interactive
```

Then choose options from menu:
1. Run all tests
2. Test specific commands
3. Show help
4. Exit

### All Tests at Once

```bash
./test/local-test-suite.sh --batch
```

### Specific Test

```bash
./test/local-test-suite.sh --test help
./test/local-test-suite.sh --test version
./test/local-test-suite.sh --test ask_simple
```

### With Gemini (requires API key)

```bash
./test/local-test-suite.sh --batch --with-gemini
```

### Verbose Output

```bash
./test/local-test-suite.sh --batch --verbose
```

## What Gets Tested

✓ Help command (`vectora --help`)
✓ Version (`vectora --version`)
✓ Ask command (`vectora ask "prompt"`)
✓ Ask with Gemini (`vectora ask --model gemini-1.5-pro`)
✓ Configuration (`vectora configure --list`)
✓ Models/Providers listing
✓ Shell completion generation
✓ Error handling (invalid commands)

## Example Output

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

## Troubleshooting

### "Vectora binary not found"

Build it:

```bash
go build -o bin/vectora ./cmd/core
```

### "API key not set" warning

It's OK to ignore - tests will work with limited functionality.

### Test timeout

Increase timeout in `.env.local`:

```
TEST_TIMEOUT=60
```

### No valid API keys

Tests will still run but skip API-dependent tests. Add at least one key to `.env.local`.

## Common Workflows

### Daily Development Testing

```bash
# Full test with verbose output
./test/local-test-suite.sh --batch --verbose

# Or quick test
./test/local-test-suite.sh --test help --test version
```

### Testing New Features

```bash
# Test specific command
./test/local-test-suite.sh --test ask --verbose

# Or interactive
./test/local-test-suite.sh --interactive
```

### Before Commit

```bash
# Full suite
./test/local-test-suite.sh --batch --with-gemini

# Save outputs for review
./test/local-test-suite.sh --batch --save-outputs --verbose
```

### Debugging Failures

```bash
# Get detailed debug info
./test/local-test-suite.sh --test ask --debug --trace

# Save output for analysis
./test/local-test-suite.sh --test ask --verbose --save-outputs
```

## File Locations

| Item | Location |
|------|----------|
| Main script | `test/local-test-suite.sh` |
| Setup script | `test/setup-test-env.sh` |
| Config template | `test/.env.local.example` |
| Your config | `test/.env.local` (git-ignored) |
| Test fixtures | `test/fixtures/` |
| Test outputs | `test/outputs/` |
| Reports | `test/reports/` |
| This guide | `test/QUICK_START.md` |
| Full docs | `test/README.md` |

## Next Steps

1. ✓ Run setup: `./test/setup-test-env.sh`
2. ✓ Edit `.env.local` with your API keys
3. ✓ Try interactive: `./test/local-test-suite.sh --interactive`
4. ✓ Run all tests: `./test/local-test-suite.sh --batch`
5. ✓ Integrate into development workflow

## Tips

- Use `--verbose` to see command outputs
- Use `--debug` to see detailed information
- Use `--trace` to see executed commands
- Use `--save-outputs` to save results
- Run specific tests during development (faster)
- Run full suite before committing

## Need Help?

1. Check `test/README.md` for detailed documentation
2. Run with `--verbose --debug --trace` for debugging
3. Check `test/outputs/` for saved outputs
4. Verify `.env.local` has valid API keys

---

**Happy Testing! 🎉**

For more details, see `test/README.md`.
