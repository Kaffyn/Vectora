# GitHub Secrets Configuration Guide

This guide explains how to configure secrets for the Vectora CI/CD pipeline.

## Overview

Secrets are encrypted environment variables used by GitHub Actions. They're never logged and are only injected at runtime.

## Required Secrets

### 1. GEMINI_API_KEY

**Purpose:** Gemini API key for running tests with Google's Gemini model

**How to obtain:**
1. Go to [Google AI Studio](https://aistudio.google.com/app/apikeys)
2. Create a new API key
3. Copy the key (starts with `AIza...`)

**Format:** `AIza[alphanumeric-string]`

**Used in:**
- Command tests (`vectora ask`, `vectora chat`)
- RAG tests
- Integration tests

### 2. ANTHROPIC_API_KEY (Optional)

**Purpose:** Anthropic API key for Claude model testing

**How to obtain:**
1. Go to [Anthropic Console](https://console.anthropic.com/)
2. Navigate to API keys
3. Create a new key
4. Copy the key (starts with `sk-ant-...`)

**Format:** `sk-ant-[alphanumeric-string]`

**Used in:**
- Alternative provider tests
- Model switching tests

### 3. OPENAI_API_KEY (Optional)

**Purpose:** OpenAI API key for GPT model testing

**How to obtain:**
1. Go to [OpenAI API Keys](https://platform.openai.com/api-keys)
2. Create a new secret key
3. Copy the key (starts with `sk-...`)

**Format:** `sk-[alphanumeric-string]`

**Used in:**
- Alternative provider tests
- GPT model tests

### 4. VOYAGE_API_KEY (Optional)

**Purpose:** Voyage AI embedding model key

**How to obtain:**
1. Go to [Voyage AI Dashboard](https://dash.voyageai.com)
2. Navigate to API keys
3. Create a new key
4. Copy the key

**Format:** `pa-[alphanumeric-string]`

**Used in:**
- RAG embedding tests

## Setup Instructions

### Via GitHub Web Interface

1. **Navigate to Secrets:**
   - Go to your repository
   - Settings → Secrets and variables → Actions

2. **Create a New Secret:**
   - Click "New repository secret"
   - Name: `GEMINI_API_KEY` (exactly as shown)
   - Value: Paste your API key
   - Click "Add secret"

3. **Repeat for other secrets:**
   - `ANTHROPIC_API_KEY`
   - `OPENAI_API_KEY`
   - `VOYAGE_API_KEY`

### Via GitHub CLI

```bash
# Login if needed
gh auth login

# Add GEMINI_API_KEY
gh secret set GEMINI_API_KEY

# Add ANTHROPIC_API_KEY
gh secret set ANTHROPIC_API_KEY

# Add OPENAI_API_KEY
gh secret set OPENAI_API_KEY

# Add VOYAGE_API_KEY
gh secret set VOYAGE_API_KEY

# List all secrets
gh secret list
```

### Via Repository File (Not Recommended)

For organization/team settings, you can configure defaults in `.github/settings.yml`:

```yaml
# .github/settings.yml
secrets:
  required:
    - GEMINI_API_KEY
  optional:
    - ANTHROPIC_API_KEY
    - OPENAI_API_KEY
    - VOYAGE_API_KEY
```

## Usage in Workflows

Secrets are accessed using the `secrets` context:

```yaml
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - name: Setup test environment
        env:
          GEMINI_API_KEY: ${{ secrets.GEMINI_API_KEY }}
        run: ./scripts/setup-test-env.sh
```

## Security Best Practices

### ✅ DO

- [ ] **Rotate keys regularly** - Change API keys every 90 days
- [ ] **Use repository secrets** - Not organization/user secrets
- [ ] **Limit key permissions** - Grant only necessary API permissions
- [ ] **Monitor key usage** - Check provider dashboards for unusual activity
- [ ] **Use separate keys** - Different keys for CI, development, production
- [ ] **Document expiry** - Track when keys expire

### ❌ DON'T

- [ ] **Commit secrets to git** - Never store in `.env` files that are tracked
- [ ] **Log secrets** - Don't print them in workflow logs
- [ ] **Share via chat/email** - Use secure key management
- [ ] **Use production keys** - Create dedicated CI/test keys
- [ ] **Disable masking** - Let GitHub mask secrets in logs
- [ ] **Add to .github/ config** - Keep secrets encrypted, not in YAML files

## Secret Expiration Checklist

| Secret | Provider | Expiry | Last Rotated | Next Rotation |
|--------|----------|--------|--------------|---------------|
| GEMINI_API_KEY | Google AI | 90 days | 2024-04-12 | 2024-07-11 |
| ANTHROPIC_API_KEY | Anthropic | 180 days | - | - |
| OPENAI_API_KEY | OpenAI | 90 days | - | - |
| VOYAGE_API_KEY | Voyage AI | 180 days | - | - |

## Troubleshooting

### Secrets Not Recognized

**Symptom:** Workflow fails with "secret not found" error

**Solutions:**
1. Check secret name matches exactly (case-sensitive)
2. Verify secret was saved (refresh page)
3. Ensure workflow has access to secrets
4. Check branch permissions (only default branch has access)

### Secret Appears in Logs

**If this happens:**
1. Immediately rotate the compromised key
2. Review key usage in provider dashboard
3. Remove the key from any public repositories
4. Consider this an exposure incident

**To prevent:**
- GitHub automatically masks secrets in logs
- Don't use `set-output` with secrets
- Avoid echoing secrets in scripts

### Failed API Calls in Tests

**Symptom:** Tests fail with "invalid API key" errors

**Solutions:**
1. Verify API key is correct
2. Check if key has correct permissions
3. Confirm key hasn't expired
4. Test key outside CI first
5. Check rate limits in provider dashboard

## Handling Multiple Environments

### Development Environment

```bash
# Local .env (not committed)
GEMINI_API_KEY=dev-key-here
```

### CI/CD Environment

```yaml
# GitHub Secrets (encrypted)
GEMINI_API_KEY=prod-key-here
```

### Separate Staging/Production Keys

For multiple environments, use GitHub environments:

1. Go to Settings → Environments
2. Create new environment (e.g., "staging", "production")
3. Add secrets per environment
4. Reference in workflow:

```yaml
jobs:
  test:
    environment: staging
    steps:
      - run: echo ${{ secrets.GEMINI_API_KEY }}
```

## Integration with Providers

### Google AI Studio (Gemini)

After creating a key:
- **Rate limit:** 60 requests/minute (free tier)
- **Cost:** Free tier available
- **Expiry:** No automatic expiry
- **Dashboard:** https://aistudio.google.com/app

### Anthropic

After creating a key:
- **Rate limit:** Depends on plan
- **Cost:** Pay-as-you-go
- **Expiry:** 180 days
- **Dashboard:** https://console.anthropic.com/

### OpenAI

After creating a key:
- **Rate limit:** Depends on plan
- **Cost:** Pay-as-you-go
- **Expiry:** Manual only
- **Dashboard:** https://platform.openai.com/

### Voyage AI

After creating a key:
- **Rate limit:** Depends on plan
- **Cost:** Free tier available
- **Expiry:** Manual only
- **Dashboard:** https://dash.voyageai.com/

## Emergency Procedures

### Compromised Secret

1. **Immediately:**
   - Go to Secrets and delete the compromised secret
   - Rotate/regenerate the key in the provider
   - Run security audit in the provider

2. **Within 1 hour:**
   - Review all workflow runs that used the secret
   - Check provider audit logs for unauthorized usage
   - Add new key to GitHub Secrets

3. **Post-incident:**
   - Document the incident
   - Review security practices
   - Consider additional monitoring
   - Implement key rotation schedule

### Lost Access to Provider

If you can't access the provider to rotate keys:

1. Delete the secret from GitHub
2. Create new API key through alternative method
3. Add new key to GitHub Secrets
4. Retry failed workflow

## References

- [GitHub Secrets Documentation](https://docs.github.com/en/actions/security-guides/encrypted-secrets)
- [Google AI Studio Documentation](https://ai.google.dev/docs)
- [Anthropic API Documentation](https://docs.anthropic.com/)
- [OpenAI API Documentation](https://platform.openai.com/docs)
- [Voyage AI Documentation](https://docs.voyageai.com/)
