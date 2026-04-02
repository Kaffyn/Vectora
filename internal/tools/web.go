package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// ----------------------------------------------------
// 1. Tool: google_search
// ----------------------------------------------------
type GoogleSearchTool struct{}

func (t *GoogleSearchTool) Name() string        { return "google_search" }
func (t *GoogleSearchTool) Description() string { return "Consome motor de busca Open Web para captar URLs e resultados orgânicos." }
func (t *GoogleSearchTool) Schema() json.RawMessage {
	return []byte(`{"type":"object","properties":{"query":{"type":"string"}},"required":["query"]}`)
}
func (t *GoogleSearchTool) Execute(ctx context.Context, args map[string]any) (ToolResult, error) {
	query, _ := args["query"].(string)

	// Fetch HTML Lite no DuckDuckGo (Alternativa Zero-Key para Agentes Open Source)
	reqURL := "https://html.duckduckgo.com/html/?q=" + url.QueryEscape(query)
	
	req, _ := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 VectoraAgent/1.0")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return ToolResult{IsError: true, Output: err.Error()}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return ToolResult{IsError: true, Output: fmt.Sprintf("Bloqueio Anti-Scraping ou Falha de Rede (Code: %d)", resp.StatusCode)}, nil
	}

	body, _ := io.ReadAll(resp.Body)
	html := string(body)

	// Parsing Mínimo sem lib externa (Captura a blocktext bruta inicial)
	// Como a LLM lida incrivelmente bem com sujeira, enviamos o HTML capado:
	if len(html) > 4000 {
		html = html[:4000] // Trunca p/ n estourar tokens
	}

	return ToolResult{Output: "DuckDuckGo RAW HTML Response:\n" + html}, nil
}

// ----------------------------------------------------
// 2. Tool: web_fetch
// ----------------------------------------------------
type WebFetchTool struct{}

func (t *WebFetchTool) Name() string        { return "web_fetch" }
func (t *WebFetchTool) Description() string { return "Lê integralmente o corpo de uma URL e injeta o texto na janela de contexto da LLM." }
func (t *WebFetchTool) Schema() json.RawMessage {
	return []byte(`{"type":"object","properties":{"url":{"type":"string"}},"required":["url"]}`)
}
func (t *WebFetchTool) Execute(ctx context.Context, args map[string]any) (ToolResult, error) {
	targetURL, _ := args["url"].(string)

	if !strings.HasPrefix(targetURL, "http") {
		targetURL = "https://" + targetURL
	}

	req, _ := http.NewRequestWithContext(ctx, "GET", targetURL, nil)
	req.Header.Set("User-Agent", "Vectora-Autonomous-Engine/1.0")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return ToolResult{IsError: true, Output: err.Error()}, nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ToolResult{IsError: true, Output: "Falha de buffer HTTP: " + err.Error()}, nil
	}

	text := string(body)
	if len(text) > 8000 {
		text = text[:8000] + "\n...(Trancated due Context Window Limit)..."
	}

	return ToolResult{Output: text}, nil
}
