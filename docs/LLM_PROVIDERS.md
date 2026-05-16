# LLM Providers Configuration

## Overview

Vectora supports multiple LLM providers configured entirely via environment variables:

- **Google Gemini** (default) — `google-genai`
- **Ollama** (local) — `ollama`
- **OpenAI** — `openai`
- **Anthropic Claude** — `anthropic`

No code changes needed — just set environment variables and Vectora auto-loads the right provider.

---

## Quick Start

### Default (Google Gemini)

```powershell
# Set your Google API key
$env:GOOGLE_API_KEY="sk-..."

# Run Vectora (uses gemini-2.0-flash by default)
uv run python vectora/run_chat.py
```

### Ollama (Local)

```powershell
# Make sure Ollama is running on localhost:11434
$env:LLM_PROVIDER="ollama"

# Run Vectora
uv run python vectora/run_chat.py
```

### OpenAI

```powershell
$env:LLM_PROVIDER="openai"
$env:OPENAI_API_KEY="sk-..."

uv run python vectora/run_chat.py
```

### Anthropic

```powershell
$env:LLM_PROVIDER="anthropic"
$env:ANTHROPIC_API_KEY="sk-ant-..."

uv run python vectora/run_chat.py
```

---

## Environment Variables

### Provider Selection

| Variable       | Values                                          | Default        |
| -------------- | ----------------------------------------------- | -------------- |
| `LLM_PROVIDER` | `google-genai`, `ollama`, `openai`, `anthropic` | `google-genai` |

### Google Gemini

| Variable         | Purpose            | Default            | Required |
| ---------------- | ------------------ | ------------------ | -------- |
| `GOOGLE_API_KEY` | API authentication | —                  | ✅ Yes   |
| `GOOGLE_MODEL`   | Model name         | `gemini-2.0-flash` | No       |

**Available models:**

- `gemini-2.0-flash` — Fast, latest (recommended)
- `gemini-1.5-pro` — More capable, slower
- `gemini-1.5-flash` — Fast alternative
- `gemini-pro` — Previous generation

```powershell
$env:GOOGLE_API_KEY="AIza..."
$env:GOOGLE_MODEL="gemini-1.5-pro"
```

### Ollama

| Variable          | Purpose           | Default                  | Required |
| ----------------- | ----------------- | ------------------------ | -------- |
| `OLLAMA_BASE_URL` | Ollama server URL | `http://127.0.0.1:11434` | No       |
| `OLLAMA_MODEL`    | Model name        | `gpt-oss:20b`            | No       |

**Available models** (depends on what's installed):

- `gpt-oss:20b` — GPT-like open model
- `qwen3-coder:30b` — Code generation
- `mistral` — General purpose
- `neural-chat` — Chat optimized
- Any model available in `ollama pull`

```powershell
$env:LLM_PROVIDER="ollama"
$env:OLLAMA_BASE_URL="http://localhost:11434"
$env:OLLAMA_MODEL="mistral"
```

### OpenAI

| Variable         | Purpose            | Default  | Required |
| ---------------- | ------------------ | -------- | -------- |
| `OPENAI_API_KEY` | API authentication | —        | ✅ Yes   |
| `OPENAI_MODEL`   | Model name         | `gpt-4o` | No       |

**Available models:**

- `gpt-4o` — Latest multimodal
- `gpt-4-turbo` — Powerful, cheaper
- `gpt-3.5-turbo` — Fast, economical

```powershell
$env:LLM_PROVIDER="openai"
$env:OPENAI_API_KEY="sk-..."
$env:OPENAI_MODEL="gpt-4o"
```

### Anthropic

| Variable            | Purpose            | Default           | Required |
| ------------------- | ------------------ | ----------------- | -------- |
| `ANTHROPIC_API_KEY` | API authentication | —                 | ✅ Yes   |
| `ANTHROPIC_MODEL`   | Model name         | `claude-opus-4-1` | No       |

**Available models:**

- `claude-opus-4-1` — Most capable
- `claude-sonnet-4-7` — Balanced
- `claude-haiku-4-5` — Fast, economical

```powershell
$env:LLM_PROVIDER="anthropic"
$env:ANTHROPIC_API_KEY="sk-ant-..."
$env:ANTHROPIC_MODEL="claude-opus-4-1"
```

### Common Settings

| Variable          | Purpose                       | Default |
| ----------------- | ----------------------------- | ------- |
| `LLM_TEMPERATURE` | Response randomness (0.0-2.0) | `0.2`   |

Lower = deterministic, Higher = creative.

```powershell
$env:LLM_TEMPERATURE="0.5"  # More creative
```

---

## Configuration via .env

Create a `.env` file in the root directory:

```env
# Provider (comment out the ones you don't use)
LLM_PROVIDER=google-genai
# LLM_PROVIDER=ollama
# LLM_PROVIDER=openai
# LLM_PROVIDER=anthropic

# Google Gemini
GOOGLE_API_KEY=AIza...
GOOGLE_MODEL=gemini-2.0-flash

# Ollama (local)
# OLLAMA_BASE_URL=http://127.0.0.1:11434
# OLLAMA_MODEL=gpt-oss:20b

# OpenAI
# OPENAI_API_KEY=sk-...
# OPENAI_MODEL=gpt-4o

# Anthropic
# ANTHROPIC_API_KEY=sk-ant-...
# ANTHROPIC_MODEL=claude-opus-4-1

# Common
LLM_TEMPERATURE=0.2
```

---

## How It Works

When Vectora starts:

1. **Read provider** → `LLM_PROVIDER` env var (default: `google-genai`)
2. **Load configuration** → Read provider-specific env vars
3. **Initialize LLM** → Call `init_chat_model()` with provider config
4. **Bind tools** → Add available tools to the LLM
5. **Ready** → Use LLM in graph invocations

**Code location:** `vectora/utils.py` — `load_llm()` function

```python
def load_llm() -> BaseChatModel:
    provider = get_env("LLM_PROVIDER", default="google-genai")

    if provider == "google-genai":
        model = init_chat_model(
            model="gemini-2.0-flash",  # ← from env or default
            model_provider="google-genai",
            api_key=get_env("GOOGLE_API_KEY"),
            temperature=0.2,  # ← from env or default
        )
    # ... other providers

    return model
```

---

## Testing Different Providers

### Test with CLI

```powershell
# Test Google Gemini
$env:LLM_PROVIDER="google-genai"
$env:GOOGLE_API_KEY="your-key"
uv run python vectora/main.py

# Test Ollama
$env:LLM_PROVIDER="ollama"
uv run python vectora/main.py

# Test OpenAI
$env:LLM_PROVIDER="openai"
$env:OPENAI_API_KEY="your-key"
uv run python vectora/main.py
```

### Test with Chat TUI

```powershell
$env:LLM_PROVIDER="google-genai"
$env:GOOGLE_API_KEY="your-key"
uv run python vectora/run_chat.py
```

### Test in Code

```python
import os
os.environ["LLM_PROVIDER"] = "google-genai"
os.environ["GOOGLE_API_KEY"] = "AIza..."

from utils import load_llm
llm = load_llm()
response = llm.invoke("Hello!")
print(response.content)
```

---

## Switching Providers

### Easy: Change environment variable

```powershell
# Was using Google
$env:LLM_PROVIDER="google-genai"

# Switch to Ollama (instantly)
$env:LLM_PROVIDER="ollama"

# Restart app
uv run python vectora/run_chat.py
```

### Persistent: Update .env file

Edit `.env` and change `LLM_PROVIDER`:

```env
# Before
LLM_PROVIDER=google-genai

# After
LLM_PROVIDER=ollama
```

Then restart.

---

## Cost Comparison

### Per-request estimates (rough)

| Provider      | Model            | Input           | Output          |
| ------------- | ---------------- | --------------- | --------------- |
| **Google**    | Gemini 2.0 Flash | $0.075/M tokens | $0.30/M tokens  |
| **Google**    | Gemini 1.5 Pro   | $1.50/M tokens  | $6.00/M tokens  |
| **Ollama**    | Any local model  | FREE            | FREE            |
| **OpenAI**    | GPT-4o           | $5.00/M tokens  | $15.00/M tokens |
| **OpenAI**    | GPT-3.5          | $0.50/M tokens  | $1.50/M tokens  |
| **Anthropic** | Claude Opus      | $3.00/M tokens  | $15.00/M tokens |
| **Anthropic** | Claude Haiku     | $0.80/M tokens  | $4.00/M tokens  |

**Cost optimization:**

- **Free local:** Use Ollama (no API costs, but needs powerful hardware)
- **Budget cloud:** Google Gemini Flash or OpenAI GPT-3.5
- **Best quality:** Claude Opus or Gemini 1.5 Pro
- **Balanced:** Gemini 2.0 Flash (fast + cheap)

---

## Troubleshooting

### "API key not found"

```
Error: GOOGLE_API_KEY not found
```

**Fix:**

```powershell
$env:GOOGLE_API_KEY="your-actual-key-here"
```

Or add to `.env`:

```env
GOOGLE_API_KEY=AIza...
```

### "Provider not supported"

```
ValueError: Unknown LLM_PROVIDER: xyz
```

**Fix:** Check spelling and use one of: `google-genai`, `ollama`, `openai`, `anthropic`

```powershell
$env:LLM_PROVIDER="google-genai"  # ✅ Correct
$env:LLM_PROVIDER="gemini"        # ❌ Wrong
```

### "Ollama connection refused"

```
Error: Failed to connect to http://127.0.0.1:11434
```

**Fix:** Make sure Ollama is running:

```powershell
# Start Ollama (macOS/Linux)
ollama serve

# Or on Windows, start the Ollama app

# Test connection
curl http://127.0.0.1:11434/api/version
```

### "Model not found" (Ollama)

```
Error: Model 'unknown-model' not found
```

**Fix:** Pull the model first:

```powershell
ollama pull mistral
ollama pull neural-chat

# List available models
ollama list
```

### Slow responses

**Solutions:**

1. Use a faster model (e.g., Gemini Flash instead of Gemini Pro)
2. Lower temperature slightly
3. Switch to local Ollama (if you have GPU)

---

## Model Selection Guide

### By Use Case

**General Chat**

- ✅ Gemini 2.0 Flash (default)
- ✅ GPT-4o
- ✅ Claude Sonnet

**Code Generation**

- ✅ GPT-4o (best code quality)
- ✅ Claude Opus
- ✅ Qwen Coder (Ollama)

**RAG / Long Context**

- ✅ Gemini 1.5 Pro (long context window)
- ✅ Claude Opus (best reasoning)

**Local / Free**

- ✅ Ollama + Mistral
- ✅ Ollama + Neural Chat

**Budget**

- ✅ Gemini Flash
- ✅ GPT-3.5
- ✅ Claude Haiku

---

## Advanced: Custom Provider

To add a new provider, edit `vectora/utils.py`:

```python
def load_llm() -> BaseChatModel:
    provider = get_env("LLM_PROVIDER", strict=False) or "google-genai"

    # ... existing providers ...

    elif provider == "custom-provider":
        model = cast(
            "BaseChatModel",
            init_chat_model(
                model=_get_env_with_default("CUSTOM_MODEL", "default-model"),
                model_provider="custom-provider",
                api_key=get_env("CUSTOM_API_KEY"),
                temperature=temperature,
                configurable_fields="any",
            ),
        )

    # ... rest of function
```

Then use:

```powershell
$env:LLM_PROVIDER="custom-provider"
$env:CUSTOM_API_KEY="your-key"
$env:CUSTOM_MODEL="your-model"
```

---

## Related Files

- **vectora/utils.py** — Provider loading logic
- **vectora/nodes.py** — LLM usage in call_llm node
- **vectora/graph.py** — Graph definition
- **.env.example** — Configuration template
- **docs/CONTEXT.md** — User preferences and model overrides

---
