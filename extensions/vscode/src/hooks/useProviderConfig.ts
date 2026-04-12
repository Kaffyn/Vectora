/**
 * useProviderConfig Hook - Phase 9
 * Gerenciamento de configuração de providers
 */

import { useEffect, useCallback, useState } from "react";
import {
  ProviderProfile,
  ProviderConfig,
  ProviderConfigBuilder,
  validateProviderProfile,
  DEFAULT_PROVIDER_CONFIG,
} from "../config/providers-config";
import { getProviderMetadata } from "../config/provider-metadata";
import { ProviderType } from "../providers/baseProvider";

// ============================================================================
// Types
// ============================================================================

export interface ProviderConfigState {
  profile: ProviderProfile | null;
  config: ProviderConfig | null;
  isValid: boolean;
  validationErrors: string[];
  validationWarnings: string[];
  isDirty: boolean;
  isSaving: boolean;
}

export interface ProviderConfigActions {
  updateApiKey: (apiKey: string) => void;
  updateBaseUrl: (baseUrl: string) => void;
  updateTimeout: (timeout: number) => void;
  updateRetryAttempts: (attempts: number) => void;
  updateMaxTokens: (tokens: number) => void;
  enableFeature: (feature: string) => void;
  disableFeature: (feature: string) => void;
  validate: () => boolean;
  save: () => Promise<void>;
  reset: () => void;
  loadProfile: (profile: ProviderProfile) => void;
}

// ============================================================================
// Hook Implementation
// ============================================================================

export function useProviderConfig(
  providerType: ProviderType,
  onSave?: (config: ProviderConfig, profile: ProviderProfile) => Promise<void>
): ProviderConfigState & ProviderConfigActions {
  const [state, setState] = useState<ProviderConfigState>({
    profile: null,
    config: null,
    isValid: false,
    validationErrors: [],
    validationWarnings: [],
    isDirty: false,
    isSaving: false,
  });

  // ============================================================================
  // Initialize with provider metadata
  // ============================================================================

  useEffect(() => {
    try {
      const metadata = getProviderMetadata(providerType);

      const initialProfile: ProviderProfile = {
        id: `${providerType}-default`,
        name: metadata.displayName,
        type: providerType,
        displayName: metadata.displayName,
        description: metadata.description,
        ...DEFAULT_PROVIDER_CONFIG,
        auth: {
          apiKey: "",
        },
      } as ProviderProfile;

      setState((prev) => ({
        ...prev,
        profile: initialProfile,
        config: {
          name: initialProfile.name,
          apiKey: initialProfile.auth.apiKey,
          baseUrl: initialProfile.network?.baseUrl,
          timeout: initialProfile.network?.timeout,
          retryAttempts: initialProfile.network?.retryAttempts,
        },
      }));
    } catch (error) {
      console.error("Failed to load provider metadata", error);
    }
  }, [providerType]);

  // ============================================================================
  // Validate configuration
  // ============================================================================

  const validate = useCallback((): boolean => {
    if (!state.profile) {
      setState((prev) => ({
        ...prev,
        isValid: false,
        validationErrors: ["Profile not loaded"],
      }));
      return false;
    }

    const validation = validateProviderProfile(state.profile);

    setState((prev) => ({
      ...prev,
      isValid: validation.valid,
      validationErrors: validation.errors,
      validationWarnings: validation.warnings,
    }));

    return validation.valid;
  }, [state.profile]);

  // ============================================================================
  // Update API Key
  // ============================================================================

  const updateApiKey = useCallback((apiKey: string) => {
    setState((prev) => {
      if (!prev.profile) return prev;

      const updatedProfile = { ...prev.profile };
      updatedProfile.auth = { ...updatedProfile.auth, apiKey };

      return {
        ...prev,
        profile: updatedProfile,
        config: prev.config
          ? { ...prev.config, apiKey }
          : { name: updatedProfile.name, apiKey },
        isDirty: true,
      };
    });
  }, []);

  // ============================================================================
  // Update Base URL
  // ============================================================================

  const updateBaseUrl = useCallback((baseUrl: string) => {
    setState((prev) => {
      if (!prev.profile) return prev;

      const updatedProfile = { ...prev.profile };
      updatedProfile.network = { ...updatedProfile.network, baseUrl };

      return {
        ...prev,
        profile: updatedProfile,
        config: prev.config
          ? { ...prev.config, baseUrl }
          : { name: updatedProfile.name, baseUrl },
        isDirty: true,
      };
    });
  }, []);

  // ============================================================================
  // Update Timeout
  // ============================================================================

  const updateTimeout = useCallback((timeout: number) => {
    setState((prev) => {
      if (!prev.profile) return prev;

      const updatedProfile = { ...prev.profile };
      updatedProfile.network = { ...updatedProfile.network, timeout };

      return {
        ...prev,
        profile: updatedProfile,
        isDirty: true,
      };
    });
  }, []);

  // ============================================================================
  // Update Retry Attempts
  // ============================================================================

  const updateRetryAttempts = useCallback((attempts: number) => {
    setState((prev) => {
      if (!prev.profile) return prev;

      const updatedProfile = { ...prev.profile };
      updatedProfile.network = { ...updatedProfile.network, retryAttempts: attempts };

      return {
        ...prev,
        profile: updatedProfile,
        isDirty: true,
      };
    });
  }, []);

  // ============================================================================
  // Update Max Tokens
  // ============================================================================

  const updateMaxTokens = useCallback((tokens: number) => {
    setState((prev) => {
      if (!prev.profile) return prev;

      const updatedProfile = { ...prev.profile };
      updatedProfile.limits = { ...updatedProfile.limits, maxTokensPerRequest: tokens };

      return {
        ...prev,
        profile: updatedProfile,
        isDirty: true,
      };
    });
  }, []);

  // ============================================================================
  // Enable Feature
  // ============================================================================

  const enableFeature = useCallback((feature: string) => {
    setState((prev) => {
      if (!prev.profile) return prev;

      const updatedProfile = { ...prev.profile };
      updatedProfile.features = updatedProfile.features || {};
      (updatedProfile.features as Record<string, boolean>)[feature] = true;

      return {
        ...prev,
        profile: updatedProfile,
        isDirty: true,
      };
    });
  }, []);

  // ============================================================================
  // Disable Feature
  // ============================================================================

  const disableFeature = useCallback((feature: string) => {
    setState((prev) => {
      if (!prev.profile) return prev;

      const updatedProfile = { ...prev.profile };
      updatedProfile.features = updatedProfile.features || {};
      (updatedProfile.features as Record<string, boolean>)[feature] = false;

      return {
        ...prev,
        profile: updatedProfile,
        isDirty: true,
      };
    });
  }, []);

  // ============================================================================
  // Save Configuration
  // ============================================================================

  const save = useCallback(async () => {
    if (!validate()) {
      throw new Error(
        `Configuration validation failed: ${state.validationErrors.join(", ")}`
      );
    }

    if (!state.profile || !state.config) {
      throw new Error("Profile or config is not loaded");
    }

    setState((prev) => ({
      ...prev,
      isSaving: true,
    }));

    try {
      if (onSave) {
        await onSave(state.config, state.profile);
      }

      setState((prev) => ({
        ...prev,
        isDirty: false,
        isSaving: false,
      }));
    } catch (error) {
      setState((prev) => ({
        ...prev,
        isSaving: false,
      }));
      throw error;
    }
  }, [state.profile, state.config, validate, onSave]);

  // ============================================================================
  // Reset Configuration
  // ============================================================================

  const reset = useCallback(() => {
    setState((prev) => {
      if (!prev.profile) return prev;

      const metadata = getProviderMetadata(providerType);
      const resetProfile: ProviderProfile = {
        id: `${providerType}-default`,
        name: metadata.displayName,
        type: providerType,
        displayName: metadata.displayName,
        ...DEFAULT_PROVIDER_CONFIG,
        auth: {
          apiKey: "",
        },
      } as ProviderProfile;

      return {
        ...prev,
        profile: resetProfile,
        config: {
          name: resetProfile.name,
          apiKey: "",
          baseUrl: resetProfile.network?.baseUrl,
        },
        isDirty: false,
        validationErrors: [],
        validationWarnings: [],
      };
    });
  }, [providerType]);

  // ============================================================================
  // Load Profile
  // ============================================================================

  const loadProfile = useCallback((profile: ProviderProfile) => {
    setState((prev) => ({
      ...prev,
      profile,
      config: {
        name: profile.name,
        apiKey: profile.auth.apiKey,
        baseUrl: profile.network?.baseUrl,
        timeout: profile.network?.timeout,
        retryAttempts: profile.network?.retryAttempts,
      },
      isDirty: false,
    }));
  }, []);

  return {
    ...state,
    updateApiKey,
    updateBaseUrl,
    updateTimeout,
    updateRetryAttempts,
    updateMaxTokens,
    enableFeature,
    disableFeature,
    validate,
    save,
    reset,
    loadProfile,
  };
}
