/**
 * ProviderSelector Component - Phase 9
 * Seletor visual de providers
 */

import React, { useMemo } from "react";
import { ProviderType } from "../../providers/baseProvider";
import {
  getProviderMetadata,
  getAllProvidersMetadata,
} from "../../config/provider-metadata";
import { useProviderManager } from "../../hooks/useProviderManager";

// ============================================================================
// Types
// ============================================================================

interface ProviderSelectorProps {
  value: ProviderType | null;
  onChange: (provider: ProviderType) => void;
  disabled?: boolean;
  showDescription?: boolean;
  className?: string;
}

// ============================================================================
// Provider Badge Component
// ============================================================================

interface ProviderBadgeProps {
  type: ProviderType;
  selected: boolean;
  onClick: () => void;
  disabled?: boolean;
}

function ProviderBadge({
  type,
  selected,
  onClick,
  disabled = false,
}: ProviderBadgeProps) {
  const metadata = getProviderMetadata(type);

  return (
    <button
      onClick={onClick}
      disabled={disabled}
      className={`
        relative px-4 py-2 rounded-lg border-2 transition-all
        ${
          selected
            ? "border-blue-500 bg-blue-50 shadow-md"
            : "border-gray-200 bg-white hover:border-gray-300"
        }
        ${disabled ? "opacity-50 cursor-not-allowed" : "cursor-pointer"}
        focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2
      `}
      title={metadata.description}
    >
      <div className="flex items-center gap-2">
        <span className="font-medium text-sm">{metadata.name}</span>
        {selected && (
          <svg
            className="w-4 h-4 text-blue-500"
            fill="currentColor"
            viewBox="0 0 20 20"
          >
            <path
              fillRule="evenodd"
              d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z"
              clipRule="evenodd"
            />
          </svg>
        )}
      </div>
    </button>
  );
}

// ============================================================================
// Provider Info Component
// ============================================================================

interface ProviderInfoProps {
  type: ProviderType;
}

function ProviderInfo({ type }: ProviderInfoProps) {
  const metadata = getProviderMetadata(type);

  return (
    <div className="bg-gray-50 p-3 rounded-lg border border-gray-200">
      <h3 className="font-semibold text-sm mb-2">{metadata.displayName}</h3>
      <p className="text-xs text-gray-600 mb-2">{metadata.description}</p>

      {/* Capabilities */}
      <div className="mb-2">
        <p className="text-xs font-medium text-gray-700 mb-1">Features:</p>
        <div className="flex flex-wrap gap-1">
          {metadata.capabilities.streaming && (
            <span className="bg-blue-100 text-blue-700 text-xs px-2 py-1 rounded">
              Streaming
            </span>
          )}
          {metadata.capabilities.vision && (
            <span className="bg-purple-100 text-purple-700 text-xs px-2 py-1 rounded">
              Vision
            </span>
          )}
          {metadata.capabilities.functionCalling && (
            <span className="bg-green-100 text-green-700 text-xs px-2 py-1 rounded">
              Function Calling
            </span>
          )}
          {metadata.capabilities.embedding && (
            <span className="bg-orange-100 text-orange-700 text-xs px-2 py-1 rounded">
              Embeddings
            </span>
          )}
          {metadata.capabilities.caching && (
            <span className="bg-pink-100 text-pink-700 text-xs px-2 py-1 rounded">
              Caching
            </span>
          )}
        </div>
      </div>

      {/* Context Window */}
      <div className="text-xs text-gray-600">
        <p>Context: {(metadata.capabilities.contextWindow / 1000).toFixed(0)}K tokens</p>
        <p>
          Cost: ${metadata.capabilities.costPer1kInputTokens}/1K input,
          ${metadata.capabilities.costPer1kOutputTokens}/1K output
        </p>
      </div>
    </div>
  );
}

// ============================================================================
// Main Component
// ============================================================================

export function ProviderSelector({
  value,
  onChange,
  disabled = false,
  showDescription = true,
  className = "",
}: ProviderSelectorProps) {
  const { availableProviders } = useProviderManager();

  const providers = useMemo(() => {
    return availableProviders.sort((a, b) => {
      const metaA = getProviderMetadata(a);
      const metaB = getProviderMetadata(b);
      return metaA.name.localeCompare(metaB.name);
    });
  }, [availableProviders]);

  return (
    <div className={`space-y-4 ${className}`}>
      {/* Provider Grid */}
      <div className="grid grid-cols-2 gap-2 sm:grid-cols-4">
        {providers.map((provider) => (
          <ProviderBadge
            key={provider}
            type={provider}
            selected={value === provider}
            onClick={() => onChange(provider)}
            disabled={disabled}
          />
        ))}
      </div>

      {/* Provider Info */}
      {showDescription && value && (
        <ProviderInfo type={value} />
      )}
    </div>
  );
}

export default ProviderSelector;
