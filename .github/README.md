# Vectora GitHub Configuration

This directory contains GitHub-specific configuration including workflows, issue templates, and documentation.

## Contents

### Workflows (`.github/workflows/`)

Automated CI/CD pipelines for testing, building, and releasing.

| Workflow | Purpose | Trigger |
|----------|---------|---------|
| `ci.yml` | Main orchestrator | Push, PR, tags |
| `tests.yml` | Command & integration tests | Push, PR, manual |
| `security.yml` | Security scanning | Push, PR, daily |
| `performance.yml` | Bundle & performance monitoring | Push, PR, manual |
| `build.yml` | Quick build validation | Manual |
| `release.yml` | Release management | Tags, manual |
| `extension-vscode.yml` | VS Code extension build | Push, PR |
| `extension-geminicli.yml` | Gemini CLI build | Push, PR |

See [workflows/README.md](workflows/README.md) for detailed information.

### Documentation

#### [CI_CD_GUIDE.md](CI_CD_GUIDE.md)
Complete guide to the CI/CD pipeline including:
- Quick start
- Architecture overview
- All workflows explained
- Testing procedures
- Secrets management
- Local development
- Troubleshooting
- Best practices

#### [SECRETS_CONFIG.md](SECRETS_CONFIG.md)
Secrets configuration and management:
- Required secrets
- How to obtain API keys
- Setup instructions
- Security best practices
- Troubleshooting
- Key rotation procedures

#### [workflows/README.md](workflows/README.md)
Workflow-specific documentation:
- Overview of all workflows
- Configuration details
- Test matrix strategy
- Performance budgets
- Artifact retention
- Debugging tips
- Monitoring and alerts

#### [workflows/TESTING.md](workflows/TESTING.md)
Testing guide:
- Overview of test pipeline
- Test scripts
- Running tests locally
- Test results format
- Coverage reports
- Troubleshooting

## Quick Links

### For First-Time Setup
1. Read [CI_CD_GUIDE.md](CI_CD_GUIDE.md#quick-start)
2. Configure secrets: [SECRETS_CONFIG.md](SECRETS_CONFIG.md#setup-instructions)
3. Run local tests: [TESTING.md](workflows/TESTING.md#running-tests-locally)

### For GitHub Maintainers
1. Workflow overview: [workflows/README.md](workflows/README.md)
2. Monitoring jobs: [workflows/README.md](workflows/README.md#workflow-status-badge)
3. Setting up rules: [workflows/README.md](workflows/README.md#branch-protection-rules)

### For Developers
1. Quick start: [CI_CD_GUIDE.md](CI_CD_GUIDE.md#quick-start)
2. Local testing: [CI_CD_GUIDE.md](CI_CD_GUIDE.md#local-development)
3. Troubleshooting: [CI_CD_GUIDE.md](CI_CD_GUIDE.md#troubleshooting)

### For DevOps
1. Architecture: [CI_CD_GUIDE.md](CI_CD_GUIDE.md#architecture)
2. Performance: [workflows/README.md](workflows/README.md#performance-tips)
3. Maintenance: [CI_CD_GUIDE.md](CI_CD_GUIDE.md#maintenance-checklist)

## Test Scripts

Located in `/scripts/` directory:

| Script | Purpose |
|--------|---------|
| `verify-help.sh` | Verify `vectora --help` command |
| `setup-test-env.sh` | Setup test environment with secrets |
| `run-command-suite.sh` | Full command test suite |
| `test-vectora-commands.sh` | Test individual command |

Usage:
```bash
./scripts/setup-test-env.sh
./scripts/run-command-suite.sh
./scripts/test-vectora-commands.sh ask "test"
```

## GitHub Secrets

Required for CI/CD (configure in Settings → Secrets):

```
GEMINI_API_KEY          # Google Gemini API key (required)
ANTHROPIC_API_KEY       # Anthropic Claude key (optional)
OPENAI_API_KEY          # OpenAI GPT key (optional)
VOYAGE_API_KEY          # Voyage AI key (optional)
```

See [SECRETS_CONFIG.md](SECRETS_CONFIG.md) for setup details.

## Environment Files

### `.env.example`
Public template with all available configuration options.
Commit this file to git.

### `.env.test`
Test environment with mock values.
Auto-generated during CI, git-ignored locally.

### `.env` (local development)
Your personal configuration.
Never commit this file.

## Key Features

### 1. Multi-Platform Testing
- Tests run on Ubuntu, macOS, and Windows
- Go versions: 1.22, 1.23
- Node versions: 20, 22
- Total: 12 parallel test configurations

### 2. Security Scanning
- Dependency vulnerability scanning
- Container/filesystem scanning (Trivy)
- Secret detection (TruffleHog)
- Code quality checks
- License compliance

### 3. Performance Monitoring
- Bundle size tracking
- Build time monitoring
- Go benchmarks
- PR comments with metrics

### 4. Comprehensive Testing
- CLI command testing
- Extension testing
- Integration testing
- Coverage reporting

### 5. Secure Secrets Handling
- Encrypted secrets
- Never logged or committed
- Only available to workflows
- Automatic masking

## Workflow Status

Current workflow status can be viewed at:
- GitHub Actions tab: `https://github.com/Kaffyn/Vectora/actions`
- Latest runs: `https://github.com/Kaffyn/Vectora/actions/workflows/ci.yml`

## Adding Status Badges

Add to README.md:

```markdown
[![CI/CD](https://github.com/Kaffyn/Vectora/actions/workflows/ci.yml/badge.svg)](https://github.com/Kaffyn/Vectora/actions/workflows/ci.yml)
[![Tests](https://github.com/Kaffyn/Vectora/actions/workflows/tests.yml/badge.svg)](https://github.com/Kaffyn/Vectora/actions/workflows/tests.yml)
[![Security](https://github.com/Kaffyn/Vectora/actions/workflows/security.yml/badge.svg)](https://github.com/Kaffyn/Vectora/actions/workflows/security.yml)
```

## Local Testing

Use GitHub CLI and act to test workflows locally:

```bash
# Install act
brew install act  # or your package manager

# Run specific workflow
act -j tests

# Run with secrets
act -s GEMINI_API_KEY=your-key

# View logs
act -v
```

## Maintenance

### Weekly Tasks
- Review failed workflows
- Check security scan results
- Monitor performance metrics

### Monthly Tasks
- Update dependencies
- Review workflow efficiency
- Audit artifacts

### Quarterly Tasks
- Rotate API keys
- Update documentation
- Performance analysis

## Troubleshooting

### Workflow Issues
See [CI_CD_GUIDE.md#troubleshooting](CI_CD_GUIDE.md#troubleshooting)

### Test Failures
See [workflows/TESTING.md#troubleshooting](workflows/TESTING.md#troubleshooting)

### Secret Problems
See [SECRETS_CONFIG.md#troubleshooting](SECRETS_CONFIG.md#troubleshooting)

## Support

For help:
1. Check relevant documentation above
2. Review GitHub Actions logs
3. Search existing issues
4. Open new issue with details
5. Contact maintainers

## Additional Resources

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Vectora Documentation](../README.md)
- [Contributing Guide](../CONTRIBUTING.md)

---

**Last Updated:** April 12, 2024
**Maintainers:** Vectora Team
