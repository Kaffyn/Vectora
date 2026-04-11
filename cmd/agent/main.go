// Package main implementa o binário Vectora ACP Agent.
// É invocado por clientes IDE (VS Code, Claude Code, Antigravity, etc.)
// para fornecer funcionalidade de agente via Agent Client Protocol (ACP).
//
// Diferente do CLI `vectora` interativo, `vectora-agent` roda como subprocess
// e se comunica via JSON-RPC em stdin/stdout com o cliente IDE.
//
// Phase 7B: CLI como ACP Agent - versão principal
package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/Kaffyn/Vectora/core/api/acp"
	"github.com/Kaffyn/Vectora/core/db"
	"github.com/Kaffyn/Vectora/core/llm"
	vecos "github.com/Kaffyn/Vectora/core/os"
)

func main() {
	// Parse flags
	verbose := flag.Bool("v", false, "Ativa logging verboso (DEBUG)")
	logFormat := flag.String("log", "text", "Formato do log: text ou json")
	workspace := flag.String("workspace", ".", "Workspace padrão para operações")
	flag.Parse()

	// Configurar logger
	logLevel := slog.LevelInfo
	if *verbose {
		logLevel = slog.LevelDebug
	}

	var logHandler slog.Handler
	if *logFormat == "json" {
		logHandler = slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
			Level: logLevel,
		})
	} else {
		logHandler = slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: logLevel,
		})
	}
	logger := slog.New(logHandler)

	logger.Info("Iniciando Vectora Agent (ACP)",
		slog.String("version", getVersion()),
		slog.String("workspace", *workspace),
	)

	// Inicializar gerenciador de SO
	osMgr, err := vecos.NewManager()
	if err != nil {
		logger.Error("Falha ao inicializar gerenciador de SO", slog.Any("error", err))
		os.Exit(1)
	}

	// Configurar diretórios de dados
	appDataDir, err := osMgr.GetAppDataDir()
	if err != nil {
		logger.Error("Falha ao obter diretório app data", slog.Any("error", err))
		os.Exit(1)
	}

	// Criar estrutura de diretórios
	os.MkdirAll(filepath.Join(appDataDir, "run"), 0755)

	dbPath := filepath.Join(appDataDir, "vectora.db")
	indexPath := filepath.Join(appDataDir, "index")

	// Inicializar armazenamento de dados
	kvStore, err := db.NewKVStoreAtPath(dbPath)
	if err != nil {
		logger.Error("Falha ao inicializar KV store", slog.Any("error", err))
		os.Exit(1)
	}
	defer kvStore.Close()

	vecStore, err := db.NewVectorStoreAtPath(indexPath)
	if err != nil {
		logger.Error("Falha ao inicializar vector store", slog.Any("error", err))
		os.Exit(1)
	}

	// Inicializar roteador LLM
	router := llm.NewRouter()

	// Inicializar serviço de mensagens
	msgService := llm.NewMessageService(kvStore)

	// Criar agente ACP
	agent := acp.NewVectoraAgent(
		"vectora-agent",
		getVersion(),
		kvStore,
		vecStore,
		router,
		msgService,
		logger,
	)

	logger.Debug("Agente ACP criado com sucesso")

	// Configurar shutdown gracioso
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		logger.Info("Sinal recebido, encerrando", slog.String("signal", sig.String()))
		cancel()
	}()

	// Iniciar agente ACP (bloqueia até cliente desconectar)
	if err := acp.StartACPAgent(ctx, agent, logger); err != nil {
		logger.Error("Erro no agente ACP", slog.Any("error", err))
		os.Exit(1)
	}

	logger.Info("Agente ACP finalizado com sucesso")
}

func getVersion() string {
	// Será preenchido no build com -ldflags
	return "0.1.0-phase7b"
}
