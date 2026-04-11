package acp

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/Kaffyn/Vectora/core/llm"
	"github.com/coder/acp-go-sdk"
)

// TestNewVectoraAgent testa criação de novo agente ACP
func TestNewVectoraAgent(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	// Criar agente
	agent := NewVectoraAgent(
		"test-agent",
		"0.1.0",
		nil,
		nil,
		llm.NewRouter(),
		nil,
		logger,
	)

	if agent == nil {
		t.Error("NewVectoraAgent retornou nil")
	}
}

// TestVectoraAgentInitialize testa inicialização do agente ACP
func TestVectoraAgentInitialize(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	agent := NewVectoraAgent(
		"test-agent",
		"0.1.0",
		nil,
		nil,
		llm.NewRouter(),
		nil,
		logger,
	)

	ctx := context.Background()
	resp, err := agent.Initialize(ctx, acp.InitializeRequest{
		ProtocolVersion: 1,
		ClientInfo: &acp.Implementation{
			Name:    "test-client",
			Version: "1.0.0",
		},
	})

	if err != nil {
		t.Errorf("Initialize falhou: %v", err)
	}

	if resp.AgentInfo == nil {
		t.Error("AgentInfo está nil")
	}

	if resp.AgentInfo.Name != "test-agent" {
		t.Errorf("Nome do agente incorreto: %s", resp.AgentInfo.Name)
	}
}

// TestVectoraAgentNewSession testa criação de nova sessão
func TestVectoraAgentNewSession(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	agent := NewVectoraAgent(
		"test-agent",
		"0.1.0",
		nil,
		nil,
		llm.NewRouter(),
		nil,
		logger,
	)

	ctx := context.Background()
	resp, err := agent.NewSession(ctx, acp.NewSessionRequest{})

	if err != nil {
		t.Errorf("NewSession falhou: %v", err)
	}

	if string(resp.SessionId) == "" {
		t.Error("SessionId está vazio")
	}
}

// TestVectoraAgentCancel testa cancelamento de sessão
func TestVectoraAgentCancel(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	agent := NewVectoraAgent(
		"test-agent",
		"0.1.0",
		nil,
		nil,
		llm.NewRouter(),
		nil,
		logger,
	)

	ctx := context.Background()

	// Criar sessão
	sessionResp, err := agent.NewSession(ctx, acp.NewSessionRequest{})
	if err != nil {
		t.Fatalf("Falha ao criar sessão: %v", err)
	}

	// Cancelar sessão
	err = agent.Cancel(ctx, acp.CancelNotification{
		SessionId: sessionResp.SessionId,
	})

	if err != nil {
		t.Errorf("Cancel falhou: %v", err)
	}
}

// BenchmarkNewSession benchmark para criar nova sessão
func BenchmarkNewSession(b *testing.B) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	agent := NewVectoraAgent(
		"bench-agent",
		"0.1.0",
		nil,
		nil,
		llm.NewRouter(),
		nil,
		logger,
	)

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = agent.NewSession(ctx, acp.NewSessionRequest{})
	}
}
