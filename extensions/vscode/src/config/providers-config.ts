/**
 * Provider Configuration Types - Phase 9
 * Schema e tipos para configuração de providers
 */

import { ProviderType, ProviderCapabilities } from "../providers/baseProvider";

// ============================================================================
// Configuration Types
// ============================================================================

export interface ProviderAuthConfig {
  apiKey: string;
  organizationId?: string;
  projectId?: string;
  customHeaders?: Record<string, string>;
}

export interface ProviderNetworkConfig {
  baseUrl?: string;
  timeout?: number;
  retryAttempts?: number;
  retryDelay?: number;
  proxyUrl?: string;
}

export interface ProviderLimitsConfig {
  maxTokensPerRequest?: number;
  maxRequestsPerMinute?: number;
  maxTokensPerMinute?: number;
  enableCaching?: boolean;
  cacheTTL?: number;
}

export interface ProviderFeatureConfig {
  enableStreaming?: boolean;
  enableVision?: boolean;
  enableFunctionCalling?: boolean;
  enableEmbedding?: boolean;
  preferredModel?: string;
}

export interface ProviderStorageConfig {
  persistConfig?: boolean;
  encryptSensitiveData?: boolean;
  storagePath?: string;
}

export interface ProviderProfile {
  id: string;
  name: string;
  type: ProviderType;
  displayName: string;
  description?: string;
  enabled: boolean;
  priority: number;
  auth: ProviderAuthConfig;
  network?: ProviderNetworkConfig;
  limits?: ProviderLimitsConfig;
  features?: ProviderFeatureConfig;
  storage?: ProviderStorageConfig;
  metadata?: Record<string, unknown>;
}

export interface ProvidersConfig {
  version: string;
  profiles: ProviderProfile[];
  defaultProvider: ProviderType;
  fallbackChain: ProviderType[];
  globalDefaults: {
    timeout?: number;
    retryAttempts?: number;
    enableStreaming?: boolean;
  };
  monitoring?: {
    trackMetrics?: boolean;
    logRequests?: boolean;
    trackCosts?: boolean;
  };
}

// ============================================================================
// Configuration Defaults
// ============================================================================

export const DEFAULT_PROVIDER_CONFIG: Partial<ProviderProfile> = {
  enabled: true,
  priority: 0,
  network: {
    timeout: 30000,
    retryAttempts: 3,
    retryDelay: 1000,
  },
  limits: {
    maxTokensPerRequest: 4000,
    maxRequestsPerMinute: 60,
    maxTokensPerMinute: 90000,
    enableCaching: true,
    cacheTTL: 3600000,
  },
  features: {
    enableStreaming: true,
    enableVision: false,
    enableFunctionCalling: false,
    enableEmbedding: false,
  },
  storage: {
    persistConfig: true,
    encryptSensitiveData: true,
  },
};

export const DEFAULT_PROVIDERS_CONFIG: ProvidersConfig = {
  version: "1.0.0",
  profiles: [],
  defaultProvider: "claude",
  fallbackChain: ["claude", "openai", "gemini", "llama"],
  globalDefaults: {
    timeout: 30000,
    retryAttempts: 3,
    enableStreaming: true,
  },
  monitoring: {
    trackMetrics: true,
    logRequests: false,
    trackCosts: true,
  },
};

// ============================================================================
// Validation Schemas
// ============================================================================

export interface ConfigValidationResult {
  valid: boolean;
  errors: string[];
  warnings: string[];
}

export function validateProviderProfile(
  profile: Partial<ProviderProfile>
): ConfigValidationResult {
  const errors: string[] = [];
  const warnings: string[] = [];

  if (!profile.id) {
    errors.push("Profile must have an id");
  }

  if (!profile.name) {
    errors.push("Profile must have a name");
  }

  if (!profile.type) {
    errors.push("Profile must have a type");
  }

  if (!profile.auth?.apiKey && profile.type !== "llama") {
    errors.push("Profile auth must include an API key");
  }

  if (profile.priority === undefined) {
    warnings.push("Profile priority not specified, using default");
  }

  if (profile.network?.timeout && profile.network.timeout < 1000) {
    warnings.push("Network timeout is very short, may cause issues");
  }

  return {
    valid: errors.length === 0,
    errors,
    warnings,
  };
}

export function validateProvidersConfig(
  config: Partial<ProvidersConfig>
): ConfigValidationResult {
  const errors: string[] = [];
  const warnings: string[] = [];

  if (!config.profiles || config.profiles.length === 0) {
    warnings.push("No provider profiles configured");
  }

  if (config.profiles) {
    for (const profile of config.profiles) {
      const validation = validateProviderProfile(profile);
      if (!validation.valid) {
        errors.push(
          `Profile "${profile.name}" has errors: ${validation.errors.join(", ")}`
        );
      }
    }
  }

  if (!config.defaultProvider) {
    warnings.push("No default provider specified");
  }

  if (!config.fallbackChain || config.fallbackChain.length === 0) {
    warnings.push("No fallback chain configured");
  }

  return {
    valid: errors.length === 0,
    errors,
    warnings,
  };
}

// ============================================================================
// Configuration Builder
// ============================================================================

export class ProviderConfigBuilder {
  private profile: Partial<ProviderProfile> = {};

  setId(id: string): this {
    this.profile.id = id;
    return this;
  }

  setName(name: string): this {
    this.profile.name = name;
    return this;
  }

  setType(type: ProviderType): this {
    this.profile.type = type;
    return this;
  }

  setDisplayName(displayName: string): this {
    this.profile.displayName = displayName;
    return this;
  }

  setApiKey(apiKey: string): this {
    if (!this.profile.auth) {
      this.profile.auth = {} as ProviderAuthConfig;
    }
    this.profile.auth.apiKey = apiKey;
    return this;
  }

  setEnabled(enabled: boolean): this {
    this.profile.enabled = enabled;
    return this;
  }

  setPriority(priority: number): this {
    this.profile.priority = priority;
    return this;
  }

  setNetworkConfig(config: ProviderNetworkConfig): this {
    this.profile.network = config;
    return this;
  }

  setLimitsConfig(config: ProviderLimitsConfig): this {
    this.profile.limits = config;
    return this;
  }

  setFeaturesConfig(config: ProviderFeatureConfig): this {
    this.profile.features = config;
    return this;
  }

  build(): ProviderProfile {
    const validation = validateProviderProfile(this.profile);
    if (!validation.valid) {
      throw new Error(
        `Invalid provider configuration: ${validation.errors.join(", ")}`
      );
    }

    return {
      ...DEFAULT_PROVIDER_CONFIG,
      ...this.profile,
      auth: this.profile.auth || ({} as ProviderAuthConfig),
    } as ProviderProfile;
  }
}

// ============================================================================
// Model Configuration
// ============================================================================

export interface ModelInfo {
  id: string;
  name: string;
  provider: ProviderType;
  contextWindow: number;
  maxTokens: number;
  costPer1kInput: number;
  costPer1kOutput: number;
  releaseDate: string;
  deprecated?: boolean;
}

export const SUPPORTED_MODELS: ModelInfo[] = [
  // Claude
  {
    id: "claude-3-5-sonnet-20241022",
    name: "Claude 3.5 Sonnet",
    provider: "claude",
    contextWindow: 200000,
    maxTokens: 4096,
    costPer1kInput: 0.003,
    costPer1kOutput: 0.015,
    releaseDate: "2024-06-20",
  },
  {
    id: "claude-3-5-haiku-20241022",
    name: "Claude 3.5 Haiku",
    provider: "claude",
    contextWindow: 200000,
    maxTokens: 1024,
    costPer1kInput: 0.0008,
    costPer1kOutput: 0.004,
    releaseDate: "2024-11-15",
  },
  {
    id: "claude-3-opus-20250219",
    name: "Claude 3 Opus",
    provider: "claude",
    contextWindow: 200000,
    maxTokens: 4096,
    costPer1kInput: 0.015,
    costPer1kOutput: 0.075,
    releaseDate: "2024-03-04",
  },

  // OpenAI
  {
    id: "gpt-4o",
    name: "GPT-4o",
    provider: "openai",
    contextWindow: 128000,
    maxTokens: 4096,
    costPer1kInput: 0.03,
    costPer1kOutput: 0.06,
    releaseDate: "2024-05-13",
  },
  {
    id: "gpt-4-turbo",
    name: "GPT-4 Turbo",
    provider: "openai",
    contextWindow: 128000,
    maxTokens: 4096,
    costPer1kInput: 0.01,
    costPer1kOutput: 0.03,
    releaseDate: "2023-11-06",
  },
  {
    id: "gpt-3.5-turbo",
    name: "GPT-3.5 Turbo",
    provider: "openai",
    contextWindow: 16384,
    maxTokens: 4096,
    costPer1kInput: 0.0005,
    costPer1kOutput: 0.0015,
    releaseDate: "2023-03-15",
  },

  // Gemini
  {
    id: "gemini-1.5-pro-latest",
    name: "Gemini 1.5 Pro",
    provider: "gemini",
    contextWindow: 1000000,
    maxTokens: 8000,
    costPer1kInput: 0.0075,
    costPer1kOutput: 0.03,
    releaseDate: "2024-06-12",
  },
  {
    id: "gemini-1.5-flash",
    name: "Gemini 1.5 Flash",
    provider: "gemini",
    contextWindow: 1000000,
    maxTokens: 4000,
    costPer1kInput: 0.00075,
    costPer1kOutput: 0.003,
    releaseDate: "2024-07-01",
  },

  // Llama (approximate pricing)
  {
    id: "llama2",
    name: "Llama 2",
    provider: "llama",
    contextWindow: 4096,
    maxTokens: 2048,
    costPer1kInput: 0.0,
    costPer1kOutput: 0.0,
    releaseDate: "2023-07-18",
  },
  {
    id: "mistral",
    name: "Mistral",
    provider: "llama",
    contextWindow: 32768,
    maxTokens: 4096,
    costPer1kInput: 0.00014,
    costPer1kOutput: 0.00042,
    releaseDate: "2023-12-11",
  },
];
