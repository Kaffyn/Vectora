# Phase 3: Testing & Validation - Test Plan

**Date**: 2026-04-12
**Status**: 🧪 IN PROGRESS
**Objective**: Validate MCP protocol implementation and all tools

---

## Test Strategy

### Test Categories

1. **Protocol Tests** (JSON-RPC 2.0 compliance)
2. **Tool Tests** (All 11 tools discoverable and callable)
3. **Error Handling Tests** (Edge cases and failures)
4. **Performance Tests** (Response times and resource usage)

### Test Approach

- **Manual Tests**: Using stdin/stdout directly with vectora mcp
- **Automated Tests**: Bash scripts simulating Claude Code
- **Validation**: Verify JSON responses match specifications

---

## Test Cases

### Category 1: Protocol Tests

#### Test 1.1: Initialize Request
```
Input:  {"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}
Expected:
  - Valid JSON-RPC 2.0 response
  - protocolVersion: "2024-11-05"
  - serverInfo.name: "Vectora Core"
  - Status: ✅ PENDING
```

#### Test 1.2: Tools List
```
Input:  {"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}
Expected:
  - Returns array of 11 tools
  - Each tool has: name, description, inputSchema
  - Status: ✅ PENDING
```

#### Test 1.3: Invalid JSON
```
Input:  invalid json without braces
Expected:
  - Error code: -32700
  - Message: "Parse error: invalid JSON - ..."
  - Status: ✅ PENDING
```

#### Test 1.4: Missing Method
```
Input:  {"jsonrpc":"2.0","id":3}
Expected:
  - Error code: -32600
  - Message: "Invalid Request: method is required"
  - Status: ✅ PENDING
```

#### Test 1.5: Wrong JSON-RPC Version
```
Input:  {"jsonrpc":"1.0","id":4,"method":"initialize"}
Expected:
  - Error code: -32600
  - Message: "Invalid Request: jsonrpc must be '2.0', got '1.0'"
  - Status: ✅ PENDING
```

#### Test 1.6: Unknown Method
```
Input:  {"jsonrpc":"2.0","id":5,"method":"unknown/method","params":{}}
Expected:
  - Error code: -32601
  - Message: Shows valid methods
  - Status: ✅ PENDING
```

---

### Category 2: Tool Tests

#### Test 2.1: Tool Discovery
```
Goal: Verify all 11 tools are discoverable
Tools Expected:
  1. embed
  2. search_database
  3. web_search_and_embed
  4. web_fetch_and_embed
  5. plan_mode
  6. refactor_with_context
  7. analyze_code_patterns
  8. knowledge_graph_analysis
  9. doc_coverage_analysis
  10. test_generation
  11. bug_pattern_detection
Status: ✅ PENDING
```

#### Test 2.2: Tool Call with Missing Required Parameter
```
Input:  {"jsonrpc":"2.0","id":6,"method":"tools/call","params":{"name":"embed","input":{}}}
Expected:
  - Tool should validate required parameters
  - Clear error message about missing field
  - Status: ✅ PENDING
```

#### Test 2.3: Tool Call with Non-existent Tool
```
Input:  {"jsonrpc":"2.0","id":7,"method":"tools/call","params":{"name":"fake_tool","input":{}}}
Expected:
  - Error shows list of available tools
  - Helpful message
  - Status: ✅ PENDING
```

---

### Category 3: Error Handling Tests

#### Test 3.1: JSON Marshaling Error
```
Goal: Verify server handles JSON errors gracefully
Action: Trigger edge case in response handling
Expected: Server continues functioning, logs error
Status: ✅ PENDING
```

#### Test 3.2: Timeout Handling
```
Goal: Verify 5-minute timeout works
Action: Monitor long-running tool execution
Expected: Times out after 5 minutes, returns error
Status: ✅ PENDING
```

#### Test 3.3: Context Cancellation
```
Goal: Verify graceful shutdown
Action: Send SIGTERM to MCP server
Expected: Logs shutdown message, exits cleanly
Status: ✅ PENDING
```

---

### Category 4: Performance Tests

#### Test 4.1: Response Time - initialize
```
Goal: Measure initialize response time
Expected: < 100ms
Status: ✅ PENDING
```

#### Test 4.2: Response Time - tools/list
```
Goal: Measure tools/list response time
Expected: < 100ms
Status: ✅ PENDING
```

#### Test 4.3: Message Throughput
```
Goal: Send 100 valid requests
Expected: All succeed, no memory leaks
Status: ✅ PENDING
```

---

## Test Execution Plan

### Setup
1. Build Vectora core: `go build -o core.exe ./cmd/core`
2. Create test workspace
3. Start MCP server in background

### Execution Order
1. Protocol tests (1.1-1.6)
2. Tool discovery test (2.1)
3. Error handling tests (3.1-3.3)
4. Performance tests (4.1-4.3)

### Validation Criteria

| Category | Pass Criteria | Weight |
|----------|---------------|--------|
| Protocol | 6/6 tests pass | 30% |
| Tools | All 11 discoverable | 30% |
| Errors | Graceful handling | 20% |
| Performance | Response times acceptable | 20% |

**Overall Pass**: 90% of tests passing

---

## Test Results Template

```markdown
# Test Results - [Date]

## Protocol Tests
- [ ] 1.1 Initialize
- [ ] 1.2 Tools List
- [ ] 1.3 Invalid JSON
- [ ] 1.4 Missing Method
- [ ] 1.5 Wrong Version
- [ ] 1.6 Unknown Method

Result: X/6 ✅

## Tool Tests
- [ ] 2.1 Tool Discovery (11/11)
- [ ] 2.2 Missing Parameter
- [ ] 2.3 Non-existent Tool

Result: X/3 ✅

## Error Handling Tests
- [ ] 3.1 JSON Error
- [ ] 3.2 Timeout
- [ ] 3.3 Graceful Shutdown

Result: X/3 ✅

## Performance Tests
- [ ] 4.1 Initialize < 100ms
- [ ] 4.2 Tools/list < 100ms
- [ ] 4.3 100 requests throughput

Result: X/3 ✅

## Summary
Total: 15/15 tests ✅ PASSED
Status: PRODUCTION READY
```

---

## Tools for Testing

### Manual Testing
```bash
# Start server
vectora mcp /workspace &

# Test via stdin
echo '{"jsonrpc":"2.0","id":1,"method":"initialize"}' | nc localhost 9000

# Or use pipes
(echo '{"jsonrpc":"2.0","id":1,"method":"initialize"}') | vectora mcp /workspace
```

### Automated Testing
Create bash script to:
1. Start server
2. Send test messages
3. Validate responses
4. Collect results
5. Generate report

---

## Success Criteria

### Must Pass ✅
- Protocol compliance (JSON-RPC 2.0)
- All 11 tools discoverable
- Error messages are helpful
- No crashes on invalid input

### Should Pass ✅
- Response times < 1 second
- Clean shutdown
- Debug logging works

### Nice to Have ✨
- Performance under load
- Memory usage acceptable
- Timeout handling verified

---

## Defect Tracking

If tests fail, document:
1. Test case ID
2. Expected vs Actual
3. Error message
4. Reproducibility steps
5. Severity (Critical/High/Medium/Low)

---

## Next Steps After Phase 3

✅ Phase 3 Complete → Phase 4: Examples & Workflows
- Create 4 workflow examples
- Document patterns
- Show advanced usage

---

_Phase 3 Testing Plan_
_Ready for execution_
