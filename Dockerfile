# Stage 1: Build stage
FROM python:3.13-slim as builder

ARG VERSION=0.1.0

WORKDIR /build

# Instalar build dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    curl \
    && rm -rf /var/lib/apt/lists/*

# Instalar UV
RUN curl -LsSf https://astral.sh/uv/install.sh | sh
ENV PATH="/root/.cargo/bin:$PATH"

# Copiar projeto
COPY . .

# Build com UV
RUN uv sync --frozen --no-dev

# Stage 2: Runtime stage
FROM python:3.13-slim

ARG VERSION=0.1.0

WORKDIR /app

# Instalar runtime dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    curl \
    && rm -rf /var/lib/apt/lists/*

# Copiar venv do builder
COPY --from=builder /build/.venv /app/.venv

# Copiar código do builder
COPY --from=builder /build/src /app/src
COPY --from=builder /build/pyproject.toml /app/pyproject.toml
COPY --from=builder /build/README.md /app/README.md

# Criar diretório de dados
RUN mkdir -p /root/.vectora && chmod 777 /root/.vectora

# Setup PATH
ENV PATH="/app/.venv/bin:$PATH" \
    PYTHONUNBUFFERED=1 \
    LOG_LEVEL=INFO \
    VERSION=${VERSION}

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=40s --retries=3 \
    CMD python -c "import sys; sys.exit(0)" || exit 1

# Default: rodar como MCP server
ENTRYPOINT ["/app/.venv/bin/python", "-m"]
CMD ["mcp.server", "src/mcp_server.py"]
