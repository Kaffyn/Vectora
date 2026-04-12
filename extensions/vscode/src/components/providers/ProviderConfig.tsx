/**
 * ProviderConfig Component - Phase 9
 * Painel de configuração completo de provider
 */

import React, { useState } from "react";
import { ProviderType } from "../../providers/baseProvider";
import { useProviderConfig } from "../../hooks/useProviderConfig";
import { APIKeyManager } from "./APIKeyManager";
import { ProviderStatus } from "./ProviderStatus";
import { useProviderManager } from "../../hooks/useProviderManager";

// ============================================================================
// Types
// ============================================================================

interface ProviderConfigPanelProps {
  provider: ProviderType;
  onConfigSaved?: (config: any) => void;
  className?: string;
}

// ============================================================================
// Component
// ============================================================================

export function ProviderConfig({
  provider,
  onConfigSaved,
  className = "",
}: ProviderConfigPanelProps) {
  const {
    profile,
    config,
    isValid,
    validationErrors,
    validationWarnings,
    isDirty,
    isSaving,
    updateApiKey,
    updateBaseUrl,
    updateTimeout,
    updateRetryAttempts,
    updateMaxTokens,
    validate,
    save,
    reset,
  } = useProviderConfig(provider, async (config) => {
    onConfigSaved?.(config);
  });

  const { activeInstance, activeProvider } = useProviderManager();
  const [expandedSections, setExpandedSections] = useState<
    Record<string, boolean>
  >({
    credentials: true,
    network: false,
    advanced: false,
  });

  // ============================================================================
  // Toggle section
  // ============================================================================

  const toggleSection = (section: string) => {
    setExpandedSections((prev) => ({
      ...prev,
      [section]: !prev[section],
    }));
  };

  if (!profile || !config) {
    return (
      <div className={`p-4 text-center text-gray-500 ${className}`}>
        Loading configuration...
      </div>
    );
  }

  return (
    <div className={`space-y-4 ${className}`}>
      {/* Status */}
      <ProviderStatus
        provider={activeProvider === provider ? activeInstance : null}
        isLoading={isSaving}
      />

      {/* Validation Messages */}
      {validationErrors.length > 0 && (
        <div className="p-3 bg-red-50 border border-red-200 rounded-lg">
          <p className="font-medium text-sm text-red-800 mb-1">Errors:</p>
          <ul className="list-disc list-inside space-y-1">
            {validationErrors.map((error) => (
              <li key={error} className="text-xs text-red-700">
                {error}
              </li>
            ))}
          </ul>
        </div>
      )}

      {validationWarnings.length > 0 && (
        <div className="p-3 bg-yellow-50 border border-yellow-200 rounded-lg">
          <p className="font-medium text-sm text-yellow-800 mb-1">Warnings:</p>
          <ul className="list-disc list-inside space-y-1">
            {validationWarnings.map((warning) => (
              <li key={warning} className="text-xs text-yellow-700">
                {warning}
              </li>
            ))}
          </ul>
        </div>
      )}

      {/* Credentials Section */}
      <div className="border border-gray-200 rounded-lg overflow-hidden">
        <button
          onClick={() => toggleSection("credentials")}
          className="w-full px-4 py-3 bg-gray-50 hover:bg-gray-100 flex items-center justify-between font-medium text-sm text-gray-700 border-b border-gray-200"
        >
          <span>Credentials</span>
          <span className="text-xs">
            {expandedSections.credentials ? "▼" : "▶"}
          </span>
        </button>

        {expandedSections.credentials && (
          <div className="p-4 space-y-4">
            <APIKeyManager
              provider={provider}
              apiKey={config.apiKey || ""}
              onApiKeyChange={updateApiKey}
              onSave={async (key) => {
                updateApiKey(key);
                await save();
              }}
              isSaving={isSaving}
            />
          </div>
        )}
      </div>

      {/* Network Section */}
      <div className="border border-gray-200 rounded-lg overflow-hidden">
        <button
          onClick={() => toggleSection("network")}
          className="w-full px-4 py-3 bg-gray-50 hover:bg-gray-100 flex items-center justify-between font-medium text-sm text-gray-700 border-b border-gray-200"
        >
          <span>Network Settings</span>
          <span className="text-xs">
            {expandedSections.network ? "▼" : "▶"}
          </span>
        </button>

        {expandedSections.network && (
          <div className="p-4 space-y-4">
            {/* Base URL */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Base URL (Optional)
              </label>
              <input
                type="text"
                value={profile.network?.baseUrl || ""}
                onChange={(e) => updateBaseUrl(e.target.value)}
                placeholder="https://api.example.com"
                className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>

            {/* Timeout */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Timeout (ms)
              </label>
              <input
                type="number"
                value={profile.network?.timeout || 30000}
                onChange={(e) => updateTimeout(parseInt(e.target.value))}
                min="1000"
                max="120000"
                step="1000"
                className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
              <p className="text-xs text-gray-500 mt-1">
                How long to wait for responses
              </p>
            </div>

            {/* Retry Attempts */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Retry Attempts
              </label>
              <input
                type="number"
                value={profile.network?.retryAttempts || 3}
                onChange={(e) => updateRetryAttempts(parseInt(e.target.value))}
                min="1"
                max="10"
                className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
              <p className="text-xs text-gray-500 mt-1">
                Number of times to retry failed requests
              </p>
            </div>
          </div>
        )}
      </div>

      {/* Advanced Section */}
      <div className="border border-gray-200 rounded-lg overflow-hidden">
        <button
          onClick={() => toggleSection("advanced")}
          className="w-full px-4 py-3 bg-gray-50 hover:bg-gray-100 flex items-center justify-between font-medium text-sm text-gray-700 border-b border-gray-200"
        >
          <span>Advanced Options</span>
          <span className="text-xs">
            {expandedSections.advanced ? "▼" : "▶"}
          </span>
        </button>

        {expandedSections.advanced && (
          <div className="p-4 space-y-4">
            {/* Max Tokens */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Max Tokens per Request
              </label>
              <input
                type="number"
                value={profile.limits?.maxTokensPerRequest || 4000}
                onChange={(e) => updateMaxTokens(parseInt(e.target.value))}
                min="100"
                max="100000"
                step="100"
                className="w-full px-3 py-2 border border-gray-300 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>

            {/* Feature Toggles */}
            <div className="space-y-2">
              <p className="text-sm font-medium text-gray-700">Features</p>
              <div className="space-y-2">
                {[
                  { key: "enableStreaming", label: "Enable Streaming" },
                  { key: "enableVision", label: "Enable Vision" },
                  {
                    key: "enableFunctionCalling",
                    label: "Enable Function Calling",
                  },
                  { key: "enableCaching", label: "Enable Caching" },
                ].map(({ key, label }) => (
                  <label
                    key={key}
                    className="flex items-center gap-2 cursor-pointer"
                  >
                    <input
                      type="checkbox"
                      checked={
                        (profile.features as Record<string, boolean>)?.[key] ??
                        false
                      }
                      onChange={(e) => {
                        if (e.target.checked) {
                          updateMaxTokens(
                            profile.limits?.maxTokensPerRequest || 4000
                          );
                        }
                      }}
                      className="w-4 h-4 rounded border-gray-300"
                    />
                    <span className="text-sm text-gray-700">{label}</span>
                  </label>
                ))}
              </div>
            </div>
          </div>
        )}
      </div>

      {/* Action Buttons */}
      <div className="flex gap-2">
        <button
          onClick={() => save()}
          disabled={!isDirty || !isValid || isSaving}
          className={`flex-1 px-4 py-2 rounded-lg font-medium text-sm transition-colors ${
            isDirty && isValid
              ? "bg-blue-600 text-white hover:bg-blue-700 disabled:opacity-50"
              : "bg-gray-300 text-gray-500 cursor-not-allowed"
          }`}
        >
          {isSaving ? "Saving..." : "Save Configuration"}
        </button>

        <button
          onClick={() => reset()}
          className="flex-1 px-4 py-2 bg-gray-200 hover:bg-gray-300 text-gray-700 rounded-lg font-medium text-sm transition-colors"
        >
          Reset
        </button>
      </div>
    </div>
  );
}

export default ProviderConfig;
