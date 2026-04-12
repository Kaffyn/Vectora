# Vectora Test Suite - Development Integration Guide

How to integrate the test suite into your daily development workflow.

## Pre-Commit Testing

Before committing changes, run the full test suite:

```bash
# Full test with Gemini
./test/local-test-suite.sh --batch --with-gemini

# Or quick smoke test
./test/local-test-suite.sh --test help --test version
```

## Git Pre-Commit Hook (Optional)

Create `.git/hooks/pre-commit`:

```bash
#!/bin/bash

# Run tests before commit
echo "Running Vectora tests..."
./test/local-test-suite.sh --batch

if [ $? -ne 0 ]; then
    echo "Tests failed! Commit aborted."
    exit 1
fi

echo "All tests passed! Proceeding with commit."
```

Make it executable:

```bash
chmod +x .git/hooks/pre-commit
```

## Continuous Development Workflow

### 1. Make Changes

Edit code:

```bash
# Edit vectora CLI or extension code
code cmd/core/...
code extensions/vscode/src/...
```

### 2. Quick Test

Test the specific feature you changed:

```bash
# Test just the command you modified
./test/local-test-suite.sh --test help --verbose

# Or interactive for quick testing
./test/local-test-suite.sh --interactive
```

### 3. Full Test Before Commit

```bash
# Run complete suite
./test/local-test-suite.sh --batch --verbose

# With Gemini if available
./test/local-test-suite.sh --batch --with-gemini
```

### 4. Commit

```bash
git add .
git commit -m "Your commit message"
```

## Performance Testing

Monitor how your changes affect performance:

```bash
# Run tests and save outputs for comparison
./test/local-test-suite.sh --batch --save-outputs --verbose

# Check outputs/last-output.log for timing information
cat test/outputs/last-output.log
```

## Adding New Tests

### 1. Create test function

Edit `test/local-test-suite.sh`:

```bash
test_my_new_feature() {
    print_section "Testing: my new feature"
    
    local output
    output=$(run_command "$VECTORA_BIN my-command" 2>&1)
    local exit_code=$?
    
    if [[ $exit_code -eq 0 ]]; then
        log_success "My feature works"
        return 0
    else
        log_error "My feature failed"
        return 1
    fi
}
```

### 2. Add to test runner

In the `run_all_tests()` function:

```bash
run_test "my_new_feature"
```

### 3. Test it

```bash
./test/local-test-suite.sh --test my_new_feature --verbose
```

## Debugging Test Failures

### Enable Debug Mode

```bash
./test/local-test-suite.sh --test failing_test --debug --trace
```

Shows:
- Full command execution
- Exit codes
- Timing information
- Variable values

### Save Outputs

```bash
./test/local-test-suite.sh --test failing_test --save-outputs --verbose
```

Check `test/outputs/last-output.log` for the full output.

### Manual Testing

Run the command directly:

```bash
# Get the actual Vectora binary
./bin/vectora --help
./bin/vectora ask "test prompt"
```

## Environment Variables for Development

Add to `.env.local` for development:

```bash
# More verbose logging
VECTORA_LOG_LEVEL=debug

# Enable timing output
VECTORA_TIMING=true

# Shorter timeout for faster feedback
TEST_TIMEOUT=15

# Always save outputs
TEST_SAVE_OUTPUTS=true

# Always verbose
TEST_VERBOSE=true
```

## IDE Integration

### VS Code

Add to `.vscode/tasks.json`:

```json
{
  "version": "2.0.0",
  "tasks": [
    {
      "label": "Vectora: Test Suite",
      "type": "shell",
      "command": "./test/local-test-suite.sh",
      "args": ["--batch", "--verbose"],
      "problemMatcher": [],
      "group": {
        "kind": "test",
        "isDefault": true
      }
    },
    {
      "label": "Vectora: Test Interactive",
      "type": "shell",
      "command": "./test/local-test-suite.sh",
      "args": ["--interactive"],
      "isBackground": false,
      "group": "test"
    },
    {
      "label": "Vectora: Quick Test",
      "type": "shell",
      "command": "./test/local-test-suite.sh",
      "args": ["--test", "help", "--test", "version"],
      "group": "test"
    }
  ]
}
```

Then run tests via:
- `Ctrl+Shift+D` (Debug) or `Ctrl+Shift+B` (Build)
- Or Command Palette: `Tasks: Run Task`

### GoLand / IntelliJ

Create Run Configuration:
1. Run → Edit Configurations
2. Click `+` to add new configuration
3. Select "Shell Script"
4. Set Script path: `test/local-test-suite.sh`
5. Set Arguments: `--batch --verbose`
6. Click OK

Then run with: `Ctrl+Shift+R` or via Run menu

## CI/CD Integration

The test suite works with GitHub Actions, GitLab CI, etc.

Example for GitHub Actions (already configured):

```yaml
- name: Run local test suite
  run: |
    chmod +x test/*.sh
    ./test/setup-test-env.sh
    ./test/local-test-suite.sh --batch --with-gemini
```

## Best Practices

1. **Run quick tests during development**
   ```bash
   ./test/local-test-suite.sh --test help
   ```

2. **Run full suite before committing**
   ```bash
   ./test/local-test-suite.sh --batch
   ```

3. **Use verbose mode when debugging**
   ```bash
   ./test/local-test-suite.sh --test failing --verbose
   ```

4. **Save outputs for review**
   ```bash
   ./test/local-test-suite.sh --batch --save-outputs
   ```

5. **Add new tests for new features**
   - Keep tests simple and isolated
   - Test both success and error cases
   - Document what each test validates

6. **Keep .env.local updated**
   - Add API keys for all providers you want to test
   - Update timeouts based on your internet speed
   - Set appropriate log levels

## Troubleshooting

### Tests hang/timeout

- Increase `TEST_TIMEOUT` in `.env.local`
- Check network connectivity
- Verify API keys are valid

### Tests pass locally but fail in CI

- Check environment variables are set in CI
- Verify same Go/Node versions
- Check API keys available in CI

### API calls not working

- Verify API keys in `.env.local`
- Check internet connectivity
- Try with `--verbose --debug` for details

## Advanced Usage

### Custom Test Metrics

```bash
# Time how long tests take
time ./test/local-test-suite.sh --batch

# Profile memory usage
/usr/bin/time -v ./test/local-test-suite.sh --batch

# Monitor with watch
watch -n 5 'ps aux | grep vectora'
```

### Batch Testing Multiple Configs

Create `test-batch.sh`:

```bash
#!/bin/bash
for provider in gemini claude openai; do
    echo "Testing with $provider..."
    DEFAULT_PROVIDER=$provider ./test/local-test-suite.sh --batch
done
```

### Performance Regression Detection

Track response times:

```bash
# Save baseline
./test/local-test-suite.sh --batch --save-outputs
cp test/outputs/last-output.log baseline.log

# After changes
./test/local-test-suite.sh --batch --save-outputs
diff baseline.log test/outputs/last-output.log
```

## Support

- Quick Start: `test/QUICK_START.md`
- Full Docs: `test/README.md`
- GitHub Actions: `.github/workflows/`

## Summary

The test suite provides:
- ✅ Quick feedback during development
- ✅ Confidence before commits
- ✅ IDE integration for convenience
- ✅ CI/CD automation support
- ✅ Detailed debugging information
- ✅ Performance tracking
- ✅ Easy extensibility

Integrate it into your workflow for better code quality and faster development!

---

**Happy coding! 🚀**
