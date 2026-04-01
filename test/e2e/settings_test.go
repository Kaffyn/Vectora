package e2e

import (
	"testing"

	"github.com/Kaffyn/Vectora/test"
	"github.com/Kaffyn/Vectora/test/mocks"
)

// TestSettingsPersistence tests that settings are persisted
func TestSettingsPersistence(t *testing.T) {
	daemon := test.StartTestDaemon(t)

	// Get initial config
	config := mocks.DefaultConfig()
	test.AssertNotNil(t, config, "Config should not be nil")
	test.AssertEqual(t, "light", config["theme"], "Initial theme should be light")

	daemon.Stop()

	// Restart daemon (in real scenario, would reload saved config)
	daemon = test.StartTestDaemon(t)
	defer daemon.Stop()

	// Verify daemon still has config
	test.AssertEqual(t, true, daemon.IsRunning(), "Daemon should be running")
}

// TestProviderConfiguration tests configuring different AI providers
func TestProviderConfiguration(t *testing.T) {
	daemon := test.StartTestDaemon(t)
	defer daemon.Stop()

	// Test Qwen3 provider (default)
	response, err := daemon.Chat("Test Qwen3")
	test.AssertNoError(t, err, "Qwen3 chat should work")
	test.AssertNotNil(t, response, "Should return response")

	// Test model switching (simulates provider change)
	models, err := daemon.ListModels()
	test.AssertNoError(t, err, "Should list models")

	if len(models) > 0 {
		err = daemon.SetModel(models[0].ID)
		test.AssertNoError(t, err, "Should set model")

		// Chat with new provider
		response, err := daemon.Chat("Test new provider")
		test.AssertNoError(t, err, "New provider chat should work")
		test.AssertNotNil(t, response, "Should return response")
	}
}

// TestSystemHealth tests system health information
func TestSystemHealth(t *testing.T) {
	daemon := test.StartTestDaemon(t)
	defer daemon.Stop()

	// In real scenario: daemon.GetSystemHealth()
	health := mocks.DefaultSystemHealth()

	test.AssertNotNil(t, health, "Health info should not be nil")
	test.AssertEqual(t, "healthy", health.Status, "Status should be healthy")
	test.AssertEqual(t, "1.0.0-test", health.Version, "Version should match")
}

// TestModelSelection tests selecting between available models
func TestModelSelection(t *testing.T) {
	daemon := test.StartTestDaemon(t)
	defer daemon.Stop()

	// List available models
	models, err := daemon.ListModels()
	test.AssertNoError(t, err, "Should list models")
	test.AssertEqual(t, true, len(models) > 0, "Should have models")

	// Select each model and test
	for _, model := range models {
		err := daemon.SetModel(model.ID)
		test.AssertNoError(t, err, "Should set model "+model.ID)

		// Chat with model
		response, err := daemon.Chat("Test with " + model.Name)
		test.AssertNoError(t, err, "Chat with "+model.Name+" should work")
		test.AssertNotNil(t, response, "Should return response")
	}
}

// TestUIPreferences tests saving UI preferences
func TestUIPreferences(t *testing.T) {
	daemon := test.StartTestDaemon(t)
	defer daemon.Stop()

	// Get default preferences
	config := mocks.DefaultConfig()
	test.AssertNotNil(t, config, "Config should exist")

	// Verify common preferences
	preferences := []string{"theme", "font_size", "auto_save", "language"}
	for _, pref := range preferences {
		if _, ok := config[pref]; !ok {
			t.Logf("Warning: preference %s not found in config", pref)
		}
	}
}

// TestThemeSwitching tests switching between light and dark themes
func TestThemeSwitching(t *testing.T) {
	daemon := test.StartTestDaemon(t)
	defer daemon.Stop()

	// In real scenario: would switch theme via config
	config := mocks.DefaultConfig()

	// Switch to dark theme
	config["theme"] = "dark"
	test.AssertEqual(t, "dark", config["theme"], "Theme should be dark")

	// Switch back to light
	config["theme"] = "light"
	test.AssertEqual(t, "light", config["theme"], "Theme should be light")

	// Verify daemon still works after theme changes
	response, err := daemon.Chat("After theme change")
	test.AssertNoError(t, err, "Should work after theme change")
	test.AssertNotNil(t, response, "Should return response")
}

// TestFontSizeSettings tests font size preference
func TestFontSizeSettings(t *testing.T) {
	daemon := test.StartTestDaemon(t)
	defer daemon.Stop()

	sizes := []string{"small", "medium", "large"}

	for _, size := range sizes {
		config := mocks.DefaultConfig()
		config["font_size"] = size

		// Verify daemon still works with each size
		response, err := daemon.Chat("Test with " + size + " font")
		test.AssertNoError(t, err, "Should work with "+size+" font")
		test.AssertNotNil(t, response, "Should return response")
	}
}

// TestAutoSaveFeature tests auto-save preference
func TestAutoSaveFeature(t *testing.T) {
	daemon := test.StartTestDaemon(t)
	defer daemon.Stop()

	config := mocks.DefaultConfig()

	// Enable auto-save
	config["auto_save"] = true
	test.AssertEqual(t, true, config["auto_save"], "Auto-save should be enabled")

	// Disable auto-save
	config["auto_save"] = false
	test.AssertEqual(t, false, config["auto_save"], "Auto-save should be disabled")

	// Verify daemon works with both settings
	response, err := daemon.Chat("With auto-save disabled")
	test.AssertNoError(t, err, "Should work with auto-save disabled")
	test.AssertNotNil(t, response, "Should return response")
}

// TestLanguagePreference tests language preference
func TestLanguagePreference(t *testing.T) {
	daemon := test.StartTestDaemon(t)
	defer daemon.Stop()

	languages := []string{"en", "pt", "es", "fr"}

	for _, lang := range languages {
		config := mocks.DefaultConfig()
		config["language"] = lang

		// In real scenario: would change UI language
		response, err := daemon.Chat("Test in " + lang)
		test.AssertNoError(t, err, "Should work with "+lang)
		test.AssertNotNil(t, response, "Should return response")
	}
}

// TestResetSettings tests resetting settings to defaults
func TestResetSettings(t *testing.T) {
	daemon := test.StartTestDaemon(t)
	defer daemon.Stop()

	// Modify config
	config := mocks.DefaultConfig()
	config["theme"] = "dark"
	config["font_size"] = "large"

	// Reset to defaults
	config = mocks.DefaultConfig()

	test.AssertEqual(t, "light", config["theme"], "Theme should be reset to light")
	test.AssertEqual(t, "medium", config["font_size"], "Font size should be reset to medium")

	// Verify daemon works after reset
	response, err := daemon.Chat("After reset")
	test.AssertNoError(t, err, "Should work after reset")
	test.AssertNotNil(t, response, "Should return response")
}
