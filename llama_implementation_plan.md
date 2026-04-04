# PLANO DE IMPLEMENTAÇÃO: LLAMA PACKAGE MANAGER (LPM)

Este documento detalha a arquitetura do Llama Package Manager do Vectora — o subsistema responsável por detectar capacidades de hardware, baixar a build correta do llama.cpp do repositório oficial, gerenciar versões instaladas e se comunicar com cada variante de binário via STDIO.

---

## 1. Definição do Problema

O `llama.cpp` não é um binário único. Cada release oficial distribui mais de uma dúzia de variantes diferenciadas por:
- **Sistema operacional**: Windows, Linux, macOS
- **CPU**: AVX, AVX2, AVX512, NEON (ARM)
- **GPU**: CUDA (v11/v12), Vulkan, Metal (macOS)

Um gerenciador de pacotes resolve a inflação do instalador e garante a performance máxima em cada hardware.

---

## 2. Estrutura do Repositório (`internal/engines/`)

- `catalog.go`: Definições de tipo e registro de versões.
- `detector.go`: Detecção de capacidades de hardware (GPU/CPU).
- `downloader.go`: Download com suporte a resume (.partial).
- `extractor.go`: Extração ZIP seletiva (apenas binário e DLLs).
- `integrity.go`: Verificação SHA256 obrigatória.
- `manager.go`: Orquestrador de ciclo de vida (Install, Switch, Active).
- `process.go`: Gerenciamento do processo via JSON-ND (STDIO).
- `paths.go`: Resolução de caminhos (~/.vectora/engines/...).

---

## 3. Fluxo de Primeiro Boot

1. **Daemon Inicia**: Detecta hardware via `detector.go`.
2. **Setup Required**: Se nenhum engine estiver instalado, emite `engine.setup_required`.
3. **Download**: UI solicita `engine.install` para o `RecommendedBackend`.
4. **Pronto**: Engine é verificado, extraído e iniciado via `process.go`.

---

## 4. Regras Críticas (RN)

- **RN-LPM-01**: Instalador Vectora < 20MB (sem embutir llama.cpp).
- **RN-LPM-02**: Verificação SHA256 obrigatória antes de cada execução.
- **RN-LPM-04**: Comunicação exclusiva via STDIO (sem portas TCP).
- **RN-LPM-06**: Armazenamento em User-Space (sem necessidade de Admin para o Engine).
- **RN-LPM-09**: Catálogo embarcado via `go:embed` (imutável em runtime).
- **RN-LPM-10**: Proibido rodar duas instâncias de llama-cli simultaneamente.

---

## 5. Próximos Passos (Sprint 1)

1. Implementar o `catalog.json` e as estruturas de dados.
2. Desenvolver o `detector.go` com probe de DLLs/Syscalls no Windows.
3. Criar o orquestrador `manager.go` e integrar ao boot do Daemon.
