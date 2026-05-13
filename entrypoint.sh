#!/bin/bash
set -e

# Colors para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Verificar variáveis obrigatórias
check_environment() {
    log_info "Verificando configurações de ambiente..."

    if [ -z "$DB_DSN" ]; then
        log_warn "DB_DSN não configurado, usando SQLite como fallback"
        export DB_DSN="sqlite:///./vectora.db"
    fi

    if [ -z "$QDRANT_URL" ]; then
        log_warn "QDRANT_URL não configurado, usando LanceDB"
        export VECTOR_STORE_TYPE="lancedb"
        export LANCEDB_DIR="./data/lancedb"
    fi

    log_success "Configurações de ambiente verificadas"
}

# Aguardar dependências
wait_for_dependencies() {
    log_info "Aguardando dependências iniciarem..."

    # PostgreSQL
    if [[ "$DB_DSN" == postgresql* ]]; then
        local postgres_host=$(echo $DB_DSN | sed -n 's/.*@\([^:]*\).*/\1/p')
        local postgres_port=$(echo $DB_DSN | sed -n 's/.*:\([0-9]*\)\/.*/\1/p')

        log_info "Aguardando PostgreSQL em $postgres_host:$postgres_port..."
        for i in {1..30}; do
            if python -c "import psycopg2; psycopg2.connect('$DB_DSN')" 2>/dev/null; then
                log_success "PostgreSQL conectado"
                break
            fi
            if [ $i -eq 30 ]; then
                log_error "PostgreSQL não respondeu após 30 segundos"
                exit 1
            fi
            sleep 1
        done
    fi

    # Qdrant
    if [ ! -z "$QDRANT_URL" ] && [ "$QDRANT_URL" != "false" ]; then
        log_info "Aguardando Qdrant em $QDRANT_URL..."
        for i in {1..30}; do
            if curl -f "$QDRANT_URL/health" >/dev/null 2>&1; then
                log_success "Qdrant conectado"
                break
            fi
            if [ $i -eq 30 ]; then
                log_error "Qdrant não respondeu após 30 segundos"
                exit 1
            fi
            sleep 1
        done
    fi

    # Valkey/Redis
    if [ ! -z "$VALKEY_HOST" ]; then
        log_info "Aguardando Valkey em $VALKEY_HOST:$VALKEY_PORT..."
        for i in {1..30}; do
            if redis-cli -h $VALKEY_HOST -p $VALKEY_PORT ping >/dev/null 2>&1; then
                log_success "Valkey conectado"
                break
            fi
            if [ $i -eq 30 ]; then
                log_warn "Valkey não respondeu, continuando sem cache"
                break
            fi
            sleep 1
        done
    fi
}

# Executar migrações do banco de dados
run_migrations() {
    log_info "Executando migrações do banco de dados..."

    if [[ "$DB_DSN" == postgresql* ]]; then
        # PostgreSQL migrations (se usar Alembic)
        if [ -d "alembic" ]; then
            log_info "Aplicando migrações Alembic..."
            alembic upgrade head || log_warn "Nenhuma migração para aplicar"
        fi
    fi

    # Inicializar banco de dados se necessário
    if [ "$1" == "--init-db" ] || [ "$2" == "--init-db" ]; then
        log_info "Inicializando banco de dados..."
        python -m src.main --init-db
        log_success "Banco de dados inicializado"
    fi
}

# Validar instalação de dependências
validate_packages() {
    log_info "Validando pacotes Python..."

    local required_packages=("fastapi" "langgraph" "psycopg2" "qdrant_client")

    for package in "${required_packages[@]}"; do
        if ! python -c "import ${package//-/_}" 2>/dev/null; then
            log_warn "Pacote $package não encontrado, mas continuando..."
        fi
    done

    log_success "Validação de pacotes concluída"
}

# Health check endpoint
create_health_endpoint() {
    log_info "Configurando health check endpoint..."
    # Health check é gerenciado pela aplicação FastAPI
}

# Criar diretórios necessários
setup_directories() {
    log_info "Configurando diretórios..."

    mkdir -p ./data
    mkdir -p ./logs
    mkdir -p ./data/lancedb
    mkdir -p ./data/embeddings

    # Set proper permissions
    chmod -R 755 ./data
    chmod -R 755 ./logs

    log_success "Diretórios criados"
}

# Main execution
main() {
    echo ""
    echo "╔══════════════════════════════════════════════════════════════╗"
    echo "║         Vectora - Agent Orchestration Platform              ║"
    echo "║                  Docker Entrypoint v1.0                     ║"
    echo "╚══════════════════════════════════════════════════════════════╝"
    echo ""

    check_environment
    setup_directories
    wait_for_dependencies
    validate_packages
    run_migrations "$@"
    create_health_endpoint

    log_success "Todas as verificações concluídas!"
    echo ""
    log_info "Iniciando aplicação Vectora..."
    echo ""

    # Execute passed command or default to API server
    if [ $# -eq 0 ]; then
        exec uvicorn src.main:app --host 0.0.0.0 --port 8000 --log-level ${LOG_LEVEL:-info}
    else
        # Handle --init-db flag
        if [ "$1" == "--init-db" ]; then
            python -m src.main --init-db
            exit 0
        else
            exec "$@"
        fi
    fi
}

# Run main function
main "$@"
