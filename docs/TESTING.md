# Testing Guide - Vectora

Guia completo para testar o Vectora, incluindo testes manuais interativos e testes automáticos.

## 🎯 Tipos de Teste

### 1. Testes Manuais Interativos (Recomendado para Dev)

**Objetivo:** Conversar diretamente com o Vectora e verificar se as respostas são corretas.

#### Setup Inicial

```bash
# Verificar que você está no diretório do projeto
cd /path/to/vectora

# Copiar arquivo de ambiente (se não existir)
cp .env.example .env

# Editar .env com suas credenciais (crítico: GOOGLE_API_KEY ou outro LLM provider)
nano .env

# Instalar dependências
uv sync

# Verificar que tudo está funcionando
uv run python -c "from src.main import create_graph; print('OK')"
```

#### Executar Chat Interativo

```bash
# Teste manual interativo
uv run python test_chat_manual.py
```

Você verá:

```
╔════════════════════════════════════════════════════════╗
║          VECTORA - Manual Chat Testing                 ║
║                                                        ║
║  Escreva suas mensagens e veja as respostas do bot.   ║
║  Digite 'sair' ou 'quit' para encerrar.               ║
║  Digite 'help' para ver comandos especiais.           ║
╚════════════════════════════════════════════════════════╝

Comandos especiais:
  help       - Mostrar este menu
  clear      - Limpar historico
  sair/quit  - Encerrar
  debug      - Ver info de debug
```

#### Exemplos de Teste Manual

**Teste 1: Saudação básica**

```
Sua mensagem: Ola, como voce funciona?

Vectora: Ola! Eu sou Vectora, um agente de IA...
Turno 1 concluido!
```

**Teste 2: Pergunta com múltiplas ferramentas**

```
Sua mensagem: Busque informações sobre Python e resuma para mim

Vectora: (usa web_search para buscar)
         (processa e retorna resumo)
Turno 2 concluido!
```

**Teste 3: Contexto de conversa**

```
Sua mensagem: Qual era o assunto anterior?

Vectora: (mantém contexto da conversa anterior)
Turno 3 concluido!
```

### 2. Testes Automáticos (Unit Tests)

**⚠️ Status Atual:** Há conflitos entre testes síncronos/async que precisam ser resolvidos.

#### Executar Testes MCP (que estão passando)

```bash
# Apenas testes MCP (10/10 passando)
uv run pytest tests/test_mcp_integration.py -v

# Resultado esperado:
# test_mcp_client_disabled PASSED
# test_mcp_client_http_transport PASSED
# test_get_mcp_tools_success PASSED
# ... (10 tests)
# ===== 10 passed in 0.23s =====
```

#### Problemas Conhecidos nos Testes

**Erro:** `'test_X' requested an async fixture 'temp_db', with no plugin or hook that handled it`

**Causa:** Testes síncronos usando fixtures async

**Solução:** Precisa refatorar conftest.py para:

- Testes async: `async def test_...`
- Fixtures async: `@pytest_asyncio.fixture` + `async def`
- Testes síncronos: usar fixtures síncronas

**Próximo Passo:** Corrigir em uma sessão dedicada a testes.

### 3. Testes via Docker Compose

**Objetivo:** Testar em ambiente isolado com todos os serviços.

#### Setup

```bash
# Build da imagem
docker compose build

# Iniciar serviços
docker compose up -d

# Aguardar 30-60 segundos
docker compose logs -f vectora
```

#### Testar Health Checks

```bash
# API health
curl http://localhost:8000/health

# PostgreSQL
docker compose exec postgres pg_isready -U vectora

# Qdrant
curl http://localhost:6333/health

# Valkey
docker compose exec valkey redis-cli ping
```

#### Testar via API HTTP

```bash
# Exemplo de chat via HTTP
curl -X POST http://localhost:8000/chat \
  -H "Content-Type: application/json" \
  -d '{
    "message": "Ola, como voce funciona?",
    "user_id": "test_user"
  }'
```

### 4. Testes de Cenários (QA)

**Objetivo:** Validar workflows específicos.

#### Teste de RAG (Retrieval-Augmented Generation)

```
Sua mensagem: Indexe este documento e depois busque informações nele

Esperado:
1. Documento indexado no Qdrant
2. Busca retorna resultados relevantes
3. LLM usa contexto do documento
```

#### Teste de Web Search

```
Sua mensagem: Busque as últimas noticias sobre Python

Esperado:
1. Web search tool é chamado
2. Retorna resultados relevantes
3. LLM processa e resume
```

#### Teste de MCP

```
Sua mensagem: Use a ferramenta MCP para [ação]

Esperado:
1. MCP client se conecta ao servidor
2. Tool é executada
3. Resultado é integrado à resposta
```

## 📋 Checklist de Testes Manuais

Use este checklist quando testar a aplicação:

### Básico

- [ ] Bot responde a saudações
- [ ] Bot responde a perguntas simples
- [ ] Bot mantém contexto em turnos múltiplos
- [ ] Mensagens de erro são claras

### Ferramentas

- [ ] Web search funciona (se habilitado)
- [ ] Fetch URL retorna conteúdo
- [ ] Database queries funcionam (se habilitado)
- [ ] MCP tools executam (se habilitado)

### RAG

- [ ] Embedding funciona
- [ ] Vector search retorna resultados
- [ ] Reranking ordena corretamente
- [ ] LLM usa contexto do RAG

### Performa

- [ ] Primeira resposta < 3 segundos
- [ ] Respostas subsequentes < 2 segundos
- [ ] Cache funciona (segunda chamada é mais rápida)
- [ ] Memory usage não cresce indefinidamente

### Edge Cases

- [ ] Mensagens muito longas são truncadas
- [ ] Unicode é suportado (português, emojis)
- [ ] Conversação é reiniciada corretamente
- [ ] Erro de API é tratado gracefully

## 🐛 Troubleshooting

### "ModuleNotFoundError: No module named 'src.main'"

```bash
# Solução: Você precisa estar no diretório raiz do projeto
cd /path/to/vectora
uv run python test_chat_manual.py
```

### "GOOGLE_API_KEY not set" ou erro de LLM provider

```bash
# Solução: Configurar arquivo .env
cp .env.example .env
nano .env  # Editar com suas chaves de API

# Verificar
echo $GOOGLE_API_KEY  # Deve mostrar sua chave
```

### "ConnectionError: Cannot connect to PostgreSQL"

```bash
# Se usando Docker:
docker compose logs postgres

# Se local:
# Certifique-se que PostgreSQL está rodando
# Ou use SQLite: DB_DSN="sqlite:///./vectora.db"
```

### Bot não responde (trava)

```bash
# 1. Verificar logs
docker compose logs vectora

# 2. Verificar configuração do LLM
echo $LLM_PROVIDER
echo $GOOGLE_MODEL

# 3. Reiniciar
docker compose restart vectora
```

## 📊 Monitoramento Durante Testes

### Ver Logs em Real-time

```bash
# Local
uv run python test_chat_manual.py 2>&1 | tee test.log

# Docker
docker compose logs -f vectora
```

### Ver Métricas

```bash
# Uso de memória
docker stats

# Conexões do banco
docker compose exec postgres "psql -U vectora vectora -c 'SELECT * FROM pg_stat_activity;'"

# Cache hits
docker compose exec valkey redis-cli INFO stats
```

## ✅ Padrão de Teste Bem-Sucedido

Um teste bem-sucedido segue este padrão:

1. **Input clara:** Mensagem específica do usuário
2. **Processamento:** Bot processa e chama ferramentas se necessário
3. **Output relevante:** Resposta direta e útil
4. **Contexto:** Mantém estado para próximos turnos
5. **Performance:** Responde em tempo aceitável

Exemplo:

```
Usuario: "Qual e a populacao do Brasil?"

Bot:
1. Chama web_search("populacao brasil")
2. Recebe resultado: "População do Brasil: ~215 milhões"
3. Processa: "A população do Brasil é de aproximadamente 215 milhões de habitantes, sendo o 7º país mais populoso do mundo."
4. Usuario pode continuar: "E de Portugal?"
```

## 🎓 Próximos Passos

1. **Fixar Testes:** Resolver conflitos async em conftest.py
2. **Adicionar Fixtures:** Para cada serviço (DB, Qdrant, Cache)
3. **Coverage:** Atingir 80%+ de coverage nos testes
4. **CI/CD:** GitHub Actions executa testes automaticamente

## 📚 Referências

- [pytest Docs](https://docs.pytest.org)
- [pytest-asyncio](https://pytest-asyncio.readthedocs.io)
- [LangGraph Testing](https://langchain-ai.github.io/langgraph/how-tos/human-in-the-loop)
- [FastAPI Testing](https://fastapi.tiangolo.com/advanced/testing-events/)
