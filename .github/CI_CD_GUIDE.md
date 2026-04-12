# Vectora CI/CD Complete Guide

Complete guide to the Vectora continuous integration and continuous deployment pipeline.

## Table of Contents

1. [Quick Start](#quick-start)
2. [Architecture](#architecture)
3. [Workflows](#workflows)
4. [Testing](#testing)
5. [Secrets Management](#secrets-management)
6. [Local Development](#local-development)
7. [Troubleshooting](#troubleshooting)
8. [Best Practices](#best-practices)

## Quick Start

### First Time Setup

```bash
# 1. Clone repository
git clone https://github.com/Kaffyn/Vectora.git
cd Vectora

# 2. Set up GitHub Secrets (see SECRETS_CONFIG.md)
# Go to: Settings → Secrets and variables → Actions
# Add: GEMINI_API_KEY

# 3. Build locally
go build -o bin/vectora ./cmd/core

# 4. Run tests
./scripts/setup-test-env.sh
./scripts/run-command-suite.sh

# 5. Push to trigger CI
git push origin feature-branch
```

### View CI Status

```bash
# GitHub CLI
gh run list
gh run view <run-id>

# Or visit GitHub Actions tab in web browser
```

## Architecture

### Pipeline Flow

```
Push to repository
        ↓
┌───────────────────────────────────────┐
│   CI.yml - Main Orchestrator          │
├───────────────────────────────────────┤
│ 1. Orchestrate (conditions)           │
│ 2. Tests (lint, type-check, build)    │
│ 3. Build (if commit, not tag)         │
│ 4. Release (if version tag)           │
│ 5. Pipeline Summary                   │
└───────────────────────────────────────┘
        ↓ (Parallel)
    ┌──────┬──────┬──────────┬────────┐
    │      │      │          │        │
    ↓      ↓      ↓          ↓        ↓
  Tests Security Perf  Extension  CLI
  (multi-os) (scan)  (monitor)  Tests Tests
    │      │      │          │        │
    └──────┴──────┴──────────┴────────┘
        ↓
   [Test Results]
        ↓
   [Deploy/Release]
```

### Job Dependencies

```
orchestrate
    ↓
  tests (depends on: orchestrate)
    ↓
  build (depends on: tests, if commit)
    ↓
  release (depends on: tests, if tag)
```

## Workflows

### CI Workflow (`ci.yml`)

Main pipeline that orchestrates all stages.

**Triggers:**
- Push to master/develop
- Pull requests
- Version tags
- Manual dispatch

**Jobs:**
1. **orchestrate** - Determine which stages to run
2. **tests** - Lint, type-check, unit tests
3. **build** - Create binaries
4. **release** - Create GitHub release
5. **pipeline-summary** - Report results

**Key Features:**
- Conditional execution based on event type
- Automatic version detection from tags
- Release notes generation

### Tests Workflow (`tests.yml`)

Comprehensive command and integration testing.

**Triggers:**
- Push to master/develop
- Pull requests
- Manual dispatch (with optional full test)

**Jobs:**
1. **command-tests** - Test all CLI commands
2. **extension-tests** - VS Code extension
3. **cli-extension-tests** - Gemini CLI
4. **integration-tests** - Core integration
5. **test-summary** - Final report

**Matrix:**
- OS: Linux, macOS, Windows
- Go: 1.22, 1.23
- Node: 20, 22

### Security Workflow (`security.yml`)

Security scanning and code quality checks.

**Triggers:**
- Push to master/develop
- Pull requests
- Daily at 2 AM UTC
- Manual dispatch

**Jobs:**
1. **dependency-scan** - npm and Go
2. **trivy-scan** - Vulnerability scan
3. **secret-scan** - Secret detection
4. **code-quality** - Lint and type checks
5. **license-check** - License compliance
6. **security-summary** - Report

### Performance Workflow (`performance.yml`)

Bundle size and build performance monitoring.

**Triggers:**
- Push to master/develop
- Pull requests
- Manual dispatch

**Jobs:**
1. **bundle-size** - Extension bundle analysis
2. **benchmarks** - Go benchmarks
3. **build-time** - Build time tracking
4. **performance-summary** - Report and PR comments

## Testing

### Test Scripts

Located in `scripts/` directory:

```bash
# Verify help command
./scripts/verify-help.sh

# Setup test environment
./scripts/setup-test-env.sh

# Run all command tests
./scripts/run-command-suite.sh

# Test individual command
./scripts/test-vectora-commands.sh ask "question"
```

### Test Results

JSON format test report:

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

### Coverage Reports

Coverage uploaded to Codecov:

```bash
# View locally
go tool cover -html=coverage.out
```

## Secrets Management

### Required Secrets

| Secret | Source | Used For |
|--------|--------|----------|
| GEMINI_API_KEY | Google AI Studio | Gemini tests |
| ANTHROPIC_API_KEY | Anthropic Console | Claude tests (optional) |
| OPENAI_API_KEY | OpenAI Dashboard | GPT tests (optional) |
| VOYAGE_API_KEY | Voyage AI Dashboard | Embedding tests (optional) |

### Setup

```bash
# Via GitHub CLI
gh secret set GEMINI_API_KEY

# Or via web interface:
# Settings → Secrets and variables → Actions → New repository secret
```

### Security

- Secrets are encrypted at rest
- Never logged in workflow output
- Masked in logs automatically
- Only available to workflows on default branch
- Can be scoped to specific environments

See `.github/SECRETS_CONFIG.md` for detailed setup.

## Local Development

### Prerequisites

```bash
# Go 1.22+
go version

# Node.js 20+
node --version
npm --version

# Git
git --version
```

### Build Locally

```bash
# Build Vectora CLI
go build -o bin/vectora ./cmd/core

# Build VS Code extension
cd extensions/vscode
npm ci
npm run compile

# Build Gemini CLI
cd extensions/geminicli
npm ci
npm run build
```

### Run Tests Locally

```bash
# Setup test environment
./scripts/setup-test-env.sh

# Run all tests
./scripts/run-command-suite.sh

# Run specific tests
go test -v ./...
cd extensions/vscode && npm test
```

### Test with CI Environment

```bash
# Install act (GitHub Actions runner)
brew install act  # or your package manager

# Run specific workflow
act -j tests

# Run with secrets
act -s GEMINI_API_KEY=your-key
```

## Troubleshooting

### Build Fails Locally

**Check:**
1. Go version: `go version` (need 1.22+)
2. Dependencies: `go mod tidy && go mod download`
3. Build script: `cat build-quick.ps1` or `bash build.sh`

```bash
# Verbose build
go build -v -x ./cmd/core
```

### Tests Fail on CI but Pass Locally

**Likely causes:**
- Different environment variables
- Missing GitHub Secrets
- Platform-specific issues
- Network connectivity

**Debug:**
1. Check GitHub Actions logs
2. Download test artifacts
3. Verify secrets are set
4. Try `act` locally

### Secret Not Found Error

**Solution:**
1. Check secret name (case-sensitive)
2. Verify it's saved in Settings
3. Check workflow references it correctly
4. Ensure on default branch (for repository secrets)

### Workflow Timeout

Default timeouts per job:
- Command tests: 10s per command
- Integration tests: 15 minutes
- Build: Varies by platform

Increase if needed in workflow YAML:

```yaml
jobs:
  test:
    timeout-minutes: 30
```

### Files Not Found in Artifact

Check:
1. Artifact path is correct in workflow
2. File was actually created
3. Retention period hasn't expired (7 days default)
4. Workflow completed successfully

## Best Practices

### Workflow Design

✅ **Do:**
- Use matrix for parallel testing
- Cache dependencies
- Use conditions to skip unnecessary jobs
- Add clear logging
- Document changes

❌ **Don't:**
- Hardcode secrets
- Commit large binaries
- Run on every event
- Skip security checks

### Testing Strategy

✅ **Do:**
- Test on multiple platforms
- Include both unit and integration tests
- Validate error cases
- Check performance budgets
- Review test coverage

❌ **Don't:**
- Skip security tests
- Ignore flaky tests
- Commit broken tests
- Test with real API keys

### Security

✅ **Do:**
- Rotate API keys periodically
- Use GitHub Secrets for sensitive data
- Limit API key permissions
- Monitor unusual activity
- Review security scan results

❌ **Don't:**
- Share secrets in chat/email
- Use production keys in CI
- Disable security checks
- Log sensitive information
- Commit .env files

### Performance

✅ **Do:**
- Cache build outputs
- Parallelize tests
- Monitor build times
- Track bundle size
- Use fast test runners

❌ **Don't:**
- Run heavy builds on every event
- Duplicate work across jobs
- Ignore performance budgets
- Upload unnecessary artifacts

## Maintenance Checklist

### Weekly
- [ ] Review failed workflows
- [ ] Check security scan results
- [ ] Monitor performance metrics

### Monthly
- [ ] Update dependencies
- [ ] Review workflow efficiency
- [ ] Audit artifact storage

### Quarterly
- [ ] Rotate API keys
- [ ] Review and update documentation
- [ ] Performance analysis
- [ ] Security audit

### Annually
- [ ] Major version upgrades
- [ ] Complete rewrite if needed
- [ ] Strategic planning
- [ ] Team training

## References

### Documentation
- [Workflow Guide](workflows/README.md)
- [Testing Guide](workflows/TESTING.md)
- [Secrets Configuration](.github/SECRETS_CONFIG.md)

### External Resources
- [GitHub Actions Docs](https://docs.github.com/en/actions)
- [Workflow Syntax](https://docs.github.com/en/actions/using-workflows)
- [Security Guide](https://docs.github.com/en/actions/security-guides)

### Tools
- [act - Local GitHub Actions](https://github.com/nektos/act)
- [GitHub CLI](https://cli.github.com/)

## Support & Contact

For issues:
1. Check troubleshooting section above
2. Review workflow logs in GitHub Actions
3. Check existing issues/discussions
4. Open new issue with details
5. Contact maintainers

---

**Last Updated:** April 12, 2024
**Version:** 1.0
**Maintainers:** Vectora Team
