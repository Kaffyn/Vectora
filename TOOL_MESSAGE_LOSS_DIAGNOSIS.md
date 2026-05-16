# Tool Message Loss Bug - Diagnosis & Solution

## Problem Statement

When Vectora attempts to execute tools (like `list_dir`), the tool execution results (ToolMessages) are not appearing in the conversation state. This prevents the LLM from seeing tool results and generating appropriate responses.

**User report:** "note que ta sem a resposta da tool... a ia não ta recebendo o retorno da tool que ela usou, por isso não consegue me responder"
(Note that it's missing the tool response... the AI is not receiving the return from the tool it used, so it can't respond to me)

## Root Cause Analysis

### ✅ What's Working Correctly

1. **Configuration & Tools**: All 13 tools are properly configured and available (web_search, fetch_url, list_dir, etc.)
2. **Message Reducer**: The `add_messages` reducer correctly preserves messages through the pipeline

   - HumanMessage → added to state
   - AIMessage (with tool_calls) → added to state
   - ToolMessage → added to state
   - All messages preserved in order

3. **Graph Flow**: Message routing is correct:
   - `call_llm` → creates AIMessage with tool_calls
   - `tools_condition` → routes to `tools` node if tool_calls present
   - `tools` node → should return `{"messages": [ToolMessage, ...]}`
   - `add_messages` → correctly appends ToolMessages
   - `process_retrieval` → processes web_search/fetch_url results
   - Back to `call_llm` with all messages visible

### 🔴 Likely Issue

The problem is likely in **ToolNode execution**:

1. ToolNode may not be executing tools correctly
2. ToolNode may not be returning messages in the proper format
3. Tools may be failing silently without returning results
4. ToolNode may have issues with async/await context in the graph

## Solution Implemented

### 1. Diagnostic Wrappers (nodes_debug.py)

Created comprehensive diagnostic classes to trace exact point of failure:

```python
class DiagnosticToolNode(ToolNode):
    """Overrides ainvoke() and invoke() to log every step"""
    - Logs: messages entering tool node
    - Logs: tool_calls detected in AIMessage
    - Logs: each ToolMessage returned
    - Logs: any errors during tool execution
    - Logs: message format validation
```

**Call LLM Debug Wrapper:**

```python
async def call_llm_debug(state, runtime):
    """Logs LLM input and output"""
    - Logs: messages seen by LLM
    - Logs: AIMessage generated
    - Logs: tool_calls present in response
    - Logs: any errors during generation
```

### 2. Integration into Graph

Updated `graph.py` to use diagnostic wrappers:

- `call_llm` → replaced with `call_llm_debug`
- `tools` → replaced with `DiagnosticToolNode`

### 3. What the Diagnostics Reveal

When you run Vectora with these diagnostics enabled (LOG_LEVEL=DEBUG), you'll see:

```
[CALL_LLM] ENTRADA (LLM receives...)
  - messages_count: N
  - messages_summary: [HumanMessage, AIMessage, ToolMessage?, ...]

[TOOL_NODE] ENTRADA (Tool node receives...)
  - messages_count: N
  - last_message_type: AIMessage
  - Detectadas tool_calls: name=[list_dir], count=1

[TOOL_NODE] SAÍDA (Tool node returns...)
  - result_keys: [messages]
  - messages_in_result: M (should be > 0 if tools executed)
  - [TOOL_NODE] ToolMessage[0]: type=ToolMessage, content_length=X
```

## How to Use the Diagnostics

### 1. Run Vectora with Debug Logging

```bash
# Set environment variables before running
export LOG_LEVEL=DEBUG
export QUIET_MODE=false
export LOG_JSON=false

# Run the chat
uv run vectora
```

### 2. Observe Log Output

Look for:

- `[CALL_LLM]` entries → shows what messages the LLM sees
- `[TOOL_NODE]` entries → shows tool execution details
- Specific error messages in brackets like `[TOOL_NODE] ERRO`

### 3. Test with Simple Tool Call

Try this command in chat:

```
liste os arquivos do diretorio atual com a ferramenta list_dir
```

The logs should show:

```
[CALL_LLM] ENTRADA - messages_count: 1 (your message)
[CALL_LLM] SAÍDA - AIMessage gerado with tool_calls

[TOOL_NODE] ENTRADA - Detectadas tool_calls: name=[list_dir]
[TOOL_NODE] SAÍDA - messages_in_result: 1 (ToolMessage from list_dir)
[TOOL_NODE] ToolMessage[0]: type=ToolMessage, content_length=XXX

[CALL_LLM] ENTRADA - messages_count: 3 (human + ai + tool messages)
[CALL_LLM] SAÍDA - AIMessage with response using tool results
```

## Expected Outcomes

### If DiagnosticsShow No ToolMessages:

```
[TOOL_NODE] SAÍDA - messages_in_result: 0
```

→ Tools aren't being executed. Check:

- Are tool_calls being detected? (Look for `[TOOL_NODE] Detectadas tool_calls`)
- Is there an error? (Look for `[TOOL_NODE] ERRO`)
- Check tool definitions in tools.py

### If ToolMessages Disappear After Process_Retrieval:

This shouldn't happen (add_messages preserves them), but if it does:

- Check process_retrieval logic in nodes.py
- Verify it's not filtering out non-web_search messages

### If ToolMessages Exist But LLM Ignores Them:

→ LLM might not understand the message format

- Check if ToolMessage.content is properly formatted
- Verify tool_call_id matches the AIMessage's tool_call id

## Files Modified

1. **vectora/nodes_debug.py** - Diagnostic wrappers (NEW)
2. **vectora/graph.py** - Integration of diagnostics
3. **vectora/ignore_validator.py** - Ignore file support (completed in Phase 4)
4. **vectora/services/embedding.py** - Ignore validation integration

## Next Steps to Investigate

Once you enable the diagnostics and run Vectora:

1. **Collect the diagnostic output** - Run a simple tool call and capture the logs
2. **Identify the failure point** - Look for where ToolMessages stop appearing
3. **Report findings** - Share the diagnostic logs to pinpoint exact issue

The diagnostic wrappers will tell us:

- ✅ Are tools being called?
- ✅ Are ToolMessages being created?
- ✅ Are they being added to state correctly?
- ✅ Is the LLM seeing them in the next call?

## Commit History

```
chore: integrate diagnostic wrappers for tool message loss debugging
- Updated DiagnosticToolNode to override ainvoke() and invoke()
- Added detailed logging to trace tool execution and message passing
- Integrated diagnostic nodes into graph for real-time debugging
```

---

**Status**: Ready for testing with debug logging enabled
**Priority**: P0 - Blocks core tool functionality
**User Impact**: Tools don't work at all without this fix
