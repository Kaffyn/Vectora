package llm

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestGeminiParsing(t *testing.T) {
	fixtureData, err := os.ReadFile("fixtures/gemini_response.json")
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(fixtureData)
	}))
	defer server.Close()

	// Inject our mock server URL into the provider logic
	// (Requires a small modification in gemini_provider to be fully testable without network,
	// but for now we can mock the http.Client behavior or use a specific test helper)

	// Override URL for testing if possible, or just validate unmarshaling directly
	var gResp geminiResponse
	if err := json.Unmarshal(fixtureData, &gResp); err != nil {
		t.Fatalf("failed to unmarshal gemini fixture: %v", err)
	}

	if len(gResp.Candidates) == 0 {
		t.Error("expected at least one candidate")
	}

	text := gResp.Candidates[0].Content.Parts[0].Text
	expected := "Olá! Eu sou o Vectora. Posso ajudar você a analisar este código Go."
	if text != expected {
		t.Errorf("expected %q, got %q", expected, text)
	}
}

func TestClaudeParsing(t *testing.T) {
	fixtureData, err := os.ReadFile("fixtures/claude_tool_call.json")
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}

	var cResp claudeResponse
	if err := json.Unmarshal(fixtureData, &cResp); err != nil {
		t.Fatalf("failed to unmarshal claude fixture: %v", err)
	}

	if len(cResp.Content) != 2 {
		t.Fatalf("expected 2 content blocks, got %d", len(cResp.Content))
	}

	foundTool := false
	for _, item := range cResp.Content {
		if item.Type == "tool_use" {
			foundTool = true
			if item.Name != "read_file" {
				t.Errorf("expected tool name 'read_file', got %q", item.Name)
			}
			var input map[string]string
			json.Unmarshal(item.Input, &input)
			if input["path"] != "main.go" {
				t.Errorf("expected path 'main.go', got %q", input["path"])
			}
		}
	}

	if !foundTool {
		t.Error("did not find tool_use block in claude response")
	}
}
