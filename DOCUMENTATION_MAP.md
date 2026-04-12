# Vectora Documentation Map

Complete guide to all Vectora documentation, organized by use case and component.

---

## 📋 Quick Navigation

| Component | Purpose | Location | Audience |
|-----------|---------|----------|----------|
| **Vectora Core** | Main service architecture & overview | [README.md](README.md) | Everyone |
| **Vectora CLI** | Command-line interface reference | [cmd/core/README.md](cmd/core/README.md) | CLI users |
| **VS Code Extension** | IDE integration guide | [extensions/vscode/README.md](extensions/vscode/README.md) | VS Code users |
| **Gemini CLI Bridge** | Google Gemini CLI integration | [extensions/geminicli/README.md](extensions/geminicli/README.md) | Gemini CLI users |
| **Testing** | Local test suite & development | [test/README.md](test/README.md) | Developers |

---

## 🏗️ Architecture & Design

### Main Project Overview
**File:** [README.md](README.md) (350 lines)

**Covers:**
- Vectora's hybrid RAG approach (semantic + structural)
- Agentic operation & sub-agent architecture
- Supported AI models (10 families, 40+ models)
- Tech stack (Go, chromem-go, BBolt, Cobra, MCP, ACP)
- Trust Folder & security model
- Agentic toolkit (read, write, search, execute)

**Read this if:**
- You're new to Vectora
- You need to understand the overall architecture
- You want to know supported models & features
- You're evaluating Vectora for your use case

**Portuguese version:** [README.pt.md](README.pt.md)

---

## 💻 User Guides

### CLI Command Reference
**File:** [cmd/core/README.md](cmd/core/README.md) (350+ lines)

**Covers:**
- All Cobra commands (start, stop, embed, ask, config, models, etc.)
- Configuration management (API keys, providers)
- Embedding & indexing
- Querying the knowledge base
- Troubleshooting common issues
- Environment variables
- Exit codes

**Commands Documented:**
```
vectora start           → Start background service
vectora stop            → Stop the service
vectora restart         → Restart service
vectora status          → Check service status
vectora reset --hard    → Wipe all data
vectora embed [path]    → Index codebase
vectora ask [query]     → Query knowledge base
vectora config          → Manage configuration
vectora models          → List available models
vectora acp             → ACP protocol server
```

**Read this if:**
- You're using Vectora from the command line
- You need to configure API keys
- You want to index a codebase
- You're troubleshooting CLI issues

---

### VS Code Extension
**File:** [extensions/vscode/README.md](extensions/vscode/README.md) (100+ lines)

**Covers:**
- Extension features (RAG chat, agentic tools, streaming)
- Two modes: Agent & Sub-Agent
- Architecture diagram (Extension ↔ Core via ACP)
- Agentic tools (read_file, grep_search, terminal_run)
- Trust Folder & file safety
- Binary management (auto-download, PATH resolution)
- Inline completion provider (planning)

**Read this if:**
- You're using Vectora in VS Code
- You want to understand how the extension works
- You're integrating with other agents (Antigravity)
- You need to configure the extension

**Portuguese version:** [extensions/vscode/README.pt.md](extensions/vscode/README.pt.md)

---

### Gemini CLI Bridge
**File:** [extensions/geminicli/README.md](extensions/geminicli/README.md) (118 lines)

**Covers:**
- Vectora as MCP sub-agent for Gemini CLI
- Installation & build instructions
- Configuration for Gemini CLI settings
- Interactive REPL for testing
- Environment variables
- Development workflows

**Read this if:**
- You're using Google's Gemini CLI
- You want Vectora as a specialized code-reasoning agent
- You're setting up MCP configuration
- You want to integrate with cloud-based Gemini

**Portuguese version:** [extensions/geminicli/README.pt.md](extensions/geminicli/README.pt.md)

---

## 🧪 Testing & Development

### Local Test Suite
**File:** [test/README.md](test/README.md) (325 lines)

**Covers:**
- Test framework overview (8 core tests)
- Usage modes (interactive, batch, specific test)
- Configuration via `.env.local`
- Test output & reports
- Troubleshooting
- Environment variables
- Best practices

**Related Files:**
- [test/QUICK_START.md](test/QUICK_START.md) — 5-minute setup guide
- [test/DEVELOPMENT_INTEGRATION.md](test/DEVELOPMENT_INTEGRATION.md) — IDE integration & workflows

**Read this if:**
- You're developing Vectora
- You want to run the test suite locally
- You're setting up pre-commit hooks
- You need CI/CD integration examples

---

### Quick Start for Testing
**File:** [test/QUICK_START.md](test/QUICK_START.md) (231 lines)

**Covers:**
- 5-minute setup (environment, API keys, run tests)
- Interactive vs. batch modes
- What gets tested (help, version, ask, config, etc.)
- Example output
- Troubleshooting common issues
- Daily development workflows

**Read this if:**
- You're new to the test suite
- You want to run tests for the first time
- You need quick reference commands

---

### Development Integration
**File:** [test/DEVELOPMENT_INTEGRATION.md](test/DEVELOPMENT_INTEGRATION.md) (370 lines)

**Covers:**
- Pre-commit testing strategies
- Git hooks setup
- Continuous development workflow
- Performance testing & regression detection
- Adding new tests
- Debugging test failures
- Environment variables for dev
- IDE integration (VS Code, GoLand, IntelliJ)
- CI/CD integration (GitHub Actions)

**Read this if:**
- You're setting up a development workflow
- You want to integrate tests with Git hooks
- You're configuring CI/CD
- You're adding new tests

---

## 📊 Project Summary & Stats

### Architecture Documentation
**File:** [FINAL_PROJECT_SUMMARY.md](FINAL_PROJECT_SUMMARY.md) (486 lines)

**Covers:**
- Complete execution summary (15 phases)
- Architecture diagrams & data flow
- Technology stack details
- Project statistics (40K+ lines, 150+ files, 50+ commits)
- Feature comparison matrix
- Future roadmap

**Read this if:**
- You want a high-level project overview
- You need architecture details
- You want statistics on the project
- You're planning future development

---

## 🔒 Security & Trust

Both the main README and the CLI documentation cover:

- **Trust Folder** — Scoped access to project directories
- **Guardian System** — Automatic blocking of sensitive files (`.env`, `.key`, `.pem`)
- **Protocol Security** — ACP (JSON-RPC 2.0) over stdio/IPC (no external network)
- **Data Privacy** — All embeddings stay local, only API calls go to cloud
- **Transactional Safety** — Git snapshots before code modifications

---

## 🌍 Multilingual Support

All primary documentation is available in both English and Portuguese:

- English: `README.md`, `DEVELOPMENT_INTEGRATION.md`, etc.
- Portuguese: `README.pt.md`, `extensions/vscode/README.pt.md`, `extensions/geminicli/README.pt.md`

---

## 📚 Reading Paths by Role

### For End Users (Using Vectora)

1. **Vectora CLI Users**
   - Start: [cmd/core/README.md](cmd/core/README.md)
   - Reference: `vectora --help`
   - Troubleshooting: CLI README section

2. **VS Code Extension Users**
   - Start: [extensions/vscode/README.md](extensions/vscode/README.md)
   - Reference: VS Code Vectora panel help
   - Integration: See Antigravity/Sub-Agent docs

3. **Gemini CLI Users**
   - Start: [extensions/geminicli/README.md](extensions/geminicli/README.md)
   - Setup: Follow MCP configuration guide
   - Development: See REPL testing section

### For Developers (Contributing to Vectora)

1. **Getting Started**
   - [README.md](README.md) — Architecture overview
   - [FINAL_PROJECT_SUMMARY.md](FINAL_PROJECT_SUMMARY.md) — Project structure

2. **Setting Up Development**
   - [test/QUICK_START.md](test/QUICK_START.md) — 5-minute setup
   - [test/DEVELOPMENT_INTEGRATION.md](test/DEVELOPMENT_INTEGRATION.md) — Workflow

3. **Running Tests**
   - [test/README.md](test/README.md) — Full test documentation
   - `cd test && ./setup-test-env.sh && ./local-test-suite.sh --interactive`

### For Architecture Review

1. [README.md](README.md) — Core architecture
2. [FINAL_PROJECT_SUMMARY.md](FINAL_PROJECT_SUMMARY.md) — Phase breakdown
3. [cmd/core/README.md](cmd/core/README.md) — CLI interface
4. [extensions/vscode/README.md](extensions/vscode/README.md) — Extension design

---

## 🔗 Cross-References

### Related Files & Configurations

| File | Purpose |
|------|---------|
| `/AGENTS.md` | Supported AI models & verification links |
| `/API_ARCHITECTURE.md` | ACP & MCP protocol specifications |
| `/core/README.md` *(if exists)* | Core library documentation |
| `/.github/workflows/` | CI/CD pipeline definitions |
| `/test/.env.local.example` | Test environment template |

---

## ✅ Documentation Checklist

- [x] Main project README (Vectora Core overview)
- [x] CLI command reference (cmd/core/README.md)
- [x] VS Code Extension guide
- [x] Gemini CLI Bridge guide
- [x] Local test suite documentation
- [x] Development integration guide
- [x] Project summary & stats
- [x] Multilingual support (Portuguese)
- [x] This documentation map

---

## 📝 How to Update Documentation

When making changes:

1. **Core changes** → Update [README.md](README.md)
2. **CLI changes** → Update [cmd/core/README.md](cmd/core/README.md)
3. **Extension changes** → Update [extensions/vscode/README.md](extensions/vscode/README.md)
4. **Testing changes** → Update [test/README.md](test/README.md)
5. **Major features** → Update [FINAL_PROJECT_SUMMARY.md](FINAL_PROJECT_SUMMARY.md)
6. **Portuguese versions** → Keep `.pt.md` files in sync

Keep all versions up-to-date with real implementation details.

---

## 🎯 Quick Links

**Getting Started:**
- New to Vectora? → [README.md](README.md)
- CLI user? → [cmd/core/README.md](cmd/core/README.md)
- VS Code user? → [extensions/vscode/README.md](extensions/vscode/README.md)
- Developer? → [test/QUICK_START.md](test/QUICK_START.md)

**Documentation:**
- Full reference → This file
- Architecture → [FINAL_PROJECT_SUMMARY.md](FINAL_PROJECT_SUMMARY.md)
- Tests → [test/README.md](test/README.md)
- Models → [AGENTS.md](AGENTS.md)
- Protocol → [API_ARCHITECTURE.md](API_ARCHITECTURE.md)

---

**Last Updated:** April 12, 2026

**Version:** Vectora 0.1.0
