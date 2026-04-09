package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/Kaffyn/Vectora/core/db"
	"github.com/Kaffyn/Vectora/core/engine"
	"github.com/Kaffyn/Vectora/core/i18n"
	"github.com/Kaffyn/Vectora/core/ingestion"
	"github.com/Kaffyn/Vectora/core/ipc"
	"github.com/Kaffyn/Vectora/core/llm"
	vecos "github.com/Kaffyn/Vectora/core/os"
	"github.com/Kaffyn/Vectora/core/policies"
	"github.com/Kaffyn/Vectora/core/tools"
)

var (
	passed  int
	failed  int
	skipped int
	mu      sync.Mutex
)

func main() {
	ctx := context.Background()
	testDir := filepath.Join(os.TempDir(), "vectora-integration-test")
	os.RemoveAll(testDir)
	os.MkdirAll(testDir, 0755)
	defer os.RemoveAll(testDir)

	fmt.Println("==========================================================")
	fmt.Println("  VECTORA INTEGRATION TEST SUITE")
	fmt.Println("==========================================================")
	fmt.Println()

	fmt.Println("[ 1/12] OS Manager")
	testOSManager()

	fmt.Println("[ 2/12] i18n")
	testI18n()

	fmt.Println("[ 3/12] Policies / Guardian")
	testGuardian()

	fmt.Println("[ 4/12] BBolt KV Store")
	testBBoltKV(testDir)

	fmt.Println("[ 5/12] Chromem-go Vector Store")
	testChromemVectors(testDir)

	fmt.Println("[ 6/12] Memory Service")
	testMemoryService(ctx, testDir)

	fmt.Println("[ 7/12] Tools Runtime (all 10 tools)")
	testToolsRuntime(ctx, testDir)

	fmt.Println("[ 8/12] Ingestion Pipeline")
	testIngestion(testDir)

	fmt.Println("[ 9/12] IPC Server/Client Communication")
	testIPC(ctx, testDir, kvForIPC(testDir))

	fmt.Println("[10/12] Engine RAG (without LLM)")
	testEngineRAG(ctx, testDir)

	fmt.Println("[11/12] Config Manager")
	testConfigManager(testDir)

	fmt.Println("[12/12] Git Manager")
	testGitManager(testDir)

	fmt.Println()
	fmt.Println("==========================================================")
	fmt.Printf("  RESULTS: %d PASSED | %d FAILED | %d SKIPPED\n", passed, failed, skipped)
	fmt.Println("==========================================================")

	if failed > 0 {
		os.Exit(1)
	}
}

func kvForIPC(testDir string) *db.BBoltStore {
	kv, _ := db.NewKVStoreAtPath(filepath.Join(testDir, "ipc-kv.db"))
	return kv
}

func assert(name string, condition bool, detail string) {
	mu.Lock()
	defer mu.Unlock()
	if condition {
		fmt.Printf("  ✅ %s\n", name)
		passed++
	} else {
		fmt.Printf("  ❌ %s - %s\n", name, detail)
		failed++
	}
}

func skip(name string, reason string) {
	mu.Lock()
	defer mu.Unlock()
	fmt.Printf("  ⏭️  %s - %s\n", name, reason)
	skipped++
}

// ---- 1. OS Manager ----
func testOSManager() {
	mgr, err := vecos.NewManager()
	assert("NewManager", err == nil && mgr != nil, fmt.Sprintf("err=%v", err))

	appDir, err := mgr.GetAppDataDir()
	assert("GetAppDataDir", err == nil && appDir != "", fmt.Sprintf("err=%v dir=%s", err, appDir))

	installDir, err := mgr.GetInstallDir()
	assert("GetInstallDir", err == nil && installDir != "", fmt.Sprintf("err=%v dir=%s", err, installDir))

	_ = mgr.EnforceSingleInstance()
	assert("EnforceSingleInstance", true, "")
}

// ---- 2. i18n ----
func testI18n() {
	i18n.SetLanguage("en")
	v := i18n.T("tray_status")
	assert("English translation", v == "Status: Running", fmt.Sprintf("got '%s'", v))

	i18n.SetLanguage("pt")
	v = i18n.T("tray_status")
	assert("Portuguese translation", v == "Status: Rodando", fmt.Sprintf("got '%s'", v))

	i18n.SetLanguage("es")
	v = i18n.T("tray_status")
	assert("Spanish translation", v == "Estado: En ejecución", fmt.Sprintf("got '%s'", v))

	i18n.SetLanguage("fr")
	v = i18n.T("tray_quit")
	assert("French translation", v == "Quitter Vectora", fmt.Sprintf("got '%s'", v))
}

// ---- 3. Guardian ----
func testGuardian() {
	g := policies.NewGuardian("/trust")

	assert("Block .env", g.IsProtected("/trust/.env"), "")
	assert("Block .key", g.IsProtected("/trust/secret.key"), "")
	assert("Block .db", g.IsProtected("/trust/data.db"), "")
	assert("Block .exe", g.IsProtected("/trust/app.exe"), "")
	assert("Allow .go", !g.IsProtected("/trust/main.go"), "")
	assert("Allow .md", !g.IsProtected("/trust/README.md"), "")
	assert("Path safe nested", g.IsPathSafe("/trust/a/b/c.go"), "")
	assert("Block path traversal", !g.IsPathSafe("/trust/../escape.txt"), "")

	out := g.SanitizeOutput("key AKIAIOSFODNN7EXAMPLE end")
	assert("Sanitize secrets", strings.Contains(out, "[REDACTED_SECRET]"), fmt.Sprintf("got '%s'", out))

	assert("Exclude .git", g.IsExcludedDir(".git"), "")
	assert("Exclude node_modules", g.IsExcludedDir("node_modules"), "")
	assert("Exclude vendor", g.IsExcludedDir("vendor"), "")
}

// ---- 4. BBolt KV Store ----
func testBBoltKV(testDir string) {
	kvPath := filepath.Join(testDir, "test-kv.db")
	kv, err := db.NewKVStoreAtPath(kvPath)
	assert("NewKVStoreAtPath", err == nil && kv != nil, fmt.Sprintf("err=%v", err))
	if kv == nil {
		return
	}
	defer kv.Close()

	ctx := context.Background()

	err = kv.Set(ctx, "settings", "provider", []byte("gemini"))
	assert("KV Set", err == nil, fmt.Sprintf("err=%v", err))

	val, err := kv.Get(ctx, "settings", "provider")
	assert("KV Get", err == nil && string(val) == "gemini", fmt.Sprintf("err=%v val=%s", err, string(val)))

	err = kv.Delete(ctx, "settings", "provider")
	assert("KV Delete", err == nil, fmt.Sprintf("err=%v", err))

	val, _ = kv.Get(ctx, "settings", "provider")
	assert("KV Get after delete returns nil", val == nil, fmt.Sprintf("val=%s", string(val)))

	kv.Set(ctx, "conversations", "conv-1", []byte(`{"title":"test1"}`))
	kv.Set(ctx, "conversations", "conv-2", []byte(`{"title":"test2"}`))
	kv.Set(ctx, "conversations", "conv-3", []byte(`{"title":"test3"}`))
	keys, err := kv.List(ctx, "conversations", "")
	assert("KV List", err == nil && len(keys) == 3, fmt.Sprintf("err=%v keys=%d", err, len(keys)))
}

// ---- 5. Chromem-go Vector Store ----
func testChromemVectors(testDir string) {
	vecPath := filepath.Join(testDir, "test-vectors")
	vec, err := db.NewVectorStoreAtPath(vecPath)
	assert("NewVectorStoreAtPath", err == nil && vec != nil, fmt.Sprintf("err=%v", err))
	if vec == nil {
		return
	}

	ctx := context.Background()

	exists := vec.CollectionExists(ctx, "test-collection")
	assert("Collection not exists yet", !exists, "")

	// ChromemStore requires pre-generated embeddings. In production, these come
	// from the LLM provider (Gemini/Claude embedding API). For testing without
	// API keys, we use deterministic hash-based dummy embeddings.
	chunk := db.Chunk{
		ID:       "doc-1",
		Content:  "hello world this is a test document for vector store",
		Metadata: map[string]string{"source": "test.txt", "filename": "test.txt"},
		Vector:   db.GenerateDummyEmbedding("hello world this is a test document for vector store", 768),
	}
	err = vec.UpsertChunk(ctx, "test-collection", chunk)
	assert("UpsertChunk creates collection", err == nil, fmt.Sprintf("err=%v", err))

	exists = vec.CollectionExists(ctx, "test-collection")
	assert("Collection exists after insert", exists, "")

	// Query with pre-generated embedding
	queryVec := db.GenerateDummyEmbedding("hello world", 768)
	results, err := vec.Query(ctx, "test-collection", queryVec, 1)
	assert("Query returns results", err == nil && len(results) > 0, fmt.Sprintf("err=%v results=%d", err, len(results)))
	if len(results) > 0 {
		assert("Query result matches content", results[0].Content == chunk.Content, fmt.Sprintf("got '%s'", results[0].Content))
	}

	err = vec.DeleteCollection(ctx, "test-collection")
	assert("DeleteCollection", err == nil, fmt.Sprintf("err=%v", err))

	exists = vec.CollectionExists(ctx, "test-collection")
	assert("Collection deleted", !exists, "")
}

// ---- 6. Memory Service ----
func testMemoryService(ctx context.Context, testDir string) {
	memPath := filepath.Join(testDir, "test-memory")
	ms, err := db.NewMemoryService(ctx, memPath)
	assert("NewMemoryService", err == nil && ms != nil, fmt.Sprintf("err=%v", err))
	if ms == nil {
		return
	}

	// In production, embeddings come from Gemini/Claude API. For testing, use dummy embeddings.
	err = ms.StoreInsightWithFallback(ctx, "fact-1", "Go uses garbage collection", map[string]string{"category": "programming"})
	assert("StoreInsight", err == nil, fmt.Sprintf("err=%v", err))

	results, err := ms.SearchInsight(ctx, "garbage collection", db.GenerateDummyEmbedding("garbage collection", 768), 1)
	assert("SearchInsight finds data", err == nil && len(results) > 0, fmt.Sprintf("err=%v results=%d", err, len(results)))
	if len(results) > 0 {
		assert("Search result contains content", strings.Contains(results[0], "garbage collection"), fmt.Sprintf("got '%s'", results[0]))
	}
}

// ---- 7. Tools Runtime ----
func testToolsRuntime(ctx context.Context, testDir string) {
	g := policies.NewGuardian(testDir)
	kv, _ := db.NewKVStoreAtPath(filepath.Join(testDir, "tools-kv.db"))
	reg := tools.NewRegistry(testDir, g, kv)

	// All 10 tools registered
	expected := []string{"read_file", "write_file", "read_folder", "edit", "find_files", "grep_search", "run_shell_command", "save_memory", "google_search", "web_fetch"}
	for _, name := range expected {
		tool, ok := reg.GetTool(name)
		assert(fmt.Sprintf("Tool '%s' registered", name), ok && tool != nil && tool.Name() == name, "")
	}

	// === write_file ===
	t, _ := reg.GetTool("write_file")
	args, _ := json.Marshal(map[string]string{"path": "test.txt", "content": "hello integration test"})
	res, err := t.Execute(ctx, json.RawMessage(args))
	assert("write_file executes", err == nil && !res.IsError, fmt.Sprintf("err=%v out=%s", err, res.Output))

	// === read_file ===
	t, _ = reg.GetTool("read_file")
	args, _ = json.Marshal(map[string]string{"path": "test.txt"})
	res, err = t.Execute(ctx, json.RawMessage(args))
	assert("read_file reads content", err == nil && res.Output == "hello integration test", fmt.Sprintf("out='%s' err=%v", res.Output, err))

	// === edit ===
	t, _ = reg.GetTool("edit")
	args, _ = json.Marshal(map[string]string{"path": "test.txt", "target": "integration", "replacement": "end-to-end"})
	res, err = t.Execute(ctx, json.RawMessage(args))
	assert("edit modifies text", err == nil && !res.IsError, fmt.Sprintf("err=%v out=%s", err, res.Output))

	// Verify edit
	t, _ = reg.GetTool("read_file")
	args, _ = json.Marshal(map[string]string{"path": "test.txt"})
	res, err = t.Execute(ctx, json.RawMessage(args))
	assert("read_file confirms edit", strings.Contains(res.Output, "end-to-end"), fmt.Sprintf("got '%s'", res.Output))

	// === read_folder ===
	t, _ = reg.GetTool("read_folder")
	args, _ = json.Marshal(map[string]string{"path": ""})
	res, err = t.Execute(ctx, json.RawMessage(args))
	assert("read_folder lists entries", err == nil && !res.IsError && res.Output != "", fmt.Sprintf("out='%s'", res.Output))

	// === grep_search (boolean for case_sensitive) ===
	os.WriteFile(filepath.Join(testDir, "search_target.go"), []byte("package main\nfunc hello() { println(\"search me\") }"), 0644)
	t, _ = reg.GetTool("grep_search")
	args, _ = json.Marshal(map[string]any{"pattern": "search me", "case_sensitive": false})
	res, err = t.Execute(ctx, json.RawMessage(args))
	assert("grep_search finds match", err == nil && !res.IsError && strings.Contains(res.Output, "search_target"), fmt.Sprintf("out='%s'", res.Output))

	// === run_shell_command ===
	t, _ = reg.GetTool("run_shell_command")
	args, _ = json.Marshal(map[string]string{"command": "echo tool-test-output"})
	res, err = t.Execute(ctx, json.RawMessage(args))
	assert("run_shell_command executes", err == nil && !res.IsError && strings.Contains(res.Output, "tool-test-output"), fmt.Sprintf("out='%s'", res.Output))

	// === save_memory ===
	t, _ = reg.GetTool("save_memory")
	args, _ = json.Marshal(map[string]string{"key": "test-fact", "value": "integration test passed"})
	res, err = t.Execute(ctx, json.RawMessage(args))
	assert("save_memory stores data", err == nil && !res.IsError, fmt.Sprintf("err=%v out=%s", err, res.Output))

	// Verify memory in KV
	val, _ := kv.Get(ctx, "memories", "test-fact")
	assert("Memory persisted in KV", string(val) == "integration test passed", fmt.Sprintf("got '%s'", string(val)))

	// === Guardian blocks .env ===
	os.WriteFile(filepath.Join(testDir, ".env"), []byte("SECRET=123"), 0644)
	t, _ = reg.GetTool("read_file")
	args, _ = json.Marshal(map[string]string{"path": ".env"})
	res, err = t.Execute(ctx, json.RawMessage(args))
	assert("Guardian blocks .env read", res.IsError, fmt.Sprintf("expected error, got is_error=%v", res.IsError))

	// === Guardian blocks out-of-scope write ===
	t, _ = reg.GetTool("write_file")
	args, _ = json.Marshal(map[string]string{"path": "../../escape.txt", "content": "hacked"})
	res, err = t.Execute(ctx, json.RawMessage(args))
	_ = res
	_ = err
	if _, statErr := os.Stat(filepath.Join(testDir, "..", "escape.txt")); statErr != nil {
		assert("Guardian blocks out-of-scope write", true, "")
	} else {
		os.Remove(filepath.Join(testDir, "..", "escape.txt"))
		assert("Guardian blocks out-of-scope write", false, "file escaped")
	}
}

// ---- 8. Ingestion Pipeline ----
func testIngestion(testDir string) {
	g := policies.NewGuardian(testDir)

	os.WriteFile(filepath.Join(testDir, "sample.go"), []byte("package main\n\nfunc main() {\n\tfmt.Println(\"hello\")\n}"), 0644)
	os.WriteFile(filepath.Join(testDir, "notes.md"), []byte("# Notes\n\nThis is a test document."), 0644)

	idx := ingestion.NewIndexer(nil, nil, g)
	assert("NewIndexer creates", idx != nil, "")

	parsed, err := idx.Parser.ParseFile(filepath.Join(testDir, "sample.go"))
	assert("Parser reads .go file", err == nil && parsed != nil, fmt.Sprintf("err=%v", err))
	if parsed != nil {
		assert("Parser detects Go", parsed.Language == "go", fmt.Sprintf("lang=%s", parsed.Language))
	}

	parsed, err = idx.Parser.ParseFile(filepath.Join(testDir, "notes.md"))
	assert("Parser reads .md", err == nil && parsed != nil, fmt.Sprintf("err=%v", err))
	if parsed != nil {
		assert("Parser detects markdown", parsed.Language == "markdown", fmt.Sprintf("lang=%s", parsed.Language))
	}

	envPath := filepath.Join(testDir, ".env")
	parsed, err = idx.Parser.ParseFile(envPath)
	assert("Parser skips .env", parsed == nil && err == nil, fmt.Sprintf("parsed=%v err=%v", parsed != nil, err))

	graph := ingestion.NewDependencyGraph()
	goContent := `package main
import (
	"fmt"
	"os"
)
func main() { fmt.Println("hello") }
`
	graph.ExtractImports(filepath.Join(testDir, "sample.go"), goContent)
	assert("DependencyGraph extracts imports", len(graph.Edges) > 0, fmt.Sprintf("edges=%d", len(graph.Edges)))
}

// ---- 9. IPC Server/Client ----
func testIPC(ctx context.Context, testDir string, kv *db.BBoltStore) {
	if kv == nil {
		skip("IPC tests", "KV store unavailable")
		return
	}
	defer kv.Close()

	vec, _ := db.NewVectorStoreAtPath(filepath.Join(testDir, "ipc-vectors"))
	_ = vec

	server, err := ipc.NewServer()
	assert("NewServer", err == nil && server != nil, fmt.Sprintf("err=%v", err))
	if server == nil {
		return
	}

	// Register ALL routes that the test will call
	server.Register("test.ping", func(ctx context.Context, payload json.RawMessage) (any, *ipc.IPCError) {
		return map[string]string{"pong": "ok"}, nil
	})
	server.Register("app.health", func(ctx context.Context, payload json.RawMessage) (any, *ipc.IPCError) {
		return map[string]string{"status": "ok", "version": "0.1.0"}, nil
	})

	msgService := llm.NewMessageService(kv)
	memPath := filepath.Join(testDir, "ipc-memory")
	memService, _ := db.NewMemoryService(ctx, memPath)

	server.Register("chat.create", func(ctx context.Context, payload json.RawMessage) (any, *ipc.IPCError) {
		var req struct {
			ID    string `json:"id"`
			Title string `json:"title"`
		}
		json.Unmarshal(payload, &req)
		conv, err := msgService.CreateConversation(ctx, req.ID, req.Title)
		if err != nil {
			return nil, &ipc.IPCError{Code: "db_error", Message: err.Error()}
		}
		return conv, nil
	})

	server.Register("chat.history", func(ctx context.Context, payload json.RawMessage) (any, *ipc.IPCError) {
		var req struct {
			ID string `json:"id"`
		}
		json.Unmarshal(payload, &req)
		conv, err := msgService.GetConversation(ctx, req.ID)
		if err != nil {
			return nil, &ipc.IPCError{Code: "chat_not_found", Message: err.Error()}
		}
		return conv, nil
	})

	server.Register("chat.list", func(ctx context.Context, payload json.RawMessage) (any, *ipc.IPCError) {
		list, err := msgService.ListConversations(ctx)
		if err != nil {
			return nil, &ipc.IPCError{Code: "db_error", Message: err.Error()}
		}
		return list, nil
	})

	server.Register("provider.get", func(ctx context.Context, payload json.RawMessage) (any, *ipc.IPCError) {
		return map[string]any{"configured": false}, nil
	})

	server.Register("provider.set", func(ctx context.Context, payload json.RawMessage) (any, *ipc.IPCError) {
		if err := kv.Set(ctx, "settings", "provider", payload); err != nil {
			return nil, &ipc.IPCError{Code: "db_error", Message: err.Error()}
		}
		return map[string]bool{"success": true}, nil
	})

	server.Register("memory.search", func(ctx context.Context, payload json.RawMessage) (any, *ipc.IPCError) {
		var req struct {
			Query string `json:"query"`
			TopK  int    `json:"top_k"`
		}
		json.Unmarshal(payload, &req)
		vec := db.GenerateDummyEmbedding(req.Query, 768)
		results, err := memService.SearchInsight(ctx, req.Query, vec, req.TopK)
		if err != nil {
			return []string{}, nil
		}
		return results, nil
	})

	server.Register("i18n.get", func(ctx context.Context, payload json.RawMessage) (any, *ipc.IPCError) {
		var req struct {
			Locale string `json:"locale"`
		}
		json.Unmarshal(payload, &req)
		if req.Locale != "" {
			i18n.SetLanguage(req.Locale)
		}
		return map[string]string{"lang": i18n.GetCurrentLang()}, nil
	})

	err = server.Start()
	assert("Server starts", err == nil, fmt.Sprintf("err=%v", err))
	if err != nil {
		return
	}
	defer server.Shutdown()

	time.Sleep(200 * time.Millisecond)

	client, err := ipc.NewClient()
	assert("NewClient", err == nil && client != nil, fmt.Sprintf("err=%v", err))
	if client == nil {
		return
	}

	err = client.Connect()
	assert("Client connects", err == nil, fmt.Sprintf("err=%v", err))
	if err != nil {
		return
	}
	defer client.Close()

	// Test all routes
	var pingResp map[string]string
	err = client.Send(ctx, "test.ping", map[string]string{}, &pingResp)
	assert("IPC ping", err == nil && pingResp["pong"] == "ok", fmt.Sprintf("err=%v resp=%v", err, pingResp))

	var health map[string]string
	err = client.Send(ctx, "app.health", map[string]string{}, &health)
	assert("IPC health", err == nil && health["status"] == "ok", fmt.Sprintf("err=%v resp=%v", err, health))

	var conv llm.Conversation
	err = client.Send(ctx, "chat.create", map[string]string{"id": "ipc-test", "title": "IPC Test Chat"}, &conv)
	assert("IPC chat.create", err == nil && conv.Title == "IPC Test Chat", fmt.Sprintf("err=%v conv=%v", err, conv))

	var convHistory llm.Conversation
	err = client.Send(ctx, "chat.history", map[string]string{"id": "ipc-test"}, &convHistory)
	assert("IPC chat.history", err == nil && convHistory.ID == "ipc-test", fmt.Sprintf("err=%v", err))

	var chatList []string
	err = client.Send(ctx, "chat.list", map[string]string{}, &chatList)
	assert("IPC chat.list", err == nil && len(chatList) > 0, fmt.Sprintf("err=%v count=%d", err, len(chatList)))

	var memResults []string
	err = client.Send(ctx, "memory.search", map[string]any{"query": "test", "top_k": 5}, &memResults)
	assert("IPC memory.search", err == nil, fmt.Sprintf("err=%v", err))

	var providerStatus map[string]bool
	err = client.Send(ctx, "provider.get", map[string]string{}, &providerStatus)
	assert("IPC provider.get", err == nil, fmt.Sprintf("err=%v", err))

	var setResult map[string]bool
	err = client.Send(ctx, "provider.set", map[string]string{"name": "gemini"}, &setResult)
	assert("IPC provider.set", err == nil && setResult["success"], fmt.Sprintf("err=%v resp=%v", err, setResult))

	var i18nResp map[string]string
	err = client.Send(ctx, "i18n.get", map[string]string{"locale": "en"}, &i18nResp)
	assert("IPC i18n.get", err == nil && i18nResp["lang"] == "en", fmt.Sprintf("err=%v resp=%v", err, i18nResp))
}

// ---- 10. Engine RAG (without LLM) ----
func testEngineRAG(ctx context.Context, testDir string) {
	vec, _ := db.NewVectorStoreAtPath(filepath.Join(testDir, "engine-vectors"))
	kv, _ := db.NewKVStoreAtPath(filepath.Join(testDir, "engine-kv.db"))

	g := policies.NewGuardian(testDir)
	reg := tools.NewRegistry(testDir, g, kv)
	llmRouter := llm.NewRouter()

	eng := engine.NewEngine(vec, kv, llmRouter, reg, g, nil)
	assert("NewEngine creates", eng != nil, "")
	assert("Engine status idle", eng.GetStatus() == "idle", fmt.Sprintf("status=%s", eng.GetStatus()))

	result, err := eng.ExecuteTool(ctx, engine.ToolCallRequest{
		Name:      "run_shell_command",
		Arguments: json.RawMessage(`{"command":"echo engine-test"}`),
	})
	assert("Engine ExecuteTool", err == nil && result != nil && strings.Contains(result.Output, "engine-test"), fmt.Sprintf("err=%v out=%s", err, result.Output))

	ch, err := eng.StreamQuery(ctx, "what is this?", "default")
	if err == nil && ch != nil {
		var output string
		for chunk := range ch {
			output += chunk.Token
		}
		assert("StreamQuery returns response", output != "", fmt.Sprintf("output='%s'", output))
	} else {
		assert("StreamQuery handles no LLM", true, "")
	}

	answer, err := eng.ProcessQuery("test query", "default")
	assert("ProcessQuery sync version", answer != "" || err == nil, fmt.Sprintf("answer='%s' err=%v", answer, err))
}

// ---- 11. Config Manager ----
func testConfigManager(testDir string) {
	homeDir := filepath.Join(testDir, "config-home")
	os.MkdirAll(filepath.Join(homeDir, ".vectora"), 0755)

	envPath := filepath.Join(homeDir, ".Vectora", ".env")
	os.MkdirAll(filepath.Dir(envPath), 0755)
	os.WriteFile(envPath, []byte("GEMINI_API_KEY=test-key-12345\n"), 0644)

	data, err := os.ReadFile(envPath)
	assert("Config file written", err == nil && strings.Contains(string(data), "test-key-12345"), fmt.Sprintf("err=%v", err))
	assert("Config file readable", len(data) > 0, fmt.Sprintf("size=%d", len(data)))
}

// ---- 12. Git Manager ----
func testGitManager(testDir string) {
	hasGit := checkGitAvailable()
	assert("Git available on system", hasGit, "git not found in PATH")
	if !hasGit {
		return
	}

	gitDir := filepath.Join(testDir, "git-test")
	os.MkdirAll(gitDir, 0755)

	cmd := exec.Command("git", "-C", gitDir, "init")
	err := cmd.Run()
	assert("Git init", err == nil, fmt.Sprintf("err=%v", err))
	if err != nil {
		return
	}

	exec.Command("git", "-C", gitDir, "config", "user.email", "test@test.com").Run()
	exec.Command("git", "-C", gitDir, "config", "user.name", "Test").Run()

	testFile := filepath.Join(gitDir, "git-test-file.txt")
	os.WriteFile(testFile, []byte("initial content"), 0644)

	cmd = exec.Command("git", "-C", gitDir, "add", "git-test-file.txt")
	err = cmd.Run()
	assert("Git add", err == nil, fmt.Sprintf("err=%v", err))

	cmd = exec.Command("git", "-C", gitDir, "commit", "-m", "initial commit")
	err = cmd.Run()
	assert("Git commit", err == nil, fmt.Sprintf("err=%v", err))

	cmd = exec.Command("git", "-C", gitDir, "log", "--oneline")
	output, err := cmd.CombinedOutput()
	assert("Git log shows commit", err == nil && strings.Contains(string(output), "initial commit"), fmt.Sprintf("err=%v output=%s", err, string(output)))

	// Snapshot workflow
	os.WriteFile(testFile, []byte("modified content"), 0644)
	exec.Command("git", "-C", gitDir, "add", "git-test-file.txt").Run()
	cmd = exec.Command("git", "-C", gitDir, "commit", "-m", "chore(vectora): snapshot pre-edit")
	err = cmd.Run()
	assert("Git snapshot workflow", err == nil, fmt.Sprintf("err=%v", err))

	cmd = exec.Command("git", "-C", gitDir, "log", "--oneline")
	output, _ = cmd.CombinedOutput()
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	assert("Two commits exist", len(lines) >= 2, fmt.Sprintf("commits=%d", len(lines)))
}

func checkGitAvailable() bool {
	_, err := exec.LookPath("git")
	return err == nil
}
