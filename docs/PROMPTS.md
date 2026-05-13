# Vectora System Prompts

## Overview

Vectora uses a language-aware system prompt that automatically detects the user's operating system language and includes it in the prompt. The AI is intelligent enough to interpret language codes (like `pt_BR`, `en_US`, etc.) and respond accordingly. This ensures that even simple queries like "47.23 * 134.97" are answered in the user's native language.

## How It Works

### Language Detection

The system automatically detects the user's language from the operating system locale:

```python
import locale

def get_system_language() -> str:
    """Detect system language from OS locale.
    
    Returns full language code like 'pt_BR', 'en_US', 'en', etc.
    """
    try:
        lang_code, _ = locale.getdefaultlocale()
        if lang_code:
            return lang_code.lower()  # Returns: pt_BR, en_US, fr_FR, etc.
    except Exception:
        pass
    return "en"  # Fallback to English
```

**Examples of detected codes:**
- `pt_BR` - Portuguese (Brazil)
- `pt_PT` - Portuguese (Portugal)
- `en_US` - English (United States)
- `en_GB` - English (Great Britain)
- `es_ES` - Spanish (Spain)
- `es_MX` - Spanish (Mexico)
- `fr_FR` - French (France)
- `de_DE` - German (Germany)
- `it_IT` - Italian (Italy)
- `ja_JP` - Japanese (Japan)
- `zh_CN` - Chinese (Simplified)
- `zh_TW` - Chinese (Traditional)

### Dynamic System Prompt

The system prompt includes the language code at the end for the AI to interpret:

```python
SYSTEM_PROMPT = """# Vectora - Advanced AI Assistant with RAG Capabilities

[... capability descriptions ...]

---

Conversation language: {language_code}
"""

def get_system_prompt() -> str:
    """Get the Vectora system prompt with language auto-detected from OS."""
    lang_code = get_system_language()
    return SYSTEM_PROMPT.format(language_code=lang_code)
```

**Result:** The AI receives a prompt ending with:
```
Conversation language: pt_BR
```

And automatically interprets this as Portuguese (Brazil) and responds accordingly.

### Integration Points

The system prompt is injected into the LLM conversation in two places:

#### 1. Production (src/nodes.py)

```python
def call_llm(state: State, runtime: Runtime[Context]) -> State:
    # ... model setup ...
    
    # Prepend Vectora system prompt with auto-detected language
    system_prompt = SystemMessage(content=get_system_prompt())
    messages_with_system = [system_prompt] + list(state["messages"])
    
    result = llm_with_config.invoke(messages_with_system)
    
    return {"messages": [result]}
```

#### 2. Testing (src/testing/fixtures.py)

```python
async def call_llm_with_mock(state: State, runtime):
    # ... mock setup ...
    
    # Prepend system prompt for consistent behavior in tests
    system_prompt = SystemMessage(content=get_system_prompt())
    messages_with_system = [system_prompt] + list(state["messages"])
    
    result = llm_with_tools.invoke(messages_with_system)
    return {"messages": [result]}
```

## What the System Prompt Contains

The Vectora system prompt specifies:

1. **Identity**: "You are Vectora, an advanced AI assistant with RAG capabilities"

2. **Core Capabilities**:
   - Information Retrieval (RAG) - vector_search, embedding
   - Tool Integration - web_search, fetch_url, query_database, call_mcp_tool
   - Advanced Features - multi-collection search, auto-reranking, embedding queue

3. **Operational Guidelines**:
   - When to use each tool
   - Response style expectations
   - **Language preference** (dynamically set)

4. **Important Notes**:
   - Error handling strategies
   - Fallback approaches
   - Context maintenance

## Adding Support for New Languages

**Great news**: You don't need to add anything! The system automatically supports any language code that your OS can produce.

Just set your system locale to any language, and Vectora will pass that code to the AI. Examples:

```bash
# Linux/Mac - Set to Hindi
export LC_ALL=hi_IN.UTF-8
export LANG=hi_IN.UTF-8

# Windows
# Settings → Region & Language → Language (choose desired language)

# Then run Vectora
uv run src/main.py
```

The AI will automatically receive `hi_IN` in the prompt and respond in Hindi.

### Supported Format

Language codes follow the standard format:
- `{ISO 639-1}_{ISO 3166-1}` (e.g., `pt_BR`, `en_US`)
- Or just `{ISO 639-1}` (e.g., `en`, `pt`)

The AI interprets all of these automatically.

## Customizing the Prompt

To modify the prompt content:

1. Edit `SYSTEM_PROMPT` in `src/prompts.py`
2. Keep the `{language_code}` placeholder at the end
3. Restart Vectora to apply changes

### Prompt Sections

- **Identity & Capabilities**: Lines 1-25
- **Tool Integration Details**: Lines 27-40
- **Operational Guidelines**: Lines 42-60
- **Important Notes**: Lines 62-70
- **Language Code**: Keep placeholder at end (line ~73)

## Debugging

### Check detected language

```python
from prompts import get_system_language, get_system_prompt

print(f"Detected language: {get_system_language()}")
print("\nSystem prompt (last 100 chars):")
print(get_system_prompt()[-100:])
```

### Check system locale

```bash
# Python
python3 -c "import locale; print(locale.getdefaultlocale())"

# Linux/Mac
locale

# Windows PowerShell
[CultureInfo]::CurrentCulture
```

## Examples

### Portuguese User (pt_BR)

When system locale is Portuguese (Brazil), the prompt ends with:

```
---

Conversation language: pt_BR
```

Result: "47.23 * 134.97 = 6.399.0751" em Português (Brasil) ✅

### Spanish User (es_ES)

When system locale is Spanish (Spain), the prompt ends with:

```
---

Conversation language: es_ES
```

Result: "47.23 * 134.97 = 6.399.0751 en Español (España)" ✅

### English User (en_US)

When system locale is English (United States), the prompt ends with:

```
---

Conversation language: en_US
```

Result: "47.23 * 134.97 = 6,399.0751" (US formatting) ✅

### Hindi User (hi_IN)

Even without explicit "Hindi" mapping, the prompt ends with:

```
---

Conversation language: hi_IN
```

The AI automatically understands this is Hindi (India) and responds in Hindi ✅

## Related Files

- **src/prompts.py** - Simple prompt template and language detection (just 2 functions!)
- **src/nodes.py** - LLM invocation with system prompt injection
- **src/testing/fixtures.py** - Test fixtures with system prompt
- **docs/PROMPTS.md** - This documentation

## Future Enhancements

- [ ] Per-user language preference override (in State or Context)
- [ ] Automatic locale detection from user messages (fallback method)
- [ ] Language-specific tool descriptions in RAG context
- [ ] Multi-lingual prompt responses for bilingual contexts
- [ ] Language preference persistence across sessions
