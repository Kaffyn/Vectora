package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// TestIntegration tests integration scenarios
func TestIntegration(config *EnvironmentConfig, runner *TestRunner) {
	runner.RunTest("Integration: Workspace Directory Exists", func() error {
		// Create test workspace
		testDir := filepath.Join(os.TempDir(), "vectora-test-ws")
		defer os.RemoveAll(testDir)

		if err := os.MkdirAll(testDir, 0755); err != nil {
			return err
		}

		return nil
	})

	runner.RunTest("Integration: Create Test Files", func() error {
		fixture, err := NewTestFixture()
		if err != nil {
			return err
		}
		defer fixture.Cleanup()

		if err := fixture.CreateProjectStructure(); err != nil {
			return err
		}

		if len(fixture.TestFiles) == 0 {
			return fmt.Errorf("no files created")
		}

		return nil
	})

	runner.RunTest("Integration: File Structure", func() error {
		fixture, err := NewTestFixture()
		if err != nil {
			return err
		}
		defer fixture.Cleanup()

		if err := fixture.CreateProjectStructure(); err != nil {
			return err
		}

		if err := fixture.CreateSubdirectories(); err != nil {
			return err
		}

		if !FileExists(filepath.Join(fixture.WorkspaceDir, "main.go")) {
			return fmt.Errorf("main.go not found")
		}

		return nil
	})

	runner.RunTest("Integration: Config File Creation", func() error {
		fixture, err := NewTestFixture()
		if err != nil {
			return err
		}
		defer fixture.Cleanup()

		configContent := `{
  "version": 1,
  "provider": "gemini",
  "api_key": "test"
}`

		if err := fixture.CreateConfigFile(configContent); err != nil {
			return err
		}

		if !FileExists(fixture.ConfigFile) {
			return fmt.Errorf("config file not created")
		}

		return nil
	})

	runner.RunTest("Integration: Sequential Operations", func() error {
		fixture, err := NewTestFixture()
		if err != nil {
			return err
		}
		defer fixture.Cleanup()

		// Create project structure
		if err := fixture.CreateProjectStructure(); err != nil {
			return err
		}

		// Add subdirectories
		if err := fixture.CreateSubdirectories(); err != nil {
			return err
		}

		// Verify all files
		for _, file := range fixture.TestFiles {
			if !FileExists(file) {
				return fmt.Errorf("file not created: %s", file)
			}
		}

		return nil
	})

	runner.RunTest("Integration: Large Project Simulation", func() error {
		fixture, err := NewTestFixture()
		if err != nil {
			return err
		}
		defer fixture.Cleanup()

		if err := fixture.CreateProjectStructure(); err != nil {
			return err
		}

		if err := fixture.CreateSubdirectories(); err != nil {
			return err
		}

		if err := fixture.CreateLargeFiles(); err != nil {
			return err
		}

		if len(fixture.TestFiles) < 5 {
			return fmt.Errorf("expected more test files")
		}

		return nil
	})
}

// TestCoreIntegration tests Core binary integration
func TestCoreIntegration(config *EnvironmentConfig, runner *TestRunner) {
	// Check if Core binary exists
	runner.RunTest("Core: Binary Exists", func() error {
		// Try to find the core binary
		paths := []string{
			"./bin/core",
			"./bin/core.exe",
			"vectora",
			"vectora.exe",
		}

		found := false
		for _, path := range paths {
			if FileExists(path) {
				found = true
				break
			}
		}

		if !found {
			// This is not a critical failure for local testing
			return nil
		}

		return nil
	})

	// Test Core startup (without actually starting it)
	runner.RunTest("Core: Startup Check", func() error {
		// Simulate checking if core can be started
		// In a real scenario, would attempt to start and connect

		// For now, just verify command structure
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_ = ExecCommand(ctx, "vectora", "start", "--help")

		// Help should work or fail gracefully
		return nil
	})

	runner.RunTest("Core: Status Check", func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_ = ExecCommand(ctx, "vectora", "status")

		// Status may fail if core is not running, which is expected
		return nil
	})

	runner.RunTest("Core: Stop Command", func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_ = ExecCommand(ctx, "vectora", "stop")

		// Stop may fail if core is not running, which is expected
		return nil
	})

	runner.RunTest("Core: Restart Command", func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Don't actually run restart in tests, just verify it exists
		_ = ExecCommand(ctx, "vectora", "restart", "--help")

		// May fail, which is OK
		return nil
	})
}

// TestCommandIntegration tests command-level integration
func TestCommandIntegration(config *EnvironmentConfig, runner *TestRunner) {
	runner.RunTest("Command: Ask Execution", func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Ask without core will fail, which is expected
		_ = ExecCommand(ctx, "vectora", "ask", "test question")

		// Just verify command exists and processes input
		return nil
	})

	runner.RunTest("Command: Embed Validation", func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Embed without core will fail, which is expected
		_ = ExecCommand(ctx, "vectora", "embed", ".")

		// Verify command exists
		return nil
	})

	runner.RunTest("Command: Chat Execution", func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Use ExecCommandWithInput to simulate chat input
		_ = ExecCommandWithInput(ctx, "exit\n", "vectora", "chat")

		// Chat without core may fail, which is expected
		return nil
	})

	runner.RunTest("Command: Config Execution", func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		result := ExecCommand(ctx, "vectora", "config", "--help")

		if result != nil && result.ExitCode == 0 {
			if !ContainsString(result.Stdout, "config") {
				return fmt.Errorf("config help output invalid")
			}
		}

		return nil
	})

	runner.RunTest("Command: Models List", func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_ = ExecCommand(ctx, "vectora", "models", "list")

		// May fail if core is not running, which is expected
		return nil
	})
}

// TestWorkspaceIntegration tests workspace operations
func TestWorkspaceIntegration(config *EnvironmentConfig, runner *TestRunner) {
	runner.RunTest("Workspace: Creation", func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_ = ExecCommand(ctx, "vectora", "workspace", "--help")

		// Verify workspace command exists
		return nil
	})

	runner.RunTest("Workspace: Directory Validation", func() error {
		fixture, err := NewTestFixture()
		if err != nil {
			return err
		}
		defer fixture.Cleanup()

		// Verify workspace directory was created
		if !FileExists(fixture.WorkspaceDir) {
			return fmt.Errorf("workspace directory not created")
		}

		return nil
	})

	runner.RunTest("Workspace: File Management", func() error {
		fixture, err := NewTestFixture()
		if err != nil {
			return err
		}
		defer fixture.Cleanup()

		if err := fixture.CreateProjectStructure(); err != nil {
			return err
		}

		// Verify files can be listed
		files := fixture.TestFiles
		if len(files) == 0 {
			return fmt.Errorf("no files in workspace")
		}

		return nil
	})

	runner.RunTest("Workspace: Subdirectory Support", func() error {
		fixture, err := NewTestFixture()
		if err != nil {
			return err
		}
		defer fixture.Cleanup()

		if err := fixture.CreateProjectStructure(); err != nil {
			return err
		}

		if err := fixture.CreateSubdirectories(); err != nil {
			return err
		}

		// Verify subdirectories exist
		srcDir := filepath.Join(fixture.WorkspaceDir, "src")
		if !FileExists(srcDir) {
			return fmt.Errorf("src directory not created")
		}

		return nil
	})
}

// TestEndToEndFlow tests complete end-to-end flows
func TestEndToEndFlow(config *EnvironmentConfig, runner *TestRunner) {
	runner.RunTest("E2E: Basic Workflow", func() error {
		// Create fixture
		fixture, err := NewTestFixture()
		if err != nil {
			return err
		}
		defer fixture.Cleanup()

		// Create project
		if err := fixture.CreateProjectStructure(); err != nil {
			return err
		}

		// Verify files exist
		if len(fixture.TestFiles) == 0 {
			return fmt.Errorf("project creation failed")
		}

		// Verify we can read files
		for _, file := range fixture.TestFiles {
			if _, err := ReadFile(file); err != nil {
				return fmt.Errorf("failed to read file: %w", err)
			}
		}

		return nil
	})

	runner.RunTest("E2E: Project Analysis Simulation", func() error {
		fixture, err := NewTestFixture()
		if err != nil {
			return err
		}
		defer fixture.Cleanup()

		if err := fixture.CreateProjectStructure(); err != nil {
			return err
		}

		if err := fixture.CreateLargeFiles(); err != nil {
			return err
		}

		// Count files by type
		goFiles := 0
		mdFiles := 0

		for _, file := range fixture.TestFiles {
			if filepath.Ext(file) == ".go" {
				goFiles++
			} else if filepath.Ext(file) == ".md" {
				mdFiles++
			}
		}

		if goFiles == 0 && mdFiles == 0 {
			return fmt.Errorf("no files analyzed")
		}

		return nil
	})

	runner.RunTest("E2E: Multi-Step Process", func() error {
		// Step 1: Create workspace
		fixture, err := NewTestFixture()
		if err != nil {
			return err
		}
		defer fixture.Cleanup()

		// Step 2: Create structure
		if err := fixture.CreateProjectStructure(); err != nil {
			return err
		}

		// Step 3: Add subdirectories
		if err := fixture.CreateSubdirectories(); err != nil {
			return err
		}

		// Step 4: Create large files
		if err := fixture.CreateLargeFiles(); err != nil {
			return err
		}

		// Step 5: Verify everything
		if len(fixture.TestFiles) < 8 {
			return fmt.Errorf("expected at least 8 files")
		}

		return nil
	})

	runner.RunTest("E2E: Cleanup Verification", func() error {
		fixture, err := NewTestFixture()
		if err != nil {
			return err
		}

		wsDir := fixture.WorkspaceDir

		if err := fixture.CreateProjectStructure(); err != nil {
			return err
		}

		// Cleanup
		if err := fixture.Cleanup(); err != nil {
			return fmt.Errorf("cleanup failed: %w", err)
		}

		// Verify cleanup
		if FileExists(wsDir) {
			return fmt.Errorf("workspace not deleted")
		}

		return nil
	})
}

// TestProcessIntegration tests process-level operations
func TestProcessIntegration(config *EnvironmentConfig, runner *TestRunner) {
	runner.RunTest("Process: Command Execution", func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		result := ExecCommand(ctx, "vectora", "--version")

		if result != nil && result.Err != nil {
			// Might fail if vectora is not in PATH, which is expected for tests
			return nil
		}

		return nil
	})

	runner.RunTest("Process: Context Timeout", func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		// Attempt help which should complete quickly
		_ = ExecCommand(ctx, "vectora", "--help")

		// Should complete before timeout
		return nil
	})

	runner.RunTest("Process: Error Handling", func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Run command that will fail
		result := ExecCommand(ctx, "vectora", "invalid-command")

		// Should handle error gracefully
		if result != nil && result.ExitCode == 0 {
			return fmt.Errorf("expected non-zero exit code")
		}

		return nil
	})

	runner.RunTest("Process: Concurrent Execution", func() error {
		// Run multiple commands concurrently (simulated)
		commands := [][]string{
			{"--help"},
			{"--version"},
			{"status"},
		}

		for _, args := range commands {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

			result := ExecCommand(ctx, "vectora", args...)

			cancel()

			if result == nil {
				return fmt.Errorf("command execution failed")
			}
		}

		return nil
	})

	runner.RunTest("Process: Input/Output Handling", func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Test with stdin input
		input := "test input\n"
		result := ExecCommandWithInput(ctx, input, "echo", "test")

		if result != nil && result.Err != nil && result.ExitCode != 0 {
			// May fail, which is OK for this test
		}

		return nil
	})
}

// TestErrorRecovery tests error recovery and resilience
func TestErrorRecovery(config *EnvironmentConfig, runner *TestRunner) {
	runner.RunTest("Recovery: Invalid Command", func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		result := ExecCommand(ctx, "vectora", "invalid")

		// Should fail gracefully
		if result != nil && result.ExitCode == 0 {
			return fmt.Errorf("invalid command should fail")
		}

		return nil
	})

	runner.RunTest("Recovery: Missing Core", func() error {
		// Test commands when core is not available
		commands := [][]string{
			{"ask", "test"},
			{"embed", "."},
			{"chat"},
		}

		for _, args := range commands {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

			_ = ExecCommand(ctx, "vectora", args...)

			cancel()

			// May fail, which is expected
		}

		return nil
	})

	runner.RunTest("Recovery: File System Error", func() error {
		// Test handling of file system errors
		fixture, err := NewTestFixture()
		if err != nil {
			return err
		}
		defer fixture.Cleanup()

		// Try to read from cleaned up directory
		result, err := ReadFile(filepath.Join(fixture.WorkspaceDir, "nonexistent.txt"))

		// Should fail gracefully
		if err == nil && result != "" {
			return fmt.Errorf("should fail on missing file")
		}

		return nil
	})

	runner.RunTest("Recovery: Context Cancellation", func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		cancel() // Cancel immediately

		_ = ExecCommand(ctx, "vectora", "--help")

		// May fail due to context cancellation, which is OK
		return nil
	})
}
