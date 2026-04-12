/**
 * RagSearchResults - Display RAG/Workspace Search Results
 *
 * Shows files and code snippets found in workspace RAG search,
 * with relevance scores and preview snippets.
 */

import React, { memo } from "react";
import { ChevronDown, File, AlertCircle } from "lucide-react";
import { cn } from "@/lib/utils";

export interface SearchResult {
  filename: string;
  content: string;
  relevance: number;
  language?: string;
  snippet?: string;
}

interface RagSearchResultsProps {
  results: SearchResult[];
  query: string;
  isLoading?: boolean;
  error?: string;
  onSelectFile?: (filename: string) => void;
}

/**
 * Displays RAG search results from workspace query
 */
export const RagSearchResults = memo(
  ({ results, query, isLoading, error, onSelectFile }: RagSearchResultsProps) => {
    if (isLoading) {
      return (
        <div className="px-3 py-2 text-xs text-vscode-descriptionForeground flex items-center gap-2">
          <span className="animate-pulse">🔍 Searching workspace...</span>
        </div>
      );
    }

    if (error) {
      return (
        <div className="px-3 py-2 text-xs text-red-400 flex items-center gap-2">
          <AlertCircle size={12} />
          <span>{error}</span>
        </div>
      );
    }

    if (results.length === 0) {
      return (
        <div className="px-3 py-2 text-xs text-vscode-descriptionForeground">
          No relevant files found for "{query}"
        </div>
      );
    }

    return (
      <div className="border-t border-vscode-widget-border">
        <div className="px-3 py-2">
          <div className="flex items-center gap-2 mb-2">
            <span className="text-xs font-bold text-vscode-descriptionForeground">
              📚 RAG Search Results ({results.length})
            </span>
          </div>

          <div className="space-y-1">
            {results.map((result) => (
              <div
                key={result.filename}
                className={cn(
                  "px-2 py-1.5 rounded text-xs hover:bg-vscode-list-hoverBackground cursor-pointer transition-colors",
                  "flex items-start gap-2"
                )}
                onClick={() => onSelectFile?.(result.filename)}
              >
                <File size={12} className="mt-0.5 flex-shrink-0 opacity-60" />

                <div className="flex-1 min-w-0">
                  <div className="font-mono text-vscode-editor-foreground truncate">
                    {result.filename}
                  </div>

                  {result.snippet && (
                    <div className="text-[10px] text-vscode-descriptionForeground mt-0.5 line-clamp-2">
                      {result.snippet}
                    </div>
                  )}

                  <div className="flex items-center gap-2 mt-1">
                    {/* Relevance score as progress bar */}
                    <div className="flex-1 h-1 bg-vscode-widget-border rounded-full overflow-hidden">
                      <div
                        className="h-full bg-vscode-button-background"
                        style={{ width: `${Math.round(result.relevance * 100)}%` }}
                      />
                    </div>
                    <span className="text-[10px] opacity-60 whitespace-nowrap">
                      {Math.round(result.relevance * 100)}%
                    </span>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>
    );
  }
);

RagSearchResults.displayName = "RagSearchResults";
