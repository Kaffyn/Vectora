/**
 * Model utility functions
 * TODO: Implement model-related utilities
 */

export type ServiceTier = "default" | "pro" | string;

export interface LongContextPricing {
	thresholdTokens: number;
	appliesToServiceTiers?: string[];
	inputPriceMultiplier?: number;
	outputPriceMultiplier?: number;
	cacheWritesPriceMultiplier?: number;
	cacheReadsPriceMultiplier?: number;
}

export interface ModelInfo {
	contextWindow: number;
	maxTokens?: number;
	supportsReasoningBudget?: boolean;
	supportsReasoningEffort?: boolean;
	inputPrice?: number;
	outputPrice?: number;
	cacheWritesPrice?: number;
	cacheReadsPrice?: number;
	longContextPricing?: LongContextPricing;
}

export function estimateTokens(text: string): number {
  // Rough estimate: 1 token ≈ 4 characters
  return Math.ceil(text.length / 4);
}

export function calculateTokensRemaining(used: number, contextWindow: number): number {
  return Math.max(0, contextWindow - used);
}

export interface TokenDistribution {
  input: number;
  output: number;
  reserved: number;
  remaining: number;
}

export function calculateTokenDistribution(
  contextWindow: number,
  usedInput: number,
  maxOutput: number
): TokenDistribution {
  const input = Math.max(0, usedInput);
  const output = Math.max(0, maxOutput);
  const reserved = output;
  const remaining = Math.max(0, contextWindow - input - reserved);

  return {
    input,
    output,
    reserved,
    remaining,
  };
}
