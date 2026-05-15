# 📋 QA Testing Guide - Vectora v0.1.0-RC

**Versão:** 0.1.0-RC (Release Candidate)  
**Data:** 2026-05-14  
**Tester Target:** 5 testers independentes  

---

## 📖 Introdução

Você foi selecionado(a) como um dos **5 testers independentes** do Vectora v0.1.0 MVP. Seu feedback é **crítico** para validar a arquitetura antes do lançamento oficial em PyPI.

### Objetivo
Executar 5 cenários de teste estruturados e reportar qualquer bug, travamento ou comportamento inesperado.

### Ambiente
- **Arquivo de configuração:** `.env` (será fornecido)
- **Banco de dados:** Local (SQLite em `data/`)
- **LLM:** Pode usar Gemini, OpenAI ou Ollama (conforme sua API key)
- **Duração esperada:** 2 horas de testing

---

## ✅ Pré-requisitos

1. **Python 3.13+**
   ```bash
   python --version  # Deve ser 3.13 ou superior
   ```

2. **Clonar repositório**
   ```bash
   git clone <repo-url>
   cd vectora
   ```

3. **Instalar dependências**
   ```bash
   uv sync --group test
   ```

4. **Configurar `.env`**
   - Você receberá um arquivo `.env` com as API keys necessárias
   - Copie para a raiz do projeto

5. **Iniciar Vectora**
   ```bash
   python src/run_chat.py
   ```

---

## 🎯 5 Cenários de Teste

### **CENÁRIO 1: TUI Responsividade (10 min)**

**Objetivo:** Validar que a interface não trava enquanto processa búscas pesadas.

**Passos:**

1. Inicie o Vectora:
   ```bash
   python src/run_chat.py
   ```

2. Digite este comando:
   ```
   Pesquise sobre Next.js 16 e salve os resultados no banco de dados local
   ```

3. **Observações críticas:**
   - [ ] Interface responde em <5 segundos? (SIM / NÃO)
   - [ ] Consegue digitar enquanto processa? (SIM / NÃO)
   - [ ] Aparece status "Document enqueued for async embedding"? (SIM / NÃO)
   - [ ] Depois de 30 segundos, a busca foi indexada? (SIM / NÃO)

4. **Teste de timeout:**
   - Desconecte a internet durante a pesquisa
   - O Vectora deve recuperar gracefully após 10 segundos?
   - [ ] SIM / [ ] NÃO

**Se alguma resposta for NÃO:**
- Salve um debug dump:
  ```bash
  python -m src.debug_dump
  ```
- Copie o arquivo `vectora_debug_*.tar.gz` gerado

---

### **CENÁRIO 2: RAG e Vector Search (15 min)**

**Objetivo:** Validar que documentos indexados são recuperáveis via busca local.

**Passos:**

1. No mesmo chat, digitar:
   ```
   Quais foram os principais features do Next.js 16 que você aprendeu na pesquisa anterior?
   ```

2. **Observações críticas:**
   - [ ] Resposta foi baseada em dados indexados (não web search)? (SIM / NÃO)
   - [ ] A resposta menciona a fonte (ex: "baseado nos documentos indexados")? (SIM / NÃO)
   - [ ] Resposta foi <3 segundos (sem chamada web)? (SIM / NÃO)

3. **Comparação:**
   - Faça uma busca diferente para contraste:
     ```
     Pesquise sobre React 19 (sem indexar)
     Depois: Como React 19 se compara a Next.js?
     ```
   - Esperado: Uma busca usa web_search, a segunda usa vector_search local

**Se alguma resposta for NÃO:**
- Copie o correlation_id dos logs (busque em `logs/mcp.log`)
- Salve o debug dump

---

### **CENÁRIO 3: Error Handling (10 min)**

**Objetivo:** Validar que erros não travam a TUI e são reportados corretamente.

**Passos:**

1. Digite comando inválido:
   ```
   Pesquise sobre "xyzabc123invalid_query_that_should_return_nothing"
   ```

2. **Observações críticas:**
   - [ ] TUI não travou? (SIM / NÃO)
   - [ ] Recebeu resposta dentro de 10 segundos? (SIM / NÃO)
   - [ ] Resposta foi graceful (ex: "Nenhum resultado encontrado, sugerindo alternativas")? (SIM / NÃO)

3. **Forçar erro de rede:**
   - Desconecte internet
   - Tente fazer uma busca web
   - [ ] Recebeu mensagem de erro clara? (SIM / NÃO)
   - [ ] Consegue reconectar após isso? (SIM / NÃO)

**Se alguma resposta for NÃO:**
- Descrição do erro:
  ```
  [Descreva o comportamento inesperado aqui]
  ```
- Salve debug dump

---

### **CENÁRIO 4: Multi-turn Conversation (15 min)**

**Objetivo:** Validar que histórico é mantido corretamente sob conversação longa.

**Passos:**

1. Faça 15+ turnos (perguntas + respostas):
   ```
   1. Pesquise sobre Rust Web Development
   2. Quais são os frameworks mais populares?
   3. Qual a diferença entre Axum e Actix?
   4. [Continue fazendo perguntas...]
   ```

2. **Observações críticas:**
   - [ ] Histórico foi mantido corretamente? (posso referir perguntas anteriores) (SIM / NÃO)
   - [ ] Não houve degradação de performance? (SIM / NÃO)
   - [ ] Mensagens estão em ordem correta? (SIM / NÃO)
   - [ ] LLM lembrou de contexto anterior? (SIM / NÃO)

3. **Teste de recuperação:**
   - Feche o Vectora (Ctrl+C)
   - Abra novamente
   - [ ] Histórico foi recuperado do banco? (SIM / NÃO)
   - [ ] Consegue continuar conversa anterior? (SIM / NÃO)

**Se alguma resposta for NÃO:**
- Salve debug dump com a conversa incluída

---

### **CENÁRIO 5: Stress Test (10 min)**

**Objetivo:** Validar que o sistema não trava sob carga.

**Passos:**

1. Faça 5 pesquisas pesadas em sequência rápida:
   ```
   Pesquise sobre:
   1. Kubernetes 2025
   2. PostgreSQL 17
   3. Docker Container
   4. RAG Systems
   5. Prompt Engineering
   ```

2. **Observações críticas:**
   - [ ] Todas as 5 pesquisas foram completadas? (SIM / NÃO)
   - [ ] Nenhuma travou ou timeout? (SIM / NÃO)
   - [ ] TUI permaneceu responsiva? (SIM / NÃO)
   - [ ] Background worker processou todos os embeddings? (SIM / NÃO)

3. **Validação de indexação:**
   - Aguarde 60 segundos após as 5 pesquisas
   - Digite:
     ```
     Resuma tudo que você aprendeu sobre os 5 tópicos acima
     ```
   - [ ] Resposta usou dados indexados (não web search)? (SIM / NÃO)
   - [ ] Menciona fontes/tópicos específicos? (SIM / NÃO)

**Se alguma resposta for NÃO:**
- Descreva qual pesquisa falhou
- Salve debug dump

---

## 🐛 Como Reportar Bugs

### Formato Estruturado

Se encontrar algum problema, siga este formato:

```markdown
# 🐛 Bug Report

**Tester:** [Seu nome]
**Cenário:** [Qual cenário: 1-5]
**Severidade:** critical | high | medium | low

## Descrição
[Descreva o que aconteceu em detalhes]

## Passos para Reproduzir
1. [Passo 1]
2. [Passo 2]
3. [Passo 3]

## Comportamento Esperado
[O que deveria ter acontecido]

## Comportamento Atual
[O que realmente aconteceu]

## Correlation ID
[Copie do arquivo logs/mcp.log]

## Screenshots/Logs
[Se possível, copie linhas relevantes]

## Debug Dump
- Arquivo: [vectora_debug_TIMESTAMP.tar.gz]
- Tamanho: [X MB]
- Incluir: [databases / logs / all]
```

### Gerar Debug Dump

Quando encontrar um bug:

```bash
# Comando para gerar dump
python -c "import asyncio; from src.debug_dump import generate_debug_dump; asyncio.run(generate_debug_dump())"

# Ou se tiver CLI integrada
vectora --debug-dump
```

O arquivo `vectora_debug_TIMESTAMP.tar.gz` conterá:
- Bancos de dados (SQLite)
- Logs estruturados (JSON)
- Configuração (sem secrets)
- Metadados de sistema

---

## 📊 Matriz de Aprovação

Para ser aprovado para lançamento, **todos os 5 cenários devem ter SIM em todas as observações críticas**:

| Cenário | Expectativa | Seu Resultado |
|---------|-------------|---------------|
| 1. TUI Responsividade | 4/4 SIM | [ ] Aprovado |
| 2. RAG Vector Search | 3/3 SIM | [ ] Aprovado |
| 3. Error Handling | 3/3 SIM | [ ] Aprovado |
| 4. Multi-turn | 4/4 SIM | [ ] Aprovado |
| 5. Stress Test | 4/4 SIM | [ ] Aprovado |

---

## 📝 Checklist Final

Antes de enviar seu relatório:

- [ ] Executei todos os 5 cenários
- [ ] Documentei cada observação crítica (SIM/NÃO)
- [ ] Salvei debug dumps para qualquer NÃO
- [ ] Copiei correlation IDs dos logs
- [ ] Descrevi qualquer comportamento inesperado
- [ ] Testei com meu LLM configurado (Gemini/OpenAI/Ollama)
- [ ] Testei recuperação de crashes (cenário 4)

---

## 📞 Contato

Se tiver dúvidas durante o testing:

- **Email:** [Será fornecido]
- **Slack:** #vectora-qa
- **Docs:** Veja `RELEASE_ENGINEERING_ROADMAP.md`

---

## 🙏 Obrigado!

Seu feedback é **essencial** para a qualidade da v0.1.0. Teste rigorosamente!

**Prazo para retorno:** 7 dias após receber o RC1

---

**Versão do Guia:** 1.0  
**Última atualização:** 2026-05-14
