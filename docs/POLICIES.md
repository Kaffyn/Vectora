# Blueprint: Motor de Políticas & Guardian Engine

**Status:** Implementado  
**Módulo:** `core/policies/`  
**Objetivo:** Garantir a segurança do usuário através de regras imutáveis de acesso ao sistema de arquivos e proteção de dados sensíveis.

No Vectora, a segurança não é baseada em "prompts" de sistema, que podem ser burlados via injeção. Ela é implementada como uma camada de software obrigatória entre o agente e o sistema operacional.

---

## 1. O Motor Guardian

O `Guardian` é o guardião de todas as operações de I/O. Ele opera sob três pilares principais:

### A. Trust Folder Enforcement

Qualquer comando ou leitura de arquivo deve estar contido no diretório de confiança (Trust Folder) definido na inicialização da sessão.

- **Prevenção de Path Traversal:** O `Guardian` normaliza caminhos e bloqueia qualquer tentativa de usar `..` para escapar do diretório.
- **Symlink Check:** Bloqueia o seguimento de links simbólicos que apontam para fora do escopo do projeto.

### B. Proteção de Arquivos Sensíveis

O sistema possui uma lista negra imutável de extensões e nomes de arquivos que nunca devem ser lidos pelo LLM:

- **Segredos:** `.env`, `.pem`, `.key`, `id_rsa`, `secrets.yml`.
- **Binários:** `.exe`, `.dll`, `.so`, `.db`, `.sqlite`.
- **Ambiente:** `.bash_history`, `.ssh/`, `.aws/`.

### C. Sanitização de Saída (Redação de Segredos)

O `Guardian` intercepta o output de ferramentas (especialmente `run_shell_command`) e aplica filtros de Regex para remover padrões de API Keys e tokens conhecidos (AWS, OpenAI, Gemini) antes que eles cheguem ao contexto do agente.

---

## 2. Estratégia de Snapshots (Git Passive)

O Vectora segue uma política de "Não Destrutivo por Padrão":

- Antes de qualquer operação de escrita ou edição via ferramenta agêntica, o sistema verifica se o diretório é um repositório Git.
- Se for, ele cria um commit automático com a tag `[vectora-snapshot]` para garantir que o usuário possa reverter qualquer mudança indesejada com um simples `git reset`.
- Se não houver Git, o sistema solicita permissão explícita para modificar arquivos sem backup.

---

## 3. Limites de Recursos

Para evitar loops infinitos ou consumo excessivo de recursos:

- **Timeouts de Comando:** Máximo de 30 segundos por execução de shell.
- **Limite de Leitura:** Truncagem automática de arquivos maiores que 50KB.
- **Limite de Memória:** O Core monitora o uso de memória e mata processos filhos que excedam os limites definidos configurados.

---

## 4. Configuração via `policies.yaml`

Embora as regras base sejam embutidas no código (go:embed), o usuário pode expandir as políticas através do arquivo de configuração global, adicionando caminhos extras a serem ignorados por motivos de privacidade.
