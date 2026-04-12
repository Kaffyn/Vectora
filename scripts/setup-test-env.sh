#!/bin/bash
# setup-test-env.sh - Setup test environment with secure .env handling
#
# Usage: ./scripts/setup-test-env.sh
#
# This script:
# 1. Creates .env.test file from .env.example
# 2. Injects secrets from GitHub Secrets (if running in CI)
# 3. Validates configuration
# 4. Sets up test directories

set -e

LOG_FILE="logs/setup-test-env.log"
ENV_TEST_FILE=".env.test"
ENV_EXAMPLE_FILE=".env.example"
TEST_DIR="test-data"

# Create directories
mkdir -p logs test-results test-data coverage

# Logging function
log() {
  echo "[$(date +'%Y-%m-%d %H:%M:%S')] $*" | tee -a "$LOG_FILE"
}

log "======================================================================"
log "Test Environment Setup"
log "======================================================================"
log ""

# Check if .env.example exists
if [ ! -f "$ENV_EXAMPLE_FILE" ]; then
  log "❌ ERROR: $ENV_EXAMPLE_FILE not found"
  exit 1
fi

log "✓ Found $ENV_EXAMPLE_FILE"

# Create .env.test from .env.example
log ""
log "Creating .env.test..."
cp "$ENV_EXAMPLE_FILE" "$ENV_TEST_FILE"
log "✓ Created $ENV_TEST_FILE"

# Load .env.example as baseline
log ""
log "Loading environment configuration..."
# shellcheck disable=SC1090
source "$ENV_EXAMPLE_FILE" 2>/dev/null || log "⚠ Warning: Could not source $ENV_EXAMPLE_FILE"

# Handle secrets from GitHub Actions
if [ -n "$GITHUB_ACTIONS" ]; then
  log ""
  log "Configuring GitHub Secrets..."

  if [ -n "$GEMINI_API_KEY" ]; then
    log "✓ Injecting GEMINI_API_KEY"
    # Update .env.test with the secret
    if grep -q "^GEMINI_API_KEY=" "$ENV_TEST_FILE"; then
      # Use a temporary file for sed (for macOS compatibility)
      sed_tmp=$(mktemp)
      sed "s/^GEMINI_API_KEY=.*/GEMINI_API_KEY=$GEMINI_API_KEY/" "$ENV_TEST_FILE" > "$sed_tmp"
      mv "$sed_tmp" "$ENV_TEST_FILE"
    else
      echo "GEMINI_API_KEY=$GEMINI_API_KEY" >> "$ENV_TEST_FILE"
    fi
  fi

  # Add other secrets as needed
  for secret in ANTHROPIC_API_KEY OPENAI_API_KEY VOYAGE_API_KEY; do
    secret_value=$(eval "echo \$$secret")
    if [ -n "$secret_value" ]; then
      log "✓ Injecting $secret"
      if grep -q "^$secret=" "$ENV_TEST_FILE"; then
        sed_tmp=$(mktemp)
        sed "s|^$secret=.*|$secret=$secret_value|" "$ENV_TEST_FILE" > "$sed_tmp"
        mv "$sed_tmp" "$ENV_TEST_FILE"
      else
        echo "$secret=$secret_value" >> "$ENV_TEST_FILE"
      fi
    fi
  done
fi

# Validate .env.test
log ""
log "Validating .env.test..."

# Check for required variables
REQUIRED_VARS=()
OPTIONAL_VARS=("GEMINI_API_KEY" "ANTHROPIC_API_KEY" "OPENAI_API_KEY")

VALID=true
for var in "${REQUIRED_VARS[@]}"; do
  if ! grep -q "^$var=" "$ENV_TEST_FILE"; then
    log "❌ Missing required variable: $var"
    VALID=false
  fi
done

if [ "$VALID" = true ]; then
  log "✓ All required variables present"
else
  log "⚠ Warning: Some variables are missing"
fi

# Check optional variables
log ""
log "Optional variables:"
for var in "${OPTIONAL_VARS[@]}"; do
  if grep -q "^$var=" "$ENV_TEST_FILE"; then
    # Extract value (hide sensitive data)
    value=$(grep "^$var=" "$ENV_TEST_FILE" | cut -d'=' -f2)
    if [ -n "$value" ]; then
      masked_value="${value:0:4}...${value: -4}"
      log "✓ $var=$masked_value"
    else
      log "⚠ $var is empty"
    fi
  else
    log "⚠ $var not configured"
  fi
done

# Setup test data directories
log ""
log "Setting up test directories..."

mkdir -p "$TEST_DIR/rag-index"
mkdir -p "$TEST_DIR/cache"
mkdir -p "$TEST_DIR/logs"

log "✓ Created test directories:"
log "  • $TEST_DIR/rag-index"
log "  • $TEST_DIR/cache"
log "  • $TEST_DIR/logs"

# Create test configuration
log ""
log "Creating test configuration..."

cat > ".env.test.json" << 'EOF'
{
  "test_mode": true,
  "timeout": 30000,
  "retry_attempts": 3,
  "log_level": "debug",
  "mock_responses": false,
  "coverage_threshold": {
    "lines": 70,
    "functions": 70,
    "branches": 65,
    "statements": 70
  }
}
EOF

log "✓ Created test configuration (.env.test.json)"

# Security check: Ensure .env.test is not tracked by git
log ""
log "Verifying .env.test is git-ignored..."

if grep -q "\.env\.test" .gitignore 2>/dev/null; then
  log "✓ .env.test is already in .gitignore"
else
  log "⚠ Adding .env.test to .gitignore"
  echo ".env.test" >> .gitignore
  echo ".env.test.json" >> .gitignore
fi

# Display summary
log ""
log "======================================================================"
log "✓ Test environment setup completed"
log "======================================================================"
log ""
log "Summary:"
log "  • .env.test created:      ✓"
log "  • Secrets injected:        ✓ (if CI environment)"
log "  • Configuration validated: ✓"
log "  • Test directories:        ✓"
log "  • Git-ignored files:       ✓"
log ""
log "To use this environment:"
log "  export VECTORA_TEST_ENV=.env.test"
log ""
log "Logs saved to: $LOG_FILE"
log ""

exit 0
