/**
 * useProviderManager Hook - Phase 9
 * Gerenciamento do provider ativo e switching
 */

import { useEffect, useCallback, useState, useRef } from "react";
import { providerFactory } from "../providers/providerFactory";
import {
  ProviderType,
  ILLMProvider,
  ProviderConfig,
} from "../providers/baseProvider";
import { fallbackStrategy } from "../utils/providers/fallbackStrategy";
import { rateLimiter } from "../utils/providers/rateLimiter";

// ============================================================================
// Types
// ============================================================================

export interface ProviderManagerState {
  activeProvider: ProviderType | null;
  activeInstance: ILLMProvider | null;
  isInitialized: boolean;
  isLoading: boolean;
  error: Error | null;
  availableProviders: ProviderType[];
}

export interface ProviderManagerActions {
  switchProvider: (type: ProviderType, config?: ProviderConfig) => Promise<void>;
  reinitialize: (config: ProviderConfig) => Promise<void>;
  getCurrentProvider: () => ILLMProvider | null;
  getAvailableProviders: () => ProviderType[];
  setFallbackChain: (chain: ProviderType[]) => void;
}

// ============================================================================
// Hook Implementation
// ============================================================================

export function useProviderManager(
  defaultProvider: ProviderType = "claude",
  defaultConfig?: ProviderConfig
): ProviderManagerState & ProviderManagerActions {
  const [state, setState] = useState<ProviderManagerState>({
    activeProvider: null,
    activeInstance: null,
    isInitialized: false,
    isLoading: false,
    error: null,
    availableProviders: [],
  });

  const previousProviderRef = useRef<ProviderType | null>(null);

  // ============================================================================
  // Initialize available providers on mount
  // ============================================================================

  useEffect(() => {
    const providers = providerFactory.getRegisteredProviders().map((p) => p.type);
    setState((prev) => ({
      ...prev,
      availableProviders: providers,
    }));

    // Initialize rate limiters
    for (const provider of providers) {
      rateLimiter.initializeProvider(provider);
    }
  }, []);

  // ============================================================================
  // Switch provider
  // ============================================================================

  const switchProvider = useCallback(
    async (type: ProviderType, config?: ProviderConfig) => {
      try {
        setState((prev) => ({
          ...prev,
          isLoading: true,
          error: null,
        }));

        // Store previous provider for fallback
        previousProviderRef.current = state.activeProvider;

        // Create new provider instance
        let provider = providerFactory.getOrCreateProvider(type, config);

        // Initialize if config provided
        if (config && !provider.isReady()) {
          await provider.initialize(config);
        }

        // Update state
        setState((prev) => ({
          ...prev,
          activeProvider: type,
          activeInstance: provider,
          isInitialized: true,
          isLoading: false,
        }));
      } catch (error) {
        const err = error instanceof Error ? error : new Error(String(error));

        // Try fallback provider
        const fallbackChain = fallbackStrategy.getConfig().chain;
        const nextProvider = fallbackChain.find(
          (p) => p !== state.activeProvider && p !== type
        );

        if (nextProvider) {
          console.warn(
            `Failed to switch to ${type}, trying fallback: ${nextProvider}`,
            err
          );
          try {
            await switchProvider(nextProvider, config);
            return;
          } catch (fallbackError) {
            console.error("Fallback provider also failed", fallbackError);
          }
        }

        setState((prev) => ({
          ...prev,
          isLoading: false,
          error: err,
        }));
      }
    },
    [state.activeProvider]
  );

  // ============================================================================
  // Reinitialize current provider
  // ============================================================================

  const reinitialize = useCallback(
    async (config: ProviderConfig) => {
      if (!state.activeProvider) {
        throw new Error("No active provider to reinitialize");
      }

      try {
        setState((prev) => ({
          ...prev,
          isLoading: true,
          error: null,
        }));

        const provider = providerFactory.getOrCreateProvider(
          state.activeProvider,
          config
        );
        await provider.initialize(config);

        setState((prev) => ({
          ...prev,
          activeInstance: provider,
          isLoading: false,
        }));
      } catch (error) {
        const err = error instanceof Error ? error : new Error(String(error));

        setState((prev) => ({
          ...prev,
          isLoading: false,
          error: err,
        }));

        throw err;
      }
    },
    [state.activeProvider]
  );

  // ============================================================================
  // Get current provider instance
  // ============================================================================

  const getCurrentProvider = useCallback((): ILLMProvider | null => {
    return state.activeInstance;
  }, [state.activeInstance]);

  // ============================================================================
  // Get available providers
  // ============================================================================

  const getAvailableProviders = useCallback((): ProviderType[] => {
    return state.availableProviders;
  }, [state.availableProviders]);

  // ============================================================================
  // Set fallback chain
  // ============================================================================

  const setFallbackChain = useCallback((chain: ProviderType[]) => {
    fallbackStrategy.setChain(chain);
  }, []);

  // ============================================================================
  // Initialize default provider on mount
  // ============================================================================

  useEffect(() => {
    if (!state.isInitialized && state.availableProviders.length > 0) {
      switchProvider(defaultProvider, defaultConfig).catch((error) => {
        console.error("Failed to initialize default provider", error);
      });
    }
  }, [state.isInitialized, state.availableProviders]);

  return {
    ...state,
    switchProvider,
    reinitialize,
    getCurrentProvider,
    getAvailableProviders,
    setFallbackChain,
  };
}
