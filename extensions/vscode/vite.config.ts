import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import { nodePolyfills } from "vite-plugin-node-polyfills";
import { resolve } from "path";

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [
    react(),
    nodePolyfills({
      include: ["buffer", "process", "util", "stream", "path", "os"], // Include path and os
    }),
  ],
  resolve: {
    alias: [
      {
        find: "@core/prompts/sections/custom-instructions",
        replacement: resolve(__dirname, "./src/core/prompts/sections/custom-instructions.browser.ts"),
      },
      { find: "@", replacement: resolve(__dirname, "./src") },
      { find: "@src", replacement: resolve(__dirname, "./src") },
      { find: "@lib", replacement: resolve(__dirname, "./src/lib") },
      { find: "@components", replacement: resolve(__dirname, "./src/components") },
      { find: "@context", replacement: resolve(__dirname, "./src/context") },
      { find: "@hooks", replacement: resolve(__dirname, "./src/hooks") },
      { find: "@utils", replacement: resolve(__dirname, "./src/utils") },
      { find: "vscode", replacement: resolve(__dirname, "./src/roo-internal/vscode-shim/src/vscode.ts") },
      { find: "ai-sdk-provider-poe/code", replacement: resolve(__dirname, "./src/stubs/poe.ts") },
      { find: "i18next", replacement: resolve(__dirname, "./src/i18n/TranslationContext.tsx") },
      { find: "react-i18next", replacement: resolve(__dirname, "./src/i18n/TranslationContext.tsx") },
      { find: /^fs\/promises$/, replacement: resolve(__dirname, "./src/stubs/fs.ts") },
      { find: /^fs$/, replacement: resolve(__dirname, "./src/stubs/fs.ts") },
      { find: /^os$/, replacement: resolve(__dirname, "./src/stubs/os.ts") },
      { find: /^path$/, replacement: resolve(__dirname, "./src/stubs/path.ts") },
      { find: /^child_process$/, replacement: resolve(__dirname, "./src/stubs/child_process.ts") },
      { find: /^readline$/, replacement: resolve(__dirname, "./src/stubs/readline.ts") },
      { find: /^crypto$/, replacement: resolve(__dirname, "./src/stubs/crypto.ts") },
      { find: "@roo", replacement: resolve(__dirname, "./src/shared") },
      { find: "@types", replacement: resolve(__dirname, "./src/types") },
      { find: "@core", replacement: resolve(__dirname, "./src/core") },
      { find: "@shared", replacement: resolve(__dirname, "./src/shared") },
      {
        find: "@services/search/file-search",
        replacement: resolve(__dirname, "./src/services/search/file-search.browser.ts"),
      },
      { find: "@services/ripgrep", replacement: resolve(__dirname, "./src/services/ripgrep/index.browser.ts") },
      { find: "@services/mcp/McpHub", replacement: resolve(__dirname, "./src/services/mcp/McpHub.browser.ts") },
      { find: "@services", replacement: resolve(__dirname, "./src/services") },
    ],
  },
  build: {
    outDir: "dist/webview",
    emptyOutDir: true,
    rollupOptions: {
      output: {
        entryFileNames: `assets/[name].js`,
        chunkFileNames: `assets/[name].js`,
        assetFileNames: `assets/[name].[ext]`,
      },
    },
  },
});
