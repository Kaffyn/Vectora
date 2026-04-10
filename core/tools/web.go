package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type GoogleSearchTool struct{}

func (t *GoogleSearchTool) Name() string        { return "google_search" }
func (t *GoogleSearchTool) Description() string { return "Searches the web via DuckDuckGo." }
func (t *GoogleSearchTool) Schema() json.RawMessage {
	return []byte(`{"type":"object","properties":{"query":{"type":"string"}},"required":["query"]}`)
}

func (t *GoogleSearchTool) Execute(ctx context.Context, args json.RawMessage) (*ToolResult, error) {
	var params struct {
		Query string `json:"query"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return &ToolResult{Output: "Invalid args", IsError: true}, nil
	}

	reqURL := "https://html.duckduckgo.com/html/?q=" + url.QueryEscape(params.Query)

	req, _ := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) VectoraAgent/1.0")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return &ToolResult{Output: err.Error(), IsError: true}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return &ToolResult{Output: fmt.Sprintf("Search failed (Code: %d)", resp.StatusCode), IsError: true}, nil
	}

	body, _ := io.ReadAll(resp.Body)
	html := string(body)
	if len(html) > 4000 {
		html = html[:4000]
	}

	return &ToolResult{Output: html}, nil
}

type WebFetchTool struct{}

func (t *WebFetchTool) Name() string        { return "web_fetch" }
func (t *WebFetchTool) Description() string { return "Fetches the content of a URL." }
func (t *WebFetchTool) Schema() json.RawMessage {
	return []byte(`{"type":"object","properties":{"url":{"type":"string"}},"required":["url"]}`)
}

func (t *WebFetchTool) Execute(ctx context.Context, args json.RawMessage) (*ToolResult, error) {
	var params struct {
		URL string `json:"url"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return &ToolResult{Output: "Invalid args", IsError: true}, nil
	}

	targetURL := params.URL
	if len(targetURL) >= 4 && targetURL[:4] != "http" {
		targetURL = "https://" + targetURL
	}

	req, _ := http.NewRequestWithContext(ctx, "GET", targetURL, nil)
	req.Header.Set("User-Agent", "Vectora-Autonomous-Engine/1.0")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return &ToolResult{Output: err.Error(), IsError: true}, nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &ToolResult{Output: "Failed to read response", IsError: true}, nil
	}

	text := string(body)
	if len(text) > 8000 {
		text = text[:8000] + "\n...(Truncated)..."
	}

	return &ToolResult{Output: text}, nil
}
