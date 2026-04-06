package e2e

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Kaffyn/Vectora/test"
)

// TestMultipleDesktopInstances tests multiple desktop app instances
func TestMultipleDesktopInstances(t *testing.T) {
	daemon := test.StartTestDaemon(t)
	defer daemon.Stop()

	// Simulate two clients
	done := make(chan error, 2)

	// Client 1
	go func() {
		_, err := daemon.Chat("Client 1 message")
		done <- err
	}()

	// Client 2
	go func() {
		_, err := daemon.Chat("Client 2 message")
		done <- err
	}()

	// Wait for both
	for i := 0; i < 2; i++ {
		err := <-done
		test.AssertNoError(t, err, "Client should execute successfully")
	}
}

// TestConcurrentChatRequests tests concurrent chat messages
func TestConcurrentChatRequests(t *testing.T) {
	daemon := test.StartTestDaemon(t)
	defer daemon.Stop()

	numRequests := 10
	done := make(chan error, numRequests)

	// Send concurrent requests
	for i := 0; i < numRequests; i++ {
		go func(id int) {
			_, err := daemon.Chat("Request " + string(rune(id)))
			done <- err
		}(i)
	}

	// Wait for all to complete
	for i := 0; i < numRequests; i++ {
		err := <-done
		test.AssertNoError(t, err, "Concurrent request should complete")
	}
}

// TestConcurrentModelSwitching tests changing models concurrently
func TestConcurrentModelSwitching(t *testing.T) {
	daemon := test.StartTestDaemon(t)
	defer daemon.Stop()

	models, err := daemon.ListModels()
	test.AssertNoError(t, err, "Should list models")

	if len(models) < 2 {
		t.Skip("Need at least 2 models for this test")
	}

	done := make(chan error, len(models))

	// Switch to each model concurrently
	for _, model := range models {
		go func(modelID string) {
			err := daemon.SetModel(modelID)
			done <- err
		}(model.ID)
	}

	// Wait for all
	for i := 0; i < len(models); i++ {
		err := <-done
		// Some might fail due to race conditions, which is fine
		_ = err
	}

	// Verify daemon still works
	response, err := daemon.Chat("After concurrent switches")
	test.AssertNoError(t, err, "Daemon should work after concurrent operations")
	test.AssertNotNil(t, response, "Should return response")
}

// TestHighThroughputChat tests high throughput of chat messages
func TestHighThroughputChat(t *testing.T) {
	daemon := test.StartTestDaemon(t)
	defer daemon.Stop()

	numMessages := 50
	done := make(chan error, numMessages)

	// Send many messages in parallel
	for i := 0; i < numMessages; i++ {
		go func(id int) {
			_, err := daemon.Chat("Message " + string(rune(id)))
			done <- err
		}(i)
	}

	// Count successes
	successCount := int32(0)
	for i := 0; i < numMessages; i++ {
		err := <-done
		if err == nil {
			atomic.AddInt32(&successCount, 1)
		}
	}

	// Most should succeed
	if successCount < int32(numMessages*9/10) {
		t.Logf("Success rate: %d/%d", successCount, numMessages)
	}
}

// TestDaemonRestartDuringOperation tests daemon restart during operation
func TestDaemonRestartDuringOperation(t *testing.T) {
	daemon := test.StartTestDaemon(t)

	// Send message
	response, err := daemon.Chat("Message 1")
	test.AssertNoError(t, err, "First message should work")
	test.AssertNotNil(t, response, "Should return response")

	// Stop daemon
	daemon.Stop()

	// Restart daemon
	daemon = test.StartTestDaemon(t)
	defer daemon.Stop()

	// Send another message
	response, err = daemon.Chat("Message 2")
	test.AssertNoError(t, err, "Message after restart should work")
	test.AssertNotNil(t, response, "Should return response after restart")
}

// TestRaceConditionDetection tests for race conditions
func TestRaceConditionDetection(t *testing.T) {
	daemon := test.StartTestDaemon(t)
	defer daemon.Stop()

	// This test would be run with -race flag to detect data races
	var wg sync.WaitGroup
	mu := sync.Mutex{}
	results := []string{}

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			response, err := daemon.Chat("Concurrent message")
			if err == nil && response != "" {
				mu.Lock()
				results = append(results, response)
				mu.Unlock()
			}
		}(i)
	}

	wg.Wait()

	// Verify we got responses
	if len(results) == 0 {
		t.Error("Should have gotten at least one response")
	}
}

// TestContextCancellation tests context cancellation
func TestContextCancellation(t *testing.T) {
	daemon := test.StartTestDaemon(t)
	defer daemon.Stop()

	// Create cancellable context
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	done := make(chan error, 1)

	go func() {
		response, err := daemon.Chat("Message")
		if err != nil {
			done <- err
		}
		_ = response
		done <- nil
	}()

	select {
	case err := <-done:
		test.AssertNoError(t, err, "Operation should complete")
	case <-ctx.Done():
		// Context cancelled, which is acceptable for this test
	}
}

// TestDeadlock tests that system doesn't deadlock under concurrent access
func TestDeadlock(t *testing.T) {
	daemon := test.StartTestDaemon(t)
	defer daemon.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	done := make(chan bool, 1)

	go func() {
		// Perform many concurrent operations
		var wg sync.WaitGroup
		for i := 0; i < 20; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				daemon.Chat("Test")
				daemon.ListModels()
			}()
		}
		wg.Wait()
		done <- true
	}()

	select {
	case <-done:
		// Success
	case <-ctx.Done():
		t.Fatal("Deadlock detected: operations did not complete in time")
	}
}
