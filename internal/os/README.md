# Vectora OS Manager (`internal/os`)

Bem-vindo à **Camada de Abstração de Sistema Operacional** do ecossistema corporativo Vectora (Kaffyn Ecosystem).
Este módulo atua como o alicerce fundamental do Core/Daemon garantindo que a infraestrutura de inteligência artificial rode nativamente e de maneira fluida em todas as arquiteturas modernas sem gargalos de compatibilidade.

---

## 🎯 Propósito e Escopo

O pacote genérico `internal/os` tem duas responsabilidades centrais e restritas:

1. **Orquestração de Binários Nativos (`llama.cpp`)**:
   A pasta guarda, gerencia e executa os motores locais de inferência pesada adaptados para `Windows`, `macOS` e `Linux`. Dependendo do sistema hospedeiro em tempo de execução, o `OS Manager` injeta os links de execução para o binário correto sem exigir que o usuário final compile códigos C++ ou baixe dependências externas.
2. **Abstração Cross-Platform**:
   Agi como um "tradutor" universal dos caprichos dos Sistemas Operacionais dentro do back-end em Go, contemplando:
   - Resoluções de diretórios agressivas (Ex: `%APPDATA%` e `%LOCALAPPDATA%` no Windows, `~/Library/Application Support` no macOS).
   - Gerenciamento de ciclo de vida de subprocessos via PIDs severos (Cleanups, Graceful Shutdowns e mitigação de Processos Zumbis).
   - Identificação de poder computacional/Hardware Local (Detecção nativa de `CUDA` para NVIDIA, `Metal` para Apple Silicon ARM64, `Vulkan` para AMD).

## 🗂️ Estrutura Proposta de Diretórios

O arcabouço local deve escalonar da seguinte forma:

- `/windows` - Lógicas dedicadas a syscalls Win32, rotas de bibliotecas dinâmicas (.dll) e executáveis `llama-server.exe`.
- `/macos` - Tratamentos de Plist, sandboxing Darwin, e compilações `arm64/Metal`.
- `/linux` - Configurações agressivas de ELF binaries e SysV / Systemd links.
- `manager.go` - O contrato/interface universal exposta que nosso Daemon (`cmd/vectora`) chama de forma abstrata, blindando toda a dor de cabeça arquitetural.

## 🚀 Fluxo de Trabalho e Desacoplamento

Quando o Ecossistema Kaffyn solicita a subida do Motor Cognitivo Local (Qwen3 / `llama.cpp`):

1. O `manager.go` investiga `runtime.GOOS` e variáveis de ambiente para rastrear qual trincheira de S.O ele está pisando.
2. Ele localiza o repositório de binários isolado para esse sistema físico.
3. Dimensiona dinamicamente as flags de aceleração (Layers GPU) que aquele S.O suporta perfeitamente por hardware.
4. Mapeia as portas e encapsula a execução do servidor IA restrito e seguro em `localhost`, retornando um canal interativo de controle para o Daemon-mestre.

## 🛠️ Regras Técnicas para Expansão e Suporte (Guidelines)

- **Nunca Hardcode Camadas de S.O:** O núcleo base do Vectora espalhado fora desta pasta **nunca** deve invocar diretórios ou chamadas OLE/Darwin diretas; chame a interface do Manager contido aqui.
- **Binários Estáticos (Statically Compiled):** Toda injeção de `llama.cpp` nestas pastas voltadas a produção devem preferencialmente dispensar links dinâmicos para prover um software real "zero-dependências" (Plug and Play).
- **Graceful CPU Degradation:** Sempre construa a falha dos bridges de GPU do S.O como alertas leves que desarmam graciosamente a IA fluindo a renderização via paralelização severa de CPU, sem travar o usuário comum ou gerar _kernel panics_.
