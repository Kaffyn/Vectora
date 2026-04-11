# Plano: Reestruturação de Diretórios Windows (Native Pathing)

## Objetivo

Migrar o Vectora do padrão de diretório oculto (`~/.Vectora`) para o padrão nativo do Windows, separando binários de execução dos dados persistentes e de cache.

## 1. Mapeamento de Pastas

### Local (Instalação)

- **Caminho**: `%LOCALAPPDATA%\Programs\Vectora`
- **Conteúdo**:
  - `vectora.exe`
  - `llama-server.exe`
  - `tray_assets/`
- **Responsável**: `extension/vscode/src/binary-manager.ts` (Download/Update) e `core/os/manager_windows.go` (`GetInstallDir`).

### Roaming (Dados e Configuração)

- **Caminho**: `%APPDATA%\Vectora`
- **Conteúdo**:
  - `config/.env`
  - `data/db/` (BBolt)
  - `data/vectors/` (Chromem)
  - `logs/`
  - `.lock` (Singleton File Lock)
- **Responsável**: `core/os/manager_windows.go` (`GetAppDataDir`).

## 2. Mudanças Propostas

### [MODIFY] [manager_windows.go](file:///c:/Users/bruno/Desktop/Vectora/core/os/manager_windows.go)

- Atualizar `GetAppDataDir()` para retornar `%APPDATA%\Vectora`.
- Atualizar `GetInstallDir()` para retornar `%LOCALAPPDATA%\Programs\Vectora`.

### [MODIFY] [binary-manager.ts](file:///c:/Users/bruno/Desktop/Vectora/extensions/vscode/src/binary-manager.ts)

- Sincronizar a constante `VECTORA_BIN_DIR` para apontar para o novo local.

## 3. Estratégia de Migração (Opcional/Segurança)

Para evitar perda de dados:

1. Ao iniciar, o Core verifica se existe a pasta antiga `~/.Vectora`.
2. Se existir e a nova estiver vazia, move os arquivos de `/data` e `.env` para a nova localização.
3. Remove a pasta antiga após sucesso.
