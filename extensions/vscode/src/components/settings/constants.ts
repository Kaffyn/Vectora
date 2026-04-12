/**
 * Provider configuration constants
 * TODO: Replace with Vectora provider configuration during Phase 2
 */

export interface ProviderInfo {
  value: string;
  label: string;
  proxy?: boolean;
}

export const PROVIDERS: ProviderInfo[] = [
  { value: "gemini", label: "Google Gemini", proxy: false },
  { value: "claude", label: "Anthropic Claude", proxy: false },
  { value: "openai", label: "OpenAI GPT", proxy: false },
  { value: "openrouter", label: "OpenRouter", proxy: true },
];
