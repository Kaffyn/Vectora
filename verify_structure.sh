#!/bin/bash

echo "Verifying Vectora Directory Structure..."
echo ""

# Expected directories
declare -a dirs=(
  "cmd/vectora"
  "cmd/vectora-app"
  "cmd/vectora-installer"
  "internal/infra"
  "internal/ipc"
  "internal/db"
  "internal/core"
  "internal/llm"
  "internal/engines"
  "internal/acp"
  "internal/tools"
  "internal/git"
  "internal/app/app"
  "internal/app/components"
  "internal/app/hooks"
  "internal/app/store"
  "internal/app/services"
  "internal/app/utils"
  "internal/app/styles"
  "internal/app/public"
  "pkg/types"
  "pkg/utils"
  "tests/integration"
  "tests/e2e"
  "tests/fixtures"
  "tests/mocks"
  "build"
  "docs"
  "scripts"
)

# Expected root files
declare -a root_files=(
  ".env"
  ".gitignore"
  "go.mod"
  "go.sum"
  "Makefile"
  "build.ps1"
  "package.json"
  "tsconfig.json"
  "next.config.js"
  "tailwind.config.js"
)

echo "=== CHECKING DIRECTORIES ==="
dirs_created=0
for dir in "${dirs[@]}"; do
  if [ -d "$dir" ]; then
    echo "✓ $dir"
    ((dirs_created++))
  else
    echo "✗ $dir (MISSING)"
  fi
done

echo ""
echo "=== CHECKING ROOT FILES ==="
files_created=0
for file in "${root_files[@]}"; do
  if [ -f "$file" ]; then
    echo "✓ $file"
    ((files_created++))
  else
    echo "✗ $file (MISSING)"
  fi
done

echo ""
echo "=== DIRECTORY TREE (First 50 lines) ==="
tree -d -L 3 2>/dev/null | head -50 || find . -mindepth 1 -type d \( -path './cmd/*' -o -path './internal/*' -o -path './pkg/*' -o -path './tests/*' -o -path './build' -o -path './docs' -o -path './scripts' \) | grep -v '\.git\|node_modules\|\.next' | sort | sed 's|^\./||' | head -50

echo ""
echo "=== SUMMARY ==="
echo "Directories verified: $dirs_created/${#dirs[@]}"
echo "Root files verified: $files_created/${#root_files[@]}"

