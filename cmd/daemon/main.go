package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/Kaffyn/Vectora/core/api/acp"
	"github.com/Kaffyn/Vectora/core/api/ipc"
	"github.com/Kaffyn/Vectora/core/db"
	"github.com/Kaffyn/Vectora/core/infra"
	"github.com/Kaffyn/Vectora/core/llm"
	vecos "github.com/Kaffyn/Vectora/core/os"
	"github.com/Kaffyn/Vectora/core/policies"
	"github.com/Kaffyn/Vectora/core/tools"
	"github.com/Kaffyn/Vectora/core/tray"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

const version = "0.1.0"

var daemonPort int

var rootCmd = &cobra.Command{
	Use:     "vectora",
	Short:   "Vectora AI - Local Engineering Assistant",
	Long:    "Vectora is an offline-first local AI assistant with knowledge base management.",
	Version: version,
	Run: func(cmd *cobra.Command, args []string) {
		runDaemon()
	},
}

var daemonCmd = &cobra.Command{
	Use:     "daemon",
	Aliases: []string{"start"},
	Short:   "Start background service (Tray)",
	Long:    "Start the Vectora daemon as a background service in the system tray.",
	Run: func(cmd *cobra.Command, args []string) {
		runDaemon()
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Verify health of micro-services",
	Long:  "Check if the Vectora daemon is running and responsive.",
	Run: func(cmd *cobra.Command, args []string) {
		runStatus()
	},
}

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Shutdown background service",
	Long:  "Stop the running Vectora daemon.",
	Run: func(cmd *cobra.Command, args []string) {
		runStop()
	},
}

var acpCmd = &cobra.Command{
	Use:   "acp",
	Short: "Start ACP server over stdio",
	Long: `Start the Vectora ACP (Agent Client Protocol) server over stdin/stdout.
This allows code editors (VS Code, JetBrains, Zed) to connect to Vectora
as an AI coding agent via the ACP protocol.

The server reads JSON-RPC 2.0 messages from stdin and writes responses to stdout.
All logging goes to stderr to avoid interfering with the protocol.

Usage:
  vectora acp              # Start ACP server for current directory
  vectora acp /path/to/proj # Start ACP server for specific workspace
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		workspace := "."
		if len(args) > 0 {
			workspace = args[0]
		}
		return runAcp(workspace)
	},
}

var (
	embedInclude   string
	embedExclude   string
	embedWorkspace string
	embedForce     bool
)

var embedCmd = &cobra.Command{
	Use:   "embed [path]",
	Short: "Embed files into vector store via LLM provider",
	Long: `Embed files into the local vector database using the configured LLM provider's embedding API.
Files are chunked, embedded via Gemini/Claude remote API, and stored locally in chromem-go.

Examples:
  vectora embed                        # Embed current directory
  vectora embed --include *.go,*.md    # Only embed Go and Markdown files
  vectora embed --exclude node_modules # Exclude node_modules directory
  vectora embed --force                # Re-embed already indexed files
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		rootPath := "."
		if len(args) > 0 {
			rootPath = args[0]
		}
		return runEmbed(rootPath)
	},
}

func init() {
	skipElevation := false
	if len(os.Args) > 1 {
		cmd := os.Args[1]
		if cmd == "embed" || cmd == "status" || cmd == "stop" || cmd == "help" || cmd == "-h" || cmd == "--help" || cmd == "--version" || cmd == "-v" {
			skipElevation = true
		}
	}

	if !skipElevation {
		systemManager, _ := vecos.NewManager()
		if systemManager != nil && !systemManager.IsRunningAsAdmin() {
			if err := elevateAdmin(); err == nil {
				os.Exit(0)
			}
		}
	}

	// Embed flags
	embedCmd.Flags().StringVar(&embedInclude, "include", "", "Comma-separated file patterns to include (e.g. *.go,*.md,*.py)")
	embedCmd.Flags().StringVar(&embedExclude, "exclude", "", "Comma-separated patterns to exclude (e.g. node_modules,.git,*.log)")
	embedCmd.Flags().StringVar(&embedWorkspace, "workspace", "default", "Workspace ID for embedding isolation")
	embedCmd.Flags().BoolVar(&embedForce, "force", false, "Re-embed files even if already indexed")

	rootCmd.AddCommand(daemonCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(stopCmd)
	rootCmd.AddCommand(embedCmd)
	rootCmd.AddCommand(acpCmd)

	daemonCmd.Flags().IntVar(&daemonPort, "port", 42780, "Custom daemon port")
	rootCmd.CompletionOptions.DisableDefaultCmd = true
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func getGeminiProvider(ctx context.Context) (llm.Provider, error) {
	cfg := infra.LoadConfig()
	if cfg.GeminiAPIKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY not configured. Set it in %USERPROFILE%\\.Vectora\\.env")
	}
	return llm.NewGeminiProvider(ctx, cfg.GeminiAPIKey)
}

func runEmbed(rootPath string) error {
	ctx := context.Background()

	absPath, err := filepath.Abs(rootPath)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	fmt.Printf("🔍 Scanning: %s\n", absPath)
	fmt.Println()

	// Initialize stores
	kvStore, err := db.NewKVStore()
	if err != nil {
		return fmt.Errorf("failed to init KV store: %w", err)
	}
	defer kvStore.Close()

	vecStore, err := db.NewVectorStore()
	if err != nil {
		return fmt.Errorf("failed to init vector store: %w", err)
	}

	// Get Gemini provider for embeddings
	provider, err := getGeminiProvider(ctx)
	if err != nil {
		return fmt.Errorf("embedding requires Gemini API: %w", err)
	}

	guardian := policies.NewGuardian(absPath)

	fmt.Printf("🤖 Embedding provider: %s\n", provider.Name())
	fmt.Println()

	// Parse include/exclude patterns
	var includePatterns []string
	if embedInclude != "" {
		includePatterns = strings.Split(embedInclude, ",")
		for i, p := range includePatterns {
			includePatterns[i] = strings.TrimSpace(p)
		}
	}

	var excludePatterns []string
	if embedExclude != "" {
		excludePatterns = strings.Split(embedExclude, ",")
		for i, p := range excludePatterns {
			excludePatterns[i] = strings.TrimSpace(p)
		}
	}

	// Collect files
	var filesToEmbed []string
	var filesSkipped int
	var filesAlreadyIndexed int

	err = filepath.WalkDir(absPath, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return nil
		}
		if d.IsDir() {
			// Check exclusions
			name := d.Name()
			for _, pattern := range excludePatterns {
				if matched, _ := filepath.Match(pattern, name); matched {
					return filepath.SkipDir
				}
			}
			// Always skip known large dirs
			if guardian.IsExcludedDir(name) {
				return filepath.SkipDir
			}
			return nil
		}

		// Check Guardian blocks
		if guardian.IsProtected(path) {
			filesSkipped++
			return nil
		}

		// Check include patterns
		if len(includePatterns) > 0 {
			name := d.Name()
			matched := false
			for _, pattern := range includePatterns {
				if m, _ := filepath.Match(pattern, name); m {
					matched = true
					break
				}
			}
			if !matched {
				filesSkipped++
				return nil
			}
		}

		// Check exclude patterns
		for _, pattern := range excludePatterns {
			if m, _ := filepath.Match(pattern, d.Name()); m {
				filesSkipped++
				return nil
			}
		}

		// Check if already indexed
		relPath, _ := filepath.Rel(absPath, path)
		content, readErr := os.ReadFile(path)
		if readErr != nil {
			filesSkipped++
			return nil
		}

		if !embedForce {
			existing, _ := kvStore.Get(ctx, "file_index", relPath)
			if existing != nil {
				// Compare hash to detect changes
				var entry db.FileIndexEntry
				if err := json.Unmarshal(existing, &entry); err == nil {
					contentHash := db.CalculateHash(string(content))
					if entry.ContentHash == contentHash {
						filesAlreadyIndexed++
						return nil
					}
				}
			}
		}

		filesToEmbed = append(filesToEmbed, path)
		return nil
	})
	if err != nil {
		return fmt.Errorf("walk error: %w", err)
	}

	if len(filesToEmbed) == 0 {
		fmt.Printf("✅ No new files to embed. (%d skipped, %d already indexed)\n", filesSkipped, filesAlreadyIndexed)
		return nil
	}

	fmt.Printf("📦 Found %d files to embed (%d skipped, %d already indexed)\n", len(filesToEmbed), filesSkipped, filesAlreadyIndexed)
	fmt.Println()

	// Embed each file
	collectionName := "ws_" + embedWorkspace
	totalEmbedded := 0
	totalChunks := 0
	totalErrors := 0

	for i, filePath := range filesToEmbed {
		relPath, _ := filepath.Abs(filePath)
		if rel, err := filepath.Rel(absPath, filePath); err == nil {
			relPath = rel
		}

		content, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Printf("  ❌ %s: read error\n", relPath)
			totalErrors++
			continue
		}

		// Chunk the content directly
		chunks := chunkContent(string(content), 800, 100)
		fileChunks := 0

		// Detect language from extension
		language := "text"
		ext := strings.ToLower(filepath.Ext(filePath))
		switch ext {
		case ".go":
			language = "go"
		case ".py":
			language = "python"
		case ".js", ".ts":
			language = "javascript"
		case ".md":
			language = "markdown"
		}

		for j, chunk := range chunks {
			// Generate embedding via Gemini API
			vec, err := provider.Embed(ctx, chunk)
			if err != nil {
				fmt.Printf("  ❌ %s[%d]: embed error: %v\n", relPath, j, err)
				totalErrors++
				break
			}

			docID := fmt.Sprintf("%s:%d", relPath, j)
			err = vecStore.UpsertChunk(ctx, collectionName, db.Chunk{
				ID:       docID,
				Content:  chunk,
				Metadata: map[string]string{"source": relPath, "filename": filepath.Base(filePath), "language": language},
				Vector:   vec,
			})
			if err != nil {
				fmt.Printf("  ❌ %s[%d]: store error: %v\n", relPath, j, err)
				totalErrors++
				break
			}
			fileChunks++
		}

		if fileChunks > 0 {
			// Save index metadata
			contentHash := db.CalculateHash(string(content))
			entry := db.FileIndexEntry{
				AbsolutePath: filePath,
				ContentHash:  contentHash,
				SizeBytes:    int64(len(content)),
			}
			entryBytes, _ := json.Marshal(entry)
			kvStore.Set(ctx, "file_index", relPath, entryBytes)

			totalEmbedded++
			totalChunks += fileChunks
			fmt.Printf("  ✅ %s → %d chunks\n", relPath, fileChunks)
		}

		// Progress indicator
		if (i+1)%10 == 0 || i+1 == len(filesToEmbed) {
			fmt.Printf("  ... %d/%d files processed\n", i+1, len(filesToEmbed))
		}
	}

	fmt.Println()
	fmt.Println("==========================================================")
	fmt.Printf("  Embedding complete!\n")
	fmt.Printf("  ✅ %d files embedded\n", totalEmbedded)
	fmt.Printf("  📊 %d total chunks\n", totalChunks)
	fmt.Printf("  ❌ %d errors\n", totalErrors)
	fmt.Printf("  📁 Workspace: %s\n", collectionName)
	fmt.Println("==========================================================")

	return nil
}

func chunkContent(content string, maxTokens int, overlap int) []string {
	if len(content) == 0 {
		return []string{}
	}

	// ~4 chars per token approximation
	maxChars := maxTokens * 4
	overlapChars := overlap * 4

	runes := []rune(content)
	totalRunes := len(runes)

	if maxChars >= totalRunes {
		return []string{content}
	}

	var chunks []string
	start := 0
	for start < totalRunes {
		end := start + maxChars
		if end > totalRunes {
			end = totalRunes
		}

		// Try to break at newline
		chunkEnd := end
		if end < totalRunes {
			for i := end; i < totalRunes && i < end+maxChars/2; i++ {
				if runes[i] == '\n' {
					chunkEnd = i + 1
					break
				}
			}
		}

		chunks = append(chunks, string(runes[start:chunkEnd]))
		start = chunkEnd - overlapChars
		if start < 0 {
			start = 0
		}
		if start >= chunkEnd {
			break
		}
	}

	return chunks
}

func runDaemon() {
	systemManager, err := vecos.NewManager()
	if err != nil {
		log.Fatalf("Critical Hardware OS Failure: %v", err)
	}

	if err := systemManager.EnforceSingleInstance(); err != nil {
		infra.NotifyOS("Vectora Running", "Vectora is already running in the system tray.")
		os.Exit(0)
	}

	infra.SetupLogger()

	appDataDir, _ := systemManager.GetAppDataDir()
	envPath := filepath.Join(appDataDir, ".env")
	if err := godotenv.Load(envPath); err != nil {
		infra.Logger().Warn(fmt.Sprintf("Global config (.env) not found at: %s. Using defaults.", envPath))
	}

	infra.Logger().Info("Starting Vectora Daemon...")

	kvStore, _ := db.NewKVStore()
	vecStore, _ := db.NewVectorStore()
	appCtx := context.Background()
	msgService := llm.NewMessageService(kvStore)
	memService, _ := db.NewMemoryService(appCtx, filepath.Join(appDataDir, "data", "memory"))

	ipcServer, _ := ipc.NewServer()

	// Provider fetcher for IPC routes
	getProvider := func() llm.Provider {
		return tray.ActiveProvider
	}

	ipc.RegisterRoutes(ipcServer, kvStore, vecStore, getProvider, msgService, memService)
	go ipcServer.Start()

	// Dev HTTP bridge for debugging
	go ipcServer.StartDevHTTP(daemonPort)

	infra.NotifyOS("Vectora", "Operational Assistant.")
	tray.Setup()
}

func runStatus() {
	fmt.Println("--- VECTORA STATUS ---")
	client, err := ipc.NewClient()
	if err != nil || client.Connect() != nil {
		fmt.Println("Status: OFFLINE (Daemon not found)")
		return
	}
	fmt.Println("Status: ONLINE (Daemon active)")
}

func runStop() {
	if os.Getenv("OS") == "Windows_NT" {
		exec.Command("taskkill", "/F", "/IM", "vectora.exe").Run()
		fmt.Println("Vectora terminated.")
	}
}

// ---- ACP Server ----

// acpEngine implements acp.Engine using Vectora's core components.
type acpEngine struct {
	provider llm.Provider
	vecStore *db.ChromemStore
	kvStore  *db.BBoltStore
	tools    *tools.Registry
	guardian *policies.Guardian
	cwd      string
}

func (e *acpEngine) Embed(ctx context.Context, text string) ([]float32, error) {
	if e.provider == nil {
		return nil, fmt.Errorf("no LLM provider configured")
	}
	return e.provider.Embed(ctx, text)
}

func (e *acpEngine) Query(ctx context.Context, query string, workspaceID string) (string, error) {
	if e.provider == nil {
		return fmt.Sprintf("No LLM provider configured. Query: %s", query), nil
	}

	// RAG: embed query, search vectors, build context, complete
	vector, err := e.provider.Embed(ctx, query)
	if err != nil {
		return e.completeWithProvider(ctx, query, "")
	}

	chunks, err := e.vecStore.Query(ctx, "ws_"+workspaceID, vector, 5)
	if err != nil {
		chunks = []db.ScoredChunk{}
	}

	contextText := ""
	for _, chunk := range chunks {
		if filename, ok := chunk.Metadata["filename"]; ok {
			contextText += "File: " + filename + "\n"
		}
		contextText += chunk.Content + "\n---\n"
	}

	return e.completeWithProvider(ctx, query, contextText)
}

func (e *acpEngine) completeWithProvider(ctx context.Context, query string, contextText string) (string, error) {
	systemPrompt := "You are Vectora, an AI coding assistant."
	if contextText != "" {
		systemPrompt += "\n\nUse the following context as your source of truth:\n" + contextText
	}

	messages := []llm.Message{
		{Role: llm.RoleSystem, Content: systemPrompt},
		{Role: llm.RoleUser, Content: query},
	}

	resp, err := e.provider.Complete(ctx, llm.CompletionRequest{
		Messages:    messages,
		MaxTokens:   2000,
		Temperature: 0.2,
	})
	if err != nil {
		return "", fmt.Errorf("LLM error: %w", err)
	}

	return resp.Content, nil
}

func (e *acpEngine) ExecuteTool(ctx context.Context, name string, args map[string]any) (acp.ToolResult, error) {
	tool, ok := e.tools.GetTool(name)
	if !ok {
		return acp.ToolResult{Output: fmt.Sprintf("Tool '%s' not found", name), IsError: true}, nil
	}

	argsJSON, _ := json.Marshal(args)
	result, err := tool.Execute(ctx, argsJSON)
	if err != nil {
		return acp.ToolResult{Output: err.Error(), IsError: true}, nil
	}

	return acp.ToolResult{Output: result.Output, IsError: result.IsError}, nil
}

func (e *acpEngine) ReadFile(ctx context.Context, path string) (string, error) {
	tool, _ := e.tools.GetTool("read_file")
	if tool == nil {
		return "", fmt.Errorf("read_file tool not available")
	}

	argsJSON, _ := json.Marshal(map[string]string{"path": path})
	result, err := tool.Execute(ctx, argsJSON)
	if err != nil {
		return "", err
	}
	if result.IsError {
		return "", fmt.Errorf(result.Output)
	}
	return result.Output, nil
}

func (e *acpEngine) WriteFile(ctx context.Context, path, content string) error {
	tool, _ := e.tools.GetTool("write_file")
	if tool == nil {
		return fmt.Errorf("write_file tool not available")
	}

	argsJSON, _ := json.Marshal(map[string]string{"path": path, "content": content})
	result, err := tool.Execute(ctx, argsJSON)
	if err != nil {
		return err
	}
	if result.IsError {
		return fmt.Errorf(result.Output)
	}
	return nil
}

func (e *acpEngine) RunCommand(ctx context.Context, cwd, command string) (string, error) {
	tool, _ := e.tools.GetTool("run_shell_command")
	if tool == nil {
		return "", fmt.Errorf("run_shell_command tool not available")
	}

	argsJSON, _ := json.Marshal(map[string]string{"command": command})
	result, err := tool.Execute(ctx, argsJSON)
	if err != nil {
		return "", err
	}
	if result.IsError {
		return result.Output, fmt.Errorf("command failed")
	}
	return result.Output, nil
}

func runAcp(workspace string) error {
	ctx := context.Background()

	absPath, err := filepath.Abs(workspace)
	if err != nil {
		return fmt.Errorf("invalid workspace path: %w", err)
	}

	// Log to stderr so it doesn't interfere with stdio protocol
	fmt.Fprintln(os.Stderr, "🚀 Vectora ACP Server starting...")
	fmt.Fprintln(os.Stderr, "📁 Workspace:", absPath)

	// Load config
	cfg := infra.LoadConfig()
	if cfg.GeminiAPIKey == "" && cfg.ClaudeAPIKey == "" {
		fmt.Fprintln(os.Stderr, "⚠️  No API key configured. Set GEMINI_API_KEY or CLAUDE_API_KEY in .env")
	}

	// Initialize stores
	kvStore, err := db.NewKVStore()
	if err != nil {
		return fmt.Errorf("failed to init KV store: %w", err)
	}
	defer kvStore.Close()

	vecStore, err := db.NewVectorStore()
	if err != nil {
		return fmt.Errorf("failed to init vector store: %w", err)
	}

	// Initialize provider (prefer Gemini, fallback to Claude)
	var provider llm.Provider
	if cfg.GeminiAPIKey != "" {
		provider, err = llm.NewGeminiProvider(ctx, cfg.GeminiAPIKey)
		if err != nil {
			fmt.Fprintf(os.Stderr, "⚠️  Failed to init Gemini: %v\n", err)
		}
	}
	if provider == nil && cfg.ClaudeAPIKey != "" {
		provider, err = llm.NewClaudeProvider(ctx, cfg.ClaudeAPIKey)
		if err != nil {
			fmt.Fprintf(os.Stderr, "⚠️  Failed to init Claude: %v\n", err)
		}
	}

	// Initialize tools
	guardian := policies.NewGuardian(absPath)
	toolRegistry := tools.NewRegistry(absPath, guardian, kvStore)

	// Create engine
	engine := &acpEngine{
		provider: provider,
		vecStore: vecStore,
		kvStore:  kvStore,
		tools:    toolRegistry,
		guardian: guardian,
		cwd:      absPath,
	}

	if provider != nil {
		fmt.Fprintln(os.Stderr, "🤖 Provider:", provider.Name())
	} else {
		fmt.Fprintln(os.Stderr, "⚠️  No LLM provider — will operate in tool-only mode")
	}

	fmt.Fprintln(os.Stderr, "📡 Listening on stdio (JSON-RPC 2.0)")
	fmt.Fprintln(os.Stderr, "========================================")

	// Run ACP server
	server := acp.NewServer(engine)
	return server.Run(ctx)
}
