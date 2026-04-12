/**
 * TokenUsageDisplay - Show API Token Usage
 *
 * Displays input/output token counts and total usage
 * for the current conversation.
 */

import React, { memo } from "react";
import { Zap } from "lucide-react";

export interface TokenUsage {
  inputTokens: number;
  outputTokens: number;
  totalTokens: number;
}

interface TokenUsageDisplayProps {
  usage?: TokenUsage;
  contextWindowSize?: number;
}

/**
 * Display token usage stats
 */
export const TokenUsageDisplay = memo(
  ({ usage, contextWindowSize = 200_000 }: TokenUsageDisplayProps) => {
    if (!usage || usage.totalTokens === 0) {
      return null;
    }

    const percentUsed = Math.round((usage.totalTokens / contextWindowSize) * 100);
    const warningLevel = percentUsed > 80 ? "warning" : percentUsed > 50 ? "info" : "normal";

    return (
      <div className={cn(
        "px-4 py-2 text-xs border-t border-vscode-widget-border",
        warningLevel === "warning" && "bg-yellow-500 bg-opacity-10",
        warningLevel === "info" && "bg-blue-500 bg-opacity-10"
      )}>
        <div className="flex items-center gap-2 mb-1">
          <Zap size={12} className={cn(
            warningLevel === "warning" && "text-yellow-500",
            warningLevel === "info" && "text-blue-500"
          )} />
          <span className="font-medium text-vscode-editor-foreground">Token Usage</span>
        </div>

        <div className="grid grid-cols-3 gap-3 text-[10px] mb-2">
          <div>
            <div className="text-vscode-descriptionForeground">Input</div>
            <div className="font-mono font-bold">{usage.inputTokens.toLocaleString()}</div>
          </div>
          <div>
            <div className="text-vscode-descriptionForeground">Output</div>
            <div className="font-mono font-bold">{usage.outputTokens.toLocaleString()}</div>
          </div>
          <div>
            <div className="text-vscode-descriptionForeground">Total</div>
            <div className="font-mono font-bold">{usage.totalTokens.toLocaleString()}</div>
          </div>
        </div>

        {/* Progress bar */}
        <div className="w-full h-1 bg-vscode-widget-border rounded-full overflow-hidden">
          <div
            className={cn(
              "h-full transition-all",
              warningLevel === "warning" && "bg-yellow-500",
              warningLevel === "info" && "bg-blue-500",
              warningLevel === "normal" && "bg-green-500"
            )}
            style={{ width: `${Math.min(percentUsed, 100)}%` }}
          />
        </div>

        <div className="text-[9px] text-vscode-descriptionForeground mt-1">
          {percentUsed}% of {(contextWindowSize / 1000).toFixed(0)}K context window used
        </div>
      </div>
    );
  }
);

TokenUsageDisplay.displayName = "TokenUsageDisplay";

// Helper for classnames
function cn(...classes: (string | false | undefined)[]) {
  return classes.filter(Boolean).join(" ");
}
