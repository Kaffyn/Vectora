/**
 * Code highlighter utilities using Shiki
 */

export interface Highlighter {
  codeToHast(code: string, options: any): Promise<any>;
}

let highlighterInstance: Highlighter | null = null;
const loadedLanguages = new Set<string>();

export async function getHighlighter(lang?: string): Promise<Highlighter> {
  // Stub implementation - returns a mock highlighter
  // Phase 2+ will implement proper Shiki integration
  if (!highlighterInstance) {
    highlighterInstance = {
      async codeToHast(code: string, options: any) {
        // Return a simple HAST structure
        return {
          type: "root",
          children: [
            {
              type: "element",
              tagName: "pre",
              properties: {
                style: "padding: 0; margin: 0;",
              },
              children: [
                {
                  type: "element",
                  tagName: "code",
                  properties: {
                    class: `hljs language-${options.lang || "text"}`,
                  },
                  children: [
                    {
                      type: "text",
                      value: code,
                    },
                  ],
                },
              ],
            },
          ],
        };
      },
    };
  }

  // Mark the language as loaded
  if (lang) {
    loadedLanguages.add(lang);
  }

  return highlighterInstance;
}

export function isLanguageLoaded(lang: string): boolean {
  return loadedLanguages.has(lang);
}

export function normalizeLanguage(lang: string): string {
  // Normalize language names
  const normalized: Record<string, string> = {
    "js": "javascript",
    "ts": "typescript",
    "py": "python",
    "sh": "bash",
    "yml": "yaml",
  };

  return normalized[lang.toLowerCase()] || lang.toLowerCase();
}
