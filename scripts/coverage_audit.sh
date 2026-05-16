#!/bin/bash
# Coverage Audit Script - Identifica gaps para 100% de cobertura

set -e

echo "🔍 AUDITORIA DE COBERTURA DE TESTES"
echo "=================================="
echo ""

# Cores para output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Verificar se pytest está instalado
if ! command -v pytest &> /dev/null; then
    echo -e "${RED}❌ pytest não está instalado${NC}"
    echo "Execute: uv pip install pytest pytest-asyncio pytest-cov"
    exit 1
fi

# Diretórios
vectora_DIR="vectora"
TESTS_DIR="tests"
COVERAGE_DIR="htmlcov"
COVERAGE_REPORT="coverage.txt"

echo "📊 Executando testes com coverage..."
echo ""

# Executar testes com coverage
python -m pytest \
    --cov=$vectora_DIR \
    --cov-report=html \
    --cov-report=term-missing:skip-covered \
    --cov-report=json \
    --tb=short \
    $TESTS_DIR \
    | tee $COVERAGE_REPORT

echo ""
echo "=================================="
echo "📈 RESULTADOS DA AUDITORIA"
echo "=================================="
echo ""

# Extrair resumo de cobertura
if [ -f ".coverage" ]; then
    python -m coverage report --skip-empty

    echo ""
    echo "🔗 Relatório HTML gerado em: $COVERAGE_DIR/index.html"
    echo ""

    # Verificar se atingiu 100%
    COVERAGE=$(python -m coverage report | grep TOTAL | awk '{print $NF}' | sed 's/%//')

    if [ "${COVERAGE%.*}" -eq 100 ] 2>/dev/null; then
        echo -e "${GREEN}✅ COBERTURA 100% ATINGIDA!${NC}"
    else
        echo -e "${YELLOW}⚠️  Cobertura atual: ${COVERAGE}%${NC}"
        echo ""
        echo "📋 Linhas não cobertas por arquivo:"
        echo "===================================="
        python -m coverage report --skip-empty | grep -v "^-" | grep -v "TOTAL" | grep -v "^$"

        echo ""
        echo "💡 Próximos passos:"
        echo "1. Abra: $COVERAGE_DIR/index.html para ver detalhes"
        echo "2. Identifique branches não testadas (red lines)"
        echo "3. Adicione testes para cobrir esses cenários"
        echo "4. Execute novamente este script para validar"
    fi
else
    echo -e "${RED}❌ Arquivo .coverage não encontrado${NC}"
    exit 1
fi

echo ""
echo "✅ Auditoria concluída"
