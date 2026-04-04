# Llama Package Manager Implementation Plan

Este plano detalha a transição do Vectora para um sistema de Gerenciamento de Pacotes Llama dinâmico. O objetivo é remover o "peso morto" dos binários do instalador principal e oferecer suporte nativo e otimizado para cada variante de hardware (NVIDIA/CUDA, AMD/VULKAN, CPU/AVX2).

## Objetivos Estratégicos

1. **Suporte Multi-Variante**: Detectar automaticamente se o usuário possui GPU NVIDIA (CUDA), suporte a Vulkan ou apenas CPU (AVX2/AVX512).
2. **Download sob Demanda**: Baixar os binários diretamente das [Releases Oficiais do llama.cpp (b8583)](https://github.com/ggml-org/llama.cpp/releases/tag/b8583).
3. **Separação de Preocupações**: O instalador do Vectora foca na interface e no motor; o Llama Manager cuida da IA local.
4. **Extensibilidade**: Facilitar a atualização para novas versões do Llama (ex: b8600) através de uma simples atualização de catálogo JSON/Map.

## Mudanças Propostas

### 1. Novo Pacote: `internal/llama/manager`

Este será o cérebro do gerenciamento de versões.

- **`discovery.go`**: Lógica de "Hardware Probing".
  - Verificar presença de `vulkan-1.dll`.
  - Verificar `nvidia-smi` ou APIs DXGI para CUDA.
  - Verificar flags de CPU via `runtime` e `cpuinfo`.
- **`catalog.go`**: Mapeamento de versões.
  - Exemplo: `b8583` -> `win-vulkan-x64.zip`, `win-avx2-x64.zip`, etc.
- **`fetcher.go`**: Motor de Download e Extração.
  - Download com barra de progresso (via CLI/App).
  - Verificação de integridade SHA256.
  - Extração para `%USERPROFILE%/.Vectora/packages/llama/b8583/[variante]`.

### 2. Integração com o Motor (`internal/os`)

- **`windows.go`**: O `StartLlamaEngine` deixará de buscar um caminho fixo. Ele consultará o `LlamaManager` para obter o `CurrentActivePath()`.
- Se o binário não for encontrado, o Vectora disparará um evento de "Requisitos Faltantes" para a interface.

### 3. Reformulação do Build System (`build.ps1`)

- **REMOVER**: O passo `[5/8] Compilando Llama-Installer`.
- **REMOVER**: A cópia de binários pesados de `internal/os/windows/llama-b8583`.
- **ADICIONAR**: Um novo comando na CLI: `vectora llama install b8583 --auto`.

## Plano de Verificação

### Testes Automatizados

- Simular diferentes saídas de hardware para validar a lógica de escolha da variante (Mock Probing).
- Validar a extração e resolução de caminhos em ambiente de teste temporário.

### Verificação Manual

1. Remover a pasta de pacotes e abrir o Vectora: Garantir que ele ofereça o download do Llama.
2. Verificar no log qual variante foi escolhida (ex: "Hardware Detectado: NVIDIA -> Baixando variant CUDA").
3. Testar o `Stop/Start` com diferentes versões instaladas simultaneamente.

---

_Assinado: Antigravity AI Engine_
