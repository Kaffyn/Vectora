# Security Policy

Vectora é uma aplicação local-first. A superfície de ataque é intencionalmente pequena — sem servidores remotos, sem autenticação de usuários, sem APIs REST expostas. Mesmo assim, o código que executa em nome do usuário precisa de proteções reais.

---

## Proteções Built-in

### Execução de Terminal (`terminal` tool)

O terminal é a ferramenta de maior risco. Proteções aplicadas:

- **Whitelist de comandos**: Bloqueados por padrão — `rm -rf`, `mkfs`, `dd if=/dev/zero`, `:(){:|:&};:` (fork bomb) e similares. A whitelist está em `vectora/services/security.py`.
- **Timeout**: 30 segundos por execução. Processos que excedem são terminados via `proc.kill()`.
- **Async execution**: Usa `asyncio.create_subprocess_shell` — nunca bloqueia o event loop ou a UI.
- **No persistent shell**: Cada execução é um processo filho novo. Não há sessão de terminal persistente entre chamadas.

### Operações de Arquivo

- **Path traversal prevention**: Todos os caminhos são validados antes de leitura/escrita. `../../../etc/passwd` é rejeitado.
- **No symlink attacks**: Validação de paths resolve symlinks antes de operar.
- **Allowed directories**: `file_read`, `file_edit` e `file_write` operam apenas em diretórios permitidos (configurável via `ENABLE_FILE_OPERATIONS`).

### Regex — Proteção anti-ReDoS

O `grep` tool aceita padrões regex do LLM. Padrões maliciosos podem causar backtracking catastrófico (ReDoS). Proteção:

- Validação de padrão em `vectora/services/security.py` (`is_safe_regex_pattern`)
- Timeout de 20s na execução do grep
- Filtros de tipos de arquivo (ignora `.pyc`, binários)

### Secrets e Logs

- **Nenhum secret em logs**: API keys são mascaradas antes de qualquer log statement.
- **`.env` não commitado**: `.gitignore` inclui `.env`, `.env.*` (exceto `.env.example`).
- **`~/.vectora/keys/`**: Chaves opcionalmente encriptadas no diretório de dados do usuário.
- **LangSmith optional**: Tracing só ativado se `LANGSMITH_API_KEY` estiver explicitamente configurada.

### MCP Server

- **stdio mode**: Comunicação via stdin/stdout. Nenhuma porta de rede aberta localmente.
- **SSE mode**: Escuta em `MCP_HOST:MCP_PORT` (default `0.0.0.0:8000`). Em produção, coloque atrás de um reverse proxy com TLS.
- **stderr para feedback**: O Rich panel de startup vai para stderr — nunca polui o canal JSON-RPC do protocolo MCP.

---

## O que NÃO está no MVP

Por design — Vectora é local-first, single-user:

- ❌ **Authentication**: Sem usuários, sem tokens de acesso, sem OAuth.
- ❌ **TLS/SSL**: Comunicação local via stdio; SSE em LAN sem TLS é aceitável no MVP.
- ❌ **Rate limiting**: Não há múltiplos usuários competindo por recursos.
- ❌ **Audit log completo**: Apenas logs estruturados em JSON.
- ❌ **Sandboxing do LLM**: O LLM pode solicitar execução de qualquer ferramenta habilitada.

Esses itens são parte do roadmap para quando Vectora for multi-tenant.

---

## Reportar Vulnerabilidade

Se você encontrar uma vulnerabilidade de segurança, **não abra uma issue pública**. Isso exporia o problema antes de haver um fix.

**Como reportar:**

1. Abra um [GitHub Security Advisory](https://github.com/brunosrz/vectora/security/advisories/new) (privado)
2. Inclua: componente afetado, tipo de vulnerabilidade, passos para reproduzir, impacto estimado

**O que incluir:**

- Versão do Vectora (`vectora --version`)
- Sistema operacional
- Passos mínimos para reproduzir
- Comportamento esperado vs. observado

**O que NÃO incluir:**

- Secrets, tokens ou credenciais reais
- Exploit técnico completo antes de coordenar disclosure

**Processo de disclosure:**

1. Maintainer confirma recebimento em 48h
2. Valida e reproduz a vulnerabilidade
3. Desenvolve e testa o fix
4. Publica fix + advisory coordenados

---

## Práticas de Desenvolvimento Seguro

Todo código submetido deve seguir:

- **Validar inputs**: Nunca confiar em dados vindos do LLM ou de `function_results`
- **Secrets fora do git**: Verificar `.gitignore` antes de qualquer commit
- **Least-privilege**: Ferramentas operam no escopo mínimo necessário
- **Dependências revisadas**: `uv audit` para verificar advisories conhecidos
- **Atenção especial** a: security.py, qualquer código que aceite paths do usuário, qualquer execução de subprocess
