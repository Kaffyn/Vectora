# Guia de Testing - Vectora MVP v0.1

Este documento descreve como rodar, escrever e entender testes no projeto Vectora.

## Estrutura de Testes

```
tests/
├── unit/           # Unit tests (isolados)
├── integration/    # Integration tests (múltiplos componentes)
└── e2e/           # End-to-end tests (fluxo completo)
```

## Rodar Testes

### Unit Tests

```bash
pytest tests/unit/ -v
```

### Integration Tests

```bash
pytest tests/integration/ -v
```

### E2E Tests

```bash
pytest tests/e2e/ -v
```

### Todos os Testes com Coverage

```bash
pytest tests/ --cov=src --cov-report=html --cov-report=term
```

## Coverage Target

Objetivo: > 80% coverage em src/

```bash
pytest tests/ --cov=src --cov-report=term-missing
```

## LangSmith Integration

Para ativar observabilidade via LangSmith:

```bash
export LANGSMITH_API_KEY=your-key-here
export LANGSMITH_PROJECT=vectora-dev
pytest tests/ -v
```

Acesse: https://smith.langchain.com

## CI/CD Pipeline

GitHub Actions roda automaticamente:

- Lint (ruff, isort)
- Build verification
- Test suites (unit, integration, E2E)
- Security scanning
- Docker build
- VPS deployment (com tag [deploy])

## Configuração para Testes

Variáveis padrão:

- `LLM_PROVIDER=google-genai`
- `GOOGLE_API_KEY=test-key`
- `VOYAGE_API_KEY=test-key`

## Checklist para PR

- [ ] `pytest tests/ -v` passa
- [ ] Coverage > 80%
- [ ] `ruff check .` limpo
- [ ] CI/CD verde
