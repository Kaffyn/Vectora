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
func (t *GoogleSearchTool) Description() string { return "Consumes Open Web search engine to capture URLs and organic results." }
func (t *GoogleSearchTool) Schema() json.RawMessage {
	return []byte(`{"type":"object","properties":{"query":{"type":"string"}},"required":["query"]}`)
}
func (t *GoogleSearchTool) Execute(ctx context.Context, args map[string]any) (ToolResult, error) {
	query, _ := args["query"].(string)

	// Fetch HTML Lite on DuckDuckGo (Zero-Key alternative for Open Source Agents)
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
		return ToolResult{IsError: true, Output: fmt.Sprintf("Anti-Scraping Block or Network Failure (Code: %d)", resp.StatusCode)}, nil
	}

	body, _ := io.ReadAll(resp.Body)
	html := string(body)

	// Minimal parsing without external lib (captures initial raw blocktext)
	// As LLMs handle noise incredibly well, we send truncated HTML:
	if len(html) > 4000 {
		html = html[:4000] // Truncate to avoid token overflow
	}

	return ToolResult{Output: "DuckDuckGo RAW HTML Response:\n" + html}, nil
}

// ----------------------------------------------------
// 2. Tool: web_fetch
// ----------------------------------------------------
type WebFetchTool struct{}

func (t *WebFetchTool) Name() string        { return "web_fetch" }
func (t *WebFetchTool) Description() string { return "Reads the entire body of a URL and injects the text into the LLM context window." }
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
		return ToolResult{IsError: true, Output: "HTTP buffer failure: " + err.Error()}, nil
	}

	text := string(body)
	if len(text) > 8000 {
		text = text[:8000] + "\n...(Truncated due to Context Window Limit)..."
	}

	return ToolResult{Output: text}, nil
}
