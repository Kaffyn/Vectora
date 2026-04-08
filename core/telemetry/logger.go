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
func ContextLogger(_ context.Context, requestID string) *slog.Logger {
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
