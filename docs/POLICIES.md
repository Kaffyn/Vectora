# Motor de Políticas Executáveis (POLICIES) - v1.1

**Status:** Fase 4 - Implementação Concluída  
**Módulo:** `core/policies/`  
**Princípio:** "Trust No One, Verify Everything."

## 1. Políticas de Escopo e Guardian (Expandido)

Além do isolamento básico, adicionamos proteção contra _Path Traversal_, _Symlink Attacks_ e vazamento de artefatos de build.

Ficheiro Técnico: [scope-guardian.yaml](file:///c:/Users/bruno/Desktop/Vectora-Dev/core/policies/rules/scope-guardian.yaml)

```yaml
policy_id: strict_scope_isolation
rule: "All I/O operations must occur within the Trust Folder."
enforcement:
  pre_condition: "resolve symlinks and verify if absolute_path is a subdirectory of Trust_Folder"
  on_violation: "blocked (Returns ErrSecurityViolation: Out of bounds)"

policy_id: symlink_security
rule: "Symlinks pointing outside the Trust Folder are treated as violations."
enforcement:
  check_method: "os.Readlink() + filepath.EvalSymlinks() before any access"
  on_violation: "blocked (Returns ErrSecurityViolation: Unsafe Symlink)"

policy_id: hard_coded_guardian
rule: "Vectora never accesses, reads, edits, or tracks sensitive metadata, binaries, or build artifacts."
enforcement:
  block_extensions:
    # Binários e Executáveis
    - ".exe", ".dll", ".so", ".dylib", ".bin", ".apk", ".ipa"
    # Bancos de Dados e State
    - ".db", ".sqlite", ".sqlitedb", ".mdb", ".accdb"
    # Chaves e Certificados
    - ".key", ".pem", ".p12", ".pfx", ".keystore"
    # Logs e Dump
    - ".log", ".dump", ".core"
  block_files:
    - ".env", ".env.local", ".env.production"
    - "secrets.yml", "credentials.json", "service-account.json"
    - "id_rsa", "id_ed25519", "authorized_keys"
  on_violation: "silent_skip on indexing; explicit block on Tooling layer"

policy_id: artifact_exclusion
rule: "Build artifacts and dependency folders are excluded from indexing to reduce noise."
enforcement:
  exclude_dirs:
    - "node_modules", "vendor", "dist", "build", "target", ".next", "out"
    - ".git", ".svn", ".hg"
  on_violation: "skip directory during walk"
```

## 2. Políticas de Mutação e GitBridge (Passivo)

Mantidas conforme definido anteriormente, com ênfase na não-invasividade.

Ficheiro Técnico: [git-passive.yaml](file:///c:/Users/bruno/Desktop/Vectora-Dev/core/policies/rules/git-passive.yaml)

```yaml
policy_id: passive_git_integration
rule: "Vectora orchestrates existing Git without invading the machine."
enforcement:
  passive_check: "run `git rev-parse --is-inside-work-tree` at startup"
  no_forced_install: "If `git` does not exist, silently disable snapshots"
  no_forced_init: "If not a git repository, NEVER execute `git init`"

policy_id: granular_snapshots
rule: "Vectora only versions files it has modified itself, keeping the human history clean."
enforcement:
  pre_write_action: "git status --porcelain <target_file>"
  action_strategy: "Execute `git add <target_file>` only. NEVER run `git add .`"
  commit_messaging: "Intermediate commits must be traceable (e.g., 'Vectora Snapshot: Edit <file>')"

policy_id: user_authorization_gate
rule: "Mutations require prior permission from the environment (IDE or CLI)."
enforcement:
  via_acp_ide: "Send JSON-RPC request asking for permission in UI. Wait for approval."
  via_mcp_cli: "Ask [Y/n] interactively or trust the auto-git-snapshot: true flag"
```

## 3. Políticas de Recuperação (RAG) e Privacidade

Adicionada política de sanitização de output para prevenir injeção de prompt ou vazamento acidental de dados estruturados sensíveis que escaparam do Guardian.

Ficheiro Técnico: [rag-priority.yaml](file:///c:/Users/bruno/Desktop/Vectora-Dev/core/policies/rules/rag-priority.yaml)

```yaml
policy_id: context_priority
rule: "Facts retrieved from the Trust Folder have ultimate priority over the cloud LLM pre-training."
enforcement:
  prompt_injection: "Contexts from chromem-go are injected as `[SYSTEM_KNOWLEDGE]`"
  block_if_no_retrieval: false
  fallback_action: "If null, the model requests via `google_search` instead of trying to guess the codebase"

policy_id: output_sanitization
rule: "Tool outputs must be sanitized before returning to the LLM to prevent prompt injection or context pollution."
enforcement:
  max_output_tokens: 8000
  truncate_strategy: "Head-and-Tail (keep start and end, remove middle if too large)"
  strip_control_chars: true
  on_violation: "truncate and append warning marker"

policy_id: privacy_shielding
rule: "Detected patterns resembling API keys or secrets in tool output must be masked."
enforcement:
  regex_patterns:
    - "AKIA[0-9A-Z]{16}" # AWS Access Key
    - "ghp_[a-zA-Z0-9]{36}" # GitHub PAT
    - "sk-[a-zA-Z0-9]{48}" # OpenAI/Generic Secret Key
  replacement_string: "[REDACTED_SECRET]"
  on_violation: "replace match with replacement string before returning to LLM"
```

## 4. Novas Políticas de Integridade do Sistema (System Integrity)

Ficheiro Técnico: [system-integrity.yaml](file:///c:/Users/bruno/Desktop/Vectora-Dev/core/policies/rules/system-integrity.yaml)

```yaml
policy_id: resource_throttling
rule: "Prevent single operations from exhausting system resources (DoS protection)."
enforcement:
  max_file_read_bytes: 5242880 # 5MB hard limit per file
  max_grep_results: 100
  terminal_timeout_seconds: 30
  on_violation: "abort operation and return error 'Resource Limit Exceeded'"

policy_id: read_only_defaults
rule: "By default, all tools are read-only unless explicitly granted write permissions by the user session."
enforcement:
  default_mode: "read_only"
  write_tools_require_auth: ["write_file", "edit", "terminal_run", "delete_file"]
  on_violation: "blocked (Returns ErrPermissionDenied: Write access required)"
```

---

## Implementação Go Atualizada (`core/policies/guardian.go`)

Aqui está como essas políticas YAML se traduzem em código Go denso e executável.

```go
package policies

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Guardian encapsula todas as regras de segurança
type Guardian struct {
	TrustFolder   string
	BlockedExts   map[string]bool
	BlockedFiles  map[string]bool
	ExcludedDirs  map[string]bool
	SecretRegexes []*regexp.Regexp
}

func NewGuardian(trustFolder string) *Guardian {
	g := &Guardian{
		TrustFolder: trustFolder,
		BlockedExts: map[string]bool{
			".db": true, ".sqlite": true, ".exe": true, ".dll": true,
			".key": true, ".pem": true, ".env": true, ".log": true,
		},
		BlockedFiles: map[string]bool{
			".env": true, "secrets.yml": true, "id_rsa": true,
		},
		ExcludedDirs: map[string]bool{
			"node_modules": true, ".git": true, "vendor": true,
			"dist": true, "build": true,
		},
	}

	// Compila regex de segredos
	g.SecretRegexes = []*regexp.Regexp{
		regexp.MustCompile(`AKIA[0-9A-Z]{16}`),
		regexp.MustCompile(`ghp_[a-zA-Z0-9]{36}`),
		regexp.MustCompile(`sk-[a-zA-Z0-9]{48}`),
	}

	return g
}

// IsPathSafe verifica escopo e symlinks
func (g *Guardian) IsPathSafe(targetPath string) bool {
	absTarget, err := filepath.Abs(targetPath)
	if err != nil {
		return false
	}

	// Resolve symlinks para evitar bypass
	realPath, err := filepath.EvalSymlinks(absTarget)
	if err != nil {
		// Se não existe, ainda podemos validar o path pretendido
		realPath = absTarget
	}

	absTrust, _ := filepath.Abs(g.TrustFolder)

	// Verifica se o path real está dentro do trust folder
	return strings.HasPrefix(realPath, absTrust+string(filepath.Separator)) || realPath == absTrust
}

// IsProtected verifica extensões e nomes de arquivo bloqueados
func (g *Guardian) IsProtected(path string) bool {
	base := filepath.Base(path)
	ext := strings.ToLower(filepath.Ext(path))

	if g.BlockedFiles[base] {
		return true
	}
	if g.BlockedExts[ext] {
		return true
	}
	return false
}

// IsExcludedDir verifica se um diretório deve ser ignorado na indexação
func (g *Guardian) IsExcludedDir(name string) bool {
	return g.ExcludedDirs[name]
}

// SanitizeOutput mascara segredos no output das tools
func (g *Guardian) SanitizeOutput(content string) string {
	for _, re := range g.SecretRegexes {
		content = re.ReplaceAllString(content, "[REDACTED_SECRET]")
	}
	return content
}
```

### Por que estas adições são críticas?

1.  **Symlink Security:** Impede que um atacante crie um link simbólico `ln -s /etc/passwd ./passwd` dentro do projeto e peça ao Vectora para ler. O `EvalSymlinks` resolve o caminho real antes da validação.
2.  **Artifact Exclusion:** Melhora drasticamente a qualidade do RAG ao impedir que o Vectora indexe milhões de linhas de `node_modules` ou arquivos de build, economizando tokens e vetorização.
3.  **Privacy Shielding:** Mesmo que um arquivo `.go` contenha uma chave API hardcoded (má prática, mas comum), o Vectora a mascara antes de enviar para a LLM na nuvem, reduzindo o risco de treinamento inadvertido ou logs expostos.
4.  **Resource Throttling:** Protege o daemon de travar ao tentar ler um arquivo de 10GB ou executar um comando infinito.

Este conjunto de políticas torna o Vectora robusto o suficiente para uso empresarial e pessoal seguro.
