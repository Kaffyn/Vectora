# Expected Log Output for Embedding Test

## What We're Testing

Running: `test_embedding_with_logs.py` with DEBUG logging to trace:

1. LLM understanding the embedding request
2. LLM deciding to use the `ingest_docs` tool
3. ToolNode executing the tool
4. ToolMessage being created and returned
5. LLM receiving the tool result and responding

## Expected Log Flow

### Phase 1: Initial Setup

```
[CALL_LLM] ENTRADA
  - messages_count: 1
  - messages_summary: [HumanMessage with embedding request]
```

### Phase 2: LLM Response with Tool Call

```
[CALL_LLM] SAÍDA
  - AIMessage gerado
  - has_tool_calls: True
  - tool_calls_count: 1
  - tool names: [ingest_docs]
```

### Phase 3: Tool Execution

```
[TOOL_NODE] ENTRADA
  - messages_count: 2 (human + ai messages)
  - last_message_type: AIMessage

[TOOL_NODE] Detectadas tool_calls
  - tool_calls_count: 1
  - tool_names: [ingest_docs]
```

### Phase 4: Tool Execution & Result

```
[TOOL_NODE] SAÍDA
  - result_keys: [messages]
  - messages_in_result: 1 (THIS IS KEY - should show ToolMessage was created)

[TOOL_NODE] ToolMessage[0]
  - type: ToolMessage
  - tool_use_id: (should match the tool call id)
  - content_length: XXX
```

### Phase 5: Next LLM Call with Tool Result

```
[CALL_LLM] ENTRADA
  - messages_count: 3 (human + ai + tool messages)
  - messages_summary shows: [HumanMessage, AIMessage, ToolMessage]

[CALL_LLM] SAÍDA
  - AIMessage with final response using embedding results
```

## Key Indicators to Look For

### ✅ Tool Execution Success

- `[TOOL_NODE] Detectadas tool_calls` appears
- `[TOOL_NODE] SAÍDA - messages_in_result: 1` (or more)
- `[TOOL_NODE] ToolMessage[0]` shows the result

### ⚠️ Tool Not Executing

- `[TOOL_NODE] Nenhuma tool_call detectada` means LLM didn't create tool_calls
- `[TOOL_NODE] SAÍDA - messages_in_result: 0` means no ToolMessages created

### 🔴 Tool Error

- `[TOOL_NODE] ERRO ao executar ferramentas` with error message
- Error will show what failed in the tool execution

## What NOT to Expect

If you see messages like:

- `"Missing required config key"` - we fixed this with DiagnosticToolNode
- `"ToolNode doesn't have __call__"` - we fixed this by using ainvoke()
- Messages disappearing between nodes - add_messages preserves them

## How to Analyze the Output

1. **Search for `[TOOL_NODE] Detectadas tool_calls`**

   - If found: Tool was detected ✅
   - If not found: LLM didn't generate tool_calls ❌

2. **Check `[TOOL_NODE] SAÍDA - messages_in_result: X`**

   - If X > 0: Tool executed and returned messages ✅
   - If X = 0: Tool node ran but didn't return messages ❌

3. **Look for `[TOOL_NODE] ToolMessage[0]`**

   - If found: ToolMessage was created ✅
   - If not found: ToolMessage wasn't created ❌

4. **Count final messages in SUMMARY**
   - Should see: 1 HumanMessage, 2 AIMessages, 1+ ToolMessages
   - If missing ToolMessages: they disappeared somewhere ❌

## Common Issues & What They Mean

| Issue               | Log Evidence                                 | Solution                                          |
| ------------------- | -------------------------------------------- | ------------------------------------------------- |
| LLM won't use tools | No `Detectadas tool_calls`                   | Check prompt, system message, or model capability |
| Tool not executing  | `messages_in_result: 0`                      | Check if tool name is recognized, args valid      |
| Tool error          | `ERRO ao executar ferramentas: ...`          | Check error message in logs                       |
| ToolMessages lost   | Appear in TOOL_NODE but not in next CALL_LLM | Bug in graph routing (shouldn't happen)           |
| Wrong tool result   | `ToolMessage[0]` shows unexpected content    | Tool executed but returned wrong data             |

## Debugging Commands

If you see issues, try:

1. **Check tool exists**: Look for `Tools initialized` log with count
2. **Check specific tool**: Search for `ingest_docs` in logs
3. **Check tool config**: Look for `enable_file_operations: True` and `enable_rag: True`
4. **Check embedding queue**: Look for `embedding_queue_enabled: True`

## Expected Final Output

```
MESSAGE SUMMARY:
  AIMessage: 2
  HumanMessage: 1
  ToolMessage: 1

ToolMessages present: True
[OK] Tool results captured in state!
```

If you get:

```
ToolMessages present: False
[ERROR] No ToolMessages found - tools may not have executed
```

Then we've identified the exact failure point that needs fixing.
