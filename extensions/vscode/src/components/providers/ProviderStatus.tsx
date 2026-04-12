/**
 * ProviderStatus Component - Phase 9
 * Exibe status e métricas do provider ativo
 */

import React, { useMemo } from "react";
import { ILLMProvider, ProviderCapabilities } from "../../providers/baseProvider";
import { getProviderMetadata } from "../../config/provider-metadata";
import { useLLMCapabilities, useRateLimitAwareness } from "../../hooks/useLLMCapabilities";

// ============================================================================
// Types
// ============================================================================

interface ProviderStatusProps {
  provider: ILLMProvider | null;
  isLoading?: boolean;
  error?: Error | null;
  className?: string;
}

// ============================================================================
// Status Badge Component
// ============================================================================

interface StatusBadgeProps {
  status: "ready" | "loading" | "error" | "rate-limited";
  message: string;
}

function StatusBadge({ status, message }: StatusBadgeProps) {
  const colors = {
    ready: "bg-green-100 text-green-800 border-green-300",
    loading: "bg-yellow-100 text-yellow-800 border-yellow-300",
    error: "bg-red-100 text-red-800 border-red-300",
    "rate-limited": "bg-orange-100 text-orange-800 border-orange-300",
  };

  const icons = {
    ready: "●",
    loading: "⟳",
    error: "⚠",
    "rate-limited": "⏱",
  };

  return (
    <div className={`inline-flex items-center gap-2 px-3 py-1 rounded-full border ${colors[status]}`}>
      <span className="text-sm font-medium">{icons[status]}</span>
      <span className="text-sm">{message}</span>
    </div>
  );
}

// ============================================================================
// Capability Indicator Component
// ============================================================================

interface CapabilityIndicatorProps {
  name: string;
  enabled: boolean;
}

function CapabilityIndicator({ name, enabled }: CapabilityIndicatorProps) {
  return (
    <div className="flex items-center gap-2 text-sm">
      <div
        className={`w-2 h-2 rounded-full ${
          enabled ? "bg-green-500" : "bg-gray-300"
        }`}
      />
      <span className={enabled ? "text-gray-800" : "text-gray-500"}>
        {name}
      </span>
    </div>
  );
}

// ============================================================================
// Rate Limit Indicator Component
// ============================================================================

interface RateLimitIndicatorProps {
  info: ReturnType<typeof useRateLimitAwareness>;
}

function RateLimitIndicator({ info }: RateLimitIndicatorProps) {
  if (!info.hasRateLimit) {
    return null;
  }

  const requestsPercent =
    (info.requestsPerMinute > 0
      ? 100 - ((info.requestsPerMinute / info.requestsPerMinute) * 100)
      : 0);

  return (
    <div className="space-y-2 pt-2 border-t border-gray-200">
      <p className="text-xs font-medium text-gray-700">Rate Limits (per minute)</p>

      {/* Requests */}
      <div className="space-y-1">
        <div className="flex justify-between text-xs">
          <span>Requests</span>
          <span className="font-mono">
            {info.requestsPerMinute}
          </span>
        </div>
        <div className="w-full h-2 bg-gray-200 rounded-full overflow-hidden">
          <div
            className="h-full bg-blue-500 transition-all"
            style={{ width: `${Math.min(100, requestsPercent)}%` }}
          />
        </div>
      </div>

      {/* Tokens */}
      <div className="space-y-1">
        <div className="flex justify-between text-xs">
          <span>Tokens</span>
          <span className="font-mono">
            {(info.tokensPerMinute / 1000).toFixed(0)}K
          </span>
        </div>
        <div className="w-full h-2 bg-gray-200 rounded-full overflow-hidden">
          <div
            className="h-full bg-blue-500 transition-all"
            style={{ width: "50%" }}
          />
        </div>
      </div>

      {info.recommendedWaitMs > 0 && (
        <div className="mt-2 p-2 bg-yellow-50 border border-yellow-200 rounded text-xs text-yellow-800">
          Recommended wait: {Math.ceil(info.recommendedWaitMs / 1000)}s
        </div>
      )}
    </div>
  );
}

// ============================================================================
// Main Component
// ============================================================================

export function ProviderStatus({
  provider,
  isLoading = false,
  error = null,
  className = "",
}: ProviderStatusProps) {
  const capabilities = useLLMCapabilities(provider);
  const rateLimitInfo = useRateLimitAwareness(provider);

  const metadata = useMemo(() => {
    return provider ? getProviderMetadata(provider.getType()) : null;
  }, [provider]);

  const status = useMemo(() => {
    if (error) return { type: "error" as const, message: "Error" };
    if (isLoading) return { type: "loading" as const, message: "Initializing..." };
    if (rateLimitInfo.recommendedWaitMs > 0) {
      return { type: "rate-limited" as const, message: "Rate limited" };
    }
    if (provider?.isReady()) return { type: "ready" as const, message: "Ready" };
    return { type: "loading" as const, message: "Not ready" };
  }, [error, isLoading, provider, rateLimitInfo.recommendedWaitMs]);

  if (!provider && !isLoading && !error) {
    return (
      <div
        className={`p-4 bg-gray-50 border border-gray-200 rounded-lg text-center text-gray-500 ${className}`}
      >
        <p className="text-sm">No provider selected</p>
      </div>
    );
  }

  return (
    <div className={`p-4 bg-white border border-gray-200 rounded-lg space-y-3 ${className}`}>
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h3 className="font-semibold text-gray-800">
            {metadata?.displayName || "Provider"}
          </h3>
          {error && (
            <p className="text-xs text-red-600 mt-1">{error.message}</p>
          )}
        </div>
        <StatusBadge status={status.type} message={status.message} />
      </div>

      {/* Provider Info */}
      {provider && !error && (
        <>
          {/* Capabilities */}
          <div className="space-y-1">
            <p className="text-xs font-medium text-gray-700">Features</p>
            <div className="grid grid-cols-2 gap-2">
              <CapabilityIndicator
                name="Streaming"
                enabled={capabilities.streaming}
              />
              <CapabilityIndicator name="Vision" enabled={capabilities.vision} />
              <CapabilityIndicator
                name="Function Calling"
                enabled={capabilities.functionCalling}
              />
              <CapabilityIndicator
                name="Embedding"
                enabled={capabilities.embedding}
              />
              <CapabilityIndicator name="Caching" enabled={capabilities.caching} />
            </div>
          </div>

          {/* Context Window */}
          <div className="space-y-1 pt-2 border-t border-gray-200">
            <p className="text-xs font-medium text-gray-700">Context Window</p>
            <div className="flex justify-between text-xs">
              <span>{(capabilities.contextWindow / 1000).toFixed(0)}K tokens</span>
              <span className="font-mono text-gray-500">
                Max: {capabilities.maxTokens}
              </span>
            </div>
            <div className="w-full h-2 bg-gray-200 rounded-full overflow-hidden">
              <div
                className="h-full bg-green-500"
                style={{
                  width: `${Math.min(100, (capabilities.maxTokens / capabilities.contextWindow) * 100)}%`,
                }}
              />
            </div>
          </div>

          {/* Pricing */}
          <div className="space-y-1 pt-2 border-t border-gray-200">
            <p className="text-xs font-medium text-gray-700">Pricing (per 1K tokens)</p>
            <div className="flex justify-between text-xs text-gray-600">
              <span>Input: ${capabilities.costPer1kInputTokens}</span>
              <span>Output: ${capabilities.costPer1kOutputTokens}</span>
            </div>
          </div>

          {/* Rate Limits */}
          <RateLimitIndicator info={rateLimitInfo} />
        </>
      )}
    </div>
  );
}

export default ProviderStatus;
