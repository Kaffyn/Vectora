# 🎯 Main Entrypoint: O Ponto de Ignição

Este diretório contém o ponto de entrada unificado para o binário do **Vectora**. Seguindo o padrão idiomático de Go, a lógica de despacho reside no `cmd/core/main.go`.

## 📂 Organização por Comandos

### 1. `core/main.go`

O binário resultante é um polimorfo que altera seu comportamento baseado em flags e comandos:

- **`vectora cli`**: Abre a interface interativa de terminal (BubbleTea/TUI).
- **`vectora web`**: Inicia o servidor de API interno e o Dashboard Web.
- **`vectora mcp`**: Inicia o servidor MCP para integração externa.
- **`vectora ingest <path>`**: Comando de utilidade para vetorizar um diretório ou arquivo específico manualmente.

## 📜 Lógica de Despacho (Dispatcher)

O `main.go` é o único responsável por ler as variáveis de ambiente e as configurações iniciais. Ele despacha a execução para as pastas correspondentes em `src/`:

- Chama `src/cli` para modo terminal.
- Chama `src/core/server` para modo daemon.
- Chama `src/web` (via embedding) para renderizar a interface visual.

## 🚀 Estratégia de Build

Com a migração para **bbolt** e **chromem-go**, o core do Vectora tornou-se **Pure Go**, permitindo compilações cruzadas fáceis e estáveis. O uso de **CGO** torna-se opcional para o binário principal, sendo mantido nos sidecars isolados do `llama.cpp` para garantir performance de inferência sem comprometer a portabilidade do orquestrador.
