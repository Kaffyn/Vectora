# Contributing to Vectora

> [!TIP]
> Read this file in another language.
> English [CONTRIBUTING.md] | Portuguese [CONTRIBUTING.pt.md]

This document outlines the standards, philosophy, and technical rules for contributing to Vectora. Our goal is to build a reliable, high-performance, and secure AI orchestrator.

## Project Philosophy

Vectora is built on three core pillars:

1. **Reliability**: The agent should never leave the system in an inconsistent state. The **Git Integration** (via the user's IDE) should be leveraged to ensure changes are trackable.
2. **Security**: Adherence to the **Trust Folder** model is non-negotiable. Tools must never escape their defined scope.
3. **Cleanliness**: We value idiomatic Go code, structured logging, and professional documentation.

**The "No Emoji" Policy:**

To maintain a professional and clean aesthetic, we do not use emojis in documentation, code comments, or commit messages. All technical communication should be clear and concise.

## Tech Stack & Standards

- **Language**: Go 1.23+ (utilizing iterators and modern standard library features).
- **Logging**: Use `log/slog` for all structured logging. Avoid `fmt.Printf` for anything other than CLI output.
- **Errors**: Always wrap errors with context using `fmt.Errorf("...: %w", err)`. Use `errors.Is` and `errors.As` for checking.
- **Concurrency**: Leverage goroutines and channels carefully. Follow CSP (Communicating Sequential Processes) principles.

## Repository Structure

Vectora is organized as a monorepo. The current MVP focus is 100% on the **core** of the engine.

- `cmd/vectora/`: The main entry point using Cobra CLI.
- `core/acp/`: Implementation of the Agent Client Protocol (JSON-RPC 2.0).
- `core/storage/`: Abstraction layer for `bbolt` and `chromem-go`.
- `core/tools/`: The agentic arsenal (File, Terminal, Knowledge).
- `core/ipc/`: Named Pipes transport implementation for Windows.
- `pkg/gemini/`: Optimized client for Gemini Pro and Embedding APIs.

## Contribution Workflow

**1. Preparation:**

- Ensure you have Go 1.23 installed.
- Configure your environment to run the core (using named pipes requires Windows).

**2. Branching and Commits:**

- Use descriptive branch names: `feat/feature-name`, `fix/bug-name`, `docs/page-name`.
- Keep commits atomic and well-documented.
- **Do not use emojis in commit messages.**

**3. Standards and Testing:**

- Run `go fmt ./...` before submitting.
- Write unit tests for new logic in the `core/` packages.
- Ensure that all ACP methods are documented and tested with JSON-RPC mocks.

**4. Pull Requests:**

- Provide a clear description of the changes and what they solve.
- Link to any relevant issues or architectural discussions.

## Release & Code Signing

### Windows Defender SmartScreen Signing

When distributing Vectora binaries on Windows, code signing helps prevent SmartScreen warnings on first execution.

**Process:**

1. **Obtain a Code Signing Certificate**
   - Purchase an EV Code Signing certificate from a trusted CA (Sectigo, DigiCert, etc.)
   - Store the `.pfx` file securely (preferably in Azure Key Vault for CI/CD)

2. **Sign the Binary**
   ```bash
   # Using signtool (Windows SDK)
   signtool sign /f certificate.pfx /p password /t http://timestamp.digicert.com vectora.exe
   ```

3. **Verify Signature**
   ```bash
   signtool verify /pa vectora.exe
   ```

4. **CI/CD Integration**
   - Store certificate and password in GitHub Secrets (or equivalent)
   - Sign binaries as part of release workflow
   - Include signed checksums in release notes

5. **Microsoft SmartScreen Reputation**
   - Initial releases may still trigger warnings due to low reputation
   - Reputation improves as more users download the signed binary
   - Monitor SmartScreen feedback through Windows Defender Feedback

**References:**
- [Microsoft Code Signing Guide](https://docs.microsoft.com/en-us/windows/win32/seccrypto/about-cryptography)
- [Sectigo Code Signing](https://www.sectigo.com/code-signing)

---

_Part of the Kaffyn open source organization._
