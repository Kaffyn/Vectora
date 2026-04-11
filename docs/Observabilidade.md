# Blueprint: Observabilidade & Log Interno

**Status:** Implementado
**Módulo:** `core/telemetry/`
**Tecnologia:** Structured Logging (`slog`)

A observabilidade no Vectora foca na depuração silenciosa ("Black Box") para garantir que o suporte técnico e a evolução do produto sejam baseados em dados reais, mantendo a privacidade total do código do usuário.

---

## 1. Logging Estruturado (Slog)

O Core utiliza o pacote `log/slog` do Go para gerar logs estruturados em formato JSON, permitindo análise programática rápida.

- **Níveis de Log:**
  - `DEBUG`: Detalhes de tráfego ACP/JSON-RPC (desativado por padrão).
  - `INFO`: Ciclo de vida da sessão, ativação de extensões e conclusão de tarefas.
  - `WARN`: Erros recuperáveis e timeouts de ferramentas.
  - `ERROR`: Falhas críticas no sistema ou pânico capturado.

---

## 2. Estratégia de Rotação "Black Box"

Para evitar que o Vectora consuma todo o espaço em disco do usuário com logs:

- **Circular Buffer:** O sistema mantém apenas as últimas 48 horas de logs de execução.
- **Auto-Rotation:** Arquivos de log são rotacionados quando atingem 10MB. O sistema mantém no máximo 5 arquivos antigos.
- **Localização:** `%LOCALAPPDATA%/Vectora/logs/daemon.log`

---

## 3. Telemetria de Ferramentas

Cada execução de ferramenta agêntica é monitorada:

- **Latência:** Tempo decorrido entre o disparo e a resposta da ferramenta.
- **Sucesso/Falha:** Taxa de erros de compilação ou falhas de sub-agentes.
- **Custo de Contexto:** Quantidade de tokens lidos e produzidos em cada interação.

---

## 4. Auditoria de Segurança (Guardian Logs)

O `Guardian` possui um canal de log dedicado e auditável que registra tentativas de acesso a arquivos fora do diretório de confiança ou detecção de segredos em outputs de ferramentas. Estes logs são essenciais para depurar políticas de segurança excessivamente restritivas ou permissivas.
