#!/bin/bash

# Cross-compilation script for Vectora
# Note: Some targets may require specific toolchains (gcc for Linux/macOS)

BIN_DIR="bin"
mkdir -p "$BIN_DIR"

echo ""
echo "╔════════════════════════════════════════════════════════════╗"
echo "║       Vectora Cross-Compilation (Go Native)               ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo ""

# Binaries that work everywhere (pure Go, no cgo)
declare -a CLI_BINARIES=("vectora-cli:cmd/vectora-cli" "mpm:cmd/mpm")

# Binaries with cgo (need specific toolchains for cross-compilation)
declare -a CGO_BINARIES=("vectora:cmd/vectora" "lpm:cmd/lpm")

# Targets
declare -a TARGETS=("windows-amd64:windows:amd64" "windows-arm:windows:arm" "linux-amd64:linux:amd64" "linux-arm:linux:arm" "macos-arm64:darwin:arm64")

COUNT=0
FAILED=0

for target in "${TARGETS[@]}"; do
    IFS=':' read -r target_name target_os target_arch <<< "$target"
    
    echo "📦 Target: $target_name"
    
    # Pure Go binaries (always compile)
    for binary in "${CLI_BINARIES[@]}"; do
        IFS=':' read -r bin_name bin_path <<< "$binary"

        ext=""
        if [ "$target_os" == "windows" ]; then
            ext=".exe"
        else
            ext=".app"
        fi

        output="$BIN_DIR/${bin_name}-${target_name}${ext}"
        
        echo -n "  Building $bin_name... "
        
        export GOOS="$target_os"
        export GOARCH="$target_arch"
        export CGO_ENABLED=0
        
        if go build -ldflags "-w -s" -o "$output" "./$bin_path" 2>/dev/null; then
            size=$(du -h "$output" | cut -f1)
            echo "✓ ($size)"
            ((COUNT++))
        else
            echo "✗"
            ((FAILED++))
        fi
    done
    
    # CGO binaries (only on native platform or with toolchains)
    for binary in "${CGO_BINARIES[@]}"; do
        IFS=':' read -r bin_name bin_path <<< "$binary"
        
        # Skip CGO binaries on non-windows unless CGO_ENABLED is explicitly set
        if [ "$target_os" != "windows" ] && [ -z "$FORCE_CGO" ]; then
            echo "  Building $bin_name... ⊘ (requires toolchain)"
            ((FAILED++))
        else
            ext=""
            if [ "$target_os" == "windows" ]; then
                ext=".exe"
            else
                ext=".app"
            fi

            output="$BIN_DIR/${bin_name}-${target_name}${ext}"
            
            echo -n "  Building $bin_name... "
            
            export GOOS="$target_os"
            export GOARCH="$target_arch"
            
            if go build -ldflags "-w -s" -o "$output" "./$bin_path" 2>/dev/null; then
                size=$(du -h "$output" | cut -f1)
                echo "✓ ($size)"
                ((COUNT++))
            else
                echo "✗"
                ((FAILED++))
            fi
        fi
    done
done

# Build installer (Windows amd64 only)
echo ""
echo "📦 Target: windows-installer"
echo -n "  Building vectora-installer... "

export GOOS="windows"
export GOARCH="amd64"
export CGO_ENABLED=1

if go build -o "$BIN_DIR/vectora-installer-windows-amd64.exe" "./cmd/vectora-installer" 2>/dev/null; then
    size=$(du -h "$BIN_DIR/vectora-installer-windows-amd64.exe" | cut -f1)
    echo "✓ ($size)"
    ((COUNT++))
else
    echo "✗"
    ((FAILED++))
fi

# Clear env
unset GOOS GOARCH CGO_ENABLED

# Summary
echo ""
echo "╔════════════════════════════════════════════════════════════╗"
echo "║              Build Resultado                              ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo ""

FILE_COUNT=$(ls -1 "$BIN_DIR" 2>/dev/null | wc -l)
TOTAL_SIZE=$(du -sh "$BIN_DIR" 2>/dev/null | cut -f1)

echo "✅ Binarios compilados: $COUNT"
echo "⊘  Skipped/Failed: $FAILED"
echo "📊 Tamanho total: $TOTAL_SIZE"
echo "📁 Localizacao: ./$BIN_DIR/ ($FILE_COUNT arquivos)"
echo ""
echo "💡 Dica: Para compilar para Linux/macOS, use uma máquina Linux ou macOS"
echo "         ou instale as toolchains de cross-compilation (gcc, etc)"
echo ""
