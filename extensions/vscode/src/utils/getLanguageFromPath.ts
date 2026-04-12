/**
 * Get syntax highlighting language from file path
 */

const languageMap: Record<string, string> = {
  ".js": "javascript",
  ".jsx": "javascript",
  ".ts": "typescript",
  ".tsx": "typescript",
  ".py": "python",
  ".java": "java",
  ".cpp": "cpp",
  ".c": "c",
  ".cc": "cpp",
  ".cxx": "cpp",
  ".h": "c",
  ".hpp": "cpp",
  ".cs": "csharp",
  ".php": "php",
  ".rb": "ruby",
  ".go": "go",
  ".rs": "rust",
  ".swift": "swift",
  ".kt": "kotlin",
  ".scala": "scala",
  ".r": "r",
  ".R": "r",
  ".sql": "sql",
  ".html": "html",
  ".htm": "html",
  ".css": "css",
  ".scss": "scss",
  ".sass": "sass",
  ".less": "less",
  ".json": "json",
  ".xml": "xml",
  ".yaml": "yaml",
  ".yml": "yaml",
  ".toml": "toml",
  ".bash": "bash",
  ".sh": "bash",
  ".zsh": "zsh",
  ".fish": "fish",
  ".ps1": "powershell",
  ".md": "markdown",
  ".tex": "latex",
};

export function getLanguageFromPath(filePath: string): string {
  if (!filePath) return "text";

  const ext = filePath.substring(filePath.lastIndexOf(".")).toLowerCase();
  return languageMap[ext] || "text";
}
