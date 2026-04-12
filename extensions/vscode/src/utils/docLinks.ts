/**
 * Documentation links for error messages and help
 */

export const DOC_LINKS = {
  // API Documentation
  apiKeySetup: "https://docs.vectora.ai/setup",
  apiRateLimit: "https://docs.vectora.ai/limits",
  apiError: "https://docs.vectora.ai/errors",

  // Provider-specific docs
  geminiSetup: "https://ai.google.dev/docs",
  claudeSetup: "https://docs.anthropic.com",
  openaiSetup: "https://platform.openai.com/docs",

  // Features
  ragFeature: "https://docs.vectora.ai/rag",
  embeddingSetup: "https://docs.vectora.ai/embeddings",
  workspaceIndexing: "https://docs.vectora.ai/indexing",

  // Troubleshooting
  troubleshoot: "https://docs.vectora.ai/troubleshooting",
  faq: "https://docs.vectora.ai/faq",
};

export function getDocLink(key: keyof typeof DOC_LINKS): string {
  return DOC_LINKS[key];
}

export function buildDocLink(path: string, anchor?: string): string {
  const baseUrl = "https://docs.vectora.ai";
  const fullPath = path.startsWith("/") ? path : `/${path}`;
  const url = `${baseUrl}${fullPath}`;
  return anchor ? `${url}#${anchor}` : url;
}
