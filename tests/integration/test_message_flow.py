"""Test the message flow through add_messages reducer."""

from langchain_core.messages import AIMessage, HumanMessage, ToolMessage
from langgraph.graph.message import add_messages

print("Testing add_messages reducer behavior:")
print("=" * 80)

# Start with initial messages
initial_messages = [
    HumanMessage(content="List files"),
]

print("\n1. Initial state:")
print(f"   Messages: {[type(m).__name__ for m in initial_messages]}")

# Simulate what call_llm returns
ai_message = AIMessage(
    content="I'll list the files",
    tool_calls=[{"name": "list_dir", "args": {"path": "."}, "id": "call_1"}],
)

print("\n2. call_llm returns AIMessage with tool_calls")
result_from_call_llm = {"messages": [ai_message]}
updated_1 = add_messages(initial_messages, result_from_call_llm["messages"])  # type: ignore[arg-type]
assert isinstance(updated_1, (list, tuple))
print(f"   After add_messages: {[type(m).__name__ for m in updated_1]}")
print(
    f"   Last message has tool_calls: {bool(getattr(updated_1[-1], 'tool_calls', None))}"
)

# Simulate what ToolNode returns
tool_message = ToolMessage(
    content="[FILE] file1.txt\n[FILE] file2.py", tool_call_id="call_1", name="list_dir"
)

print("\n3. tools node (ToolNode) returns ToolMessage")
result_from_tools = {"messages": [tool_message]}
updated_2 = add_messages(updated_1, result_from_tools["messages"])  # type: ignore[arg-type]
print(f"   After add_messages: {[type(m).__name__ for m in updated_2]}")
print(f"   ToolMessage present: {any(isinstance(m, ToolMessage) for m in updated_2)}")

# Check what process_retrieval would return
print("\n4. process_retrieval returns retrieval_results")
result_from_process = {"retrieval_results": {}}
# Note: process_retrieval doesn't add messages, only updates retrieval_results
print("   Retrieval updates state but not messages")

# Simulate the next call_llm seeing all messages
print("\n5. call_llm sees all messages in next call:")
all_messages_at_next_call = updated_2
print(f"   Messages available: {[type(m).__name__ for m in all_messages_at_next_call]}")

# Verify the ToolMessage is still there
tool_messages = [m for m in all_messages_at_next_call if isinstance(m, ToolMessage)]
print(f"\n6. ToolMessages in final state: {len(tool_messages)}")
if tool_messages:
    for tm in tool_messages:
        print(f"   - {tm.name}: {tm.content[:50]}")
    print("\n[OK] ToolMessages are preserved through add_messages!")
else:
    print("[ERROR] No ToolMessages found!")
