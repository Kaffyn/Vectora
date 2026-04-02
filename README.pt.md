# Vectora

> [!TIP]
> Read this file in another language.
> English [README.md] | Portuguese [README.pt.md]

**Um NotebookLM privado que roda inteiramente na sua máquina.**

Vectora é um assistente de IA local que aprende com o que você fornece — documentos, código, papers, imagens — e responde perguntas baseado estritamente nesse conteúdo. Pense no Google NotebookLM, mas rodando no seu hardware, com seus dados nunca saindo da sua máquina.

Sem dependência de nuvem. Sem custo recorrente. Sem dados saindo da sua máquina.

---

## O Problema

Sabe quando você pergunta para uma IA algo muito específico — uma versão particular de um framework, um documento interno, um paper de pesquisa de nicho — e ela ou inventa uma resposta ou te dá algo genérico que não resolve?

Isso acontece porque a IA não tem acesso ao _seu_ contexto. O Vectora resolve isso. Alimente-o com seus arquivos, aponte para uma base de conhecimento, e ele responde exatamente com base nisso — nada mais, nada menos.

---

## Como Funciona

O Vectora embute seus arquivos e bases de conhecimento baixadas em bancos de dados vetoriais locais isolados. Quando você faz uma pergunta, ele recupera o contexto mais semanticamente relevante dos workspaces que você tem ativos e envia tudo — junto com sua pergunta — para o modelo de linguagem.

```markdown
Modelo Base de IA (Qwen / Gemini)   ← sempre presente, totalmente generalista
    + Workspace: Godot 4.2          ← baixado do Vectora Index
    + Workspace: Papers de Física   ← baixado do Vectora Index
    + Workspace: Seus Arquivos      ← adicionado por você, privado
              ↓
         Sua Pergunta
              ↓
        Recuperação nos workspaces ativos
              ↓
           Resposta
```

Cada workspace é um namespace completamente isolado. Os contextos nunca se misturam. Você controla quais workspaces estão ativos por sessão.

---

## Vectora Index

O Index é um marketplace curado de bases de conhecimento — datasets vetoriais pré-construídos, publicados pela comunidade e revisados pela Kaffyn antes de ficarem disponíveis para download.

De dentro do app Vectora, você pode navegar pelo catálogo completo com busca e filtros, ler um README leve de cada dataset descrevendo seu conteúdo, baixar qualquer dataset diretamente para o seu Vectora local como um novo workspace, e publicar suas próprias bases de conhecimento para outros usarem.

**Exemplos do que você encontrará no Index:**

- Documentação do Godot 4.x (por versão)
- Referências de frameworks frontend e backend
- Papers de engenharia, física e ciência da computação
- Recursos de game design, especificações de linguagens e muito mais

Todo dataset baixado do Index é embarcado e armazenado localmente. Após o download, nenhuma requisição de rede é feita no momento da consulta.

---

## O que Você Pode Fazer com Ele?

**Estudo e Pesquisa**
Coloque PDFs, papers ou anotações em um workspace. Peça ao Vectora para explicar, resumir, cruzar referências ou te fazer perguntas. Tudo permanece local e privado.

**Desenvolvimento**
Combine um workspace de documentação de engine com um workspace da sua base de código. Obtenha respostas que consideram tanto o contrato da API quanto sua implementação real.

**Trabalho Intenso**
Use o modo Gemini para indexar imagens, PDFs e áudios junto com texto — tudo processado e armazenado localmente após a indexação.

**Integração com IDE**
Exponha qualquer workspace como um servidor MCP, alimentando contexto preciso diretamente em ferramentas como Cursor, VS Code ou Claude Code.

---

## Provedores de IA

O Vectora suporta dois provedores nativamente, com o motor construído para acomodar mais no futuro:

**Qwen (Local / Offline)**
Roda inteiramente no seu hardware via `llama.cpp`. Sem internet. Suporta texto e código usando os modelos Qwen3 (veja a seção abaixo para detalhes). Ideal para fluxos de trabalho totalmente privados.

**Gemini (Cloud / Multimodal)**
Usa sua própria Gemini API Key, armazenada apenas na sua configuração local. Desbloqueia indexação multimodal — PDFs, imagens e áudios são todos suportados. A chave nunca sai da sua máquina.

Ambos os provedores incluem modelos de embedding dedicados. O Vectora não depende de um serviço de embedding separado.

## Modelos Oficiais Qwen

O Vectora suporta a nova linhagem **Qwen3**, otimizada para diferentes frentes de desenvolvimento:

**Codificação e Raciocínio**

- **Qwen3-Coder-Next (80B):** O estado da arte para refatoração massiva e arquitetura de sistemas.
- **Qwen3-4B-Thinking (2507):** Modelo de raciocínio lógico (Chain-of-Thought) para resolução de bugs complexos.

**Visão e Multimodal (Thinking VL)**

- **Qwen3-VL-Thinking (2B/8B):** Modelos de visão que "pensam" sobre a imagem, ideais para analisar print-screens de bugs de UI ou diagramas de arquitetura.
- **Qwen3-VL-Embedding (2B):** Vetorização de ativos visuais e diagramas para busca semântica em GDDs.

**Áudio e Fala (ASR/TTS)**

- **Qwen3-ASR (0.6B):** Transcrição ultra-rápida de reuniões de sprint e áudios de feedback.
- **Qwen3-TTS-VoiceDesign (1.7B):** Síntese de voz de alta fidelidade (12Hz) para prototipagem de diálogos em tempo real.

**RAG e Embeddings**

- **Qwen3-Embedding (0.6B/4B/8B):** Os motores de busca vetorial que alimentam o chromem-go. **Recomendamos a versão 0.6B** para o limite estrito de 2GB de RAM, garantindo que o contexto do seu código seja recuperado com precisão sem comprometer a performance do sistema.

---

## Interfaces

O Vectora não é um único app — é um ecossistema de interfaces compartilhando um núcleo comum via IPC, tudo orquestrado por um daemon leve no systray:

| Interface           | Descrição                                                                                                                 |
| ------------------- | ------------------------------------------------------------------------------------------------------------------------- |
| **Systray**         | O daemon central. Fica próximo ao relógio, orquestra tudo, ~5MB de RAM.                                                   |
| **Web UI (Wails)**  | App desktop local baseado em Next.js. Interface de chat, gerenciamento de workspaces, configurações e navegação no Index. |
| **CLI (Bubbletea)** | Interface de terminal. Consumo mínimo, resposta imediata.                                                                 |
| **Servidor MCP**    | Expõe o conhecimento do Vectora para ferramentas e IDEs externas.                                                         |
| **Agente ACP**      | Modo agente autônomo com acesso ao sistema de arquivos e terminal.                                                        |

---

## Toolkit Agêntico

Ao operar em modo MCP ou ACP, o Vectora expõe um conjunto compartilhado de ferramentas construídas do zero em Go:

- **Sistema de Arquivos:** `read_file`, `write_file`, `read_folder`, `edit`
- **Busca:** `find_files`, `grep_search`, `google_search`, `web_fetch`
- **Sistema:** `run_shell_command`
- **Memória:** `save_memory`, `enter_plan_mode`

---

## Arquitetura

O Vectora é escrito inteiramente em Go. O núcleo roda como um daemon leve via systray e inicializa as demais interfaces sob demanda via IPC.

| Componente        | Tecnologia                 | Papel                                                        |
| ----------------- | -------------------------- | ------------------------------------------------------------ |
| Banco Vetorial    | chromem-go                 | Busca semântica e embeddings                                 |
| Banco Chave-Valor | bbolt                      | Histórico de chat, logs, configuração                        |
| Motor de IA       | langchaingo                | Abstração de provedores LLM e embedding (Gemini, extensível) |
| Inferência Local  | llama.cpp (sidecar)        | Execução offline de modelos (Qwen)                           |
| Instalador        | Fyne                       | Wizard de setup multiplataforma                              |
| Tray              | systray                    | Daemon central e orquestrador                                |
| Web UI            | Wails + Next.js (estático) | Interface de chat desktop local                              |
| CLI               | Bubbletea                  | Interface de terminal                                        |
| Index Server      | Go (net/http)              | Catálogo e distribuição de datasets vetoriais                |

O Web UI é construído com Next.js em modo de export estático, embarcado no binário Wails via `go:embed`. O frontend se comunica com o backend Go através de Wails bindings — sem servidor HTTP, sem runtime Node.js, chamadas JS→Go diretas.

Projetado para operar com menos de **4GB de RAM** em hardware modesto.

---

## Roadmap

- [ ] Integração ponta-a-ponta completa (em andamento)
- [ ] Primeiro release público
- [ ] Lançamento público do Vectora Index
- [ ] Indexação multimodal (imagens, PDFs) via Gemini
- [ ] Transcrição e indexação de áudio
- [ ] Site e documentação do Vectora

---

_Parte da organização open source [Kaffyn](https://github.com/Kaffyn)._
