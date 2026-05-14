FROM python:3.12-slim-bookworm

# Instala uv para gerenciamento de dependências ultra-rápido
COPY --from=ghcr.io/astral-sh/uv:latest /uv /uvx /bin/

# Configura ambiente
ENV PYTHONUNBUFFERED=1 \
    PYTHONDONTWRITEBYTECODE=1 \
    UV_COMPILE_BYTECODE=1 \
    UV_LINK_MODE=copy

WORKDIR /app

# Instala dependências de sistema para LanceDB/PyArrow
RUN apt-get update && apt-get install -y --no-install-recommends \
    build-essential \
    curl \
    && rm -rf /var/lib/apt/lists/*

# Copia arquivos de configuração de dependências
COPY pyproject.toml uv.lock ./

# Instala dependências do projeto (usa cache do docker para camadas)
RUN uv sync --frozen --no-install-project --no-dev

# Copia código fonte
COPY src/ ./src/
COPY .env.example ./.env

# Adiciona diretório src ao PYTHONPATH
ENV PYTHONPATH="/app/src"

# Cria diretórios de dados
RUN mkdir -p /app/data /app/logs

# Expõe porta do MCP/API
EXPOSE 8000

# Comando padrão
CMD ["uv", "run", "python", "src/main.py"]
