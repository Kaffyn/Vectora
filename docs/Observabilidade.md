# Blueprint: Observabilidade e Auditoria (The Black Box)

**Status:** Fase 4 - Implementação Concluída  
**Módulo:** `core/telemetry/`  
**Dependencies:** `log/slog` (Stdlib), `io`, `os`, `fmt`

## 1. Arquitetura do Logger Rotativo

Como o `slog` nativo não possui rotação de arquivos embutida, implementamos um `io.Writer` customizado (`RotatingWriter`) que intercepta os bytes antes de serem escritos no disco. Se o arquivo atual exceder o limite (ex: 10MB), ele é renomeado para `.old` e um novo arquivo é criado. Mantemos apenas 1 backup para economizar espaço, conforme solicitado.

### Estrutura de Arquivos

```text
core/telemetry/
├── logger.go       # Inicialização do slog com handler JSON e writer rotativo
└── rotation.go     # Lógica pura de rotação de arquivos (io.Writer implementation)
```

## 2. Implementação do Writer Rotativo (`rotation.go`)

Este componente é agnóstico ao formato do log. Ele apenas gerencia o ciclo de vida do arquivo físico.

```go
package telemetry

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
)

const (
	DefaultMaxSize = 10 * 1024 * 1024 // 10 MB
	BackupSuffix   = ".old"
)

// RotatingWriter implementa io.Writer com rotação simples de arquivos.
type RotatingWriter struct {
	filepath string
	maxSize  int64
	current  *os.File
	mu       sync.Mutex
}

func NewRotatingWriter(path string, maxSize int64) (*RotatingWriter, error) {
	if maxSize <= 0 {
		maxSize = DefaultMaxSize
	}

	rw := &RotatingWriter{
		filepath: path,
		maxSize:  maxSize,
	}

	// Garante que o diretório existe
	if err := os.MkdirAll(filepath.Dir(path), 0750); err != nil {
		return nil, err
	}

	// Abre ou cria o arquivo inicial
	if err := rw.openFile(); err != nil {
		return nil, err
	}

	return rw, nil
}

func (rw *RotatingWriter) openFile() error {
	f, err := os.OpenFile(rw.filepath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return err
	}
	rw.current = f
	return nil
}

// rotate executa a lógica de renomear o atual para .old e criar um novo.
func (rw *RotatingWriter) rotate() error {
	if rw.current != nil {
		rw.current.Close()
	}

	backupPath := rw.filepath + BackupSuffix

	// Remove o backup antigo se existir (mantém apenas 1 geração)
	os.Remove(backupPath)

	// Renomeia o atual para backup
	if err := os.Rename(rw.filepath, backupPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to rotate log: %w", err)
	}

	// Cria novo arquivo limpo
	return rw.openFile()
}

func (rw *RotatingWriter) Write(p []byte) (n int, err error) {
	rw.mu.Lock()
	defer rw.mu.Unlock()

	// Verifica tamanho atual
	info, err := rw.current.Stat()
	if err != nil {
		return 0, err
	}

	// Se escrevermos p, ultrapassaremos o limite?
	if info.Size()+int64(len(p)) > rw.maxSize {
		if err := rw.rotate(); err != nil {
			return 0, err
		}
	}

	return rw.current.Write(p)
}

func (rw *RotatingWriter) Close() error {
	rw.mu.Lock()
	defer rw.mu.Unlock()
	if rw.current != nil {
		return rw.current.Close()
	}
	return nil
}
```

## 3. Inicialização do Logger Estruturado (`logger.go`)

Configura o `slog` para usar JSON (machine-readable) e injeta o `RotatingWriter`. Também fornece helpers para contextos específicos.

```go
package telemetry

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
)

var GlobalLogger *slog.Logger

// InitLogger configura o logger global do Vectora Core.
// Deve ser chamado uma única vez no startup do daemon.
func InitLogger(logDir string) error {
	logPath := filepath.Join(logDir, "vectora-core.log")

	writer, err := NewRotatingWriter(logPath, DefaultMaxSize)
	if err != nil {
		return err
	}

	// Handler JSON para facilitar parsing por ferramentas externas ou grep
	handler := slog.NewJSONHandler(writer, &slog.HandlerOptions{
		Level: slog.LevelDebug, // Captura tudo, filtramos via output se necessário
	})

	GlobalLogger = slog.New(handler)

	// Log de inicialização para confirmar que o sistema de logs está vivo
	GlobalLogger.Info("Telemetry initialized", "path", logPath, "max_size_mb", DefaultMaxSize/1024/1024)

	return nil
}

// GetLogger retorna o logger global. Se não inicializado, usa stderr como fallback seguro.
func GetLogger() *slog.Logger {
	if GlobalLogger == nil {
		// Fallback de emergência: nunca silenciar erros críticos
		return slog.New(slog.NewTextHandler(os.Stderr, nil))
	}
	return GlobalLogger
}

// ContextLogger injeta atributos padrão no contexto para tracing de requisições.
func ContextLogger(ctx context.Context, requestID string) *slog.Logger {
	return GetLogger().With("request_id", requestID)
}

// Helpers de Níveis Específicos para Auditoria

func LogSecurityViolation(path string, reason string) {
	GetLogger().Error("SECURITY VIOLATION BLOCKED",
		"action", "file_access",
		"path", path,
		"reason", reason,
		"severity", "CRITICAL")
}

func LogToolExecution(toolName string, success bool, durationMs int64) {
	level := slog.LevelInfo
	if !success {
		level = slog.LevelWarn
	}

	GetLogger().Log(context.Background(), level, "Tool Execution",
		"tool", toolName,
		"success", success,
		"duration_ms", durationMs)
}

func LogLLMInteraction(provider string, tokensIn int, tokensOut int, costUsd float64) {
	GetLogger().Debug("LLM Interaction Metrics",
		"provider", provider,
		"tokens_in", tokensIn,
		"tokens_out", tokensOut,
		"estimated_cost_usd", costUsd)
}
```

## 4. Integração no Ciclo de Vida do Core

No `main.go` ou no ponto de entrada do daemon (`cmd/vectora/main.go`):

```go
package main

import (
	"log"
	"os"
	"path/filepath"
	"vectora/core/telemetry"
)

func main() {
	homeDir, _ := os.UserHomeDir()
	logDir := filepath.Join(homeDir, ".vectora", "logs")

	if err := telemetry.InitLogger(logDir); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	logger := telemetry.GetLogger()
	logger.Info("Vectora Core starting up...")

	// ... resto da inicialização do daemon ...

	// No shutdown:
	defer func() {
		logger.Info("Vectora Core shutting down.")
		// Fechar writers se necessário (o RotatingWriter deve ter Close)
	}()
}
```

## 5. Estratégia de Logs por Módulo

Para manter a sanidade mental ao ler os logs, cada módulo deve usar atributos consistentes:

| Módulo       | Atributos Chave (Keys)            | Exemplo de Mensagem                       |
| :----------- | :-------------------------------- | :---------------------------------------- |
| **Storage**  | `bucket`, `key`, `operation`      | `"BBolt transaction committed"`           |
| **Tools**    | `tool_name`, `path`, `status`     | `"Tool read_file executed successfully"`  |
| **API/IPC**  | `protocol`, `method`, `client_id` | `"JSON-RPC request received: tools/call"` |
| **Security** | `violation_type`, `blocked_path`  | `"Guardian blocked access to .env"`       |
| **LLM**      | `provider`, `model`, `latency_ms` | `"Gemini response received in 450ms"`     |

---

### Por que esta implementação é superior?

1.  **Zero Dependências Externas:** Usa apenas a stdlib do Go. Isso reduz a superfície de ataque e o tamanho do binário final.
2.  **Segurança de Disco:** O `RotatingWriter` garante que o log nunca cresça indefinidamente, protegendo o usuário de encher o HD com logs de debug.
3.  **Machine-Readable:** O formato JSON permite que ferramentas futuras (ou até o próprio Vectora) analisem seus próprios logs para detectar padrões de erro ou uso indevido.
4.  **Fallback Seguro:** Se o sistema de arquivos estiver corrompido ou inacessível, o fallback para `stderr` garante que erros críticos ainda sejam visíveis para quem iniciou o processo via terminal.
