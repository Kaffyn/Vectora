# Blueprint: LLM Inference Engine (Zero-Port)

**Status:** Implementation Complete / Industrial Grade
**Module:** `internal/llm/`

---

## 1. Overview
The inference engine abstracts all model logic behind a unified `llm.Provider` interface. It is designed to handle both cloud (Gemini) and local (Qwen) execution without leaking implementation details to other modules.

## 2. Zero-Port Architecture
To ensure maximum security and avoid opening local network ports, Vectora uses a "Zero-Port" protocol for local models:

### 2.1 The Llama Sidecar
- **Binary**: `llama-cli` (subprocess).
- **Communication**: Communication is performed via `os/exec` pipes (Stdin/Stdout).
- **Isolation**: The inference process has no network access and is automatically reaped if the Daemon crashes.
- **Protocol**: Custom JSON-ND (Newline Delimited) over standard pipes.

## 3. Providers
Vectora currently supports two main providers:

### 3.1 Qwen (Local / Offline)
- **Engine**: llama.cpp (sidecar).
- **State**: Interactive. History is maintained in the KV Cache of the subprocess.
- **Goal**: Full privacy for code and personal documents.

### 3.2 Gemini (Cloud / Multimodal)
- **Engine**: Google AI (LangChainGo).
- **Capabilities**: High-speed reasoning and multimodal tokenization.

## 4. Key Constraints (BR-LLM)
- **Encapsulation**: No package outside `internal/llm` may import a provider SDK.
- **Cancellable**: All requests must be context-aware (able to abort if the user cancels).
- **Embedded Prompting**: Master prompt instructions are embedded in the binary to ensure consistency across versions.
