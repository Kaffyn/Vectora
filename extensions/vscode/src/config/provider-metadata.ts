/**
 * Provider Metadata - Phase 9
 * Metadata descritivo para cada provider
 */

import { ProviderType, ProviderCapabilities } from "../providers/baseProvider";

// ============================================================================
// Types
// ============================================================================

export interface ProviderMetadata {
  type: ProviderType;
  name: string;
  displayName: string;
  description: string;
  website: string;
  documentation: string;
  apiKeyFormat: string;
  apiKeyLength: { min: number; max: number };
  supportEmail: string;
  status: "stable" | "beta" | "deprecated";
  capabilities: ProviderCapabilities;
  regionAvailability: string[];
  pricingModel: string;
  freeTrialAvailable: boolean;
  rateLimit: {
    requestsPerMinute: number;
    tokensPerMinute: number;
  };
  supportedLanguages: string[];
  authenticationType: "api_key" | "oauth" | "bearer_token" | "custom";
  customHeaders?: Record<string, string>;
  notes?: string;
}

// ============================================================================
// Provider Metadata Registry
// ============================================================================

export const PROVIDER_METADATA: Record<ProviderType, ProviderMetadata> = {
  claude: {
    type: "claude",
    name: "Claude",
    displayName: "Claude (Anthropic)",
    description: "Anthropic's advanced AI assistant with strong reasoning capabilities",
    website: "https://anthropic.com",
    documentation: "https://docs.anthropic.com",
    apiKeyFormat: "sk-ant-*",
    apiKeyLength: { min: 30, max: 100 },
    supportEmail: "support@anthropic.com",
    status: "stable",
    capabilities: {
      streaming: true,
      vision: true,
      functionCalling: true,
      embedding: false,
      caching: true,
      maxTokens: 4096,
      contextWindow: 200000,
      costPer1kInputTokens: 0.003,
      costPer1kOutputTokens: 0.015,
    },
    regionAvailability: ["US", "EU"],
    pricingModel: "Pay per use - per 1K tokens",
    freeTrialAvailable: true,
    rateLimit: {
      requestsPerMinute: 60,
      tokensPerMinute: 1000000,
    },
    supportedLanguages: ["English", "Multilingual"],
    authenticationType: "api_key",
    customHeaders: {
      "anthropic-version": "2023-06-01",
    },
    notes: "Supports message caching for repeated prompts",
  },

  openai: {
    type: "openai",
    name: "OpenAI",
    displayName: "OpenAI (GPT)",
    description: "OpenAI's GPT models including GPT-4 with advanced capabilities",
    website: "https://openai.com",
    documentation: "https://platform.openai.com/docs",
    apiKeyFormat: "sk-*",
    apiKeyLength: { min: 30, max: 100 },
    supportEmail: "support@openai.com",
    status: "stable",
    capabilities: {
      streaming: true,
      vision: true,
      functionCalling: true,
      embedding: true,
      caching: false,
      maxTokens: 4096,
      contextWindow: 128000,
      costPer1kInputTokens: 0.03,
      costPer1kOutputTokens: 0.06,
    },
    regionAvailability: ["US", "EU", "APAC"],
    pricingModel: "Pay per use - per 1K tokens",
    freeTrialAvailable: true,
    rateLimit: {
      requestsPerMinute: 200,
      tokensPerMinute: 90000,
    },
    supportedLanguages: ["English", "Multilingual"],
    authenticationType: "api_key",
    notes: "Requires organization ID for team accounts",
  },

  llama: {
    type: "llama",
    name: "Llama",
    displayName: "Llama (Meta)",
    description: "Meta's open-source language models - local or cloud via Groq",
    website: "https://www.llama.com",
    documentation: "https://github.com/meta-llama/llama",
    apiKeyFormat: "Optional - for Groq cloud",
    apiKeyLength: { min: 0, max: 100 },
    supportEmail: "support@groq.com",
    status: "stable",
    capabilities: {
      streaming: true,
      vision: false,
      functionCalling: false,
      embedding: false,
      caching: false,
      maxTokens: 2048,
      contextWindow: 4096,
      costPer1kInputTokens: 0.0,
      costPer1kOutputTokens: 0.0,
    },
    regionAvailability: ["Global"],
    pricingModel: "Free (local) or pay per use (Groq)",
    freeTrialAvailable: true,
    rateLimit: {
      requestsPerMinute: 30,
      tokensPerMinute: 300000,
    },
    supportedLanguages: ["English"],
    authenticationType: "api_key",
    notes: "Local: Install Ollama. Cloud: Use Groq API",
  },

  gemini: {
    type: "gemini",
    name: "Gemini",
    displayName: "Gemini (Google)",
    description: "Google's multimodal AI with 1M token context window",
    website: "https://gemini.google.com",
    documentation: "https://ai.google.dev",
    apiKeyFormat: "AIza*",
    apiKeyLength: { min: 30, max: 100 },
    supportEmail: "support@google.com",
    status: "stable",
    capabilities: {
      streaming: true,
      vision: true,
      functionCalling: false,
      embedding: true,
      caching: true,
      maxTokens: 8000,
      contextWindow: 1000000,
      costPer1kInputTokens: 0.0075,
      costPer1kOutputTokens: 0.03,
    },
    regionAvailability: ["Global"],
    pricingModel: "Free tier available, then pay per use",
    freeTrialAvailable: true,
    rateLimit: {
      requestsPerMinute: 60,
      tokensPerMinute: 1000000,
    },
    supportedLanguages: ["English", "Multilingual"],
    authenticationType: "api_key",
    notes: "Excellent for long context tasks. Free tier has rate limits.",
  },
};

// ============================================================================
// Helper Functions
// ============================================================================

/**
 * Get metadata for a provider type
 */
export function getProviderMetadata(type: ProviderType): ProviderMetadata {
  const metadata = PROVIDER_METADATA[type];
  if (!metadata) {
    throw new Error(`Unknown provider type: ${type}`);
  }
  return metadata;
}

/**
 * Get all provider metadata
 */
export function getAllProvidersMetadata(): ProviderMetadata[] {
  return Object.values(PROVIDER_METADATA);
}

/**
 * Check if provider supports a capability
 */
export function providerSupportsCapability(
  type: ProviderType,
  capability: keyof ProviderCapabilities
): boolean {
  const metadata = getProviderMetadata(type);
  const value = metadata.capabilities[capability];
  return (value as boolean | number) !== false && (value as boolean | number) !== 0;
}

/**
 * Compare capabilities between providers
 */
export function compareProviders(
  types: ProviderType[]
): Record<string, Record<ProviderType, boolean | number>> {
  const comparison: Record<string, Record<ProviderType, boolean | number>> = {};

  const capabilityKeys = [
    "streaming",
    "vision",
    "functionCalling",
    "embedding",
    "caching",
    "maxTokens",
    "contextWindow",
  ];

  for (const capability of capabilityKeys) {
    comparison[capability] = {};
    for (const type of types) {
      const metadata = getProviderMetadata(type);
      comparison[capability][type] =
        metadata.capabilities[capability as keyof ProviderCapabilities];
    }
  }

  return comparison;
}

/**
 * Get providers sorted by a metric
 */
export function getProvidersSortedBy(
  metric: "cost" | "speed" | "context" | "capabilities"
): ProviderType[] {
  const all = Object.keys(PROVIDER_METADATA) as ProviderType[];

  switch (metric) {
    case "cost":
      return all.sort((a, b) => {
        const costA =
          PROVIDER_METADATA[a].capabilities.costPer1kInputTokens +
          PROVIDER_METADATA[a].capabilities.costPer1kOutputTokens;
        const costB =
          PROVIDER_METADATA[b].capabilities.costPer1kInputTokens +
          PROVIDER_METADATA[b].capabilities.costPer1kOutputTokens;
        return costA - costB;
      });

    case "context":
      return all.sort(
        (a, b) =>
          PROVIDER_METADATA[b].capabilities.contextWindow -
          PROVIDER_METADATA[a].capabilities.contextWindow
      );

    case "capabilities":
      return all.sort((a, b) => {
        const countA = Object.values(PROVIDER_METADATA[a].capabilities).filter(
          (v) => v === true
        ).length;
        const countB = Object.values(PROVIDER_METADATA[b].capabilities).filter(
          (v) => v === true
        ).length;
        return countB - countA;
      });

    case "speed":
      // Estimate based on model size/efficiency
      return all.sort((a, b) => {
        const speedA = PROVIDER_METADATA[a].rateLimit.tokensPerMinute;
        const speedB = PROVIDER_METADATA[b].rateLimit.tokensPerMinute;
        return speedB - speedA;
      });

    default:
      return all;
  }
}

/**
 * Get providers with specific capability
 */
export function getProvidersWithCapability(
  capability: keyof ProviderCapabilities
): ProviderType[] {
  return (Object.keys(PROVIDER_METADATA) as ProviderType[]).filter((type) =>
    providerSupportsCapability(type, capability)
  );
}

/**
 * Find best provider for use case
 */
export function findBestProvider(requirements: {
  vision?: boolean;
  embedding?: boolean;
  functionCalling?: boolean;
  streaming?: boolean;
  caching?: boolean;
  maxContextWindow?: number;
  preferFree?: boolean;
}): ProviderType | null {
  const candidates = Object.keys(PROVIDER_METADATA) as ProviderType[];

  for (const type of candidates) {
    const metadata = PROVIDER_METADATA[type];

    if (requirements.vision && !metadata.capabilities.vision) continue;
    if (requirements.embedding && !metadata.capabilities.embedding) continue;
    if (requirements.functionCalling && !metadata.capabilities.functionCalling)
      continue;
    if (requirements.streaming && !metadata.capabilities.streaming) continue;
    if (requirements.caching && !metadata.capabilities.caching) continue;
    if (
      requirements.maxContextWindow &&
      metadata.capabilities.contextWindow < requirements.maxContextWindow
    )
      continue;
    if (requirements.preferFree && metadata.pricingModel !== "Free (local)") {
      continue;
    }

    return type;
  }

  return null;
}
