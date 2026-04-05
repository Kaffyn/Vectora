#!/bin/bash

# Vectora Build Script - Compiles all binaries to bin/ directory
# Always overwrites existing exe files with same names

set -e

BIN_DIR="bin"
mkdir -p "$BIN_DIR"

echo "╔════════════════════════════════════════════════════════════╗"
echo "║         Vectora Build (Consolidated)                      ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo ""

# Daemon
echo "📦 Building vectora (daemon)..."
go build -o "$BIN_DIR/vectora.exe" ./cmd/vectora 2>/dev/null
if [ -f "$BIN_DIR/vectora.exe" ]; then
	size=$(du -h "$BIN_DIR/vectora.exe" | cut -f1)
	echo "  ✓ vectora.exe ($size)"
else
	echo "  ✗ Failed"
	exit 1
fi

# CLI
echo "📦 Building vectora-cli..."
go build -o "$BIN_DIR/vectora-cli.exe" ./cmd/vectora-cli 2>/dev/null
if [ -f "$BIN_DIR/vectora-cli.exe" ]; then
	size=$(du -h "$BIN_DIR/vectora-cli.exe" | cut -f1)
	echo "  ✓ vectora-cli.exe ($size)"
else
	echo "  ✗ Failed"
	exit 1
fi

# Installer
echo "📦 Building vectora-installer..."
go build -o "$BIN_DIR/vectora-installer.exe" ./cmd/vectora-installer 2>/dev/null
if [ -f "$BIN_DIR/vectora-installer.exe" ]; then
	size=$(du -h "$BIN_DIR/vectora-installer.exe" | cut -f1)
	echo "  ✓ vectora-installer.exe ($size)"
else
	echo "  ✗ Failed"
	exit 1
fi

# MPM
echo "📦 Building mpm..."
go build -o "$BIN_DIR/mpm.exe" ./cmd/mpm 2>/dev/null
if [ -f "$BIN_DIR/mpm.exe" ]; then
	size=$(du -h "$BIN_DIR/mpm.exe" | cut -f1)
	echo "  ✓ mpm.exe ($size)"
else
	echo "  ✗ Failed"
	exit 1
fi

# LPM
echo "📦 Building lpm..."
go build -o "$BIN_DIR/lpm.exe" ./cmd/lpm 2>/dev/null
if [ -f "$BIN_DIR/lpm.exe" ]; then
	size=$(du -h "$BIN_DIR/lpm.exe" | cut -f1)
	echo "  ✓ lpm.exe ($size)"
else
	echo "  ✗ Failed"
	exit 1
fi

echo ""
echo "╔════════════════════════════════════════════════════════════╗"
echo "║              ✅ Build Successful!                         ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo ""

total=$(du -sh "$BIN_DIR" | cut -f1)
count=$(ls -1 "$BIN_DIR"/*.exe 2>/dev/null | wc -l)
echo "📊 Total: $total ($count executables)"
echo "📁 Location: ./$BIN_DIR/"
echo ""
