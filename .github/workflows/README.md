# GitHub Actions Workflows

This directory contains GitHub Actions workflows for the Vectora project CI/CD pipeline.

## Workflows Overview

### 1. `ci.yml` - Main CI Pipeline

**Purpose:** Core CI/CD orchestrator for lint, build, and tests

**When triggered:**
- Push to `master` or `develop`
- Pull requests to `master` or `develop`
- Version tags (`v*`)

**Jobs:**
1. **Orchestrate** - Determines which pipeline stages should run
2. **Tests** - Lint, type-check, unit tests, and builds
3. **Build** - Cross-platform binary building
4. **Release** - GitHub release creation for version tags
5. **Pipeline Summary** - Final status report

**Timeout:** ~15 minutes

**Artifacts:**
- Binary executables (7 day retention)
- Coverage reports (via Codecov)

### 2. `tests.yml` - Comprehensive Command & Integration Tests

**Purpose:** Test all Vectora commands and integrations

**When triggered:**
- Push to `master` or `develop`
- Pull requests to `master` or `develop`
- Manual workflow dispatch with optional full test

**Jobs:**
1. **command-tests** - Multi-platform command testing
2. **extension-tests** - VS Code extension
3. **cli-extension-tests** - Gemini CLI extension
4. **integration-tests** - Core integration
5. **test-summary** - Final test report

**Matrix Testing:**
- OS: ubuntu-latest, macos-latest, windows-latest
- Go: 1.22, 1.23
- Node: 20, 22

**Timeout:** ~30 minutes (per OS combination)

**Artifacts:**
- Test results (JSON format)
- Coverage reports
- Execution logs

### 3. `security.yml` - Security Scanning

**Purpose:** Dependency, vulnerability, and code quality scanning

**When triggered:**
- Push to `master` or `develop`
- Pull requests to `master` or `develop`
- Daily schedule (2 AM UTC)
- Manual workflow dispatch

**Jobs:**
1. **dependency-scan** - npm and Go dependency audit
2. **trivy-scan** - Container/filesystem vulnerability scan
3. **secret-scan** - Secret detection (TruffleHog)
4. **code-quality** - golangci-lint and ESLint
5. **license-check** - License compliance
6. **security-summary** - Summary report

**Timeout:** ~20 minutes

**Artifacts:**
- SARIF format security reports
- Audit logs

### 4. `performance.yml` - Performance Monitoring

**Purpose:** Monitor bundle size, build time, and performance

**When triggered:**
- Push to `master` or `develop`
- Pull requests to `master` or `develop`
- Manual workflow dispatch

**Jobs:**
1. **bundle-size** - VS Code extension bundle analysis
2. **benchmarks** - Go performance benchmarks
3. **build-time** - Build time tracking
4. **performance-summary** - Summary and PR comments

**Timeout:** ~10 minutes

**Artifacts:**
- Bundle statistics
- Benchmark results
- Build time logs

### 5. `build.yml` - Quick Build

**Purpose:** Fast build validation without full testing

**When triggered:**
- Manually or from CI orchestration

**Jobs:**
- Single cross-platform build

**Timeout:** ~5 minutes

### 6. `extension-vscode.yml` - VS Code Extension Build

**Purpose:** Build and validate VS Code extension

**Jobs:**
- Lint
- Type check
- Build

**Timeout:** ~5 minutes

### 7. `extension-geminicli.yml` - Gemini CLI Build

**Purpose:** Build and validate Gemini CLI extension

**Jobs:**
- Lint
- Type check
- Build

**Timeout:** ~5 minutes

### 8. `release.yml` - Release Management

**Purpose:** Automated release workflow

**When triggered:**
- Manual workflow dispatch
- Version tag push

**Jobs:**
- Create GitHub release
- Build artifacts
- Generate release notes

**Timeout:** ~10 minutes

## Configuration

### Secrets Required

Add these secrets to your GitHub repository (Settings → Secrets):

```
GEMINI_API_KEY          # Required for testing
ANTHROPIC_API_KEY       # Optional
OPENAI_API_KEY          # Optional
VOYAGE_API_KEY          # Optional
```

See `.github/SECRETS_CONFIG.md` for setup instructions.

### Environment Variables

Key variables used in workflows:

```yaml
GO_VERSION: "1.22"
NODE_VERSION: "20"
RUST_BACKTRACE: 1
```

## Test Matrix Strategy

### Command Tests

Runs tests on multiple combinations:

```
OS:              ubuntu-latest, macos-latest, windows-latest
Go Version:      1.22, 1.23
Node Version:    20, 22
```

This creates 12 parallel test runs for comprehensive coverage.

### Why Multi-Matrix?

1. **OS Differences** - Binary paths, environment variables differ
2. **Go Versions** - Compatibility validation
3. **Node Versions** - Extension compatibility
4. **Parallel Execution** - Faster feedback (12 jobs in parallel)

## Performance Budgets

Monitored by `performance.yml`:

| Metric | Limit | Warning |
|--------|-------|---------|
| Extension Bundle | 2 MB | >1.8 MB |
| Webview Bundle | 3 MB | >2.7 MB |
| Total Build | 5 MB | >4.5 MB |
| Go Build Time | 60s | >50s |
| Node Build Time | 120s | >100s |

## Artifact Retention

- **Build artifacts:** 7 days
- **Test results:** 7 days
- **Coverage reports:** 7 days
- **Release artifacts:** Indefinite

## Debugging Workflow Issues

### 1. Review Workflow Logs

```bash
# View workflow run details
gh run view <run-id>

# View specific job logs
gh run view <run-id> --log
```

### 2. Download Artifacts

```bash
# List artifacts
gh run download <run-id>

# Download specific artifact
gh run download <run-id> -n artifact-name
```

### 3. Common Issues

| Issue | Solution |
|-------|----------|
| Secret not found | Check spelling, save again |
| Timeout error | Increase timeout in workflow YAML |
| File not found | Check working directory |
| Permission denied | Run `chmod +x scripts/*.sh` |

### 4. Local Testing

Test workflows locally using [act](https://github.com/nektos/act):

```bash
# Install act
brew install act  # or your package manager

# Run specific workflow
act -j tests

# Run with secrets
act -s GEMINI_API_KEY=your-key
```

## Workflow Status Badge

Add to README.md:

```markdown
[![CI/CD](https://github.com/Kaffyn/Vectora/actions/workflows/ci.yml/badge.svg)](https://github.com/Kaffyn/Vectora/actions/workflows/ci.yml)
[![Tests](https://github.com/Kaffyn/Vectora/actions/workflows/tests.yml/badge.svg)](https://github.com/Kaffyn/Vectora/actions/workflows/tests.yml)
[![Security](https://github.com/Kaffyn/Vectora/actions/workflows/security.yml/badge.svg)](https://github.com/Kaffyn/Vectora/actions/workflows/security.yml)
```

## Monitoring & Alerts

### GitHub Status Page

- Monitor workflow runs at: GitHub repository → Actions tab
- Set up email notifications: Settings → Notifications

### Branch Protection Rules

Recommended rules for `master` branch:

1. Require status checks to pass:
   - `tests / Tests - Required`
   - Other critical jobs

2. Require code reviews: 1+

3. Require branch to be up to date

4. Include administrators in restrictions

## Performance Tips

### Faster Builds

1. **Cache dependencies:**
   ```yaml
   cache: npm
   cache-dependency-path: "**/package-lock.json"
   ```

2. **Parallel jobs:** Use matrix strategy

3. **Conditional steps:** Skip non-critical steps in PRs

4. **Artifact caching:** Reuse build artifacts

### Faster Tests

1. Run only affected tests
2. Parallelize test execution
3. Cache test fixtures
4. Use fast test runners (Jest, Go test -race)

## Best Practices

### ✅ Do

- [ ] Keep workflows DRY - Use reusable workflows
- [ ] Cache dependencies - Speed up builds
- [ ] Fail fast - Use conditions to skip unnecessary jobs
- [ ] Clear logging - Make it easy to debug issues
- [ ] Document changes - Update this README
- [ ] Test locally - Use `act` before pushing

### ❌ Don't

- [ ] Hardcode secrets - Use GitHub Secrets
- [ ] Commit large artifacts - Use artifact upload
- [ ] Run on every event - Use appropriate triggers
- [ ] Ignore failures - Review and fix errors
- [ ] Skip security checks - Always scan dependencies
- [ ] Disable branch protection - Enforce code quality

## Maintenance

### Regular Tasks

- **Weekly:** Review failed workflows
- **Monthly:** Update dependencies
- **Quarterly:** Review and optimize workflows
- **Annually:** Audit secrets and permissions

### Update Workflows

When adding new features:

1. Update relevant workflow(s)
2. Test with `act` locally
3. Create PR with workflow changes
4. Get review from maintainers
5. Merge and verify in CI

## References

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Workflow Syntax](https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions)
- [Security Best Practices](https://docs.github.com/en/actions/security-guides)
- [Testing Guide](./TESTING.md)
- [Secrets Configuration](../SECRETS_CONFIG.md)

## Support

For issues or questions:

1. Check [TESTING.md](./TESTING.md) for test-specific help
2. Review workflow logs in GitHub Actions
3. Open an issue with workflow details
4. Contact maintainers
