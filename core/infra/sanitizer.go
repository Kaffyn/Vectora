package infra

import (
	"context"
	"log/slog"
	"regexp"
)

// SanitizingHandler wraps an slog.Handler and redacts sensitive data
// (API keys, tokens, passwords, PII) from all log records.
type SanitizingHandler struct {
	handler  slog.Handler
	patterns []*regexp.Regexp
}

// NewSanitizingHandler creates a handler that wraps the given slog.Handler.
func NewSanitizingHandler(handler slog.Handler) *SanitizingHandler {
	return &SanitizingHandler{
		handler: handler,
		patterns: []*regexp.Regexp{
			// API key patterns: sk_*, sk-*, gsk_*, etc.
			regexp.MustCompile(`(?i)(sk[_-][a-zA-Z0-9]{20,}|sk[_-][a-zA-Z0-9]{48,})`),
			// Bearer tokens in Authorization headers
			regexp.MustCompile(`(?i)bearer\s+[a-zA-Z0-9_.-]+`),
			// UUID-like tokens (common in some APIs)
			regexp.MustCompile(`[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}`),
			// Generic API_KEY or secret patterns
			regexp.MustCompile(`(?i)(api[_-]?key|secret|password|token|auth)\s*[:=]\s*[^\s"']+`),
		},
	}
}

// Enabled reports whether the handler handles records at the given level.
func (h *SanitizingHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

// Handle sanitizes the record and forwards it to the wrapped handler.
func (h *SanitizingHandler) Handle(ctx context.Context, record slog.Record) error {
	// Sanitize the message
	record.Message = h.sanitize(record.Message)

	// Sanitize all attributes
	record.Attrs(func(attr slog.Attr) bool {
		attr.Value = slog.StringValue(h.sanitize(attr.Value.String()))
		return true
	})

	return h.handler.Handle(ctx, record)
}

// WithAttrs returns a new handler with additional attributes.
func (h *SanitizingHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &SanitizingHandler{
		handler:  h.handler.WithAttrs(attrs),
		patterns: h.patterns,
	}
}

// WithGroup returns a new handler with the group name set.
func (h *SanitizingHandler) WithGroup(name string) slog.Handler {
	return &SanitizingHandler{
		handler:  h.handler.WithGroup(name),
		patterns: h.patterns,
	}
}

// sanitize redacts sensitive data from a string.
func (h *SanitizingHandler) sanitize(s string) string {
	for _, pattern := range h.patterns {
		s = pattern.ReplaceAllString(s, "[REDACTED]")
	}

	// Additional patterns for common PII/sensitive strings
	s = h.redactEmailAddresses(s)
	s = h.redactIPAddresses(s)

	return s
}

// redactEmailAddresses replaces email addresses with [REDACTED].
func (h *SanitizingHandler) redactEmailAddresses(s string) string {
	return regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`).
		ReplaceAllString(s, "[REDACTED-EMAIL]")
}

// redactIPAddresses replaces private IP addresses with [REDACTED].
func (h *SanitizingHandler) redactIPAddresses(s string) string {
	// Redact common private IP patterns (10.*, 192.168.*, 172.16.*)
	return regexp.MustCompile(`(?:10\.\d{1,3}|192\.168\.\d{1,3}|172\.(?:1[6-9]|2\d|3[01])\.\d{1,3})\.\d{1,3}`).
		ReplaceAllString(s, "[REDACTED-IP]")
}
