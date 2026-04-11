# Blueprint: Executores de Ferramentas (Tier 1 Tools)

**Status:** Implementado
**Módulo:** `core/tools/`
**Segurança:** Guardian-Mediated

Este blueprint detalha a suite de ferramentas de sistema que o Vectora Core expõe para seus agentes e sub-agentes. Todas as ferramentas são escritas em Go nativo para máxima portabilidade e performance.

---

## 1. O Registro de Ferramentas (`registry.go`)

O `Registry` é o componente central que injeta as dependências de segurança e estado em cada ferramenta antes de sua execução.

- **Trust Folder:** Todas as ferramentas de arquivo operam exclusivamente dentro do diretório de confiança.
- **Guardian Integration:** Antes de ler ou escrever, o executor consulta o `Guardian` para validar permissões de arquivo.

---

## 2. Ferramentas de Sistema de Arquivos (FS)

### `read_file`

- **Capacidade:** Lê o conteúdo de arquivos de texto.
- **Proteção:** Bloqueia automaticamente segredos e binários. Possui truncagem automática em 50KB para preservar o contexto do LLM.

### `write_file`

- **Capacidade:** Cria ou sobrescreve arquivos.
- **Integridade:** Integra-se ao `git` para criar um commit de backup automático antes da alteração, permitindo reversão total se o agente falhar.

### `edit` (Search & Replace)

- **Capacidade:** Realiza edições locais e precisas.
- **Vantagem:** Evita o custo de tokens de ler e escrever arquivos inteiros quando apenas uma linha precisa mudar.

### `read_folder`

- **Capacidade:** Escaneamento recursivo de diretórios.
- **Filtro:** Ignora automaticamente `.git`, `node_modules` e pastas definidas como protegidas no `POLICIES.md`.

---

## 3. Ferramentas de Navegação e Busca

### `grep_search`

- **Capacidade:** Busca rápida por padrões Regex em todo o workspace.
- **Uso:** Essencial para o sub-agente localizar declarações de funções ou referências cruzadas sem precisar indexar tudo.

### `find_files`

- **Capacidade:** Localiza arquivos por nome ou glob patterns (ex: `*.go`).

---

## 4. Ferramentas de Sistema e Memória

### `run_shell_command`

- **Capacidade:** Execução de scripts no terminal do usuário.
- **Sandbox:** Roda no diretório do projeto com timeout rígido (30s) e captura de stdout/stderr.
- **Segurança:** O `Guardian` filtra o output para remover segredos expostos acidentalmente.

### `save_memory`

- **Capacidade:** Persistência de fatos importantes sobre o workspace.
- **Armazenamento:** Salvo no banco de dados local (BBolt) e reutilizado em sessões futuras para reduzir alucinações sobre a arquitetura do projeto.

---

## 5. Garantia de Segurança (Guardian Workflow)

Toda execução segue este fluxo:
`LLM Call` -> `Registry` -> `Guardian.IsAllowed?` -> `Tool.Execute` -> `Guardian.SanitizeOutput` -> `ACP Response`.
