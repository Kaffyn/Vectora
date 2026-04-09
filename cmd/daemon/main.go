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

	"github.com/Kaffyn/Vectora/core/api/ipc"
	"github.com/Kaffyn/Vectora/core/db"
	"github.com/Kaffyn/Vectora/core/infra"
	"github.com/Kaffyn/Vectora/core/llm"
	vecos "github.com/Kaffyn/Vectora/core/os"
	"github.com/Kaffyn/Vectora/core/policies"
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
		infra.Logger.Warn(fmt.Sprintf("Global config (.env) not found at: %s. Using defaults.", envPath))
	}

	infra.Logger.Info("Starting Vectora Daemon...")

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
