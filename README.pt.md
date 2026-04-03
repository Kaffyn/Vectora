# Vectora

> [!TIP]
> Leia esse arquivo em outro idioma.
> Inglês [README.md] | Português [README.pt.md]

**Um NotebookLM privado que roda inteiramente na sua máquina.**

O Vectora é um assistente de IA local que aprende com o que você fornecer — documentos, código, artigos, imagens — e responde perguntas estritamente baseadas nesse conteúdo. Pense no Google NotebookLM, mas rodando no seu hardware, com seus dados nunca saindo da sua máquina.

Sem dependência de nuvem. Sem custo recorrente. Nenhum dado sai da sua máquina.

---

## O Problema

Sabe quando você pergunta a uma IA sobre algo muito específico — uma versão particular de um framework, um documento interno, um artigo técnico de nicho — e ela inventa algo ou dá uma resposta genérica que erra o ponto completamente?

Isso acontece porque a IA não tem acesso ao _seu_ contexto. O Vectora resolve isso. Forneça seus arquivos, aponte para uma base de conhecimento e ele responderá exatamente a partir disso — nada mais, nada menos.

---

## Como Funciona

O Vectora faz o embedding de seus arquivos e bases de conhecimento baixadas em bancos de dados vetoriais locais isolados. Quando você faz uma pergunta, ele recupera o contexto semanticamente mais relevante de quaisquer workspaces ativos e envia tudo — junto com sua pergunta — para o modelo de linguagem.

```markdown
Modelo Base de IA (Qwen / Gemini)  ← sempre presente, generalista
    + Workspace: Godot 4.2         ← baixado do Vectora Index
    + Workspace: Artigos de Física ← baixado do Vectora Index
    + Workspace: Seus Arquivos     ← adicionado por você, privado
              ↓
         Sua Pergunta
              ↓
       Recuperação entre workspaces ativos
              ↓
           Resposta
```

Cada workspace é um namespace completamente isolado. Contextos nunca vazam de um para outro. Você controla quais workspaces estão ativos por sessão.

---

## Vectora Index

O Index é um marketplace curado de bases de conhecimento — datasets vetoriais pré-construídos publicados pela comunidade e revisados pela Kaffyn antes de estarem disponíveis para download.

De dentro do app Vectora, você pode navegar pelo catálogo completo com busca e filtros, ler um README resumido para cada dataset descrevendo seu conteúdo, baixar qualquer dataset diretamente para seu Vectora local como um novo workspace e publicar suas próprias bases de conhecimento para outros usarem.

**Exemplos do que você encontrará no Index:**

- Documentação do Godot 4.x (por versão)
- Referências de frameworks de frontend e backend
- Artigos de engenharia, física e ciência da computação
- Recursos de game design, especificações de linguagens e mais

Todo dataset baixado do Index é indexado e armazenado localmente. Após o download, nenhuma requisição de rede é feita no momento da consulta.

---

## O Que Você Pode Fazer Com Ele?

**Estudo & Pesquisa**
Arraste PDFs, artigos ou notas para um workspace. Peça ao Vectora para explicar, resumir, correlacionar ou testar seus conhecimentos. Tudo permanece local e privado.

**Desenvolvimento**
Combine um workspace de documentação de motor com o workspace do seu próprio código. Obtenha respostas que conhecem tanto o contrato da API quanto sua implementação real.

**Trabalho Profundo**
Use o modo Gemini para indexar imagens, PDFs e áudio junto com texto — tudo processado e armazenado localmente após a indexação.

**Integração com IDE**
Exponha qualquer workspace como um servidor MCP, fornecendo contexto preciso diretamente para ferramentas como Cursor, VS Code ou Claude Code.

---

## Provedores de IA

O Vectora suporta dois provedores nativamente, com o motor construído para acomodar mais no futuro:

**Qwen3 (Local / Offline)**
Roda inteiramente no seu hardware via `llama-cli` usando a arquitetura Zero-Port de pipes. Sem necessidade de internet. Suporta texto e código usando os modelos Qwen3 (veja seção abaixo para detalhes). Ideal para fluxos de trabalho totalmente privados.

**Gemini (Nuvem / Multimodal)**
Usa sua própria chave de API Gemini, armazenada apenas na sua config local. Desbloqueia indexação multimodal — PDFs, imagens e áudio são todos suportados. A chave nunca sai da sua máquina.

Ambos os provedores incluem modelos de embedding dedicados. O Vectora não depende de um serviço de embedding separado.

## Modelos Oficiais Qwen3

O Vectora suporta a nova linhagem **Qwen3**, otimizada para diferentes frentes de desenvolvimento:

**Código & Raciocínio**

- **Qwen3-Coder-Next (80B):** O estado da arte para refatoração massiva e arquitetura de sistemas.
- **Qwen3-4B-Thinking (2507):** Modelo de raciocínio lógico (Chain-of-Thought) para resolução de bugs complexos.

**Visão & Multimodal (Thinking VL)**

- **Qwen3-VL-Thinking (2B/8B):** Modelos de visão que "pensam" sobre a imagem, ideais para analisar screenshots de bugs de interface ou diagramas de arquitetura.
- **Qwen3-VL-Embedding (2B):** Vetorização de ativos visuais e diagramas para busca semântica em GDDs.

**Áudio & Voz (ASR/TTS)**

- **Qwen3-ASR (0.6B):** Transcrição ultrarrápida de reuniões de sprint e áudios de feedback.
- **Qwen3-TTS-VoiceDesign (1.7B):** Síntese de voz de alta fidelidade (12Hz) para prototipagem de diálogos em tempo real.

**RAG & Embeddings**

- **Qwen3-Embedding (0.6B/4B/8B):** Os motores de busca vetorial que alimentam o chromem-go. **Recomendamos a versão 0.6B** para o limite rigoroso de 2GB de RAM, garantindo que o contexto do seu código seja recuperado com precisão sem comprometer a performance do sistema.

---

## Interfaces

O Vectora não é um único app — é um ecossistema de interfaces compartilhando um core comum via IPC, tudo orquestrado por um daemon leve no systray:

| Interface           | Descrição                                                                                             |
| ------------------- | ----------------------------------------------------------------------------------------------------- |
| **Systray**         | O daemon central. Vive perto do relógio, orquestra tudo, consome ~100MB de RAM.                       |
| **Web UI (Wails)**  | App desktop local powered by Next.js. Interface de chat, gestão de workspaces, config e navegação no Index. |
| **CLI (Bubbletea)** | Interface de terminal. Footprint mínimo, resposta instantânea.                                        |
| **Servidor MCP**    | Expõe o conhecimento do Vectora para ferramentas de IA externas e IDEs.                               |
| **Agente ACP**      | Modo agente autônomo com acesso ao sistema de arquivos e terminal.                                    |

---

## Toolkit Agêntico

Ao operar em modo MCP ou ACP, o Vectora expõe um conjunto compartilhado de ferramentas construídas do zero em Go:

- **Filesystem:** `read_file`, `write_file`, `read_folder`, `edit`
- **Search:** `find_files`, `grep_search`, `google_search`, `web_fetch`
- **System:** `run_shell_command`
- **Memory:** `save_memory`, `enter_plan_mode`

> [!IMPORTANT]
> Toda ação de escrita ou shell dispara um snapshot automático via `GitBridge` em `internal/git` antes da execução. Qualquer ação agêntica pode ser totalmente revertida com um único comando `undo`.

---

## Arquitetura

O Vectora é escrito inteiramente em Go. O core roda como um daemon leve no systray e inicia outras interfaces sob demanda via IPC.

| Componente      | Tecnologia               | Papel                                                       |
| --------------- | ------------------------ | ----------------------------------------------------------- |
| Vector DB       | chromem-go               | Busca semântica e embeddings                                |
| Key-Value DB    | bbolt                    | Histórico de chat, logs, configuração                       |
| Motor de IA     | langchaingo              | Abstração de LLM e provedor de embedding (Gemini, expansível) |
| Inferência Local| llama-cli (pipes)        | Execução de modelos offline (Qwen3)                         |
| Instalador      | Fyne                     | Assistente de configuração multiplataforma                  |
| Tray            | systray                  | Daemon central e orquestrador                               |
| Web UI          | Wails + Next.js (estático) | Interface de chat desktop local (em `internal/app`)         |
| CLI             | Bubbletea                | Interface de terminal                                       |
| Index Server    | Go (net/http)            | Catálogo e distribuição de datasets vetoriais               |

O Web UI é construído com Next.js em modo de exportação estática a partir de `internal/app`, embarcado no binário Wails via `go:embed`. O frontend se comunica com o backend em Go através de bindings do Wails — sem servidor HTTP, sem runtime Node.js, chamadas diretas de funções JS→Go.

Projetado para operar com **menos de 4GB de RAM** em hardware modesto.

---

## Roadmap

- [ ] Integração completa end-to-end (em progresso)
- [ ] Primeiro release público
- [ ] Lançamento público do Vectora Index
- [ ] Indexação multimodal (imagens, PDFs) via Gemini
- [ ] Transcrição e indexação de áudio
- [ ] Site e documentação do Vectora

---

_Parte da organização open source [Kaffyn](https://github.com/Kaffyn)._
