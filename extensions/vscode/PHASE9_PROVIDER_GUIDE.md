# Phase 9: Advanced LLM Provider Support - Complete Guide

## Overview

Phase 9 introduces a comprehensive multi-provider system for Vectora, enabling seamless switching between multiple LLM providers with intelligent fallback handling, rate limiting, and token management.

### Supported Providers

1. **Claude** (Anthropic) - Primary provider
2. **OpenAI** (GPT models)
3. **Llama** (Meta - Local or Groq cloud)
4. **Gemini** (Google)

## Architecture

```
┌─────────────────────────────────────────┐
│      UI Layer (React Components)        │
├─────────────────────────────────────────┤
│  ProviderSelector │ ProviderStatus      │
│  ProviderConfig   │ APIKeyManager       │
├─────────────────────────────────────────┤
│      Hooks Layer (State Management)     │
├─────────────────────────────────────────┤
│  useProviderManager                     │
│  useProviderConfig                      │
│  useLLMCapabilities                     │
├─────────────────────────────────────────┤
│      Provider Abstraction Layer         │
├─────────────────────────────────────────┤
│  BaseProvider (Abstract)                │
│  ├── ClaudeProvider                     │
│  ├── OpenAIProvider                     │
│  ├── LlamaProvider                      │
│  └── GeminiProvider                     │
├─────────────────────────────────────────┤
│      Factory & Management               │
├─────────────────────────────────────────┤
│  ProviderFactory                        │
│  FallbackStrategy                       │
│  RateLimiter                            │
│  TokenCounter                           │
└─────────────────────────────────────────┘
```

## Core Components

### 1. Provider Abstraction (`baseProvider.ts`)

All providers implement a common interface:

```typescript
interface ILLMProvider {
  initialize(config: ProviderConfig): Promise<void>
  validateConfig(): ValidationResult
  sendPrompt(request: PromptRequest): Promise<PromptResponse>
  streamPrompt(request: PromptRequest): AsyncGenerator<string>
  getTokenCount(text: string): TokenCountResult
  getRateLimitInfo(): RateLimitInfo | null
  estimateCost(inputTokens, outputTokens): CostInfo
  hasCapability(capability: string): boolean
}
```

### 2. Provider Implementations

#### Claude Provider
- Models: Claude 3.5 Sonnet, Haiku, Opus
- Features: Streaming, Vision, Function Calling, Caching
- Context: 200K tokens
- Pricing: $0.003/$0.015 per 1K input/output tokens

#### OpenAI Provider
- Models: GPT-4o, GPT-4 Turbo, GPT-3.5
- Features: Streaming, Vision, Function Calling, Embeddings
- Context: 128K tokens (GPT-4)
- Pricing: $0.03/$0.06 per 1K input/output tokens

#### Llama Provider
- Models: Llama 2, Llama 3, Mistral
- Deployment: Local (Ollama) or Cloud (Groq)
- Features: Streaming, Open source
- Cost: Free (local) or minimal (Groq)

#### Gemini Provider
- Models: Gemini 1.5 Pro, Flash
- Features: Streaming, Vision, Embeddings, 1M context
- Context: 1M tokens
- Pricing: $0.0075/$0.03 per 1K input/output tokens

### 3. Factory Pattern (`providerFactory.ts`)

The factory manages provider creation and lifecycle:

```typescript
// Get singleton instance
const factory = ProviderFactory.getInstance()

// Create provider
const provider = factory.createProvider("claude", config)

// Create and initialize
const provider = await factory.createAndInitializeProvider("openai", config)

// Get active instance
const provider = factory.getOrCreateProvider("gemini")
```

### 4. Configuration Management

#### Provider Profile
```typescript
interface ProviderProfile {
  id: string
  name: string
  type: ProviderType
  enabled: boolean
  priority: number
  auth: ProviderAuthConfig
  network?: ProviderNetworkConfig
  limits?: ProviderLimitsConfig
  features?: ProviderFeatureConfig
}
```

#### Configuration Builder
```typescript
const config = new ProviderConfigBuilder()
  .setId("claude-primary")
  .setType("claude")
  .setApiKey(process.env.ANTHROPIC_API_KEY)
  .setNetworkConfig({ timeout: 30000, retryAttempts: 3 })
  .build()
```

## Usage Guide

### Basic Provider Switching

```typescript
import { useProviderManager } from "@/hooks/useProviderManager"

export function MyComponent() {
  const {
    activeProvider,
    switchProvider,
    getAvailableProviders
  } = useProviderManager("claude")

  const handleSwitch = async () => {
    await switchProvider("openai", {
      apiKey: process.env.OPENAI_API_KEY
    })
  }

  return (
    <div>
      <p>Current: {activeProvider}</p>
      <button onClick={handleSwitch}>Switch to OpenAI</button>
    </div>
  )
}
```

### Configuring Providers

```typescript
import { useProviderConfig } from "@/hooks/useProviderConfig"

export function ProviderConfigPanel() {
  const {
    profile,
    updateApiKey,
    validate,
    save
  } = useProviderConfig("claude")

  return (
    <div>
      <input
        value={profile?.auth.apiKey}
        onChange={(e) => updateApiKey(e.target.value)}
        placeholder="Paste API key..."
      />
      <button onClick={save} disabled={!validate()}>
        Save
      </button>
    </div>
  )
}
```

### Using Provider Capabilities

```typescript
import { useLLMCapabilities } from "@/hooks/useLLMCapabilities"
import { useProviderManager } from "@/hooks/useProviderManager"

export function CapabilityDisplay() {
  const { activeInstance } = useProviderManager()
  const capabilities = useLLMCapabilities(activeInstance)

  return (
    <div>
      <p>Streaming: {capabilities.streaming ? "✓" : "✗"}</p>
      <p>Vision: {capabilities.vision ? "✓" : "✗"}</p>
      <p>Context: {capabilities.contextWindow / 1000}K tokens</p>
      <p>Cost: ${capabilities.costPer1kInputTokens} per 1K input</p>
    </div>
  )
}
```

### Fallback Strategy

```typescript
import { fallbackStrategy } from "@/utils/providers/fallbackStrategy"

// Configure fallback chain
fallbackStrategy.setChain(["claude", "openai", "gemini", "llama"])
fallbackStrategy.setMaxRetries(3)

// Execute with automatic fallback
const result = await fallbackStrategy.executeWithFallback(
  async (provider) => {
    const instance = providerFactory.getOrCreateProvider(provider, config)
    return instance.sendPrompt(request)
  }
)

// Check provider health
const health = fallbackStrategy.getHealth("claude")
console.log(`Claude success rate: ${health?.successRate}%`)
```

### Token Management

```typescript
import { TokenCounter } from "@/utils/providers/tokenCounter"

// Count tokens for specific provider
const tokens = TokenCounter.countTokens(text, "claude")

// Estimate cost
const cost = TokenCounter.estimateCost(
  100, // input tokens
  50,  // output tokens
  "openai",
  0.03,  // cost per 1k input
  0.06   // cost per 1k output
)

// Check if text fits in context
const fit = TokenCounter.fits(text, "gemini", 1000000)
console.log(`Text fits: ${fit.fits}, Usage: ${fit.percentage}%`)
```

### Rate Limiting

```typescript
import { rateLimiter } from "@/utils/providers/rateLimiter"

// Initialize provider limits
rateLimiter.initializeProvider("claude", 60, 1000000)

// Check if request can be made
if (!rateLimiter.canMakeRequest("claude")) {
  console.log("Rate limited, waiting...")
}

// Queue request with automatic rate limiting
const result = await rateLimiter.queueRequest(
  "claude",
  () => provider.sendPrompt(request),
  estimatedTokens
)

// Get current status
const status = rateLimiter.getStatus("openai")
console.log(`Requests remaining: ${status.requestsRemaining}`)
```

## Configuration Examples

### Claude Configuration
```typescript
{
  id: "claude-prod",
  name: "Claude Production",
  type: "claude",
  enabled: true,
  priority: 1,
  auth: {
    apiKey: process.env.ANTHROPIC_API_KEY
  },
  network: {
    timeout: 30000,
    retryAttempts: 3
  },
  limits: {
    maxTokensPerRequest: 4000,
    maxRequestsPerMinute: 60,
    enableCaching: true
  }
}
```

### Local Llama (Ollama)
```typescript
{
  id: "llama-local",
  name: "Local Llama",
  type: "llama",
  enabled: true,
  priority: 3,
  auth: { apiKey: "" },
  network: {
    baseUrl: "http://localhost:11434",
    timeout: 60000,
    retryAttempts: 1
  }
}
```

### Groq Cloud
```typescript
{
  id: "groq-cloud",
  name: "Groq API",
  type: "llama",
  enabled: true,
  priority: 2,
  auth: {
    apiKey: process.env.GROQ_API_KEY
  },
  network: {
    baseUrl: "https://api.groq.com/openai/v1",
    timeout: 30000
  }
}
```

## Features

### Dynamic Provider Switching
- Switch providers on-the-fly
- Automatic fallback on failures
- Session continuity maintained

### Intelligent Fallback
- Configurable fallback chain
- Health tracking per provider
- Automatic retry with exponential backoff
- Request history and metrics

### Rate Limiting
- Per-provider limits
- Token-aware rate limiting
- Adaptive backoff strategies
- Automatic queue management

### Token Management
- Provider-specific token counting
- Cost estimation
- Context window monitoring
- Optimization suggestions

### UI Components
- Provider selector with details
- Status dashboard
- Configuration panel
- API key manager with validation

## Best Practices

### 1. Configuration Management
```typescript
// Store sensitive data securely
const config = {
  apiKey: process.env.PROVIDER_API_KEY,
  // Don't hardcode secrets!
}

// Validate before use
const validation = provider.validateConfig()
if (!validation.valid) {
  throw new Error(validation.errors.join(", "))
}
```

### 2. Error Handling
```typescript
try {
  const response = await provider.sendPrompt(request)
} catch (error) {
  // Use fallback strategy
  const result = await fallbackStrategy.executeWithFallback(
    (type) => getProvider(type).sendPrompt(request)
  )
}
```

### 3. Token Management
```typescript
// Always check context window
const fit = TokenCounter.fits(text, provider, contextWindow)
if (!fit.fits) {
  // Summarize or split message
  console.warn(`Message too long: ${fit.percentage}% of context`)
}

// Monitor costs
const cost = TokenCounter.estimateCost(input, output, provider, costPerInput, costPerOutput)
console.log(`Request cost: ${cost.totalCost}`)
```

### 4. Rate Limit Handling
```typescript
// Respect rate limits
const status = rateLimiter.getStatus(provider)
if (!status.canMakeRequest) {
  await new Promise(r => setTimeout(r, status.requestsResetIn))
}

// Use queue for batch requests
for (const request of requests) {
  await rateLimiter.queueRequest(provider, () => send(request))
}
```

## Troubleshooting

### Provider Not Initializing
```typescript
const validation = provider.validateConfig()
if (!validation.valid) {
  console.error("Config errors:", validation.errors)
  console.warn("Config warnings:", validation.warnings)
}
```

### High Latency
```typescript
// Check rate limit status
const status = rateLimiter.getStatus(provider)
console.log("Requests remaining:", status.requestsRemaining)

// Check fallback health
const health = fallbackStrategy.getHealth(provider)
console.log("Success rate:", health?.successRate)
```

### Token Overflow
```typescript
// Monitor usage
const fit = TokenCounter.fits(text, provider, maxContext)
if (!fit.fits) {
  console.error("Message exceeds context", fit.percentage)
}

// Get optimization suggestions
const suggestions = AdvancedTokenCounter.findOptimizations(messages, provider)
suggestions.forEach(s => console.log(s.suggestion))
```

## Migration from Previous Phases

Phase 9 is backward compatible with existing code. To migrate:

1. Replace provider calls with new abstraction
2. Use hooks for state management
3. Update configuration format
4. Test fallback behavior

```typescript
// Old way
const response = await fetch(apiUrl, { /* ... */ })

// New way
const provider = providerFactory.getOrCreateProvider("claude", config)
const response = await provider.sendPrompt(request)
```

## Performance Considerations

- **Caching**: Enable provider caching for repeated prompts
- **Streaming**: Use streaming for real-time responses
- **Batch Processing**: Queue requests to respect rate limits
- **Token Counting**: Cache token counts for common texts
- **Fallback**: Disable fallback for speed-critical operations

## Security

- API keys encrypted at rest
- No keys in logs or error messages
- Secure clipboard operations
- HTTPS only for API calls
- Rate limit protection

## Future Enhancements

- [ ] Provider cost tracking dashboard
- [ ] Automatic provider selection based on requirements
- [ ] Custom provider support
- [ ] Provider performance benchmarking
- [ ] Advanced caching strategies
- [ ] Multi-provider request balancing

## Support & Resources

- **Anthropic Claude**: https://docs.anthropic.com
- **OpenAI GPT**: https://platform.openai.com/docs
- **Meta Llama**: https://www.llama.com
- **Google Gemini**: https://ai.google.dev
