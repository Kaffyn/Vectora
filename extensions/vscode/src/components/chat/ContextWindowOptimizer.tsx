/**
 * ContextWindowOptimizer - Token Budget & Context Management
 *
 * Helps manage context window size and suggests optimizations
 * when approaching token limits.
 */

import React, { memo, useMemo } from "react";
import { AlertCircle, TrendingDown } from "lucide-react";

export interface ContextWindowState {
  inputTokens: number;
  outputTokens: number;
  maxTokens: number;
  estimatedRemaining: number;
}

interface ContextWindowOptimizerProps {
  state: ContextWindowState;
  onSuggestOptimization?: (suggestion: string) => void;
}

/**
 * Analyze context window usage and suggest optimizations
 */
export const ContextWindowOptimizer = memo(
  ({ state, onSuggestOptimization }: ContextWindowOptimizerProps) => {
    const analysis = useMemo(() => {
      const usedTokens = state.inputTokens + state.outputTokens;
      const percentUsed = (usedTokens / state.maxTokens) * 100;
      const tokensPerMessage = state.inputTokens > 0 ? state.outputTokens / state.inputTokens : 0;

      const suggestions: string[] = [];

      // Generate suggestions based on usage
      if (percentUsed > 90) {
        suggestions.push("Context window nearly full - consider summarizing old messages");
      } else if (percentUsed > 75) {
        suggestions.push("Context window is getting full - optimize conversation");
      }

      if (tokensPerMessage > 2 && state.outputTokens > 500) {
        suggestions.push("Responses are very long - try being more specific in requests");
      }

      if (state.estimatedRemaining < 1000) {
        suggestions.push("Few tokens remaining - consider starting a new conversation");
      }

      return {
        usedTokens,
        percentUsed,
        tokensPerMessage,
        suggestions,
        warningLevel:
          percentUsed > 90 ? "critical" : percentUsed > 75 ? "warning" : "normal",
      };
    }, [state]);

    const getWarningColor = () => {
      switch (analysis.warningLevel) {
        case "critical":
          return "border-red-500 bg-red-500 bg-opacity-5";
        case "warning":
          return "border-yellow-500 bg-yellow-500 bg-opacity-5";
        default:
          return "border-vscode-widget-border bg-vscode-list-hoverBackground";
      }
    };

    const getProgressColor = () => {
      switch (analysis.warningLevel) {
        case "critical":
          return "bg-red-500";
        case "warning":
          return "bg-yellow-500";
        default:
          return "bg-green-500";
      }
    };

    return (
      <div className={`border rounded-lg p-3 ${getWarningColor()} transition-all`}>
        <div className="space-y-2">
          {/* Header */}
          <div className="flex items-center justify-between">
            <span className="text-xs font-bold text-vscode-editor-foreground">
              Context Window Status
            </span>
            <span className="text-xs font-mono">
              {analysis.percentUsed.toFixed(0)}%
            </span>
          </div>

          {/* Progress bar */}
          <div className="h-2 bg-vscode-widget-border rounded-full overflow-hidden">
            <div
              className={`h-full transition-all ${getProgressColor()}`}
              style={{ width: `${Math.min(analysis.percentUsed, 100)}%` }}
            />
          </div>

          {/* Token breakdown */}
          <div className="grid grid-cols-3 gap-2 text-[10px]">
            <div>
              <div className="text-vscode-descriptionForeground">Input</div>
              <div className="font-mono font-bold">
                {state.inputTokens.toLocaleString()}
              </div>
            </div>
            <div>
              <div className="text-vscode-descriptionForeground">Output</div>
              <div className="font-mono font-bold">
                {state.outputTokens.toLocaleString()}
              </div>
            </div>
            <div>
              <div className="text-vscode-descriptionForeground">Remaining</div>
              <div className="font-mono font-bold">
                {state.estimatedRemaining.toLocaleString()}
              </div>
            </div>
          </div>

          {/* Suggestions */}
          {analysis.suggestions.length > 0 && (
            <div className="mt-2 space-y-1 pt-2 border-t border-current border-opacity-20">
              {analysis.suggestions.map((suggestion, idx) => (
                <div key={idx} className="flex items-start gap-2 text-[9px]">
                  {analysis.warningLevel === "critical" ? (
                    <AlertCircle size={10} className="mt-0.5 flex-shrink-0 text-red-500" />
                  ) : (
                    <TrendingDown size={10} className="mt-0.5 flex-shrink-0 text-yellow-500" />
                  )}
                  <span className="text-vscode-descriptionForeground">{suggestion}</span>
                </div>
              ))}
            </div>
          )}

          {/* Efficiency metric */}
          <div className="text-[9px] text-vscode-descriptionForeground pt-1 border-t border-current border-opacity-20">
            <span>Output/Input ratio: {analysis.tokensPerMessage.toFixed(2)}</span>
          </div>
        </div>
      </div>
    );
  }
);

ContextWindowOptimizer.displayName = "ContextWindowOptimizer";
