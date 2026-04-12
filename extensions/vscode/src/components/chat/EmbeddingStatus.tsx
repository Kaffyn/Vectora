/**
 * EmbeddingStatus - Display RAG Indexing Progress
 *
 * Shows the status of workspace embedding/indexing process,
 * including progress percentage and vector count.
 */

import React, { memo } from "react";
import { Database, Zap } from "lucide-react";
import { cn } from "@/lib/utils";

export interface EmbeddingStatusProps {
  isIndexing: boolean;
  progress: number; // 0-100
  vectorCount: number;
  onToggleIndexing?: (enable: boolean) => void;
}

/**
 * Displays RAG/embedding indexing status
 */
export const EmbeddingStatus = memo(
  ({ isIndexing, progress, vectorCount, onToggleIndexing }: EmbeddingStatusProps) => {
    if (!isIndexing && vectorCount === 0) {
      return (
        <div className="px-4 py-2 text-xs border-t border-vscode-widget-border hover:bg-vscode-list-hoverBackground cursor-pointer transition-colors"
             onClick={() => onToggleIndexing?.(true)}>
          <div className="flex items-center gap-2 text-vscode-descriptionForeground">
            <Database size={12} />
            <span>Enable RAG (Workspace indexing not started)</span>
          </div>
        </div>
      );
    }

    if (isIndexing) {
      return (
        <div className="px-4 py-2 border-t border-vscode-widget-border">
          <div className="flex items-center gap-2 mb-1">
            <Zap size={12} className="text-yellow-500 animate-pulse" />
            <span className="text-xs font-medium text-vscode-editor-foreground">
              Indexing Workspace...
            </span>
          </div>

          {/* Progress bar */}
          <div className="w-full h-1.5 bg-vscode-widget-border rounded-full overflow-hidden mb-1">
            <div
              className="h-full bg-yellow-500 transition-all duration-300"
              style={{ width: `${progress}%` }}
            />
          </div>

          <div className="flex items-center justify-between">
            <span className="text-[10px] text-vscode-descriptionForeground">
              {progress}% complete • {vectorCount.toLocaleString()} vectors
            </span>
            <button
              onClick={() => onToggleIndexing?.(false)}
              className="text-[10px] text-vscode-descriptionForeground hover:text-vscode-editor-foreground px-1 py-0.5"
            >
              Stop
            </button>
          </div>
        </div>
      );
    }

    // Indexing complete
    return (
      <div className="px-4 py-2 text-xs border-t border-vscode-widget-border bg-vscode-list-hoverBackground">
        <div className="flex items-center gap-2">
          <Database size={12} className="text-green-500" />
          <span className="text-vscode-editor-foreground font-medium">✓ RAG Ready</span>
          <span className="text-vscode-descriptionForeground ml-auto">
            {vectorCount.toLocaleString()} vectors indexed
          </span>
        </div>
      </div>
    );
  }
);

EmbeddingStatus.displayName = "EmbeddingStatus";
