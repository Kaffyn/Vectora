# 🎨 Vectora Wails: O Maestro Visual

Este diretório é o ponto de unificação da interface nativa do **Vectora**. Através do framework **Wails**, orquestramos dois ambientes distintos que coexistem para oferecer performance extrema e interatividade rica.

---

## 🏛️ A Estrutura Unificada

Para manter o Vectora leve (<2GB RAM) e ao mesmo tempo poderoso, dividimos a interface em duas frentes modulares:

### 1. `studio/` (Studio Nativo - Master App)

Implementado em **Go + Go-app (Wasm)**.

- **Papel:** É a interface de controle mestre que se expande a partir da bandeja do sistema (Tray).
- **Por que:** Como é compilado em WebAssembly via Go, ele não carrega o overhead de um runtime JavaScript pesado. É instantâneo e foca em operações críticas:
  - **Data Library:** Gestão de downloads e injeção de conhecimentos (Backserver).
  - **Models:** Instalação e monitoramento de térmicas/RAM dos modelos Qwen3 GGUF.
  - **Settings:** Configurações globais e troca de modos (Local/Cloud).
- **Aesthetics:** Glassmorphism via CSS `backdrop-filter` nativo.

### 2. `chat/` (Rich Chat Dashboard - Next.js)

Implementado em **Next.js + TypeScript**.

- **Papel:** É a janela de diagnóstico profundo e interação multimodal.
- **Por que:** Mantemos o ecossistema Next.js aqui para aproveitar as visualizações de dados complexas e a interface de chat premium que você já construiu. O Wails carrega este componente sob demanda quando o usuário deseja uma conversa aprofundada.
- **Funcionamento:** O Wails pode atuar como um "container" para o código Next.js compilado, eliminando a necessidade de rodar o `bun dev` em produção.

---

## 🔄 Orquestração e Mensageria

O Wails atua como a ponte (Bridge) entre o seu **Core Go** e as interfaces:

1. **Bindings de Go:** Todas as funcionalidades do core (Busca RAG, Ingestão, Status do Llama) são expostas como métodos de Go para JS/Wasm.
2. **Eventos em Tempo Real:** O Wails gerencia eventos de streaming de tokens e logs em background, entregando-os simultaneamente para o Studio ou para o Chat.
3. **Distribuição:** O projeto é compilado para um binário único. Os ativos de `studio/` e `chat/` são incluídos via `go:embed`.

## 🚀 Como Executar (Desenvolvimento)

Para rodar o ecossistema Wails em modo dev:

```bash
# De dentro da raiz do projeto
wails dev
```

---

_O Wails transforma o Vectora no "Tauri do Go", unificando o poder da sua infraestrutura com o design de ponta._
