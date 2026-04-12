/**
 * Jest Configuration for Vectora VS Code Extension
 * Phase 6: Runtime Testing & Vectora Core Integration
 */

module.exports = {
  displayName: "vectora-extension",
  testEnvironment: "node",
  roots: ["<rootDir>/src"],
  testMatch: ["**/__tests__/**/*.test.ts"],
  moduleFileExtensions: ["ts", "tsx", "js", "jsx", "json"],
  transform: {
    "^.+\\.tsx?$": [
      "ts-jest",
      {
        tsconfig: {
          module: "commonjs",
          lib: ["es2020"],
          target: "es2020",
          jsx: "react",
          esModuleInterop: true,
          allowSyntheticDefaultImports: true,
        },
      },
    ],
  },

  // Module resolution
  moduleNameMapper: {
    "^@/(.*)$": "<rootDir>/src/$1",
    "^@tests/(.*)$": "<rootDir>/src/__tests__/$1",
    "^@fixtures/(.*)$": "<rootDir>/src/__tests__/fixtures/$1",
  },

  // Coverage configuration
  collectCoverage: true,
  collectCoverageFrom: [
    "src/**/*.ts",
    "src/**/*.tsx",
    "!src/**/*.d.ts",
    "!src/__tests__/**",
    "!src/types/**",
  ],
  coverageDirectory: "<rootDir>/coverage",
  coverageReporters: ["text", "lcov", "html", "json"],
  coverageThreshold: {
    global: {
      branches: 60,
      functions: 60,
      lines: 60,
      statements: 60,
    },
  },

  // Setup files
  setupFilesAfterEnv: ["<rootDir>/src/__tests__/setup.ts"],

  // Timeout
  testTimeout: 10000,

  // Verbose
  verbose: true,

  // Clear mocks between tests
  clearMocks: true,

  // Restore mocks between tests
  restoreMocks: true,

  // Test name pattern
  testNamePattern: ".*",

  // No cover pattern
  testPathIgnorePatterns: ["/node_modules/", "/dist/"],

  // Globals
  globals: {
    "ts-jest": {
      isolatedModules: true,
    },
  },
};
