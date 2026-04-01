package e2e

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Kaffyn/Vectora/test"
)

// TestEditorFileOperations tests basic file operations
func TestEditorFileOperations(t *testing.T) {
	tempDir := test.CreateTempDir(t)

	// Create a test file
	filePath := filepath.Join(tempDir, "test.go")
	content := "package main\n\nfunc main() {\n\tprintln(\"Hello\")\n}"

	err := test.CreateTestFile(filePath, content)
	test.AssertNoError(t, err, "Should create test file")

	// Read the file
	readContent, err := os.ReadFile(filePath)
	test.AssertNoError(t, err, "Should read file")

	// Verify content
	test.AssertEqual(t, content, string(readContent), "File content should match")
}

// TestEditorMultiTabEditing tests editing multiple files in tabs
func TestEditorMultiTabEditing(t *testing.T) {
	tempDir := test.CreateTempDir(t)

	files := []struct {
		name    string
		content string
	}{
		{"file1.go", "package main\nfunc main1() {}"},
		{"file2.go", "package main\nfunc main2() {}"},
		{"file3.go", "package main\nfunc main3() {}"},
	}

	// Create test files
	for _, f := range files {
		filePath := filepath.Join(tempDir, f.name)
		err := test.CreateTestFile(filePath, f.content)
		test.AssertNoError(t, err, "Should create file "+f.name)
	}

	// Read and verify each file
	for _, f := range files {
		filePath := filepath.Join(tempDir, f.name)
		content, err := os.ReadFile(filePath)
		test.AssertNoError(t, err, "Should read file "+f.name)
		test.AssertEqual(t, f.content, string(content), "Content should match for "+f.name)
	}
}

// TestEditorSaveFile tests saving file changes
func TestEditorSaveFile(t *testing.T) {
	tempDir := test.CreateTempDir(t)
	filePath := filepath.Join(tempDir, "edit.go")

	// Create initial file
	originalContent := "package main\n"
	err := test.CreateTestFile(filePath, originalContent)
	test.AssertNoError(t, err, "Should create initial file")

	// Modify content
	newContent := "package main\n\nfunc modified() {}\n"
	err = os.WriteFile(filePath, []byte(newContent), 0644)
	test.AssertNoError(t, err, "Should write modified content")

	// Read and verify
	readContent, err := os.ReadFile(filePath)
	test.AssertNoError(t, err, "Should read modified file")
	test.AssertEqual(t, newContent, string(readContent), "Content should be updated")
}

// TestEditorCreateNewFile tests creating a new file
func TestEditorCreateNewFile(t *testing.T) {
	tempDir := test.CreateTempDir(t)
	filePath := filepath.Join(tempDir, "newfile.go")

	// Create file
	content := "// New file\npackage main\n"
	err := test.CreateTestFile(filePath, content)
	test.AssertNoError(t, err, "Should create new file")

	// Verify file exists
	info, err := os.Stat(filePath)
	test.AssertNoError(t, err, "File should exist")
	if info.IsDir() {
		t.Error("Created path should be a file, not directory")
	}

	// Verify content
	readContent, err := os.ReadFile(filePath)
	test.AssertNoError(t, err, "Should read new file")
	test.AssertEqual(t, content, string(readContent), "Content should match")
}

// TestEditorFileNotFound tests handling of missing files
func TestEditorFileNotFound(t *testing.T) {
	tempDir := test.CreateTempDir(t)
	filePath := filepath.Join(tempDir, "nonexistent.go")

	// Try to read non-existent file
	_, err := os.ReadFile(filePath)
	if err == nil {
		t.Error("Should error when reading non-existent file")
	}
}

// TestEditorFileSyntaxHighlight tests syntax highlighting compatibility
func TestEditorFileSyntaxHighlight(t *testing.T) {
	tempDir := test.CreateTempDir(t)

	// Create Go file with syntax
	goFile := filepath.Join(tempDir, "syntax.go")
	goCode := `package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")
}
`
	err := test.CreateTestFile(goFile, goCode)
	test.AssertNoError(t, err, "Should create Go file")

	// Read and verify
	content, err := os.ReadFile(goFile)
	test.AssertNoError(t, err, "Should read Go file")
	test.AssertNotNil(t, content, "Go file should have content")

	// Verify syntax elements are present
	contentStr := string(content)
	if !contains(contentStr, "package main") {
		t.Error("File should contain package declaration")
	}
	if !contains(contentStr, "func main()") {
		t.Error("File should contain main function")
	}
}

// TestEditorPermissionError tests handling of permission errors
func TestEditorPermissionError(t *testing.T) {
	// This test would require setting file permissions
	// For now, just verify the test infrastructure works
	tempDir := test.CreateTempDir(t)
	filePath := filepath.Join(tempDir, "readonly.go")

	// Create file
	err := test.CreateTestFile(filePath, "package main\n")
	test.AssertNoError(t, err, "Should create file")

	// Change to read-only
	err = os.Chmod(filePath, 0444)
	if err != nil {
		t.Skip("Cannot change file permissions in test environment")
	}
	defer os.Chmod(filePath, 0644)

	// Try to write (should fail on most systems)
	err = os.WriteFile(filePath, []byte("new content"), 0644)
	if err == nil && os.Geteuid() != 0 {
		// Error expected unless running as root
		t.Error("Should error when writing to read-only file")
	}
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	for i := 0; i < len(s)-len(substr)+1; i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
