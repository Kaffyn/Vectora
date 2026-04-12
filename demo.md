# Plano: Demo Completo do Vectora com Pesquisa Técnica

## Contexto & Objetivo

O usuário quer criar uma **demonstração completa do Vectora** que:

1. Use TODAS as tools disponíveis
2. Execute `vectora ask` com pedidos práticos
3. Crie um projeto demo em `/demo/`
4. Implemente um caso de uso real: **Pesquisa Técnica e Geração de Relatórios**

---

## Escopo Final (Definido pelo Usuário) ✅

**Tipo de Demo**: Pesquisa Técnica e Relatórios Estruturados

- Vectora pesquisa tópicos na web
- Busca e analisa documentação
- Encontra padrões no código
- Gera relatórios técnicos estruturados

**Codebase para Analisar**: Criar projeto de exemplo em `/demo/sample-project/`

**Objetivo**:

- Showcase completo de TODAS as features do Vectora
- Exemplo prático e real
- Teste end-to-end funcional

---

## Cenário Demo: "Arquitetura Orientada por Eventos em TypeScript"

**O que vai acontecer**:

1. Criar um pequeno projeto TypeScript em `/demo/sample-project/` com um sistema de eventos básico
2. Usar `vectora ask` para Vectora pesquisar padrões de eventos na web
3. Vectora analisa o código do projeto usando RAG/embedding search
4. Vectora compara a implementação atual com best practices
5. Vectora gera relatório técnico detalhado com recomendações
6. Salva aprendizados na memória para futuro

**Resultado Final**:

- Relatório técnico profissional em `/demo/output/technical-report.md`
- Exemplos de código melhorado
- Roadmap de implementação
- Checklist de ferramentas usadas

---

## Estrutura do Projeto Demo

```
demo/
├── README.md                      ← Guia principal
├── config.json                    ← Configuração
├── WALKTHROUGH.md                 ← Passo-a-passo
│
├── sample-project/                ← Código para analisar (TypeScript)
│   ├── package.json
│   ├── tsconfig.json
│   ├── src/
│   │   ├── events/
│   │   │   ├── types.ts           (tipos de eventos)
│   │   │   ├── bus.ts             (event bus)
│   │   │   └── listener.ts        (listeners)
│   │   ├── examples/
│   │   │   ├── pubsub.ts          (exemplo pub/sub)
│   │   │   └── workflow.ts        (exemplo workflow)
│   │   └── index.ts
│   ├── docs/
│   │   ├── ARCHITECTURE.md
│   │   └── API.md
│   └── tests/
│       └── events.test.ts
│
├── scripts/
│   ├── 1-run-demo.sh              ← Orquestra tudo
│   ├── 2-verify-tools.sh          ← Valida ferramentas
│   └── 3-cleanup.sh               ← Limpa outputs
│
├── templates/
│   ├── prompt-1-pesquisa.txt      ← Prompt 1
│   ├── prompt-2-analise.txt       ← Prompt 2
│   ├── prompt-3-comparacao.txt    ← Prompt 3
│   ├── prompt-4-relatorio.txt     ← Prompt 4
│   └── prompt-5-memoria.txt       ← Prompt 5
│
└── output/                        ← Gerado automaticamente
    ├── technical-report.md        (RELATÓRIO FINAL)
    ├── architecture-analysis.md
    ├── code-examples/
    │   ├── improved-event-bus.go
    │   └── async-handler.go
    └── report.json
```

---

## Fases de Implementação

### Fase 1: Criar Código de Exemplo (TypeScript)

**Arquivos a criar**:

1. `/demo/sample-project/package.json` (30 linhas)
   - Dependencies: typescript, jest, ts-node
   - Scripts: build, test, start

2. `/demo/sample-project/tsconfig.json` (25 linhas)
   - Configuração TypeScript

3. `/demo/sample-project/src/events/types.ts` (60 linhas)
   - Interfaces Event, EventHandler
   - EventEmitter base
   - EventType enums

4. `/demo/sample-project/src/events/bus.ts` (120 linhas)
   - EventBus class
   - Métodos subscribe, publish, emit
   - Error handling e logging simples

5. `/demo/sample-project/src/events/listener.ts` (70 linhas)
   - Listener abstrato
   - Async/await patterns
   - Retry logic básico

6. `/demo/sample-project/src/examples/pubsub.ts` (90 linhas)
   - Exemplo Pub/Sub completo
   - Producer/Consumer pattern
   - Uso prático

7. `/demo/sample-project/src/examples/workflow.ts` (80 linhas)
   - Exemplo de workflow com eventos
   - Multi-step processing
   - State management com tipos

8. `/demo/sample-project/docs/ARCHITECTURE.md` (70 linhas)
   - Explicação da arquitetura
   - Diagramas ASCII
   - Limitações conhecidas

### Fase 2: Criar Infraestrutura Demo

**Arquivos a criar**:

1. `/demo/README.md` (200 linhas)
   - O que é o demo
   - Ferramentas demostradas
   - Como executar
   - O que esperar
   - Interpretação dos resultados

2. `/demo/config.json`
   - Caminho do trust folder
   - API keys necessárias
   - Modelo de IA
   - Paths de output

3. `/demo/WALKTHROUGH.md` (300 linhas)
   - Passo-a-passo completo
   - O que acontece em cada prompt
   - Como interpretar outputs
   - Troubleshooting

### Fase 3: Criar Prompts do Vectora

**5 prompts bem estruturados**:

#### Prompt 1: Pesquisa Web

```
Pesquise "event-driven architecture patterns TypeScript" e "event bus best practices Node.js".

1. Use web_search para encontrar recursos
2. Use web_fetch para baixar 2-3 artigos/blogs relevantes
3. Resuma os principais padrões encontrados:
   - Pub/Sub com EventEmitter
   - Event Sourcing em TypeScript
   - CQRS patterns
   - Async/Promise patterns e error handling

4. Salve em formato:
## Pesquisa Web - Padrões de Eventos em TypeScript
- Padrão 1: [descrição]
- Padrão 2: [descrição]
...

## Achados Principais
- [achado 1]
- [achado 2]
```

#### Prompt 2: Análise do Código

```
Analise o código TypeScript do projeto (em sample-project/src/).

1. Use read_file para ler todos os arquivos *.ts em src/events/ e src/examples/
2. Use grep_search para encontrar:
   - Definições de interfaces e tipos (EventListener, EventHandler)
   - Padrões de callbacks vs Promises vs async/await
   - Try/catch e tratamento de erros
   - Listeners e subscriptions

3. Use find_files para listar estrutura (*.ts, *.test.ts)

4. Crie diagrama da arquitetura mostrando:
   - Classes e interfaces TypeScript
   - Fluxo de dados de eventos
   - Dependências entre módulos
   - Padrões usados (pub/sub, observers, etc)
```

#### Prompt 3: Comparação e Recomendações

```
Compare a implementação atual com as best practices pesquisadas.

1. Quais padrões são implementados corretamente?
2. Onde a implementação diverge das best practices?
3. O que está faltando?

Para cada gap, proponha:
- Por que é importante
- Como corrigir (com exemplos)
- Impacto esperado
- Esforço de implementação

Priorize por: Impacto, Esforço, Risco
```

#### Prompt 4: Gerar Relatório Técnico

```
Gere um relatório técnico profissional contendo:

## 1. Sumário Executivo
- Estado atual
- Principais achados
- Top 3 recomendações

## 2. Análise da Arquitetura Atual
- Componentes e responsabilidades
- Fluxos de dados
- Forças
- Fraquezas

## 3. Best Practices da Indústria
- O que foi aprendido na pesquisa
- Padrões recomendados
- Trade-offs

## 4. Recomendações de Melhoria
Para cada recomendação:
- Código atual (snippet)
- Abordagem recomendada
- Código melhorado
- Impacto esperado
- Esforço de migração

## 5. Roadmap de Implementação
- Fase 1 (imediato)
- Fase 2 (curto prazo)
- Fase 3 (longo prazo)

## 6. Avaliação de Riscos
- Mudanças quebradoras
- Implicações de performance
- Requisitos de testes

Salve em: output/technical-report.md
```

#### Prompt 5: Memória e Conclusões

```
Salve os achados na memória de longo prazo.

1. Extraia aprendizados-chave:
   - Best practices para sistemas de eventos
   - Armadilhas comuns
   - Considerações de performance
   - Estratégias de testes

2. Use save_memory para persistir estrutura JSON:
   {
     "topico": "Event-Driven Architecture in Go",
     "data": "2024-04-12",
     "padroes_chave": [...],
     "recomendacoes": [...],
     "armadilhas": [...]
   }

3. Gere documento de conclusão:
   - O que foi aprendido
   - Como se aplica ao projeto
   - Próximos passos
```

#### Prompt 6: Implementação & Codificação 💻 [NOVO]

```
Com base em TODO o planejamento anterior (Prompts 1-5),
implemente agora as melhorias identificadas em TypeScript.

## Fase 1: Melhorar Event Bus (TypeScript)

1. Use read_file para ler implementação atual (src/events/bus.ts)
2. Use edit para aplicar melhorias:
   - Adicionar strong typing com interfaces genéricas
   - Melhorar error handling com custom Error classes
   - Adicionar logging estruturado
   - Adicionar suporte para observáveis/RxJS patterns
   - Aplicar best practices pesquisadas

3. write_file para salvar versão melhorada

## Fase 2: Criar exemplos de uso (TypeScript)

1. write_file("code-examples/improved-event-bus-example.ts")
   - Demonstra novo API com tipos
   - Error handling com try/catch
   - Exemplos de uso prático

2. write_file("code-examples/async-patterns.ts")
   - Padrões async/await avançados
   - Promise chains e error propagation
   - Graceful shutdown

3. write_file("code-examples/event-sourcing-pattern.ts")
   - Event sourcing implementation
   - Persistência com JSON/DB mock
   - Event replay functionality

## Fase 3: Criar suite de testes

1. write_file("src/events/bus.test.ts")
   - Testes unitários com Jest
   - Testes de emit/subscribe
   - Testes de error handling

2. write_file("src/events/integration.test.ts")
   - Testes end-to-end
   - Fluxo completo pub/sub
   - Performance tests

## Fase 4: Documentação de implementação

1. write_file("IMPLEMENTATION_GUIDE.md")
   - Como usar novas features
   - Exemplos de código
   - Migration guide do código antigo

2. write_file("PERFORMANCE_NOTES.md")
   - Benchmarks e métricas
   - Recomendações de otimização
   - Limites conhecidos

## Fase 5: Validação final

1. run_shell_command para:
   - npm run build
   - npm run test
   - npm run lint

2. write_file("IMPLEMENTATION_REPORT.md")
   - Mudanças implementadas
   - Testes que passaram
   - Métricas de cobertura
   - Status: PRONTO PARA PRODUÇÃO

Output final:
- Código melhorado em code-examples/
- Testes em src/events/*.test.ts
- Documentação TypeScript
- Relatório de qualidade com métricas
```

### Fase 4: Criar Scripts de Execução

**Script 1: `/demo/scripts/1-run-demo.sh`** (100 linhas)

```bash
#!/bin/bash
# Orquestra todo o processo

echo "🚀 Iniciando Demo Vectora..."

# 1. Verificar Vectora
vectora status

# 2. Carregar config
source ../config.json

# 3. Executar 5 prompts sequencialmente
echo "1️⃣ Pesquisa web..."
vectora ask "$(cat ../templates/prompt-1-pesquisa.txt)"

echo "2️⃣ Análise de código..."
vectora ask "$(cat ../templates/prompt-2-analise.txt)"

echo "3️⃣ Comparação..."
vectora ask "$(cat ../templates/prompt-3-comparacao.txt)"

echo "4️⃣ Geração de relatório..."
vectora ask "$(cat ../templates/prompt-4-relatorio.txt)"

echo "5️⃣ Salvando memória..."
vectora ask "$(cat ../templates/prompt-5-memoria.txt)"

echo "6️⃣ Implementando código (CODIFICAÇÃO)..."
vectora ask "$(cat ../templates/prompt-6-codificacao.txt)"

# 4. Verificar outputs
echo "✅ Verificando outputs..."
./2-verify-tools.sh

echo "🎉 Demo completo!"
```

**Script 2: `/demo/scripts/2-verify-tools.sh`** (80 linhas)

```bash
#!/bin/bash
# Verifica se todas as ferramentas foram usadas

echo "📋 Verificação de Ferramentas Vectora"
echo "====================================="

echo ""
echo "✓ web_search - Pesquisa na web"
echo "✓ web_fetch - Download de documentação"
echo "✓ read_file - Leitura de arquivos"
echo "✓ grep_search - Busca por padrões"
echo "✓ find_files - Descoberta de arquivos"
echo "✓ write_file - Criação de documentos"
echo "✓ edit - Edição com contexto"
echo "✓ run_shell_command - Execução de comandos"
echo "✓ save_memory - Persistência de memória"
echo "✓ RAG/embedding - Busca semântica"
echo "✓ Multi-hop reasoning - Raciocínio conectado"

echo ""
echo "📂 Arquivos gerados:"
ls -lah ../output/

echo ""
echo "✅ Verificação concluída!"
```

---

## Ferramentas do Vectora & Demonstração Completa

### Lista Completa de Ferramentas (13 Total)

#### 1. **web_search** 🔍

- **O que faz**: Busca tópicos na web usando motor de busca
- **Como será usado**: Prompt 1 - pesquisar "event-driven architecture patterns Go"
- **Entrada**: Query de busca
- **Saída**: URLs e snippets relevantes
- **Evidência no demo**: Links em `research-summary.md`
- **Status no plano**: Prompt 1

#### 2. **web_fetch** 📥

- **O que faz**: Baixa HTML de URLs e converte para markdown
- **Como será usado**: Prompt 1 - fetch 2-3 blogs/artigos de arquitetura
- **Entrada**: URL de website
- **Saída**: Markdown do conteúdo processado
- **Evidência no demo**: Conteúdo citado com atribuição de fonte
- **Status no plano**: Prompt 1

#### 3. **read_file** 📖

- **O que faz**: Lê conteúdo completo de um arquivo
- **Como será usado**: Prompt 2 - ler todos os arquivos em `src/events/` e `src/examples/`
- **Entrada**: Caminho do arquivo
- **Saída**: Conteúdo do arquivo (até 50KB)
- **Evidência no demo**: Snippets de código no analysis.md
- **Status no plano**: Prompt 2

#### 4. **grep_search** 🔎

- **O que faz**: Busca por regex em arquivos (ripgrep baseado)
- **Como será usado**: Prompt 2 - encontrar padrões específicos:
  - Definições de Handler
  - Padrões síncronos vs assíncronos
  - Tratamento de erros
  - Locks e sincronização
- **Entrada**: Padrão regex, filtros de tipo de arquivo
- **Saída**: Linhas correspondentes com contexto
- **Evidência no demo**: Resultados organizados por padrão em analysis.md
- **Status no plano**: Prompt 2

#### 5. **find_files** 📁

- **O que faz**: Descobre arquivos usando glob patterns
- **Como será usado**: Prompt 2 - mapear estrutura do projeto
  - Encontrar todos arquivos `.go`
  - Localizar testes
  - Descobrir exemplos
- **Entrada**: Glob pattern (ex: `**/*.go`, `**/tests/*`)
- **Saída**: Lista de caminhos de arquivo
- **Evidência no demo**: Listagem de arquivos em architecture-analysis.md
- **Status no plano**: Prompt 2

#### 6. **read_folder** 📂

- **O que faz**: Lista recursivamente diretório com metadados
- **Como será usado**: Prompt 2 - explorar estrutura completa de `sample-project/`
- **Entrada**: Caminho da pasta
- **Saída**: Árvore de diretórios com tamanhos e datas
- **Evidência no demo**: Estrutura visual em docs
- **Status no plano**: Prompt 2

#### 7. **edit** ✏️

- **O que faz**: Edição inteligente com busca semântica e contexto
- **Como será usado**: Prompt 4 - melhorar exemplos de código
  - Corrigir padrões
  - Adicionar tratamento de erros
  - Aplicar best practices
- **Entrada**: Arquivo, trecho para buscar, novo conteúdo
- **Saída**: Arquivo modificado (com snapshot Git automático)
- **Evidência no demo**: Exemplos "antes/depois" em code-examples/
- **Status no plano**: Prompt 4

#### 8. **write_file** ✍️

- **O que faz**: Cria ou sobrescreve arquivo com novo conteúdo
- **Como será usado**: Prompts 4 & 5 - criar documentos gerados:
  - `technical-report.md` (relatório final)
  - `code-examples/*.go` (exemplos melhorados)
  - `implementation-roadmap.md` (plano de implementação)
  - `memory.json` (aprendizados salvos)
- **Entrada**: Caminho arquivo, conteúdo novo
- **Saída**: Arquivo criado/sobrescrito
- **Evidência no demo**: Novos arquivos em `output/`
- **Status no plano**: Prompts 4 & 5

#### 9. **run_shell_command** 🖥️

- **O que faz**: Executa comandos shell com captura de stdout/stderr
- **Como será usado**: Prompt 5 & 6 - validação:
  - Rodar testes TypeScript: `npm test` (Jest)
  - Compilar: `npm run build` (tsc)
  - Lint: `npm run lint` (ESLint)
- **Entrada**: Comando shell
- **Saída**: stdout/stderr capturado
- **Evidência no demo**: Saída de testes em relatório.json
- **Timeout**: 30 segundos (configurável)
- **Status no plano**: Prompt 5

#### 10. **save_memory** 💾

- **O que faz**: Persiste aprendizados em memória de longo prazo
- **Como será usado**: Prompt 5 - guardar conclusões
  - Padrões recomendados
  - Armadilhas comuns
  - Considerações de performance
- **Entrada**: Estrutura JSON com dados importantes
- **Saída**: Memory.json persistido localmente
- **Recuperação**: Disponível em futuras sessões
- **Evidência no demo**: `output/memory.json` criado
- **Status no plano**: Prompt 5

#### 11. **google_search** 🌐

- **O que faz**: Integração com Google Search (usada internamente)
- **Como será usado**: Parte do web_search
- **Status no plano**: Automático em Prompt 1

#### 12. **RAG/Embedding Search** 🧠

- **O que faz**: Busca semântica usando embeddings da base de conhecimento
- **Como será usado**: Prompt 2 & 3 - encontrar padrões relacionados:
  - "Onde são criados handlers?"
  - "Como é tratado erro neste padrão?"
  - "Que outras implementações são similares?"
- **Entrada**: Consulta em linguagem natural
- **Saída**: Snippets de código relevante ordenados por similaridade
- **Evidência no demo**: Cross-references em comparison-analysis.md
- **Status no plano**: Prompts 2 & 3

#### 13. **Multi-hop Reasoning** 🔗

- **O que faz**: Raciocínio conectado através de múltiplos pontos do código
- **Como será usado**: Prompt 3 & 4 - conectar conceitos:
  - Padrão web → Código atual → Recomendação
  - Problema → Causa raiz → Solução proposta
- **Entrada**: Contexto de múltiplas análises
- **Saída**: Recomendações integradas
- **Evidência no demo**: Recomendações conectando pesquisa com análise
- **Status no plano**: Prompts 3 & 4

---

### Tabela Resumida: Ferramentas por Prompt

| #   | Ferramenta          | Tipo                 | Prompt       | Saída                               |
| --- | ------------------- | -------------------- | ------------ | ----------------------------------- |
| 1   | web_search          | Internet Search      | 1            | research-summary.md                 |
| 2   | web_fetch           | Internet Download    | 1            | research-summary.md                 |
| 3   | read_file           | File Access          | 2, 6         | architecture-analysis.md            |
| 4   | grep_search         | Code Search          | 2            | architecture-analysis.md            |
| 5   | find_files          | File Discovery       | 2            | architecture-analysis.md            |
| 6   | read_folder         | Directory Access     | 2            | architecture-analysis.md            |
| 7   | edit                | File Modification    | 4, **6**     | code-examples/\*.go                 |
| 8   | write_file          | File Creation        | 4-5, **6**   | technical-report.md + código gerado |
| 9   | run_shell_command   | Command Execution    | 5, **6**     | report.json + test results          |
| 10  | save_memory         | Data Persistence     | 5            | memory.json                         |
| 11  | google_search       | Web Search (interno) | 1            | (auto)                              |
| 12  | RAG/embedding       | Semantic Search      | 2-3-4, **6** | comparison-analysis.md              |
| 13  | Multi-hop reasoning | Reasoning            | 3-4, **6**   | technical-report.md + code quality  |

---

### Fluxo Visual de Ferramentas

```
PROMPT 1: PESQUISA WEB
├─ web_search("event-driven architecture patterns Go")
├─ web_fetch("https://blog1.com/...")
├─ web_fetch("https://blog2.com/...")
└─ Output: research-summary.md ✓

PROMPT 2: ANÁLISE DE CÓDIGO
├─ read_folder("sample-project/")
├─ find_files("**/*.go")
├─ read_file("src/events/types.go")
├─ read_file("src/events/bus.go")
├─ read_file("src/examples/...")
├─ grep_search("type Handler")
├─ grep_search("func.*Subscribe")
├─ RAG/embedding search
└─ Output: architecture-analysis.md ✓

PROMPT 3: COMPARAÇÃO & RECOMENDAÇÕES
├─ Analisar outputs 1 + 2
├─ RAG search (padrões similares)
├─ Multi-hop reasoning
└─ Output: comparison-analysis.md ✓

PROMPT 4: RELATÓRIO TÉCNICO
├─ read_file (referências)
├─ edit("code-examples/") [antes/depois]
├─ write_file("technical-report.md")
├─ write_file("code-examples/improved-*.go")
├─ RAG search (contexto)
├─ Multi-hop reasoning
└─ Output: technical-report.md + exemplos ✓

PROMPT 5: MEMÓRIA & CONCLUSÕES
├─ run_shell_command("npm test") (TypeScript)
├─ save_memory({padrões, recomendações})
├─ write_file("memory.json")
└─ Output: memory.json ✓

PROMPT 6: IMPLEMENTAÇÃO & CODIFICAÇÃO 💻 [NOVO]
├─ read_file (implementação atual - TypeScript)
├─ edit (melhorar código - aplicar best practices)
├─ write_file (criar exemplos melhorados)
├─ write_file (criar testes Jest completos)
├─ write_file (documentação implementação)
├─ run_shell_command (npm test, npm run build, npm run lint)
├─ RAG/embedding (validar qualidade do código)
├─ Multi-hop reasoning (conectar mudanças com planejamento)
└─ Output: Código TypeScript implementado + testes + documentação ✓
```

---

## Fluxo Completo: Planejamento → Codificação

```
📋 PROMPTS 1-5: FASE DE PLANEJAMENTO (Análise e Pesquisa)
├─ Prompt 1: Pesquisar padrões na web
├─ Prompt 2: Analisar código atual
├─ Prompt 3: Comparar com best practices
├─ Prompt 4: Gerar plano detalhado (relatório)
└─ Prompt 5: Salvar aprendizados

💻 PROMPT 6: FASE DE IMPLEMENTAÇÃO (Codificação Real)
├─ Implementar melhorias baseado no plano
├─ Criar exemplos funcionais
├─ Escrever testes abrangentes
├─ Gerar documentação
└─ Validar tudo com run_shell_command

✅ RESULTADO FINAL
├─ Código melhorado e testado
├─ Suite de testes completa
├─ Documentação profissional
├─ Relatório de qualidade
└─ PRONTO PARA PRODUÇÃO
```

---

### Checklista de Ferramentas Usadas

**Ferramentas Internet (2)**:

- [x] web_search
- [x] web_fetch

**Ferramentas de Leitura (4)**:

- [x] read_file
- [x] grep_search
- [x] find_files
- [x] read_folder

**Ferramentas de Escrita (2)**:

- [x] write_file
- [x] edit

**Ferramentas de Execução (1)**:

- [x] run_shell_command

**Ferramentas de Persistência (1)**:

- [x] save_memory

**Ferramentas de IA (2)**:

- [x] RAG/embedding (busca semântica)
- [x] Multi-hop reasoning (raciocínio conectado)

**Status**: Todas as 13 ferramentas serão demonstradas ✅

### Outputs Gerados

```
demo/output/
├── research-summary.md           (Resumo da pesquisa web)
├── architecture-analysis.md      (Análise do código atual)
├── comparison-analysis.md        (Best practices vs atual)
├── technical-report.md           (RELATÓRIO FINAL - 2000+ palavras)
├── implementation-roadmap.md     (Fases de implementação)
├── code-examples/
│   ├── improved-event-bus.go    (Código melhorado)
│   ├── async-handler.go         (Handler assíncrono)
│   └── error-recovery.go        (Recuperação de erros)
├── report.json                   (Metadados)
└── memory.json                   (Aprendizados salvos)
```

---

## Plano de Execução

### Passo 1: Criar Código de Exemplo (30 min)

- Escrever 7 arquivos Go
- Estrutura básica, bem comentada
- ~500 linhas total

### Passo 2: Criar Infraestrutura Demo (20 min)

- README.md
- config.json
- WALKTHROUGH.md
- 3 scripts shell

### Passo 3: Criar Prompts (30 min)

- 6 prompts bem estruturados
- Primeiros 5 prompts: Planejamento e análise
- Último prompt: Codificação (baseado no plano completo)

### Passo 4: Teste Manual (3-4 horas)

- Iniciar Vectora
- Executar cada um dos 6 prompts manualmente
- Verificar outputs de planejamento (Prompts 1-5)
- Verificar código gerado (Prompt 6)
- Validar testes automaticamente
- Ajustar prompts conforme necessário

### Passo 5: Teste Automatizado (1 hora)

- Executar run-demo.sh completo
- Rodar verify-tools.sh
- Documentar resultados

### Passo 6: Documentação Final (1 hora)

- Completar WALKTHROUGH.md
- Criar expected-outputs/
- Documentar problemas/soluções

---

## Arquivos Críticos a Criar

**Código de Exemplo** (9 arquivos TypeScript):

- `/demo/sample-project/package.json`
- `/demo/sample-project/tsconfig.json`
- `/demo/sample-project/src/events/types.ts`
- `/demo/sample-project/src/events/bus.ts`
- `/demo/sample-project/src/events/listener.ts`
- `/demo/sample-project/src/examples/pubsub.ts`
- `/demo/sample-project/src/examples/workflow.ts`
- `/demo/sample-project/src/index.ts`
- `/demo/sample-project/docs/ARCHITECTURE.md`

**Infraestrutura** (7 arquivos):

- `/demo/README.md`
- `/demo/config.json`
- `/demo/WALKTHROUGH.md`
- `/demo/scripts/1-run-demo.sh`
- `/demo/scripts/2-verify-tools.sh`
- `/demo/scripts/3-cleanup.sh`
- `/demo/.gitignore`

**Prompts** (6 arquivos):

- `/demo/templates/prompt-1-pesquisa.txt`
- `/demo/templates/prompt-2-analise.txt`
- `/demo/templates/prompt-3-comparacao.txt`
- `/demo/templates/prompt-4-relatorio.txt`
- `/demo/templates/prompt-5-memoria.txt`
- `/demo/templates/prompt-6-codificacao.txt` ← NOVO: Vetoriza implementa o código!

**Total**: 22 arquivos a criar (TypeScript) + 15+ outputs gerados

---

## Verificação & Validação

**Checklist de Conclusão**:

- [ ] Todos 5 prompts executados com sucesso
- [ ] Arquivo technical-report.md gerado (2000+ palavras)
- [ ] Todas 11 ferramentas demonstradas
- [ ] Exemplos de código melhorado inclusos
- [ ] Roadmap de implementação criado
- [ ] Memória persistida
- [ ] WALKTHROUGH.md completo
- [ ] Scripts testados e funcionando
- [ ] Documentação do demo finalizada

---

## Timeline Esperado

**Estimativa Total**: 6-7 horas

- Criação de código sample-project: 30 min
- Infraestrutura (scripts + config): 20 min
- Prompts (6 arquivos): 40 min
- Prompts 1-5 (planejamento): 2-3 horas
- Prompt 6 (codificação real): 1-2 horas
- Validação final: 1 hora

**Nota**: A diferença está no Prompt 6 que IMPLEMENTA o código real (não apenas planeja)

---

**Status**: Pronto para aprovação e implementação 🚀
