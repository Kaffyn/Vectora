/**
 * VS Code API Shim
 * Provides a mock interface to the actual VS Code API
 */

// Mock VS Code API for browser context
export const vscode = {
  postMessage: (message: any) => {
    // Will be replaced with actual vscode.postMessage in extension context
    console.log("postMessage (shim):", message);
  },

  getState: () => {
    return null;
  },

  setState: (state: any) => {
    // Will be replaced in extension context
  },

  acquireVsCodeApi: () => {
    // Return this object itself as the API
    return globalThis as any;
  }
};

// Fallback for cases where VS Code API is not available
if (typeof (globalThis as any).acquireVsCodeApi === 'function') {
  Object.assign(vscode, (globalThis as any).acquireVsCodeApi());
}
