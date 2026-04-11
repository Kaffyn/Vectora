package telemetry

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
)

var GlobalLogger *slog.Logger

// InitLogger configura o logger global do Vectora Core.
// Deve ser chamado uma única vez no startup do core.
func InitLogger(logDir string) error {
	logPath := filepath.Join(logDir, "vectora-core.log")

	writer, err := NewRotatingWriter(logPath, DefaultMaxSize)
	if err != nil {
		return err
	}

	// Handler JSON para facilitar parsing por ferramentas externas ou grep
	jsonHandler := slog.NewJSONHandler(writer, &slog.HandlerOptions{
		Level: slog.LevelDebug, // Captura tudo, filtramos via output se necessário
	})

	// Wrap with sanitizing handler to redact API keys, tokens, and PII
	sanitizingHandler := newSanitizingHandler(jsonHandler)

	GlobalLogger = slog.New(sanitizingHandler)

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

// sanitizingHandler wraps slog.Handler to redact sensitive data from logs.
type sanitizingHandler struct {
	handler  slog.Handler
	patterns []*regexp.Regexp
}

// newSanitizingHandler creates a handler that redacts API keys, tokens, and PII.
func newSanitizingHandler(handler slog.Handler) slog.Handler {
	return &sanitizingHandler{
		handler: handler,
		patterns: []*regexp.Regexp{
			// API key patterns
			regexp.MustCompile(`(?i)(sk[_-][a-zA-Z0-9]{20,}|sk[_-][a-zA-Z0-9]{48,})`),
			// Bearer tokens
			regexp.MustCompile(`(?i)bearer\s+[a-zA-Z0-9_.-]+`),
			// Generic secrets
			regexp.MustCompile(`(?i)(api[_-]?key|secret|password|token|auth)\s*[:=]\s*[^\s"']+`),
		},
	}
}

func (h *sanitizingHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h *sanitizingHandler) Handle(ctx context.Context, record slog.Record) error {
	record.Message = h.sanitize(record.Message)
	record.Attrs(func(attr slog.Attr) bool {
		attr.Value = slog.StringValue(h.sanitize(attr.Value.String()))
		return true
	})
	return h.handler.Handle(ctx, record)
}

func (h *sanitizingHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &sanitizingHandler{
		handler:  h.handler.WithAttrs(attrs),
		patterns: h.patterns,
	}
}

func (h *sanitizingHandler) WithGroup(name string) slog.Handler {
	return &sanitizingHandler{
		handler:  h.handler.WithGroup(name),
		patterns: h.patterns,
	}
}

func (h *sanitizingHandler) sanitize(s string) string {
	for _, pattern := range h.patterns {
		s = pattern.ReplaceAllString(s, "[REDACTED]")
	}
	return s
}
