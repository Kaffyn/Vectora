# Vectora Context System

## Overview

The **Context** is an immutable data structure passed to the LangGraph conversation that contains:
- **User identification** — who is using Vectora
- **Session metadata** — conversation tracking and persistence
- **User preferences** — language, search settings, model preferences
- **Feature flags** — which tools the user can access

Context is **immutable during a conversation** but can be **created anew between sessions** with different values.

---

## Structure

### Context (Main Class)

```python
@dataclass(frozen=True)
class Context:
    user_id: str = "default"
    user_type: Literal["plus", "enterprise", "pro"] = "plus"
    thread_id: int  # Required: conversation thread ID
    conversation_id: str | None = None
    created_at: str | None = None  # Auto-set to current ISO timestamp
    
    preferences: UserPreferences = field(default_factory=UserPreferences)
    features: FeatureFlags = field(default_factory=FeatureFlags)
```

### UserPreferences (Optional Customization)

```python
@dataclass(frozen=True)
class UserPreferences:
    language: str = "en"                      # Language code (pt_BR, en_US, etc)
    max_search_results: int = 10              # RAG search limit
    min_score_threshold: float = 0.5          # Minimum relevance score
    preferred_model: str | None = None        # Override default LLM model
```

### FeatureFlags (Access Control)

```python
@dataclass(frozen=True)
class FeatureFlags:
    enable_web_search: bool = True
    enable_url_fetch: bool = True
    enable_database: bool = True
    enable_rag: bool = True
    enable_mcp: bool = False
```

---

## Usage Examples

### Basic Context (Minimal)

```python
from context import Context

context = Context(user_type="plus", thread_id=1)
# All other fields use defaults
```

### Context with Custom Preferences

```python
from context import Context, UserPreferences

preferences = UserPreferences(
    language="pt_BR",
    max_search_results=5,
    min_score_threshold=0.7,
    preferred_model="gemini-2.0-pro"
)

context = Context(
    user_id="user_123",
    user_type="enterprise",
    thread_id=1,
    preferences=preferences
)
```

### Context with Feature Access Control

```python
from context import Context, FeatureFlags

features = FeatureFlags(
    enable_web_search=False,  # Disable web search
    enable_database=False,    # Disable database queries
    enable_rag=True,          # Allow RAG
)

context = Context(
    user_id="user_456",
    user_type="plus",
    thread_id=42,
    features=features
)
```

### Passing to LangGraph

```python
# Create context once at conversation start
context = Context(user_type="plus", thread_id=thread_id)

# Pass to graph invocation
result = await graph.ainvoke(
    {"messages": [user_message]},
    config=config,
    context=context  # ← Context passed here
)
```

---

## User Types

The `user_type` field controls which features/models are available:

| Type | Description | Typical Features |
|------|-------------|------------------|
| `plus` | Paid individual tier | All tools enabled |
| `pro` | Premium tier | All tools, faster models |
| `enterprise` | Enterprise tier | All tools, custom models, higher limits |

Custom logic can be added in nodes to route based on user_type:

```python
def call_llm(state: State, runtime: Runtime[Context]) -> State:
    user_type = runtime.context.user_type
    
    if user_type == "enterprise":
        model = "claude-opus-4-1"
    elif user_type == "pro":
        model = "gemini-2.0-pro"
    else:
        model = "gemini-2.0-flash"
    
    # ... rest of node logic
```

---

## Language Preferences

The `preferences.language` field stores the user's language preference:

```python
from context import Context, UserPreferences

prefs = UserPreferences(language="pt_BR")
context = Context(user_type="plus", thread_id=1, preferences=prefs)

# In system prompt injection:
# "Conversation language: pt_BR"
# → AI responds in Portuguese (Brazil)
```

Language codes follow `ISO 639-1_ISO 3166-1` format (e.g., `pt_BR`, `en_US`, `es_ES`).

See [PROMPTS.md](./PROMPTS.md) for detailed language detection and system prompt integration.

---

## Model Override

Users can override the default LLM model:

```python
from context import Context, UserPreferences

prefs = UserPreferences(preferred_model="claude-opus-4-1")
context = Context(user_type="plus", thread_id=1, preferences=prefs)
```

In the `call_llm` node, check for override:

```python
def call_llm(state: State, runtime: Runtime[Context]) -> State:
    model = runtime.context.preferences.preferred_model or get_default_model()
    # ... use model
```

---

## RAG Search Limits

RAG behavior can be tuned per-user:

```python
prefs = UserPreferences(
    max_search_results=20,      # Return up to 20 results
    min_score_threshold=0.6     # Only results ≥ 0.6 similarity
)
context = Context(
    user_id="user_789",
    thread_id=1,
    preferences=prefs
)
```

In the `embedding` or `vector_search` tool, respect these settings:

```python
@tool
async def vector_search(
    query: str,
    collections: list[str] | None = None,
    top_k: int = 10,
    min_score: float = 0.5,
    runtime: ToolRuntime[Context, State] | None = None,
) -> str:
    # Override defaults with user preferences
    if runtime and runtime.context:
        top_k = runtime.context.preferences.max_search_results
        min_score = runtime.context.preferences.min_score_threshold
    
    # ... rest of tool logic
```

---

## Feature Access Control

Disable tools per-user via feature flags:

```python
features = FeatureFlags(
    enable_web_search=False,   # No web access
    enable_database=False,     # No SQL queries
    enable_rag=True,           # Allow RAG
)
context = Context(
    user_id="restricted_user",
    thread_id=1,
    features=features
)
```

In tools, check feature flags:

```python
@tool
async def web_search(query: str, runtime: ToolRuntime[Context, State]) -> str:
    if runtime and not runtime.context.features.enable_web_search:
        return "Web search is disabled for your account"
    # ... rest of tool logic
```

---

## Session Persistence

The `thread_id` and `conversation_id` fields enable multi-turn conversations:

```python
context = Context(
    user_id="user_100",
    thread_id=42,
    conversation_id="conv_abc123",  # External tracking ID
)

# Pass to checkpointer for state persistence
result = await graph.ainvoke(
    {"messages": messages},
    config=RunnableConfig(configurable={"thread_id": context.thread_id}),
    context=context
)

# Later, resume same conversation
context_v2 = Context(
    user_id="user_100",
    thread_id=42,  # Same thread → loads history from checkpointer
)

result_v2 = await graph.ainvoke(
    {"messages": [new_message]},
    config=RunnableConfig(configurable={"thread_id": context_v2.thread_id}),
    context=context_v2
)
```

---

## Immutability and Design

### Why Frozen?

Context is `frozen=True` (immutable) per LangGraph semantics:
- Prevents accidental mutations mid-conversation
- Ensures consistency across node executions
- Makes reasoning about state easier

### Changing Context

To update context between turns:

```python
# Turn 1
context_v1 = Context(user_type="plus", thread_id=1)
result_v1 = await graph.ainvoke(msg1, context=context_v1)

# Turn 2: Create new context with updated values
context_v2 = Context(
    user_type="pro",  # ← Changed
    thread_id=1,      # ← Same (continue conversation)
)
result_v2 = await graph.ainvoke(msg2, context=context_v2)
```

**Do NOT try to mutate context:**

```python
# ❌ This will fail (frozen dataclass)
context.user_type = "pro"

# ✅ Do this instead
context = Context(..., user_type="pro", ...)
```

---

## Accessing Context in Nodes

### In Graph Nodes

```python
def call_llm(state: State, runtime: Runtime[Context]) -> State:
    ctx = runtime.context
    user_type = ctx.user_type
    language = ctx.preferences.language
    can_search = ctx.features.enable_web_search
    # ... use values
```

### In Tools

```python
@tool
async def some_tool(
    arg: str,
    runtime: ToolRuntime[Context, State] | None = None
) -> str:
    if runtime and runtime.context:
        user_id = runtime.context.user_id
        user_type = runtime.context.user_type
        # ... use values
    # ... rest of tool
```

---

## Extending Context

### Adding New User Properties

To add a new user property, extend `Context`:

```python
# Before
class Context:
    user_id: str = "default"

# After
class Context:
    user_id: str = "default"
    email: str | None = None  # ← New field
    organization_id: str | None = None  # ← New field
```

Update all Context instantiations to include the new fields (with defaults for backward compatibility).

### Adding New Preferences

```python
@dataclass(frozen=True)
class UserPreferences:
    language: str = "en"
    max_search_results: int = 10
    # ... existing fields ...
    
    # New fields
    enable_summarization: bool = True
    response_style: Literal["concise", "detailed"] = "detailed"
```

Then update nodes/tools to use the new preferences.

### Adding New Feature Flags

```python
@dataclass(frozen=True)
class FeatureFlags:
    # ... existing fields ...
    
    # New flags
    enable_voice_output: bool = False
    enable_streaming: bool = True
```

---

## Testing with Context

### Creating Test Contexts

```python
import pytest
from context import Context

@pytest.fixture
def basic_context():
    return Context(user_type="plus", thread_id=1)

@pytest.fixture
def enterprise_context():
    return Context(
        user_id="ent_123",
        user_type="enterprise",
        thread_id=1,
    )

def test_something(basic_context):
    # Use context in test
    assert basic_context.user_type == "plus"
```

### Mocking Runtime

```python
from unittest.mock import Mock
from langgraph.prebuilt.tool_node import ToolRuntime
from context import Context

def test_tool_with_context():
    context = Context(user_type="plus", thread_id=1)
    runtime = Mock(spec=ToolRuntime)
    runtime.context = context
    
    # Call tool with mocked runtime
    result = my_tool(arg="test", runtime=runtime)
```

---

## Migration from Old System

Old system had only:
```python
@dataclass
class Context:
    user_type: Literal["plus", "enterprise"] = "plus"
```

Migration path:

1. **Update context.py** — Done ✓
2. **Update instantiations:**
   ```python
   # Old
   context = Context(user_type="plus")
   
   # New
   context = Context(user_type="plus", thread_id=thread_id)
   ```
3. **Update tools (optional)** — Start using feature flags when needed
4. **Update nodes (optional)** — Access preferences for model selection

---

## Related Files

- **src/context.py** — Context class definitions
- **src/main.py** — Context instantiation for CLI
- **src/chat.py** — Context instantiation for TUI
- **src/nodes.py** — Accessing context in call_llm node
- **src/tools.py** — Using context in tools
- **src/testing/fixtures.py** — Test fixture
- **docs/PROMPTS.md** — Language detection and system prompt integration

---

## Best Practices

1. **Create context once per session** — Set it before graph invocation
2. **Use typed preferences** — Don't pass raw dicts
3. **Respect feature flags** — Check before calling restricted tools
4. **Default gracefully** — Tools should work if runtime is None
5. **Log context metadata** — Include user_id, thread_id in logs
6. **Test with fixtures** — Use pytest fixtures for consistent test contexts
7. **Don't mutate** — Create new Context instances for changes
8. **Document extensions** — When adding fields, update this doc

---
