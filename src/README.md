# 🛠️ Source Core: Vectora Engine

Este diretório contém a implementação completa do ecossistema **Vectora**. O projeto segue uma arquitetura modular em Go, onde o core gerencia o ciclo de vida dos dados e da inferência.

## 📂 Estrutura de Pastas

### 1. `core/` (O Cérebro em Go)

Contém a lógica de negócio pura e a orquestração de sistemas.

- **`domain/`**: Entidades e interfaces fundamentais (Repositories, Providers). É a camada mais interna e agnóstica.
- **`rag/`**: Implementação dos serviços de Ingestão e Busca. Aqui reside a inteligência de como os pedaços (chunks) de código são cruzados com a API da Engine.
- **`ai/`**: Gerenciador de Sidecars. Controla o processo `llama.cpp` e faz a ponte com a Gemini API para recursos multimodais.
- **`db/`**: Gerenciamento de conexões NoSQL e orquestração de transações.
- **`config/`**: Gestão de caminhos dinâmicos baseados no diretório do usuário (`~/.Vectora`).

- **`os/`**: Orquestração nativa para Windows, Linux e macOS. Contém a lógica de ativação de sidecars, gestão de processos e métricas do sistema.

### 2. `lib/` (Implementações de Infraestrutura)

Contém os adaptadores pesados que interagem com o mundo exterior ou bancos de dados específicos.

- **`chromem-go/`**: Motor vetorial 100% Go para RAG.
- **`bbolt/`**: Repositório industrial NoSQL (KV) para chats, logs e persistência ACID.
- **`filestore/`**: Gestão de arquivos e cache em JSON.

### 3. `studio/` (App Nativo - Wails)

Interface master do ecossistema, construída em Go e Wasm.

- **Papel:** É o aplicativo Vectora em si. Gerencia o ecossistema, Data Library e configurações globais.
- **Arquitetura:** Utiliza Wails com Go-app para garantir que a UI seja nativa com o mínimo de overhead, operando sob o limite de 2GB de RAM.

### 4. `web/` (Sidecar Web - Next.js)

Dashboard especializado e interface de chat via browser.

- **Papel:** Funciona como um **Sidecar opcional**. Provê uma visualização rica e acessível via rede local para monitoramento.
- **Arquitetura:** Frontend em Next.js (TypeScript) compilado e embutido, mas operando como um módulo isolado do core visual.

### 5. `cli/` (Modo CLI)

Interface de linha de comando baseada em Bubbletea.

- Foca em velocidade e comandos de manutenção (re-indexação, troca de modelo).

---

## 📜 Regras de Desenvolvimento Interno

- **Pureza de Domínio**: Nunca importe bibliotecas de infraestrutura (como bbolt ou chromem-go) dentro de `core/domain`.
- **Injeção de Dependência**: Serviços em `core/rag` devem receber interfaces, permitindo mocks fáceis para testes de 300%.
- **Zero-Daemon**: Nenhuma parte do código deve assumir a existência de um serviço rodando fora do controle do Vectora.
