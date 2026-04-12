#!/bin/bash

################################################################################
# Setup Test Environment for Vectora Local Testing
# Prepares environment variables and validates configuration
################################################################################

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
ENV_LOCAL="${SCRIPT_DIR}/.env.local"
ENV_EXAMPLE="${SCRIPT_DIR}/.env.local.example"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}Vectora Test Environment Setup${NC}\n"

# Check if .env.local exists
if [[ ! -f "$ENV_LOCAL" ]]; then
    echo -e "${YELLOW}⚠ .env.local not found${NC}"
    echo "Creating from template..."
    cp "$ENV_EXAMPLE" "$ENV_LOCAL"
    echo -e "${GREEN}✓ Created $ENV_LOCAL${NC}"
    echo -e "${YELLOW}Please edit $ENV_LOCAL and add your API keys${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Found $ENV_LOCAL${NC}"

# Load environment
set -a
source "$ENV_LOCAL"
set +a

# Validate API keys
echo ""
echo "Checking API keys..."

check_api_key() {
    local key_name="$1"
    local key_value="${!key_name}"

    if [[ -z "$key_value" ]]; then
        echo -e "${YELLOW}⚠ $key_name not set${NC}"
        return 1
    elif [[ "$key_value" == "YOUR_KEY_HERE" ]] || [[ "$key_value" == "sk-" ]] || [[ ${#key_value} -lt 10 ]]; then
        echo -e "${YELLOW}⚠ $key_name appears to be invalid (placeholder)${NC}"
        return 1
    else
        local masked="${key_value:0:7}...${key_value: -4}"
        echo -e "${GREEN}✓ $key_name set ($masked)${NC}"
        return 0
    fi
}

GEMINI_OK=false
ANTHROPIC_OK=false
OPENAI_OK=false

check_api_key "GEMINI_API_KEY" && GEMINI_OK=true || true
check_api_key "ANTHROPIC_API_KEY" && ANTHROPIC_OK=true || true
check_api_key "OPENAI_API_KEY" && OPENAI_OK=true || true

# Check Vectora binary
echo ""
echo "Checking Vectora binary..."

if [[ ! -f "${PROJECT_ROOT}/bin/vectora" ]]; then
    echo -e "${YELLOW}⚠ Vectora binary not found at ${PROJECT_ROOT}/bin/vectora${NC}"
    echo "Building Vectora..."
    cd "$PROJECT_ROOT"
    go build -o bin/vectora ./cmd/core
    echo -e "${GREEN}✓ Built Vectora${NC}"
else
    echo -e "${GREEN}✓ Found Vectora binary${NC}"
fi

# Check Go installation
echo ""
echo "Checking Go installation..."
if ! command -v go &> /dev/null; then
    echo -e "${RED}✗ Go not installed${NC}"
    exit 1
fi
GO_VERSION=$(go version | awk '{print $3}')
echo -e "${GREEN}✓ Go $GO_VERSION${NC}"

# Final status
echo ""
echo -e "${BLUE}Environment Status:${NC}"

if [[ $GEMINI_OK == true ]] || [[ $ANTHROPIC_OK == true ]] || [[ $OPENAI_OK == true ]]; then
    echo -e "${GREEN}✓ At least one API key configured${NC}"
    echo -e "${GREEN}✓ Ready to run tests${NC}"
else
    echo -e "${YELLOW}⚠ No valid API keys configured${NC}"
    echo -e "${YELLOW}Tests will run but some features will be limited${NC}"
fi

echo ""
echo -e "${BLUE}Test Configuration:${NC}"
echo "  Project Root: $PROJECT_ROOT"
echo "  Test Directory: $SCRIPT_DIR"
echo "  Environment File: $ENV_LOCAL"
echo "  Log Level: ${VECTORA_LOG_LEVEL:-info}"
echo "  Mode: ${VECTORA_MODE:-normal}"
echo "  Test Timeout: ${TEST_TIMEOUT:-30}s"

echo ""
echo -e "${GREEN}Setup complete!${NC}"
echo ""
echo "Next steps:"
echo "  1. Run interactive tests: ./test/local-test-suite.sh --interactive"
echo "  2. Run all tests: ./test/local-test-suite.sh --batch"
echo "  3. Run specific test: ./test/local-test-suite.sh --test ask"
echo "  4. Enable Gemini tests: ./test/local-test-suite.sh --batch --with-gemini"
echo ""
