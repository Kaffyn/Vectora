#!/bin/bash

# Vectora Directory Structure Setup Script
# Creates all directories and scaffold files

set -e

echo "Starting Vectora directory structure setup..."

# Create all directories using mkdir -p (can run in parallel)
mkdir -p \
  cmd/vectora \
  cmd/vectora-app \
  cmd/vectora-installer \
  internal/infra \
  internal/ipc \
  internal/db \
  internal/core \
  internal/llm \
  internal/engines \
  internal/acp \
  internal/tools \
  internal/git \
  internal/app/app \
  internal/app/components \
  internal/app/hooks \
  internal/app/store \
  internal/app/services \
  internal/app/utils \
  internal/app/styles \
  internal/app/public \
  pkg/types \
  pkg/utils \
  tests/integration \
  tests/e2e \
  tests/fixtures \
  tests/mocks \
  build \
  docs \
  scripts

echo "Directories created successfully"

# Create root level files
touch .env .gitignore go.mod go.sum Makefile build.ps1 package.json tsconfig.json next.config.js tailwind.config.js

# Create cmd/vectora files
touch cmd/vectora/main.go cmd/vectora/app.go cmd/vectora/cli.go cmd/vectora/flags.go

# Create cmd/vectora-app files
touch cmd/vectora-app/main.go cmd/vectora-app/frontend.go

# Create cmd/vectora-installer files
touch cmd/vectora-installer/main.go cmd/vectora-installer/wizard.go

# Create internal/infra files
touch internal/infra/config.go internal/infra/logger.go internal/infra/env.go

# Create internal/ipc files
touch internal/ipc/server.go internal/ipc/message.go internal/ipc/handlers.go

# Create internal/db files
touch internal/db/db.go internal/db/migrations.go internal/db/query.go

# Create internal/core files
touch internal/core/engine.go internal/core/workspace.go internal/core/registry.go

# Create internal/llm files
touch internal/llm/provider.go internal/llm/gemini.go internal/llm/qwen.go

# Create internal/engines files
touch internal/engines/rag.go internal/engines/cache.go internal/engines/fallback.go

# Create internal/acp files
touch internal/acp/tool.go internal/acp/security.go internal/acp/executor.go

# Create internal/tools files
touch internal/tools/filesystem.go internal/tools/web.go internal/tools/shell.go

# Create internal/git files
touch internal/git/bridge.go internal/git/operations.go internal/git/snapshot.go

# Create internal/app/app files
touch internal/app/app/page.tsx internal/app/app/layout.tsx internal/app/app/globals.css

# Create internal/app/components files
touch internal/app/components/Header.tsx internal/app/components/Sidebar.tsx internal/app/components/ChatPanel.tsx

# Create internal/app/hooks files
touch internal/app/hooks/useChat.ts internal/app/hooks/useWorkspace.ts

# Create internal/app/store files
touch internal/app/store/store.ts internal/app/store/chat.ts internal/app/store/workspace.ts

# Create internal/app/services files
touch internal/app/services/api.ts internal/app/services/ipc.ts

# Create internal/app/utils files
touch internal/app/utils/formatters.ts internal/app/utils/validators.ts

# Create internal/app/styles files
touch internal/app/styles/theme.css internal/app/styles/components.css

# Create internal/app/public files
touch internal/app/public/favicon.ico

# Create pkg/types files
touch pkg/types/types.go pkg/types/messages.go pkg/types/config.go

# Create pkg/utils files
touch pkg/utils/string.go pkg/utils/file.go pkg/utils/validation.go

# Create tests files
touch tests/integration/integration_test.go tests/e2e/e2e_test.go
touch tests/fixtures/fixtures.go tests/mocks/mocks.go

# Create scripts files
touch scripts/build.sh scripts/test.sh scripts/deploy.sh

# Create docs files
touch docs/README.md docs/ARCHITECTURE.md docs/API.md

echo "All scaffold files created successfully"

# Count results
DIR_COUNT=$(find . -mindepth 1 -maxdepth 3 -type d | wc -l)
FILE_COUNT=$(find . -mindepth 1 -maxdepth 4 -type f -not -path './.git/*' | wc -l)

echo ""
echo "========================================="
echo "SETUP COMPLETE"
echo "========================================="
echo "Total directories created: $DIR_COUNT"
echo "Total files created: $FILE_COUNT"
echo ""
echo "Directory structure tree:"
tree -d -L 3 2>/dev/null || find . -mindepth 1 -type d | grep -v '\.git' | sort | sed 's|^./||' | sed 's|[^/]*/|  |g'

