import React from "react";
import { Check, Loader2 } from "lucide-react";

interface ThinkingIndicatorProps {
  isThinking: boolean;
  contextCount: number;
}

export function ThinkingIndicator({
  isThinking,
  contextCount,
}: ThinkingIndicatorProps) {
  return (
    <div className="flex items-center gap-2 mb-4 px-3 py-1.5 rounded-full bg-white/5 border border-white/10 w-fit">
      {isThinking ? (
        <Loader2 className="w-3.5 h-3.5 text-orange-400 animate-spin" />
      ) : (
        <Check className="w-3.5 h-3.5 text-green-400" />
      )}
      <span className="text-xs font-medium text-white/70">
        {isThinking
          ? "Recuperando contextos..."
          : `${contextCount} contextos recuperados`}
      </span>
    </div>
  );
}
