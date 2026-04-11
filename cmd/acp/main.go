// cmd/acp é um binário ACP agent independente para testes e produção.
// Implementa o Agent Client Protocol (ACP) sobre stdio.
// É invocado por clientes IDE (VS Code, Claude Code, etc.)
// e se comunica via JSON-RPC 2.0 em stdin/stdout.
//
// Phase 7B: CLI como ACP Agent - versão melhorada com suporte a múltiplos provedores
package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/Kaffyn/Vectora/core/api/acp"
	"github.com/Kaffyn/Vectora/core/db"
	"github.com/Kaffyn/Vectora/core/llm"
	vecos "github.com/Kaffyn/Vectora/core/os"
)

func main() {
	// Parse flags com opções expandidas para Phase 7B
	verbose := flag.Bool("v", false, "Ativa logging verboso (DEBUG)")
	logFormat := flag.String("log", "text", "Formato do log: text ou json")
	workspace := flag.String("workspace", ".", "Workspace padrão para operações")
	defaultProvider := flag.String("provider", "claude", "Provedor LLM padrão: claude, gemini, ou voyage")
	flag.Parse()

	// Configurar logger em stderr (stdout é reservado para protocolo ACP)
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

	logger.Info("Iniciando Vectora ACP Agent",
		slog.String("version", getVersion()),
		slog.String("workspace", *workspace),
		slog.String("provider", *defaultProvider),
	)

	// Configurar shutdown gracioso
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

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

	// Criar estrutura de diretórios se não existir
	os.MkdirAll(filepath.Join(appDataDir, "run"), 0755)

	dbPath := filepath.Join(appDataDir, "vectora.db")
	indexPath := filepath.Join(appDataDir, "index")

	logger.Debug("Paths de dados",
		slog.String("db", dbPath),
		slog.String("index", indexPath),
	)

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

	// Inicializar roteador LLM e registrar provedores disponíveis
	router := llm.NewRouter()

	// Registrar provedores baseados em variáveis de ambiente
	registrarProvedores(ctx, router, *defaultProvider, logger)

	// Inicializar serviço de mensagens
	msgService := llm.NewMessageService(kvStore)

	// Criar agente ACP
	agent := acp.NewVectoraAgent(
		"vectora-acp",
		getVersion(),
		kvStore,
		vecStore,
		router,
		msgService,
		logger,
	)

	logger.Debug("Agente ACP criado com sucesso")

	// Iniciar agente ACP (bloqueia até cliente desconectar)
	if err := acp.StartACPAgent(ctx, agent, logger); err != nil {
		logger.Error("Erro no agente ACP", slog.Any("error", err))
		os.Exit(1)
	}

	logger.Info("Agente ACP finalizado com sucesso")
}

// registrarProvedores inicializa provedores LLM baseado em variáveis de ambiente
func registrarProvedores(ctx context.Context, router *llm.Router, defaultProvider string, logger *slog.Logger) {
	provedoresRegistrados := []string{}

	// Claude via CLAUDE_API_KEY
	if claudeKey := os.Getenv("CLAUDE_API_KEY"); claudeKey != "" {
		// Aqui registraríamos o provedor Claude quando tiver suporte
		logger.Info("Provedor Claude detectado via CLAUDE_API_KEY")
		provedoresRegistrados = append(provedoresRegistrados, "claude")
	}

	// Gemini via GEMINI_API_KEY
	if geminiKey := os.Getenv("GEMINI_API_KEY"); geminiKey != "" {
		// Aqui registraríamos o provedor Gemini quando tiver suporte
		logger.Info("Provedor Gemini detectado via GEMINI_API_KEY")
		provedoresRegistrados = append(provedoresRegistrados, "gemini")
	}

	// Voyage via VOYAGE_API_KEY (para embeddings)
	if voyageKey := os.Getenv("VOYAGE_API_KEY"); voyageKey != "" {
		logger.Info("Provedor Voyage detectado via VOYAGE_API_KEY")
		provedoresRegistrados = append(provedoresRegistrados, "voyage")
	}

	// Validar provedor padrão
	if !strings.Contains(strings.Join(provedoresRegistrados, ","), defaultProvider) {
		logger.Warn("Provedor padrão não está configurado",
			slog.String("solicitado", defaultProvider),
			slog.String("disponíveis", strings.Join(provedoresRegistrados, ", ")),
		)
	}

	logger.Info("Provedores LLM registrados",
		slog.String("padrão", defaultProvider),
		slog.Int("quantidade", len(provedoresRegistrados)),
	)
}

func getVersion() string {
	// Será preenchido no build com -ldflags
	// Por enquanto, retorna versão padrão
	return "0.1.0-phase7b"
}
