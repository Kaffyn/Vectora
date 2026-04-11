# Blueprint: Gerenciador de Configuração & Segredos

**Status:** Implementado
**Módulo:** `core/config/`
**Segurança:** AES-GCM Encryption para chaves de API

O `ConfigManager` é o componente responsável pela persistência das preferências do usuário, chaves de API e configurações de provedores LLM. Ele garante que dados sensíveis sejam armazenados de forma segura no ambiente local.

---

## 1. Localização e Estrutura dos Arquivos

O Vectora segue os padrões do sistema operacional para armazenamento de dados de aplicação:

- **Windows:** `%LOCALAPPDATA%/Vectora/config.yaml`
- **Linux:** `~/.config/vectora/config.yaml`
- **macOS:** `~/Library/Application Support/Vectora/config.yaml`

### Estrutura do `config.yaml`:

```yaml
app:
  language: "pt"
  theme: "dark"
providers:
  gemini:
    enabled: true
    api_key: "AES-GCM-ENCRYPTED-K..."
    model: "gemini-1.5-pro"
  openai:
    enabled: false
    api_key: ""
workspaces:
  - path: "/home/user/project1"
    trust_level: "high"
```

---

## 2. Segurança de Segredos (Encryption at Rest)

Para mitigar o risco de vazamento de chaves de API caso o arquivo `config.yaml` seja acessado, o Vectora implementa criptografia simétrica:

- **Algoritmo:** **AES-256-GCM**.
- **Chave de Criptografia:** Gerada unicamente para cada máquina a partir de um identificador de hardware único (Machine ID), garantindo que o arquivo de configuração não funcione se copiado para outro computador.
- **Processo:** Toda chave de API é criptografada antes da escrita no disco e descriptografada apenas em memória durante a inicialização do `LLMProvider`.

---

## 3. Isolamento e Gerenciamento de Workspaces

O `ConfigManager` rastreia todos os "Trust Folders" aprovados pelo usuário.

- **Workspace Hash:** Cada diretório é identificado por um hash SHA-256 de seu caminho absoluto.
- **Configurações Locais:** Permite sobrescrever modelos de LLM padrão para projetos específicos (ex: usar um modelo mais barato para documentação e um mais potente para refatoração de código).

---

## 4. Integração com o Core

O `ConfigManager` é injetado globalmente no `Engine`. Qualquer mudança detectada no arquivo de configuração dispara um sinal de recarregamento (Hot Reload) para os provedores de LLM e motor de busca, permitindo mudanças dinâmicas sem a necessidade de reiniciar o daemon.
