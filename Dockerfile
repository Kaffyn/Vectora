# Stage 1: Builder
FROM python:3.13-slim as builder

WORKDIR /build

# Instalar uv
RUN pip install --no-cache-dir uv

# Copiar arquivos de dependências
COPY pyproject.toml uv.lock ./

# Compilar dependências (pre-compile wheels)
RUN uv pip compile pyproject.toml --output-file requirements.txt && \
    uv pip compile pyproject.toml --all-extras --output-file requirements-all.txt

# Stage 2: Runtime
FROM python:3.13-slim

WORKDIR /app

# Instalar dependências do sistema (build essentials para compilações, curl/wget para healthchecks)
RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential \
    curl \
    git \
    && rm -rf /var/lib/apt/lists/*

# Instalar uv
RUN pip install --no-cache-dir uv

# Copiar requirements do builder
COPY --from=builder /build/requirements-all.txt /tmp/

# Instalar dependências Python com uv
RUN uv pip install -r /tmp/requirements-all.txt

# Copiar código-fonte
COPY . .

# Criar usuário não-root para segurança
RUN useradd -m -u 1000 vectora && \
    chown -R vectora:vectora /app

# Mudar para usuário não-root
USER vectora

# Variáveis de ambiente padrão
ENV PYTHONUNBUFFERED=1 \
    PYTHONDONTWRITEBYTECODE=1 \
    PATH="/home/vectora/.local/bin:${PATH}" \
    LOG_LEVEL="INFO"

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=10s --retries=3 \
    CMD curl -f http://localhost:8000/health || exit 1

# Port padrão da API
EXPOSE 8000

# Entry point
COPY entrypoint.sh /app/
RUN chmod +x /app/entrypoint.sh

ENTRYPOINT ["/app/entrypoint.sh"]
CMD ["uvicorn", "src.main:app", "--host", "0.0.0.0", "--port", "8000"]
