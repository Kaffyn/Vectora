# Vectora System Prompts

## Overview

Vectora uses a dynamic, language-aware system prompt that automatically detects the user's operating system language and responds accordingly. This ensures that even simple queries like "47.23 * 134.97" are answered in the user's native language.

## How It Works

### Language Detection

The system automatically detects the user's language from the operating system locale:

```python
import locale

def get_system_language() -> str:
    """Detect system language from OS locale."""
    try:
        lang_code, _ = locale.getdefaultlocale()
        if lang_code:
            return lang_code.split("_")[0].lower()  # e.g., 'pt', 'en', 'es'
    except Exception:
        pass
    return "en"  # Fallback to English
```

**Supported languages** (src/prompts.py):
- `pt` - Portuguese (Português)
- `en` - English
- `es` - Spanish (Español)
- `fr` - French (Français)
- `de` - German (Deutsch)
- `it` - Italian (Italiano)
- `ja` - Japanese (日本語)
- `zh` - Chinese (中文)

### Dynamic System Prompt

The system prompt is a template that gets formatted at runtime:

```python
SYSTEM_PROMPT_TEMPLATE = """# Vectora - Advanced AI Assistant with RAG Capabilities

...

**Language:**
Respond in {language_name} ({language_code_upper}). 
Always match the user's language preference.

...
"""

def get_system_prompt() -> str:
    """Get the Vectora system prompt with language auto-detected from OS."""
    lang_code = get_system_language()
    lang_name = LANGUAGE_NAMES.get(lang_code, "English")
    
    return SYSTEM_PROMPT_TEMPLATE.format(
        language_name=lang_name,
        language_code_upper=lang_code.upper(),
    )
```

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

To add support for a new language:

### 1. Add language mapping (src/prompts.py)

```python
LANGUAGE_NAMES: dict[str, str] = {
    # ... existing languages ...
    "hi": "Hindi (हिन्दी)",  # Example: add Hindi
}
```

### 2. Verify locale code

The locale code should be the ISO 639-1 two-letter code:
- `pt` for Portuguese
- `en` for English
- `es` for Spanish
- `hi` for Hindi
- `tr` for Turkish
- etc.

### 3. Test locally

Set your system locale and verify:

```bash
# On Linux/Mac
export LC_ALL=pt_BR.UTF-8
export LANG=pt_BR.UTF-8

# On Windows
# Settings → Region & Language → Language
```

Then run:

```bash
uv run src/main.py
# or test the prompt directly:
python3 -c "from prompts import get_system_prompt; print(get_system_prompt())"
```

## Customizing the Prompt

To modify the prompt content:

1. Edit `SYSTEM_PROMPT_TEMPLATE` in `src/prompts.py`
2. Maintain the `{language_name}` and `{language_code_upper}` placeholders
3. Restart Vectora to apply changes

### Prompt Sections

- **Identity & Capabilities**: Lines 1-25
- **Tool Integration Details**: Lines 27-40
- **Operational Guidelines**: Lines 42-60
- **Important Notes**: Lines 62-70
- **Language Section**: Line 72-73 (keep placeholders!)

## Debugging

### Check detected language

```python
from prompts import get_system_language, get_system_prompt

print(f"Detected language: {get_system_language()}")
print("\nSystem prompt:")
print(get_system_prompt())
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

### Portuguese User (pt)

When system locale is Portuguese, the prompt ends with:

```
**Language:**
Respond in Portuguese (Português). Always match the user's language preference.
```

Result: "47.23 * 134.97 = 6.399.0751" em Português ✅

### Spanish User (es)

When system locale is Spanish, the prompt ends with:

```
**Language:**
Respond in Spanish (Español). Always match the user's language preference.
```

Result: "47.23 * 134.97 = 6.399.0751 en Español ✅

### Unsupported Language (e.g., Icelandic)

Falls back to English since 'is' is not in LANGUAGE_NAMES.

To support Icelandic, add to LANGUAGE_NAMES:

```python
LANGUAGE_NAMES = {
    # ... existing ...
    "is": "Icelandic (Íslenska)",
}
```

## Related Files

- **src/prompts.py** - Prompt templates and language detection
- **src/nodes.py** - LLM invocation with system prompt injection
- **src/testing/fixtures.py** - Test fixtures with system prompt
- **docs/PROMPTS.md** - This documentation

## Future Enhancements

- [ ] Per-user language preference override (in State or Context)
- [ ] Language-specific system prompts (different templates per language)
- [ ] Automatic locale detection from user messages (fallback method)
- [ ] Language-specific tool descriptions in RAG context
- [ ] Multi-lingual prompt responses for bilingual contexts
