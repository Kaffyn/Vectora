# Provider Integration Guide - Phase 9

## Adding a New Provider

This guide explains how to add support for a new LLM provider to Vectora.

## Step 1: Create Provider Class

Create a new file `src/providers/newProvider.ts`:

```typescript
import {
  BaseProvider,
  ProviderCapabilities,
  ProviderConfig,
  PromptRequest,
  PromptResponse,
  ValidationResult,
  TokenCountResult,
  RateLimitInfo,
} from "./baseProvider"

export class NewProvider extends BaseProvider {
  private apiKey: string = ""
  private baseUrl: string = "https://api.example.com"

  // Define capabilities
  private static CAPABILITIES: ProviderCapabilities = {
    streaming: true,
    vision: false,
    functionCalling: false,
    embedding: false,
    caching: false,
    maxTokens: 2048,
    contextWindow: 4096,
    costPer1kInputTokens: 0.001,
    costPer1kOutputTokens: 0.002,
  }

  constructor(config?: ProviderConfig) {
    super("NewProvider", "newprovider", NewProvider.CAPABILITIES, config || {})
    // Initialize from config
  }

  async initialize(config: ProviderConfig): Promise<void> {
    // Validate and store configuration
    const validation = this.validateConfig()
    if (!validation.valid) {
      throw new Error(`Initialization failed: ${validation.errors.join(", ")}`)
    }
    this.isInitialized = true
  }

  validateConfig(): ValidationResult {
    const errors: string[] = []
    const warnings: string[] = []

    if (!this.apiKey) {
      errors.push("API key is required")
    }

    return { valid: errors.length === 0, errors, warnings }
  }

  async sendPrompt(request: PromptRequest): Promise<PromptResponse> {
    // Implement prompt sending
    const response = await fetch(`${this.baseUrl}/chat`, {
      method: "POST",
      headers: {
        "Authorization": `Bearer ${this.apiKey}`,
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        messages: request.messages,
        max_tokens: request.maxTokens,
        temperature: request.temperature,
      }),
    })

    const data = await response.json()

    return {
      content: data.choices[0].message.content,
      usage: {
        inputTokens: data.usage.prompt_tokens,
        outputTokens: data.usage.completion_tokens,
        totalTokens: data.usage.total_tokens,
      },
    }
  }

  async *streamPrompt(
    request: PromptRequest
  ): AsyncGenerator<string, void, unknown> {
    // Implement streaming
    const response = await fetch(`${this.baseUrl}/chat/stream`, {
      method: "POST",
      headers: {
        "Authorization": `Bearer ${this.apiKey}`,
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        messages: request.messages,
        stream: true,
      }),
    })

    const reader = response.body?.getReader()
    const decoder = new TextDecoder()

    while (reader) {
      const { done, value } = await reader.read()
      if (done) break

      const text = decoder.decode(value)
      yield text
    }
  }

  getTokenCount(text: string): TokenCountResult {
    // Approximate token counting
    const tokens = Math.ceil(text.length / 4)
    return { tokens, characters: text.length }
  }

  getRateLimitInfo(): RateLimitInfo | null {
    // Return rate limit info if available
    return null
  }
}
```

## Step 2: Register Provider in Factory

Update `src/providers/providerFactory.ts`:

```typescript
import { NewProvider } from "./newProvider"

// In ProviderFactory.initializeProviders():

this.registerProvider({
  name: "NewProvider",
  type: "newprovider" as ProviderType,
  displayName: "New Provider",
  description: "Description of your provider",
  capabilities: {
    streaming: true,
    vision: false,
    functionCalling: false,
    embedding: false,
    caching: false,
    maxTokens: 2048,
    contextWindow: 4096,
    costPer1kInputTokens: 0.001,
    costPer1kOutputTokens: 0.002,
  },
  create: (config) => new NewProvider(config),
})
```

## Step 3: Add Metadata

Update `src/config/provider-metadata.ts`:

```typescript
import { ProviderType } from "../providers/baseProvider"

// Add to PROVIDER_METADATA:

newprovider: {
  type: "newprovider",
  name: "New Provider",
  displayName: "New Provider",
  description: "Your provider description",
  website: "https://example.com",
  documentation: "https://example.com/docs",
  apiKeyFormat: "api_*",
  apiKeyLength: { min: 20, max: 100 },
  supportEmail: "support@example.com",
  status: "stable",
  capabilities: {
    streaming: true,
    vision: false,
    functionCalling: false,
    embedding: false,
    caching: false,
    maxTokens: 2048,
    contextWindow: 4096,
    costPer1kInputTokens: 0.001,
    costPer1kOutputTokens: 0.002,
  },
  regionAvailability: ["US", "EU"],
  pricingModel: "Pay per use",
  freeTrialAvailable: true,
  rateLimit: {
    requestsPerMinute: 100,
    tokensPerMinute: 100000,
  },
  supportedLanguages: ["English"],
  authenticationType: "api_key",
  notes: "Optional notes about the provider",
}
```

## Step 4: Update Types

Add to type definitions if needed:

```typescript
// In baseProvider.ts, update ProviderType:
export type ProviderType =
  | "claude"
  | "openai"
  | "llama"
  | "gemini"
  | "newprovider"
  | "custom"
```

## Step 5: Add Configuration

Update `src/config/providers-config.ts`:

```typescript
// Add model information:
export const SUPPORTED_MODELS: ModelInfo[] = [
  // ... existing models ...
  {
    id: "newprovider-model-v1",
    name: "NewProvider Model v1",
    provider: "newprovider",
    contextWindow: 4096,
    maxTokens: 2048,
    costPer1kInput: 0.001,
    costPer1kOutput: 0.002,
    releaseDate: "2024-01-01",
  },
]
```

## Step 6: Create Tests

Create `src/__tests__/providers/newProvider.test.ts`:

```typescript
import { NewProvider } from "../../providers/newProvider"

describe("NewProvider", () => {
  let provider: NewProvider

  beforeEach(() => {
    provider = new NewProvider({
      apiKey: "test-key",
    })
  })

  it("should initialize with valid config", async () => {
    await provider.initialize({
      apiKey: "test-key",
    })
    expect(provider.isReady()).toBe(true)
  })

  it("should validate config", () => {
    const validation = provider.validateConfig()
    expect(validation.valid).toBe(true)
  })

  it("should count tokens", () => {
    const result = provider.getTokenCount("Hello, world!")
    expect(result.tokens).toBeGreaterThan(0)
  })

  it("should send prompt", async () => {
    // Mock implementation
    const response = await provider.sendPrompt({
      messages: [{ role: "user", content: "Test" }],
    })
    expect(response.content).toBeDefined()
  })
})
```

## Step 7: Update Exports

Update `src/providers/index.ts`:

```typescript
export { NewProvider } from "./newProvider"
```

## Step 8: Add UI Support (Optional)

If you want custom UI, create `src/components/providers/NewProviderConfig.tsx`:

```typescript
import React from "react"
import { APIKeyManager } from "./APIKeyManager"

export function NewProviderConfig({ /* props */ }) {
  return (
    <div>
      <APIKeyManager
        provider="newprovider"
        // ... other props
      />
      {/* Custom fields specific to NewProvider */}
    </div>
  )
}
```

## Implementation Checklist

- [ ] Create provider class extending BaseProvider
- [ ] Implement all abstract methods
- [ ] Register in ProviderFactory
- [ ] Add metadata
- [ ] Update type definitions
- [ ] Add model information
- [ ] Create unit tests
- [ ] Add to exports
- [ ] Test with real API
- [ ] Document API key format
- [ ] Document rate limits
- [ ] Add integration tests

## Testing New Provider

```typescript
import { providerFactory } from "@/providers/providerFactory"

// Test instantiation
const provider = providerFactory.createProvider("newprovider", {
  apiKey: process.env.NEW_PROVIDER_API_KEY,
})

// Test initialization
await provider.initialize({
  apiKey: process.env.NEW_PROVIDER_API_KEY,
})

// Test validation
const validation = provider.validateConfig()
console.log("Valid:", validation.valid)

// Test prompt
const response = await provider.sendPrompt({
  messages: [
    {
      role: "user",
      content: "Hello!",
    },
  ],
})

console.log("Response:", response.content)

// Test streaming
for await (const chunk of provider.streamPrompt({
  messages: [{ role: "user", content: "Hello!" }],
})) {
  console.log(chunk)
}

// Test capabilities
console.log("Capabilities:", provider.getCapabilities())
```

## Common Issues and Solutions

### Issue: API Key Not Working

```typescript
// Debug validation
const validation = provider.validateConfig()
console.log("Validation errors:", validation.errors)
console.log("Validation warnings:", validation.warnings)

// Check API key format
if (!apiKey.match(expectedFormat)) {
  console.error("API key format incorrect")
}
```

### Issue: High Latency

```typescript
// Check rate limits
const rateLimit = provider.getRateLimitInfo()
console.log("Rate limit info:", rateLimit)

// Use rate limiter
import { rateLimiter } from "@/utils/providers/rateLimiter"
const result = await rateLimiter.queueRequest(
  "newprovider",
  () => provider.sendPrompt(request)
)
```

### Issue: Token Counting Inaccurate

```typescript
// Override token counting if needed
getTokenCount(text: string): TokenCountResult {
  // Use provider-specific algorithm
  // Can call API if needed
  const tokens = customCountLogic(text)
  return { tokens, characters: text.length }
}
```

## Provider Best Practices

1. **Error Handling**: Always provide clear error messages
2. **Timeout**: Set reasonable timeout values (30-60 seconds)
3. **Retry**: Implement retry logic with exponential backoff
4. **Validation**: Validate config before initialization
5. **Logging**: Log important events for debugging
6. **Rate Limiting**: Respect provider rate limits
7. **Streaming**: Implement proper stream error handling
8. **Security**: Never log API keys

## Example: Adding Error Handling

```typescript
async sendPrompt(request: PromptRequest): Promise<PromptResponse> {
  try {
    const response = await fetch(/* ... */)

    if (!response.ok) {
      if (response.status === 401) {
        throw new Error("Invalid API key")
      }
      if (response.status === 429) {
        throw new Error("Rate limited - wait before retrying")
      }
      throw new Error(`API error: ${response.statusText}`)
    }

    return { /* ... */ }
  } catch (error) {
    if (error instanceof TypeError) {
      throw new Error(`Network error: ${error.message}`)
    }
    throw error
  }
}
```

## Support

For help adding new providers:
1. Check existing provider implementations
2. Review provider API documentation
3. Test thoroughly before submitting PR
4. Include unit tests and documentation
