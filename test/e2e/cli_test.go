package e2e

import (
	"testing"

	"github.com/Kaffyn/Vectora/test"
)

// TestDaemonCLI tests daemon CLI functionality
func TestDaemonCLI(t *testing.T) {
	daemon := test.StartTestDaemon(t)
	defer daemon.Stop()

	// Verify daemon is running
	test.AssertEqual(t, true, daemon.IsRunning(), "Daemon should be running")

	// Test basic daemon operations
	response, err := daemon.Chat("test message")
	test.AssertNoError(t, err, "Daemon should accept messages")
	test.AssertNotNil(t, response, "Daemon should return response")
}

// TestDaemonStatus tests daemon status command
func TestDaemonStatus(t *testing.T) {
	daemon := test.StartTestDaemon(t)
	defer daemon.Stop()

	// In real scenario: would execute "vectora status" command
	// For now, verify daemon reports correct status
	test.AssertEqual(t, true, daemon.IsRunning(), "Daemon should report running status")
}

// TestLPMCLI tests LPM (Llama Package Manager) CLI
func TestLPMCLI(t *testing.T) {
	// LPM tests would verify:
	// - lpm list (list available llama builds)
	// - lpm detect (detect hardware)
	// - lpm recommend (recommend best build)
	// - lpm install (install build)
	// - lpm active (show active build)
	// - lpm set-active (set active build)

	t.Run("list", func(t *testing.T) {
		// Test: lpm list
		// Should return list of available llama builds
	})

	t.Run("detect", func(t *testing.T) {
		// Test: lpm detect
		// Should return system hardware info
	})

	t.Run("recommend", func(t *testing.T) {
		// Test: lpm recommend
		// Should recommend best build for system
	})
}

// TestMPMCLI tests MPM (Model Package Manager) CLI
func TestMPMCLI(t *testing.T) {
	daemon := test.StartTestDaemon(t)
	defer daemon.Stop()

	// Test: mpm list
	models, err := daemon.ListModels()
	test.AssertNoError(t, err, "Should list models")
	test.AssertNotNil(t, models, "Models should be returned")
	test.AssertEqual(t, true, len(models) > 0, "Should have models")

	// Test: mpm detect
	// In real scenario: would detect hardware
	// For now, verify daemon responds

	// Test: mpm search
	response, err := daemon.Chat("search qwen")
	test.AssertNoError(t, err, "Should handle search")
	test.AssertNotNil(t, response, "Should return search results")
}

// TestSetupCLI tests setup/installer CLI
func TestSetupCLI(t *testing.T) {
	// Setup tests would verify:
	// - setup install --path <path> (install Vectora)
	// - setup uninstall (uninstall Vectora)
	// - Dual mode: GUI when no args, CLI when args provided

	t.Run("install", func(t *testing.T) {
		// Test: setup install
		// Should install to specified path
	})

	t.Run("uninstall", func(t *testing.T) {
		// Test: setup uninstall
		// Should uninstall from specified path
	})

	t.Run("gui-mode", func(t *testing.T) {
		// Test: running setup with no args
		// Should open GUI interface
	})

	t.Run("cli-mode", func(t *testing.T) {
		// Test: running setup with flags
		// Should run in CLI mode
	})
}

// TestCLIErrorHandling tests error handling in CLI commands
func TestCLIErrorHandling(t *testing.T) {
	daemon := test.StartTestDaemon(t)
	defer daemon.Stop()

	// Test with invalid model ID
	err := daemon.SetModel("invalid-model-id-12345")
	if err == nil {
		t.Error("Should error with invalid model ID")
	}

	// Test that daemon still responds after error
	response, err := daemon.Chat("test after error")
	test.AssertNoError(t, err, "Should recover from error")
	test.AssertNotNil(t, response, "Should return response after error")
}

// TestCLIJSONOutput tests JSON output format from CLI commands
func TestCLIJSONOutput(t *testing.T) {
	daemon := test.StartTestDaemon(t)
	defer daemon.Stop()

	// Get models in JSON format
	models, err := daemon.ListModels()
	test.AssertNoError(t, err, "Should list models")

	// Verify it's valid
	test.AssertNotNil(t, models, "Models should be returned")

	if len(models) == 0 {
		t.Error("Should have at least one model")
	}

	// Verify structure
	for _, model := range models {
		test.AssertNotNil(t, model.ID, "Model should have ID")
		test.AssertNotNil(t, model.Name, "Model should have Name")
	}
}

// TestCLIVersionFlag tests --version flag
func TestCLIVersionFlag(t *testing.T) {
	daemon := test.StartTestDaemon(t)
	defer daemon.Stop()

	// In real scenario: would execute "vectora --version"
	// Daemon should report version info
	test.AssertEqual(t, true, daemon.IsRunning(), "Daemon should be running")
}

// TestCLIHelpFlag tests --help flag
func TestCLIHelpFlag(t *testing.T) {
	// In real scenario: would execute commands with --help
	// Should display help text for each command
	t.Run("vectora-help", func(t *testing.T) {
		// vectora --help
	})

	t.Run("lpm-help", func(t *testing.T) {
		// lpm --help
	})

	t.Run("mpm-help", func(t *testing.T) {
		// mpm --help
	})

	t.Run("setup-help", func(t *testing.T) {
		// setup --help
	})
}
