/**
 * Sub-Agent Manifest for Vectora
 *
 * This constant contains the instructions for external orchestrators
 * (like Antigravity, Cursor, or Copilot) on how to effectively delegate
 * tasks to the Vectora Core.
 */
export const VECTORA_SUB_AGENT_MANIFEST = `
## Tool: Vectora Core (Codebase Expert)

You have access to a specialized sub-agent named **Vectora**.
Delegate tasks requiring deep codebase understanding, global refactoring, or semantic navigation to Vectora.

### When to Call Vectora:
- **Global Context:** When answers depend on files not currently open.
- **Semantic RAG:** To find function definitions/usages across the whole project.
- **Safe Refactoring:** For multi-file changes requiring Git snapshots.
- **Architecture Analysis:** To explain folder structures and dependency graphs.

### How to Call Vectora (ACP Protocol):
Use the exposed ACP tools:
1. \`vectora.search_codebase(query: string)\`: Deep semantic search.
2. \`vectora.read_file(path: string)\`: Structure-aware file reading.
3. \`vectora.refactor(task: string, scope: string[])\`: Multi-file refactoring with safety checks.
4. \`vectora.explain_architecture()\`: Project structure summary.

### Interaction Rules:
1. **Do Not Duplicate Effort:** If Vectora returns file content, do not re-read it with your standard tools.
2. **Trust the Index:** Vectora's semantic index is authoritative for finding obscure definitions.
3. **Security Scope:** Vectora only acts within the 'Trust Folder'.
4. **Clear Delegation:** Be specific. Example: "Use Vectora to find all implementations of 'UserService'."
`;
