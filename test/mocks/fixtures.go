package mocks

import "time"

// TestFixtures provides common test data
type TestFixtures struct {
	Models  []ModelInfo
	Indices []IndexInfo
	Messages []ChatMessage
}

// NewTestFixtures creates new test fixtures
func NewTestFixtures() *TestFixtures {
	return &TestFixtures{
		Models: []ModelInfo{
			{
				ID:     "qwen3-0.6b",
				Name:   "Qwen 3 0.6B",
				Active: false,
				Size:   "600MB",
				Type:   "chat",
			},
			{
				ID:     "qwen3-1.7b",
				Name:   "Qwen 3 1.7B",
				Active: true,
				Size:   "1.7GB",
				Type:   "chat",
			},
			{
				ID:     "qwen3-4b",
				Name:   "Qwen 3 4B",
				Active: false,
				Size:   "4GB",
				Type:   "chat",
			},
		},
		Indices: []IndexInfo{
			{
				ID:            "idx-1",
				Name:          "Documentation",
				DocumentCount: 100,
				Size:          1024 * 1024 * 50, // 50MB
				CreatedAt:     time.Now().Add(-24 * time.Hour),
				UpdatedAt:     time.Now(),
			},
			{
				ID:            "idx-2",
				Name:          "Code Repository",
				DocumentCount: 500,
				Size:          1024 * 1024 * 200, // 200MB
				CreatedAt:     time.Now().Add(-72 * time.Hour),
				UpdatedAt:     time.Now(),
			},
		},
		Messages: []ChatMessage{
			{
				Role:    "user",
				Content: "Hello, how are you?",
				Model:   "qwen3-1.7b",
			},
			{
				Role:    "assistant",
				Content: "I'm doing well, thanks for asking!",
				Model:   "qwen3-1.7b",
			},
			{
				Role:    "user",
				Content: "What's the weather like?",
				Model:   "qwen3-1.7b",
			},
		},
	}
}

// DefaultConfig returns default configuration
func DefaultConfig() map[string]interface{} {
	return map[string]interface{}{
		"theme":      "light",
		"font_size":  "medium",
		"auto_save":  true,
		"language":   "en",
	}
}

// DefaultSystemHealth returns default system health info
func DefaultSystemHealth() SystemHealth {
	return SystemHealth{
		Status:     "healthy",
		Uptime:     3600,
		ModelsPath: "/home/test/.vectora/models",
		Version:    "1.0.0-test",
	}
}
