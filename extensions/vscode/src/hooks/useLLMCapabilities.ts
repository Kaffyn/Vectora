/**
 * useLLMCapabilities Hook - Phase 9
 * Hook para acessar capacidades do LLM ativo
 */

import { useMemo, useCallback } from "react";
import { ILLMProvider, ProviderCapabilities } from "../providers/baseProvider";
import { getProviderMetadata, providerSupportsCapability } from "../config/provider-metadata";
import { ProviderType } from "../providers/baseProvider";

// ============================================================================
// Types
// ============================================================================

export interface LLMCapabilities extends ProviderCapabilities {
  providerName: string;
  displayName: string;
}

// ============================================================================
// Hook Implementation
// ============================================================================

export function useLLMCapabilities(
  provider: ILLMProvider | null
): LLMCapabilities {
  const capabilities = useMemo(() => {
    if (!provider) {
      return {
        providerName: "none",
        displayName: "None",
        streaming: false,
        vision: false,
        functionCalling: false,
        embedding: false,
        caching: false,
        maxTokens: 0,
        contextWindow: 0,
        costPer1kInputTokens: 0,
        costPer1kOutputTokens: 0,
      };
    }

    const caps = provider.getCapabilities();
    const metadata = getProviderMetadata(provider.getType());

    return {
      ...caps,
      providerName: provider.getName(),
      displayName: metadata.displayName,
    };
  }, [provider]);

  return capabilities;
}

// ============================================================================
// Helper hook for checking specific capabilities
// ============================================================================

export function useHasCapability(
  provider: ILLMProvider | null,
  capability: keyof ProviderCapabilities
): boolean {
  return useMemo(() => {
    if (!provider) return false;
    return provider.hasCapability(capability);
  }, [provider, capability]);
}

// ============================================================================
// Helper hook for capability comparison
// ============================================================================

export function useCapabilityComparison(providers: ILLMProvider[]) {
  return useMemo(() => {
    const comparison: Record<string, Record<string, boolean | number>> = {};

    const capabilities: (keyof ProviderCapabilities)[] = [
      "streaming",
      "vision",
      "functionCalling",
      "embedding",
      "caching",
      "maxTokens",
      "contextWindow",
    ];

    for (const capability of capabilities) {
      comparison[capability] = {};
      for (const provider of providers) {
        const caps = provider.getCapabilities();
        comparison[capability][provider.getName()] = caps[capability];
      }
    }

    return comparison;
  }, [providers]);
}

// ============================================================================
// Helper hook for context window warnings
// ============================================================================

export interface ContextWindowWarning {
  warning: boolean;
  critical: boolean;
  usage: number;
  remaining: number;
  message: string;
}

export function useContextWindowWarning(
  provider: ILLMProvider | null,
  currentUsage: number
): ContextWindowWarning {
  return useMemo(() => {
    if (!provider) {
      return {
        warning: false,
        critical: false,
        usage: 0,
        remaining: 0,
        message: "",
      };
    }

    const contextWindow = provider.getMaxContextWindow();
    const remaining = contextWindow - currentUsage;
    const usagePercent = (currentUsage / contextWindow) * 100;

    let warning = false;
    let critical = false;
    let message = "";

    if (usagePercent > 90) {
      critical = true;
      message = `Critical: Using ${usagePercent.toFixed(1)}% of context window (${remaining} tokens remaining)`;
    } else if (usagePercent > 75) {
      warning = true;
      message = `Warning: Using ${usagePercent.toFixed(1)}% of context window (${remaining} tokens remaining)`;
    }

    return {
      warning,
      critical,
      usage: currentUsage,
      remaining: Math.max(0, remaining),
      message,
    };
  }, [provider, currentUsage]);
}

// ============================================================================
// Helper hook for cost estimation
// ============================================================================

export interface CostEstimate {
  inputCost: number;
  outputCost: number;
  totalCost: number;
  formatted: string;
}

export function useCostEstimate(
  provider: ILLMProvider | null,
  inputTokens: number,
  outputTokens: number
): CostEstimate {
  return useMemo(() => {
    if (!provider) {
      return {
        inputCost: 0,
        outputCost: 0,
        totalCost: 0,
        formatted: "$0.00",
      };
    }

    const cost = provider.estimateCost(inputTokens, outputTokens);

    return {
      ...cost,
      formatted: `$${cost.totalCost.toFixed(4)}`,
    };
  }, [provider, inputTokens, outputTokens]);
}

// ============================================================================
// Helper hook for model information
// ============================================================================

export interface ModelInfo {
  provider: ProviderType;
  displayName: string;
  contextWindow: number;
  maxTokens: number;
  streaming: boolean;
  vision: boolean;
  costPerInput: number;
  costPerOutput: number;
}

export function useModelInfo(provider: ILLMProvider | null): ModelInfo | null {
  return useMemo(() => {
    if (!provider) return null;

    const caps = provider.getCapabilities();
    const metadata = getProviderMetadata(provider.getType());

    return {
      provider: provider.getType(),
      displayName: metadata.displayName,
      contextWindow: caps.contextWindow,
      maxTokens: caps.maxTokens,
      streaming: caps.streaming,
      vision: caps.vision,
      costPerInput: caps.costPer1kInputTokens,
      costPerOutput: caps.costPer1kOutputTokens,
    };
  }, [provider]);
}

// ============================================================================
// Helper hook for feature availability
// ============================================================================

export interface FeatureAvailability {
  streaming: boolean;
  vision: boolean;
  functionCalling: boolean;
  embedding: boolean;
  caching: boolean;
  allFeatures: Record<string, boolean>;
}

export function useFeatureAvailability(
  provider: ILLMProvider | null
): FeatureAvailability {
  return useMemo(() => {
    if (!provider) {
      return {
        streaming: false,
        vision: false,
        functionCalling: false,
        embedding: false,
        caching: false,
        allFeatures: {},
      };
    }

    const caps = provider.getCapabilities();

    return {
      streaming: caps.streaming,
      vision: caps.vision,
      functionCalling: caps.functionCalling,
      embedding: caps.embedding,
      caching: caps.caching,
      allFeatures: {
        streaming: caps.streaming,
        vision: caps.vision,
        functionCalling: caps.functionCalling,
        embedding: caps.embedding,
        caching: caps.caching,
      },
    };
  }, [provider]);
}

// ============================================================================
// Helper hook for rate limit awareness
// ============================================================================

export interface RateLimitAwareness {
  hasRateLimit: boolean;
  requestsPerMinute: number;
  tokensPerMinute: number;
  recommendedWaitMs: number;
}

export function useRateLimitAwareness(
  provider: ILLMProvider | null,
  tokensToUse: number = 0
): RateLimitAwareness {
  return useMemo(() => {
    if (!provider) {
      return {
        hasRateLimit: false,
        requestsPerMinute: 999999,
        tokensPerMinute: 999999,
        recommendedWaitMs: 0,
      };
    }

    const rateLimitInfo = provider.getRateLimitInfo();

    if (!rateLimitInfo) {
      return {
        hasRateLimit: false,
        requestsPerMinute: 999999,
        tokensPerMinute: 999999,
        recommendedWaitMs: 0,
      };
    }

    // Calculate recommended wait time
    let recommendedWaitMs = 0;

    if (rateLimitInfo.requestsRemaining < 5) {
      const timeToReset =
        rateLimitInfo.resetAt.getTime() - Date.now();
      recommendedWaitMs = Math.max(0, timeToReset);
    }

    if (rateLimitInfo.tokensRemaining < tokensToUse * 2) {
      const timeToReset =
        rateLimitInfo.resetAt.getTime() - Date.now();
      recommendedWaitMs = Math.max(recommendedWaitMs, timeToReset);
    }

    return {
      hasRateLimit: true,
      requestsPerMinute: rateLimitInfo.requestsPerMinute,
      tokensPerMinute: rateLimitInfo.tokensPerMinute,
      recommendedWaitMs,
    };
  }, [provider, tokensToUse]);
}
