
export interface ParsedApiReqStartedTextType {
    request?: string;
    tokens?: number;
}

export function getApiMetrics() {
    return {
        totalInputTokens: 0,
        totalOutputTokens: 0,
        totalCost: 0
    };
}

export function hasTokenUsageChanged() {
    return false;
}

export function hasToolUsageChanged() {
    return false;
}
