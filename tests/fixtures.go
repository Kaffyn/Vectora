package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// TestFixture holds test data
type TestFixture struct {
	WorkspaceDir string
	ConfigFile   string
	TestFiles    []string
}

// NewTestFixture creates a new test fixture
func NewTestFixture() (*TestFixture, error) {
	ws, err := TempDir("vectora-test-")
	if err != nil {
		return nil, err
	}

	fixture := &TestFixture{
		WorkspaceDir: ws,
		ConfigFile:   filepath.Join(ws, ".vectora", "config.json"),
		TestFiles:    []string{},
	}

	return fixture, nil
}

// CreateProjectStructure creates a test project
func (f *TestFixture) CreateProjectStructure() error {
	files := map[string]string{
		"main.go": `package main

import "fmt"

func main() {
	fmt.Println("Hello from test project")
}
`,
		"lib.go": `package main

func Add(a, b int) int {
	return a + b
}

func Subtract(a, b int) int {
	return a - b
}
`,
		"README.md": `# Test Project

This is a test project for Vectora.

## Features
- Feature 1
- Feature 2
- Feature 3

## Installation
Just run the code.
`,
		"config.yaml": `version: 1
app:
  name: TestApp
  version: 1.0.0
`,
		"utils.go": `package main

import "strings"

func StringUtils(s string) string {
	return strings.ToUpper(s)
}
`,
	}

	for filename, content := range files {
		path := filepath.Join(f.WorkspaceDir, filename)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return err
		}
		f.TestFiles = append(f.TestFiles, path)
	}

	return nil
}

// CreateSubdirectories creates nested directories
func (f *TestFixture) CreateSubdirectories() error {
	dirs := []string{
		"src",
		"src/modules",
		"tests",
		"docs",
	}

	for _, dir := range dirs {
		path := filepath.Join(f.WorkspaceDir, dir)
		if err := os.MkdirAll(path, 0755); err != nil {
			return err
		}
	}

	// Create files in subdirectories
	subFiles := map[string]string{
		"src/main.go":           "package src\n\nfunc Main() {}\n",
		"src/modules/module.go": "package modules\n\nfunc Init() {}\n",
		"tests/main_test.go":    "package main\n\nimport \"testing\"\n\nfunc TestMain(t *testing.T) {}\n",
		"docs/README.md":        "# Documentation\n\nThis is the documentation.\n",
	}

	for filename, content := range subFiles {
		path := filepath.Join(f.WorkspaceDir, filename)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return err
		}
		f.TestFiles = append(f.TestFiles, path)
	}

	return nil
}

// CreateLargeFiles creates files for indexing tests
func (f *TestFixture) CreateLargeFiles() error {
	// Create a file with code content
	largeFile := `package main

import (
	"context"
	"fmt"
	"log"
	"sync"
)

type DataStore interface {
	Get(ctx context.Context, key string) (interface{}, error)
	Set(ctx context.Context, key string, value interface{}) error
	Delete(ctx context.Context, key string) error
}

type MemoryStore struct {
	mu    sync.RWMutex
	data  map[string]interface{}
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		data: make(map[string]interface{}),
	}
}

func (s *MemoryStore) Get(ctx context.Context, key string) (interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if val, ok := s.data[key]; ok {
		return val, nil
	}
	return nil, fmt.Errorf("key not found: %s", key)
}

func (s *MemoryStore) Set(ctx context.Context, key string, value interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data[key] = value
	return nil
}

func (s *MemoryStore) Delete(ctx context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.data, key)
	return nil
}

func main() {
	store := NewMemoryStore()
	ctx := context.Background()

	if err := store.Set(ctx, "test", "value"); err != nil {
		log.Fatal(err)
	}

	val, err := store.Get(ctx, "test")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Value:", val)
}
`

	path := filepath.Join(f.WorkspaceDir, "store.go")
	if err := os.WriteFile(path, []byte(largeFile), 0644); err != nil {
		return err
	}
	f.TestFiles = append(f.TestFiles, path)

	return nil
}

// CreateConfigFile creates a config file
func (f *TestFixture) CreateConfigFile(content string) error {
	configDir := filepath.Dir(f.ConfigFile)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	return os.WriteFile(f.ConfigFile, []byte(content), 0644)
}

// Cleanup removes all test files
func (f *TestFixture) Cleanup() error {
	return os.RemoveAll(f.WorkspaceDir)
}

// GetTestData returns various test data
type TestData struct {
	SimplePrompts       []string
	ComplexPrompts      []string
	InvalidInputs       []string
	FilePatterns        []string
	ExcludePatterns     []string
	SearchQueries       []string
	ChatMessages        []string
}

// NewTestData creates test data
func NewTestData() *TestData {
	return &TestData{
		SimplePrompts: []string{
			"What is the purpose of main.go?",
			"How does the Add function work?",
			"What packages are imported?",
			"List all Go files in the project",
			"Show me the README content",
		},
		ComplexPrompts: []string{
			"Analyze the architecture and suggest improvements",
			"Find all database operations and their performance implications",
			"Identify potential security vulnerabilities",
			"Create a test plan for this project",
			"Generate documentation from the code",
		},
		InvalidInputs: []string{
			"",
			"   ",
			"\n\n\n",
			string([]byte{0x00, 0xFF, 0xFE}), // Invalid UTF-8
		},
		FilePatterns: []string{
			"*.go",
			"*.md",
			"*.yaml",
			"*.json",
			"*.txt",
			"**/*.go",
			"**/test*.go",
		},
		ExcludePatterns: []string{
			"*.log",
			"*.tmp",
			".git",
			"node_modules",
			".DS_Store",
			"__pycache__",
		},
		SearchQueries: []string{
			"function Add",
			"package main",
			"import fmt",
			"documentation",
			"test",
			"error handling",
		},
		ChatMessages: []string{
			"Hello, can you help me understand this code?",
			"What are the best practices here?",
			"How can I optimize this?",
			"Can you explain the logic?",
			"What dependencies does this need?",
		},
	}
}

// CreateMockVectoraProject creates a mock Vectora project structure
func CreateMockVectoraProject(dir string) error {
	structure := map[string]string{
		"go.mod": `module github.com/test/project

go 1.21
`,
		"go.sum": `github.com/some/package v1.0.0 h1:...
`,
		"main.go": `package main

import "fmt"

func main() {
	fmt.Println("Test")
}
`,
		".env": `API_KEY=test123
LOG_LEVEL=debug
`,
		".gitignore": `*.log
.env.local
/bin
/dist
`,
	}

	for filename, content := range structure {
		path := filepath.Join(dir, filename)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to create %s: %w", filename, err)
		}
	}

	return nil
}

// CreateCorruptedFiles creates files with invalid content
func CreateCorruptedFiles(dir string) error {
	files := map[string][]byte{
		"corrupted.json": []byte(`{"invalid": json}`),
		"corrupted.yaml": []byte(`invalid: yaml: content: :`),
		"binary.bin":     []byte{0xFF, 0xFE, 0x00, 0x00, 0xFF, 0xFE},
	}

	for filename, content := range files {
		path := filepath.Join(dir, filename)
		if err := os.WriteFile(path, content, 0644); err != nil {
			return err
		}
	}

	return nil
}

// EnvironmentConfig holds test environment configuration
type EnvironmentConfig struct {
	CoreBinaryPath string
	GeminiAPIKey   string
	VectoraPort    int
	WorkspaceDir   string
	LogLevel       string
	TimeoutSeconds int
}

// NewEnvironmentConfig creates config from environment
func NewEnvironmentConfig() *EnvironmentConfig {
	return &EnvironmentConfig{
		CoreBinaryPath: GetCoreBinaryPath(),
		GeminiAPIKey:   os.Getenv("GEMINI_API_KEY"),
		VectoraPort:    42780,
		WorkspaceDir:   GetWorkspaceDir(),
		LogLevel:       GetLogLevel(),
		TimeoutSeconds: 30,
	}
}

// LoadTestConfig loads test configuration from .testenv
func LoadTestConfig(path string) *EnvironmentConfig {
	config := ParseTestConfig(path)

	return &EnvironmentConfig{
		CoreBinaryPath: getOrDefault(config, "VECTORA_CORE_BIN", GetCoreBinaryPath()),
		GeminiAPIKey:   getOrDefault(config, "GEMINI_API_KEY", os.Getenv("GEMINI_API_KEY")),
		VectoraPort:    42780,
		WorkspaceDir:   getOrDefault(config, "TEST_WORKSPACE", GetWorkspaceDir()),
		LogLevel:       getOrDefault(config, "TEST_LOG_LEVEL", "info"),
		TimeoutSeconds: 30,
	}
}

func getOrDefault(m map[string]string, key, defaultVal string) string {
	if val, ok := m[key]; ok && val != "" {
		return val
	}
	return defaultVal
}
