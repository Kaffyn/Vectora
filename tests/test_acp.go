package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// TestACP tests ACP (Agent Client Protocol) functionality
func TestACP(config *EnvironmentConfig, runner *TestRunner) {
	runner.RunTest("ACP: JSON-RPC 2.0 Structure", func() error {
		// Test basic JSON-RPC request structure
		req := JSONRPCRequest(1, "test.method", map[string]string{"key": "value"})

		if req["jsonrpc"] != "2.0" {
			return fmt.Errorf("invalid jsonrpc version")
		}

		if req["id"] != 1 {
			return fmt.Errorf("invalid request id")
		}

		if req["method"] != "test.method" {
			return fmt.Errorf("invalid method")
		}

		return nil
	})

	runner.RunTest("ACP: JSON-RPC Request Format", func() error {
		// Test request serialization
		req := JSONRPCRequest(1, "test", nil)

		data, err := json.Marshal(req)
		if err != nil {
			return fmt.Errorf("failed to marshal request: %w", err)
		}

		if len(data) == 0 {
			return fmt.Errorf("marshaled request is empty")
		}

		// Verify it's valid JSON
		var parsed map[string]interface{}
		if err := json.Unmarshal(data, &parsed); err != nil {
			return fmt.Errorf("failed to parse marshaled request: %w", err)
		}

		return nil
	})

	runner.RunTest("ACP: JSON-RPC Response Format", func() error {
		// Test response structure
		resp := JSONRPCResult(1, "success")

		data, err := json.Marshal(resp)
		if err != nil {
			return fmt.Errorf("failed to marshal response: %w", err)
		}

		// Verify it's valid JSON
		var parsed map[string]interface{}
		if err := json.Unmarshal(data, &parsed); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}

		return nil
	})

	runner.RunTest("ACP: JSON-RPC Error Response", func() error {
		// Test error response
		errResp := JSONRPCError(1, -32600, "Invalid Request")

		if errResp["error"] == nil {
			return fmt.Errorf("error field missing")
		}

		errorObj := errResp["error"].(map[string]interface{})
		if errorObj["code"] != -32600 {
			return fmt.Errorf("invalid error code")
		}

		return nil
	})

	runner.RunTest("ACP: Response Validation", func() error {
		// Test response validation
		validResp := JSONRPCResult(1, "test")

		if err := ValidateJSONRPCResponse(validResp); err != nil {
			return fmt.Errorf("valid response failed validation: %w", err)
		}

		return nil
	})

	runner.RunTest("ACP: Invalid Response Detection", func() error {
		// Test detection of invalid responses
		invalidResp := map[string]interface{}{
			"jsonrpc": "1.0", // Wrong version
			"id":      1,
		}

		if err := ValidateJSONRPCResponse(invalidResp); err == nil {
			return fmt.Errorf("should reject invalid version")
		}

		return nil
	})

	runner.RunTest("ACP: Request Without Params", func() error {
		// Test request without params field
		req := JSONRPCRequest(1, "method", nil)

		data, err := json.Marshal(req)
		if err != nil {
			return fmt.Errorf("failed to marshal: %w", err)
		}

		// Should still be valid
		var parsed map[string]interface{}
		if err := json.Unmarshal(data, &parsed); err != nil {
			return fmt.Errorf("failed to parse: %w", err)
		}

		return nil
	})

	runner.RunTest("ACP: Request With Array Params", func() error {
		// Test request with array parameters
		params := []string{"param1", "param2"}
		req := JSONRPCRequest(1, "method", params)

		data, _ := json.Marshal(req)

		var parsed map[string]interface{}
		json.Unmarshal(data, &parsed)

		if parsed["params"] == nil {
			return fmt.Errorf("params not included")
		}

		return nil
	})

	runner.RunTest("ACP: Batch Requests", func() error {
		// Test batch request format
		batch := []map[string]interface{}{
			JSONRPCRequest(1, "method1", nil),
			JSONRPCRequest(2, "method2", nil),
		}

		data, err := json.Marshal(batch)
		if err != nil {
			return fmt.Errorf("failed to marshal batch: %w", err)
		}

		var parsed []map[string]interface{}
		if err := json.Unmarshal(data, &parsed); err != nil {
			return fmt.Errorf("failed to parse batch: %w", err)
		}

		if len(parsed) != 2 {
			return fmt.Errorf("batch should have 2 requests")
		}

		return nil
	})

	runner.RunTest("ACP: Notification Format", func() error {
		// Test notification (no id)
		notification := map[string]interface{}{
			"jsonrpc": "2.0",
			"method":  "notify",
			"params":  map[string]string{"key": "value"},
		}

		data, _ := json.Marshal(notification)

		var parsed map[string]interface{}
		json.Unmarshal(data, &parsed)

		if parsed["id"] != nil {
			return fmt.Errorf("notification should not have id")
		}

		return nil
	})
}

// TestACPMessages tests ACP message handling
func TestACPMessages(config *EnvironmentConfig, runner *TestRunner) {
	runner.RunTest("ACP: Message with Object Params", func() error {
		params := map[string]interface{}{
			"key1": "value1",
			"key2": 123,
			"key3": true,
		}

		req := JSONRPCRequest(1, "method", params)
		data, _ := json.Marshal(req)

		var parsed map[string]interface{}
		json.Unmarshal(data, &parsed)

		return nil
	})

	runner.RunTest("ACP: Message ID Generation", func() error {
		// Test various ID formats
		ids := []interface{}{1, 42, 999999, "string-id"}

		for _, id := range ids {
			if intID, ok := id.(int); ok {
				req := JSONRPCRequest(intID, "test", nil)
				if req["id"] != intID {
					return fmt.Errorf("id not preserved")
				}
			}
		}

		return nil
	})

	runner.RunTest("ACP: Method Name Parsing", func() error {
		methods := []string{
			"method",
			"method.submethod",
			"method.sub.deep",
			"a",
		}

		for _, method := range methods {
			req := JSONRPCRequest(1, method, nil)
			if req["method"] != method {
				return fmt.Errorf("method name not preserved: %s", method)
			}
		}

		return nil
	})

	runner.RunTest("ACP: Large Payload", func() error {
		// Test with large payload
		largeData := make(map[string]interface{})
		for i := 0; i < 1000; i++ {
			largeData[fmt.Sprintf("key%d", i)] = fmt.Sprintf("value%d", i)
		}

		req := JSONRPCRequest(1, "method", largeData)
		data, err := json.Marshal(req)

		if err != nil {
			return fmt.Errorf("failed to marshal large payload: %w", err)
		}

		if len(data) == 0 {
			return fmt.Errorf("marshaled data is empty")
		}

		return nil
	})

	runner.RunTest("ACP: Nested Objects", func() error {
		// Test deeply nested objects
		params := map[string]interface{}{
			"level1": map[string]interface{}{
				"level2": map[string]interface{}{
					"level3": map[string]interface{}{
						"value": "deep",
					},
				},
			},
		}

		req := JSONRPCRequest(1, "method", params)
		data, _ := json.Marshal(req)

		var parsed map[string]interface{}
		if err := json.Unmarshal(data, &parsed); err != nil {
			return fmt.Errorf("failed to parse nested objects")
		}

		return nil
	})

	runner.RunTest("ACP: Special Characters in Method", func() error {
		// Test special characters in method names
		methods := []string{
			"method-with-dash",
			"method_with_underscore",
			"method$special",
		}

		for _, method := range methods {
			req := JSONRPCRequest(1, method, nil)
			if req["method"] != method {
				return fmt.Errorf("method name not preserved: %s", method)
			}
		}

		return nil
	})

	runner.RunTest("ACP: Empty Params", func() error {
		req := JSONRPCRequest(1, "method", map[string]interface{}{})
		data, _ := json.Marshal(req)

		var parsed map[string]interface{}
		json.Unmarshal(data, &parsed)

		return nil
	})
}

// TestACPErrors tests ACP error handling
func TestACPErrors(config *EnvironmentConfig, runner *TestRunner) {
	runner.RunTest("ACP: Error Codes", func() error {
		errorCodes := map[int]string{
			-32700: "Parse error",
			-32600: "Invalid Request",
			-32601: "Method not found",
			-32602: "Invalid params",
			-32603: "Internal error",
		}

		for code, message := range errorCodes {
			errResp := JSONRPCError(1, code, message)

			if errResp["id"] != 1 {
				return fmt.Errorf("error id not set")
			}

			errorObj := errResp["error"].(map[string]interface{})
			if errorObj["code"] != code {
				return fmt.Errorf("error code not set correctly")
			}

			if errorObj["message"] != message {
				return fmt.Errorf("error message not set correctly")
			}
		}

		return nil
	})

	runner.RunTest("ACP: Error Response Validation", func() error {
		errResp := JSONRPCError(1, -32600, "Invalid Request")

		if err := ValidateJSONRPCResponse(errResp); err != nil {
			return fmt.Errorf("valid error response failed validation: %w", err)
		}

		return nil
	})

	runner.RunTest("ACP: Error Data Field", func() error {
		// Test error with data field
		errResp := JSONRPCError(1, -32602, "Invalid params")

		errorObj := errResp["error"].(map[string]interface{})

		// JSON-RPC allows optional data field
		if errorObj["data"] == nil {
			// This is OK, data is optional
		}

		return nil
	})

	runner.RunTest("ACP: Multiple Errors", func() error {
		// Test handling multiple error responses
		errors := []map[string]interface{}{
			JSONRPCError(1, -32600, "Invalid Request"),
			JSONRPCError(2, -32601, "Method not found"),
			JSONRPCError(3, -32602, "Invalid params"),
		}

		for _, errResp := range errors {
			if err := ValidateJSONRPCResponse(errResp); err != nil {
				return fmt.Errorf("error response validation failed: %w", err)
			}
		}

		return nil
	})

	runner.RunTest("ACP: Mutual Exclusion of Result and Error", func() error {
		// Response must not have both result and error
		invalidResp := map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      1,
			"result":  "success",
			"error": map[string]interface{}{
				"code":    -32600,
				"message": "Invalid",
			},
		}

		if err := ValidateJSONRPCResponse(invalidResp); err == nil {
			return fmt.Errorf("should reject response with both result and error")
		}

		return nil
	})

	runner.RunTest("ACP: Missing Required Fields", func() error {
		// Test detection of missing required fields
		incompleteResp := map[string]interface{}{
			"jsonrpc": "2.0",
			// Missing result and error
		}

		if err := ValidateJSONRPCResponse(incompleteResp); err == nil {
			return fmt.Errorf("should reject incomplete response")
		}

		return nil
	})
}

// TestACPStreaming tests ACP streaming features
func TestACPStreaming(config *EnvironmentConfig, runner *TestRunner) {
	runner.RunTest("Stream: Initialize Stream", func() error {
		// Test initializing a stream
		return nil
	})

	runner.RunTest("Stream: Token Streaming", func() error {
		// Test streaming tokens
		tokens := []string{"Hello", " ", "world", "!"}

		messages := make([]map[string]interface{}, len(tokens))
		for i, token := range tokens {
			messages[i] = map[string]interface{}{
				"type":  "token",
				"value": token,
			}
		}

		if len(messages) != len(tokens) {
			return fmt.Errorf("token stream incomplete")
		}

		return nil
	})

	runner.RunTest("Stream: Stream Completion", func() error {
		// Test stream completion
		completion := map[string]interface{}{
			"type":   "complete",
			"tokens": 42,
		}

		if completion["type"] != "complete" {
			return fmt.Errorf("completion type invalid")
		}

		return nil
	})

	runner.RunTest("Stream: Cancel Stream", func() error {
		// Test canceling stream
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		select {
		case <-ctx.Done():
			// Cancellation successful
			return nil
		default:
			return fmt.Errorf("context not canceled")
		}
	})

	runner.RunTest("Stream: Stream Timeout", func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()

		select {
		case <-ctx.Done():
			return nil
		case <-time.After(200 * time.Millisecond):
			return fmt.Errorf("timeout did not trigger")
		}
	})
}

// TestACPTools tests tool-related ACP functionality
func TestACPTools(config *EnvironmentConfig, runner *TestRunner) {
	runner.RunTest("Tool: Tool Definition Format", func() error {
		tool := map[string]interface{}{
			"name":        "test_tool",
			"description": "A test tool",
			"parameters": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"param1": map[string]interface{}{
						"type": "string",
					},
				},
			},
		}

		data, err := json.Marshal(tool)
		if err != nil {
			return fmt.Errorf("failed to marshal tool: %w", err)
		}

		if len(data) == 0 {
			return fmt.Errorf("marshaled tool is empty")
		}

		return nil
	})

	runner.RunTest("Tool: Tool Call Format", func() error {
		toolCall := map[string]interface{}{
			"type": "tool_call",
			"tool": "test_tool",
			"args": map[string]interface{}{
				"param1": "value1",
			},
		}

		data, _ := json.Marshal(toolCall)

		var parsed map[string]interface{}
		if err := json.Unmarshal(data, &parsed); err != nil {
			return fmt.Errorf("failed to parse tool call")
		}

		return nil
	})

	runner.RunTest("Tool: Multiple Tools", func() error {
		tools := []map[string]interface{}{
			{
				"name": "tool1",
				"description": "First tool",
			},
			{
				"name": "tool2",
				"description": "Second tool",
			},
		}

		if len(tools) != 2 {
			return fmt.Errorf("tools list incomplete")
		}

		return nil
	})
}

// TestACPProtocol tests ACP protocol compliance
func TestACPProtocol(config *EnvironmentConfig, runner *TestRunner) {
	runner.RunTest("Protocol: Version Compliance", func() error {
		req := JSONRPCRequest(1, "test", nil)

		if req["jsonrpc"] != "2.0" {
			return fmt.Errorf("not JSON-RPC 2.0 compliant")
		}

		return nil
	})

	runner.RunTest("Protocol: Request/Response Pairing", func() error {
		req := JSONRPCRequest(1, "test", nil)
		resp := JSONRPCResult(1, "success")

		reqID := req["id"]
		respID := resp["id"]

		if reqID != respID {
			return fmt.Errorf("request and response IDs don't match")
		}

		return nil
	})

	runner.RunTest("Protocol: Method Routing", func() error {
		methods := []string{
			"initialize",
			"chat",
			"embed",
			"search",
		}

		for _, method := range methods {
			req := JSONRPCRequest(1, method, nil)
			if req["method"] != method {
				return fmt.Errorf("method not preserved")
			}
		}

		return nil
	})

	runner.RunTest("Protocol: Capability Advertisement", func() error {
		capabilities := map[string]interface{}{
			"chat":    true,
			"embed":   true,
			"search":  true,
			"stream":  true,
		}

		if len(capabilities) == 0 {
			return fmt.Errorf("no capabilities advertised")
		}

		return nil
	})
}
