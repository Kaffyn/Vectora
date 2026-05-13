#!/bin/bash

echo "═══════════════════════════════════════════════════════════"
echo "  Vectora Setup Verification"
echo "═══════════════════════════════════════════════════════════"
echo ""

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

check() {
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓${NC} $1"
        return 0
    else
        echo -e "${RED}✗${NC} $1"
        return 1
    fi
}

warn() {
    echo -e "${YELLOW}⚠${NC} $1"
}

# Check Python
echo "Verificando Python..."
python --version > /dev/null 2>&1
check "Python instalado"

# Check uv
echo ""
echo "Verificando uv..."
uv --version > /dev/null 2>&1
check "uv instalado"

# Check .env
echo ""
echo "Verificando configuracao..."
if [ -f ".env" ]; then
    echo -e "${GREEN}✓${NC} Arquivo .env encontrado"
    if grep -q "GOOGLE_API_KEY" .env; then
        check "GOOGLE_API_KEY configurado"
    else
        warn "GOOGLE_API_KEY nao configurado (necessario para testes)"
    fi
else
    warn ".env nao encontrado - copie de .env.example"
fi

# Check dependencies
echo ""
echo "Verificando dependencias..."
uv run python -c "import src.main; print('Importacao OK')" > /dev/null 2>&1
check "Dependencias Python instaladas"

# Check Docker
echo ""
echo "Verificando Docker..."
docker --version > /dev/null 2>&1
check "Docker instalado"

docker compose version > /dev/null 2>&1
check "Docker Compose instalado"

# Summary
echo ""
echo "═══════════════════════════════════════════════════════════"
echo ""
echo "Proximos passos:"
echo "  1. Editar .env com suas credenciais (se ainda nao feito)"
echo "  2. Executar testes manuais:"
echo "     uv run python test_chat_manual.py"
echo "  3. Ou iniciar com Docker:"
echo "     docker compose up -d"
echo ""
