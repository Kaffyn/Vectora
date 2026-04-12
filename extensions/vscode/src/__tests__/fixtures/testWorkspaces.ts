/**
 * Test Workspaces - Fixture for Testing
 * Simula diferentes estruturas de workspace para testes
 */

import type { WorkspaceContext } from "@/types/core";

// ============================================================================
// Basic Workspace Fixtures
// ============================================================================

export const mockWorkspaceBasic: WorkspaceContext = {
  workspaceId: "test-workspace-001",
  rootPath: "/home/user/projects/test-app",
};

export const mockWorkspaceWithFiles: WorkspaceContext = {
  workspaceId: "test-workspace-002",
  rootPath: "/home/user/projects/typescript-app",
  files: [
    "src/main.ts",
    "src/utils.ts",
    "src/components/Button.tsx",
    "src/components/Form.tsx",
    "tests/main.test.ts",
    "package.json",
    "tsconfig.json",
  ],
};

export const mockWorkspaceWithEmbeddings: WorkspaceContext = {
  workspaceId: "test-workspace-003",
  rootPath: "/home/user/projects/rag-enabled-app",
  files: [
    "src/index.ts",
    "src/service.ts",
    "src/models/User.ts",
    "src/db/connection.ts",
  ],
  embeddingEnabled: true,
  vectorCount: 1250,
};

// ============================================================================
// Different Project Types
// ============================================================================

export const mockReactProject: WorkspaceContext = {
  workspaceId: "react-project",
  rootPath: "/home/user/projects/react-app",
  files: [
    "src/App.tsx",
    "src/App.css",
    "src/components/Header.tsx",
    "src/components/Footer.tsx",
    "src/hooks/useData.ts",
    "src/services/api.ts",
    "public/index.html",
    "package.json",
  ],
};

export const mockNodeProject: WorkspaceContext = {
  workspaceId: "node-project",
  rootPath: "/home/user/projects/node-api",
  files: [
    "src/index.ts",
    "src/routes/users.ts",
    "src/routes/products.ts",
    "src/middleware/auth.ts",
    "src/models/User.ts",
    "src/models/Product.ts",
    "src/db/connection.ts",
    "package.json",
  ],
};

export const mockFullStackProject: WorkspaceContext = {
  workspaceId: "fullstack-project",
  rootPath: "/home/user/projects/fullstack-app",
  files: [
    // Frontend
    "frontend/src/App.tsx",
    "frontend/src/components/Button.tsx",
    "frontend/src/pages/Home.tsx",
    "frontend/package.json",
    // Backend
    "backend/src/index.ts",
    "backend/src/routes/api.ts",
    "backend/src/models/User.ts",
    "backend/package.json",
    // Shared
    "shared/types.ts",
    "README.md",
  ],
  embeddingEnabled: true,
  vectorCount: 3500,
};

// ============================================================================
// Large Workspace
// ============================================================================

export const mockLargeWorkspace: WorkspaceContext = {
  workspaceId: "large-workspace",
  rootPath: "/home/user/projects/enterprise-app",
  files: generateMockFileList(500),
  embeddingEnabled: true,
  vectorCount: 25000,
};

function generateMockFileList(count: number): string[] {
  const files: string[] = [];
  const dirs = [
    "src",
    "src/components",
    "src/services",
    "src/models",
    "src/utils",
    "src/hooks",
    "tests",
    "tests/unit",
    "tests/integration",
  ];
  const extensions = [".ts", ".tsx", ".js", ".jsx", ".css", ".json"];

  for (let i = 0; i < count; i++) {
    const dir = dirs[Math.floor(Math.random() * dirs.length)];
    const ext = extensions[Math.floor(Math.random() * extensions.length)];
    files.push(`${dir}/file-${i}${ext}`);
  }
  return files;
}

// ============================================================================
// Monorepo Workspace
// ============================================================================

export const mockMonorepoWorkspace: WorkspaceContext = {
  workspaceId: "monorepo-workspace",
  rootPath: "/home/user/projects/monorepo",
  files: [
    // Package 1
    "packages/core/src/index.ts",
    "packages/core/src/types.ts",
    "packages/core/package.json",
    // Package 2
    "packages/ui/src/Button.tsx",
    "packages/ui/src/Form.tsx",
    "packages/ui/package.json",
    // Package 3
    "packages/api/src/routes.ts",
    "packages/api/src/middleware.ts",
    "packages/api/package.json",
    // Root
    "pnpm-workspace.yaml",
    "tsconfig.json",
    "turbo.json",
  ],
  embeddingEnabled: true,
  vectorCount: 5000,
};

// ============================================================================
// Minimal Workspace
// ============================================================================

export const mockMinimalWorkspace: WorkspaceContext = {
  workspaceId: "minimal-workspace",
  rootPath: "/home/user/projects/simple-app",
  files: ["main.ts", "package.json", "README.md"],
};

// ============================================================================
// Workspace with Different File Structures
// ============================================================================

export const mockWorkspaceWithMixedFiles: WorkspaceContext = {
  workspaceId: "mixed-files-workspace",
  rootPath: "/home/user/projects/mixed-app",
  files: [
    // Source files
    "src/index.ts",
    "src/app.tsx",
    "src/styles.css",
    // Config files
    ".env",
    ".env.local",
    "tsconfig.json",
    "webpack.config.js",
    "jest.config.js",
    // Documentation
    "README.md",
    "docs/API.md",
    "docs/ARCHITECTURE.md",
    // Build artifacts (should be ignored)
    "dist/index.js",
    "build/app.js",
  ],
};

// ============================================================================
// Python Project (for cross-language testing)
// ============================================================================

export const mockPythonProject: WorkspaceContext = {
  workspaceId: "python-project",
  rootPath: "/home/user/projects/python-app",
  files: [
    "src/main.py",
    "src/utils.py",
    "src/models/user.py",
    "tests/test_main.py",
    "requirements.txt",
    "setup.py",
    ".env",
  ],
};

// ============================================================================
// Go Project
// ============================================================================

export const mockGoProject: WorkspaceContext = {
  workspaceId: "go-project",
  rootPath: "/home/user/projects/go-app",
  files: [
    "main.go",
    "internal/server.go",
    "internal/handlers/user.go",
    "pkg/models/user.go",
    "go.mod",
    "go.sum",
  ],
};

// ============================================================================
// Workspace Configurations
// ============================================================================

export const mockWorkspaceConfigs = {
  typescript: mockReactProject,
  nodejs: mockNodeProject,
  fullstack: mockFullStackProject,
  monorepo: mockMonorepoWorkspace,
  python: mockPythonProject,
  go: mockGoProject,
};

// ============================================================================
// Workspace Comparison Fixtures
// ============================================================================

export const workspaceSequence: WorkspaceContext[] = [
  mockMinimalWorkspace,
  mockWorkspaceBasic,
  mockWorkspaceWithFiles,
  mockWorkspaceWithEmbeddings,
];

// ============================================================================
// Invalid/Edge Case Workspaces
// ============================================================================

export const mockInvalidWorkspace: Partial<WorkspaceContext> = {
  workspaceId: "",
  rootPath: "",
};

export const mockWorkspaceWithInvalidPath: WorkspaceContext = {
  workspaceId: "invalid-path-workspace",
  rootPath: "relative/path/that/is/invalid",
  files: [],
};

export const mockEmptyWorkspace: WorkspaceContext = {
  workspaceId: "empty-workspace",
  rootPath: "/home/user/projects/empty",
  files: [],
  embeddingEnabled: false,
  vectorCount: 0,
};
