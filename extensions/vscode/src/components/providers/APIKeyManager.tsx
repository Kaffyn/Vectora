/**
 * APIKeyManager Component - Phase 9
 * Gerenciador seguro de API keys para providers
 */

import React, { useState, useCallback } from "react";
import { ProviderType } from "../../providers/baseProvider";
import { getProviderMetadata } from "../../config/provider-metadata";

// ============================================================================
// Types
// ============================================================================

interface APIKeyManagerProps {
  provider: ProviderType;
  apiKey?: string;
  onApiKeyChange: (apiKey: string) => void;
  onSave?: (apiKey: string) => Promise<void>;
  isSaving?: boolean;
  className?: string;
}

// ============================================================================
// Component
// ============================================================================

export function APIKeyManager({
  provider,
  apiKey = "",
  onApiKeyChange,
  onSave,
  isSaving = false,
  className = "",
}: APIKeyManagerProps) {
  const [showKey, setShowKey] = useState(false);
  const [localValue, setLocalValue] = useState(apiKey);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState(false);

  const metadata = getProviderMetadata(provider);

  // ============================================================================
  // Handle input change
  // ============================================================================

  const handleChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      const value = e.target.value;
      setLocalValue(value);
      onApiKeyChange(value);
      setError(null);
    },
    [onApiKeyChange]
  );

  // ============================================================================
  // Handle save
  // ============================================================================

  const handleSave = useCallback(async () => {
    try {
      setError(null);

      if (!localValue.trim()) {
        setError("API key is required");
        return;
      }

      if (localValue.length < metadata.apiKeyLength.min) {
        setError(
          `API key must be at least ${metadata.apiKeyLength.min} characters`
        );
        return;
      }

      if (localValue.length > metadata.apiKeyLength.max) {
        setError(
          `API key must not exceed ${metadata.apiKeyLength.max} characters`
        );
        return;
      }

      if (onSave) {
        await onSave(localValue);
      }

      setSuccess(true);
      setTimeout(() => setSuccess(false), 3000);
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : String(err);
      setError(errorMessage);
    }
  }, [localValue, metadata, onSave]);

  // ============================================================================
  // Handle paste
  // ============================================================================

  const handlePaste = useCallback(async () => {
    try {
      const text = await navigator.clipboard.readText();
      setLocalValue(text);
      onApiKeyChange(text);
      setError(null);
    } catch (err) {
      setError("Failed to paste from clipboard");
    }
  }, [onApiKeyChange]);

  // ============================================================================
  // Handle copy
  // ============================================================================

  const handleCopy = useCallback(() => {
    if (!localValue) return;
    navigator.clipboard.writeText(localValue).then(() => {
      setSuccess(true);
      setTimeout(() => setSuccess(false), 2000);
    });
  }, [localValue]);

  // ============================================================================
  // Validate format
  // ============================================================================

  const validateFormat = useCallback(() => {
    if (!localValue) return true;

    if (
      metadata.apiKeyFormat &&
      metadata.apiKeyFormat !== "Optional - for Groq cloud"
    ) {
      const pattern = metadata.apiKeyFormat.replace("*", ".+");
      const regex = new RegExp(`^${pattern}$`);
      return regex.test(localValue);
    }

    return true;
  }, [localValue, metadata.apiKeyFormat]);

  const isValidFormat = validateFormat();

  return (
    <div className={`space-y-3 ${className}`}>
      {/* Header */}
      <div className="flex items-center justify-between">
        <label className="block text-sm font-medium text-gray-700">
          API Key for {metadata.displayName}
        </label>
        {isValidFormat && !error ? (
          <span className="text-xs text-green-600">●</span>
        ) : (
          <span className="text-xs text-gray-400">●</span>
        )}
      </div>

      {/* Info */}
      <div className="text-xs text-gray-500 bg-gray-50 p-2 rounded border border-gray-200">
        <p>
          Format: <code className="font-mono">{metadata.apiKeyFormat}</code>
        </p>
        <p className="mt-1">
          Get your API key at:{" "}
          <a
            href={metadata.website}
            target="_blank"
            rel="noopener noreferrer"
            className="text-blue-600 hover:underline"
          >
            {metadata.name}
          </a>
        </p>
      </div>

      {/* Input */}
      <div className="relative">
        <input
          type={showKey ? "text" : "password"}
          value={localValue}
          onChange={handleChange}
          placeholder="Paste your API key here..."
          className={`w-full px-3 py-2 border rounded-lg text-sm font-mono focus:outline-none focus:ring-2 ${
            error
              ? "border-red-300 focus:ring-red-500"
              : isValidFormat
              ? "border-green-300 focus:ring-green-500"
              : "border-gray-300 focus:ring-blue-500"
          }`}
        />

        {/* Actions */}
        <div className="absolute right-2 top-1/2 -translate-y-1/2 flex gap-1">
          <button
            onClick={() => setShowKey(!showKey)}
            className="p-1 text-gray-400 hover:text-gray-600 focus:outline-none"
            title={showKey ? "Hide" : "Show"}
            type="button"
          >
            {showKey ? "●" : "○"}
          </button>

          {localValue && (
            <button
              onClick={handleCopy}
              className="p-1 text-gray-400 hover:text-gray-600 focus:outline-none text-xs"
              title="Copy"
              type="button"
            >
              📋
            </button>
          )}

          <button
            onClick={handlePaste}
            className="p-1 text-gray-400 hover:text-gray-600 focus:outline-none text-xs"
            title="Paste"
            type="button"
          >
            📌
          </button>
        </div>
      </div>

      {/* Status Messages */}
      {error && (
        <div className="p-2 bg-red-50 border border-red-200 rounded text-xs text-red-700 flex items-center gap-2">
          <span>⚠</span>
          <span>{error}</span>
        </div>
      )}

      {success && (
        <div className="p-2 bg-green-50 border border-green-200 rounded text-xs text-green-700 flex items-center gap-2">
          <span>✓</span>
          <span>Success!</span>
        </div>
      )}

      {!isValidFormat && localValue && !error && (
        <div className="p-2 bg-yellow-50 border border-yellow-200 rounded text-xs text-yellow-700 flex items-center gap-2">
          <span>⚠</span>
          <span>API key format may be incorrect</span>
        </div>
      )}

      {/* Save Button */}
      {onSave && (
        <button
          onClick={handleSave}
          disabled={!localValue || !isValidFormat || isSaving}
          className={`w-full px-4 py-2 rounded-lg font-medium text-sm transition-colors ${
            localValue && isValidFormat
              ? "bg-blue-600 text-white hover:bg-blue-700 disabled:opacity-50"
              : "bg-gray-300 text-gray-500 cursor-not-allowed"
          }`}
        >
          {isSaving ? "Saving..." : "Save API Key"}
        </button>
      )}

      {/* Security Notice */}
      <div className="text-xs text-gray-500 bg-blue-50 p-2 rounded border border-blue-100">
        <p className="font-medium text-gray-600 mb-1">Security:</p>
        <ul className="list-disc list-inside space-y-1">
          <li>API keys are encrypted before storage</li>
          <li>Never share your API key publicly</li>
          <li>Rotate keys regularly for security</li>
        </ul>
      </div>
    </div>
  );
}

export default APIKeyManager;
