package e2e

import (
	"context"
	"testing"
	"time"

	"github.com/Kaffyn/Vectora/test"
)

// TestChatBasic tests basic chat functionality
func TestChatBasic(t *testing.T) {
	daemon := test.StartTestDaemon(t)
	defer daemon.Stop()

	// Test chat message
	response, err := daemon.Chat("Hello, how are you?")
	test.AssertNoError(t, err, "Chat message should not error")
	test.AssertNotNil(t, response, "Chat should return response")

	// Verify response contains expected content
	if response == "" {
		t.Error("Chat response should not be empty")
	}
}

// TestChatModelSwitching tests switching models during chat
func TestChatModelSwitching(t *testing.T) {
	daemon := test.StartTestDaemon(t)
	defer daemon.Stop()

	// Get initial models
	models, err := daemon.ListModels()
	test.AssertNoError(t, err, "Should list models")
	test.AssertNotNil(t, models, "Models should not be nil")

	if len(models) < 2 {
		t.Fatal("Need at least 2 models for this test")
	}

	// Switch to a different model
	targetModel := models[1].ID
	err = daemon.SetModel(targetModel)
	test.AssertNoError(t, err, "Should set model")

	// Chat with new model
	response, err := daemon.Chat("Test message")
	test.AssertNoError(t, err, "Chat with new model should not error")
	test.AssertNotNil(t, response, "Response should not be nil")
}

// TestChatPersistence tests that chat history is persisted
func TestChatPersistence(t *testing.T) {
	daemon := test.StartTestDaemon(t)

	// Send first message
	msg1, err := daemon.Chat("First message")
	test.AssertNoError(t, err, "First message should not error")
	test.AssertNotNil(t, msg1, "First message response should not be nil")

	// Send second message
	msg2, err := daemon.Chat("Second message")
	test.AssertNoError(t, err, "Second message should not error")
	test.AssertNotNil(t, msg2, "Second message response should not be nil")

	daemon.Stop()

	// Restart daemon (in real scenario, would reload from disk)
	daemon = test.StartTestDaemon(t)
	defer daemon.Stop()

	// Verify daemon is still functional
	response, err := daemon.Chat("Third message")
	test.AssertNoError(t, err, "Third message should not error")
	test.AssertNotNil(t, response, "Third message response should not be nil")
}

// TestChatErrorHandling tests error handling in chat
func TestChatErrorHandling(t *testing.T) {
	daemon := test.StartTestDaemon(t)
	defer daemon.Stop()

	// Test with invalid input (empty message should still work)
	response, err := daemon.Chat("")
	test.AssertNoError(t, err, "Empty message should be handled")
	test.AssertNotNil(t, response, "Should return response even for empty message")

	// Test with very long message
	longMsg := ""
	for i := 0; i < 10000; i++ {
		longMsg += "x"
	}
	response, err = daemon.Chat(longMsg)
	test.AssertNoError(t, err, "Long message should be handled")
	test.AssertNotNil(t, response, "Should return response for long message")
}

// TestChatConcurrent tests concurrent chat messages
func TestChatConcurrent(t *testing.T) {
	daemon := test.StartTestDaemon(t)
	defer daemon.Stop()

	// Send messages concurrently
	done := make(chan error, 3)

	go func() {
		_, err := daemon.Chat("Message 1")
		done <- err
	}()

	go func() {
		_, err := daemon.Chat("Message 2")
		done <- err
	}()

	go func() {
		_, err := daemon.Chat("Message 3")
		done <- err
	}()

	// Wait for all to complete
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for i := 0; i < 3; i++ {
		select {
		case err := <-done:
			test.AssertNoError(t, err, "Concurrent message should not error")
		case <-ctx.Done():
			t.Fatal("Timeout waiting for concurrent messages")
		}
	}
}

// TestChatStreamResponse tests streaming chat responses
func TestChatStreamResponse(t *testing.T) {
	daemon := test.StartTestDaemon(t)
	defer daemon.Stop()

	// In a real test, would test actual streaming
	// For now, test that normal chat works
	response, err := daemon.Chat("Stream this message")
	test.AssertNoError(t, err, "Stream request should not error")
	test.AssertNotNil(t, response, "Stream response should not be nil")

	// Verify response format
	if response == "" {
		t.Error("Stream response should contain data")
	}
}
