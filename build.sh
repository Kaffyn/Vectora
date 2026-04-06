#!/bin/bash

# Vectora Build Script - Unified cross-platform compilation
# REGRAS RÍGIDAS:
# 1. Binários APENAS em bin/ com convenção: {app}-{os}-{arch}{suffix}
# 2. Executável APENAS via build.sh
# 3. CGO=1 apenas para Fyne (setup, desktop)
# 4. Limpar bin/ antes de compilar

set -e

# Cores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

BIN_DIR="bin"

# Limpar bin/ completamente
if [ -d "$BIN_DIR" ]; then
	rm -rf "$BIN_DIR"
	echo -e "${YELLOW}🗑️  Limpando bin/...${NC}"
fi
mkdir -p "$BIN_DIR"

# Detectar OS
OS=$(uname -s 2>/dev/null || echo "Windows")
ARCH=$(uname -m 2>/dev/null || echo "x86_64")

# Map ARCH para padrão Go
case "$ARCH" in
	x86_64) GOARCH="amd64" ;;
	aarch64) GOARCH="arm64" ;;
	armv7l) GOARCH="armv7" ;;
	*) GOARCH="$ARCH" ;;
esac

echo ""
echo "╔════════════════════════════════════════════════════════════╗"
echo "║       Vectora Build - Compilação Unificada Universal      ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo ""
echo "🖥️  Sistema: $OS ($ARCH → $GOARCH)"
echo ""

# Function to build a single binary
build_binary() {
	local cmd=$1        # vectora, lpm, mpm, etc
	local path=$2       # cmd/daemon, cmd/lpm, etc
	local goos=$3       # windows, linux, darwin
	local goarch=$4     # amd64, arm64, etc
	local suffix=$5     # .exe para Windows, "" para Unix

	export GOOS=$goos
	export GOARCH=$goarch

	# CGO_ENABLED=1 APENAS para Fyne apps (setup, desktop)
	if [ "$cmd" = "vectora-setup" ] || [ "$cmd" = "vectora-desktop" ]; then
		export CGO_ENABLED=1
	else
		export CGO_ENABLED=0
	fi

	# Nomear com convenção: {app}-{os}-{arch}{suffix}
	local output="$BIN_DIR/${cmd}-${goos}-${goarch}${suffix}"

	printf "  %-40s" "Building ${cmd}-${goos}-${goarch}..."

	if go build -v -ldflags="-s -w" -o "$output" "./$path" 2>/dev/null; then
		if [ -f "$output" ]; then
			local size
			if command -v du &>/dev/null; then
				size=$(du -h "$output" 2>/dev/null | cut -f1)
			else
				size=$(ls -lh "$output" 2>/dev/null | awk '{print $5}')
			fi
			printf "${GREEN}✓${NC} ($size)\n"
			return 0
		fi
	fi
	printf "${RED}✗${NC}\n"
	return 1
}

# ==============================================================================
# COMPILAR PARA O SISTEMA DETECTADO (nativa apenas)
# ==============================================================================

case "$OS" in
	Darwin) # macOS
		echo "📦 Compilando para macOS (nativo)..."
		echo ""

		# amd64
		build_binary "vectora" "cmd/daemon" "darwin" "amd64" ""
		build_binary "vectora-tui" "cmd/tui" "darwin" "amd64" ""
		build_binary "vectora-setup" "cmd/setup" "darwin" "amd64" ""
		build_binary "vectora-desktop" "cmd/desktop" "darwin" "amd64" ""
		build_binary "mpm" "cmd/mpm" "darwin" "amd64" ""
		build_binary "lpm" "cmd/lpm" "darwin" "amd64" ""

		# arm64
		build_binary "vectora" "cmd/daemon" "darwin" "arm64" ""
		build_binary "vectora-tui" "cmd/tui" "darwin" "arm64" ""
		build_binary "vectora-setup" "cmd/setup" "darwin" "arm64" ""
		build_binary "vectora-desktop" "cmd/desktop" "darwin" "arm64" ""
		build_binary "mpm" "cmd/mpm" "darwin" "arm64" ""
		build_binary "lpm" "cmd/lpm" "darwin" "arm64" ""
		;;

	Linux)
		echo "📦 Compilando para Linux (nativo)..."
		echo ""

		# amd64
		build_binary "vectora" "cmd/daemon" "linux" "amd64" ""
		build_binary "vectora-tui" "cmd/tui" "linux" "amd64" ""
		build_binary "vectora-setup" "cmd/setup" "linux" "amd64" ""
		build_binary "vectora-desktop" "cmd/desktop" "linux" "amd64" ""
		build_binary "mpm" "cmd/mpm" "linux" "amd64" ""
		build_binary "lpm" "cmd/lpm" "linux" "amd64" ""

		# arm64
		build_binary "vectora" "cmd/daemon" "linux" "arm64" ""
		build_binary "vectora-tui" "cmd/tui" "linux" "arm64" ""
		build_binary "vectora-setup" "cmd/setup" "linux" "arm64" ""
		build_binary "vectora-desktop" "cmd/desktop" "linux" "arm64" ""
		build_binary "mpm" "cmd/mpm" "linux" "arm64" ""
		build_binary "lpm" "cmd/lpm" "linux" "arm64" ""
		;;

	MINGW* | MSYS* | CYGWIN* | Windows_NT | *)
		echo "📦 Compilando para Windows (nativo)..."
		echo ""

		# amd64 (x86_64) - todos os binários
		build_binary "vectora" "cmd/daemon" "windows" "amd64" ".exe"
		build_binary "vectora-tui" "cmd/tui" "windows" "amd64" ".exe"
		build_binary "vectora-setup" "cmd/setup" "windows" "amd64" ".exe"
		build_binary "vectora-desktop" "cmd/desktop" "windows" "amd64" ".exe"
		build_binary "mpm" "cmd/mpm" "windows" "amd64" ".exe"
		build_binary "lpm" "cmd/lpm" "windows" "amd64" ".exe"

		# arm64 - apenas CLI tools (sem CGO/Fyne)
		# Nota: Setup e Desktop (Fyne) com CGO_ENABLED=1 não compilam bem para ARM64 no Windows
		build_binary "vectora" "cmd/daemon" "windows" "arm64" ".exe"
		build_binary "vectora-tui" "cmd/tui" "windows" "arm64" ".exe"
		build_binary "mpm" "cmd/mpm" "windows" "arm64" ".exe"
		build_binary "lpm" "cmd/lpm" "windows" "arm64" ".exe"
		;;
esac

# Limpar variáveis de environment
unset GOOS
unset GOARCH
unset CGO_ENABLED

echo ""
echo "╔════════════════════════════════════════════════════════════╗"
echo "║            ✅ Build Concluído com Sucesso!                ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo ""

# Sumário final
if [ -d "$BIN_DIR" ]; then
	count=$(find "$BIN_DIR" -type f 2>/dev/null | wc -l)
	total=$(du -sh "$BIN_DIR" 2>/dev/null | cut -f1)
	echo "📊 Resumo:"
	echo "  ✓ Binários compilados: $count"
	echo "  ✓ Tamanho total: $total"
	echo "  ✓ Local: ./$BIN_DIR/"
	echo ""
	echo "📋 Binários gerados:"
	ls -1 "$BIN_DIR" | sed 's/^/  - /'
	echo ""
else
	echo -e "${RED}⚠️  Nenhum binário encontrado em $BIN_DIR${NC}"
fi
