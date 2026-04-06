#!/bin/bash

# Vectora Build Script - Unified cross-platform compilation
# Automatically detects OS and compiles for native target
# Replaces: build-all.sh, build.ps1, and previous build.sh
# Works on Windows (Git Bash, MSYS2, WSL), Linux, and macOS

set -e

BIN_DIR="bin"
mkdir -p "$BIN_DIR"

# Detect OS
OS=$(uname -s 2>/dev/null || echo "Windows")
ARCH=$(uname -m 2>/dev/null || echo "x86_64")

echo ""
echo "╔════════════════════════════════════════════════════════════╗"
echo "║         Vectora Build (Unified Universal Script)          ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo ""
echo "🖥️  Detected: $OS ($ARCH)"
echo ""

# Function to build a binary
build_binary() {
	local name=$1
	local path=$2
	local goos=$3
	local goarch=$4
	local suffix=$5

	export GOOS=$goos
	export GOARCH=$goarch
	export CGO_ENABLED=0

	local output="$BIN_DIR/${name}${suffix}"

	printf "  %-30s" "Building $name..."

	if go build -v -ldflags="-s -w" -o "$output" "./$path" 2>/dev/null; then
		if [ -f "$output" ]; then
			local size
			if command -v du &>/dev/null; then
				size=$(du -h "$output" 2>/dev/null | cut -f1)
			else
				size=$(ls -lh "$output" 2>/dev/null | awk '{print $5}')
			fi
			printf "✓ ($size)\n"
			return 0
		fi
	fi
	printf "✗\n"
	return 1
}

# Determine build targets based on detected OS
case "$OS" in
	Darwin) # macOS
		echo "📦 Building for macOS (native)..."
		build_binary "vectora" "cmd/vectora" "darwin" "amd64" ""
		build_binary "vectora" "cmd/vectora" "darwin" "arm64" "-arm64"
		build_binary "vectora-cli" "cmd/vectora-cli" "darwin" "amd64" ""
		build_binary "vectora-cli" "cmd/vectora-cli" "darwin" "arm64" "-arm64"
		build_binary "vectora-installer" "cmd/vectora-installer" "darwin" "amd64" ""
		build_binary "vectora-installer" "cmd/vectora-installer" "darwin" "arm64" "-arm64"
		build_binary "mpm" "cmd/mpm" "darwin" "amd64" ""
		build_binary "mpm" "cmd/mpm" "darwin" "arm64" "-arm64"
		build_binary "lpm" "cmd/lpm" "darwin" "amd64" ""
		build_binary "lpm" "cmd/lpm" "darwin" "arm64" "-arm64"
		;;

	Linux)
		echo "📦 Building for Linux (native)..."
		build_binary "vectora" "cmd/vectora" "linux" "amd64" ""
		build_binary "vectora" "cmd/vectora" "linux" "arm64" "-arm64"
		build_binary "vectora-cli" "cmd/vectora-cli" "linux" "amd64" ""
		build_binary "vectora-cli" "cmd/vectora-cli" "linux" "arm64" "-arm64"
		build_binary "vectora-installer" "cmd/vectora-installer" "linux" "amd64" ""
		build_binary "vectora-installer" "cmd/vectora-installer" "linux" "arm64" "-arm64"
		build_binary "mpm" "cmd/mpm" "linux" "amd64" ""
		build_binary "mpm" "cmd/mpm" "linux" "arm64" "-arm64"
		build_binary "lpm" "cmd/lpm" "linux" "amd64" ""
		build_binary "lpm" "cmd/lpm" "linux" "arm64" "-arm64"
		;;

	MINGW* | MSYS* | CYGWIN* | Windows_NT | *)
		# Windows
		echo "📦 Building for Windows (native)..."
		build_binary "vectora" "cmd/vectora" "windows" "amd64" ".exe"
		build_binary "vectora-cli" "cmd/vectora-cli" "windows" "amd64" ".exe"
		build_binary "vectora-installer" "cmd/vectora-installer" "windows" "amd64" ".exe"
		build_binary "mpm" "cmd/mpm" "windows" "amd64" ".exe"
		build_binary "lpm" "cmd/lpm" "windows" "amd64" ".exe"
		;;
esac

# Unset Go environment variables
unset GOOS
unset GOARCH
unset CGO_ENABLED

echo ""
echo "╔════════════════════════════════════════════════════════════╗"
echo "║              ✅ Build Successful!                         ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo ""

# Summary
if [ -d "$BIN_DIR" ]; then
	count=$(find "$BIN_DIR" -type f 2>/dev/null | wc -l)
	total=$(du -sh "$BIN_DIR" 2>/dev/null | cut -f1)
	echo "📊 Summary:"
	echo "  ✓ Binaries compiled: $count"
	echo "  ✓ Total size: $total"
	echo "  ✓ Location: ./$BIN_DIR/"
	echo ""
else
	echo "⚠️  No binaries found in $BIN_DIR"
fi
