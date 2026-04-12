/**
 * RelevanceAnalysis - Vector Similarity & Distance Metrics
 *
 * Shows detailed relevance analysis with vector distance and similarity scores.
 */

import React, { memo, useState } from "react";
import { ChevronDown, Eye, EyeOff } from "lucide-react";

export interface RelevanceMetrics {
  similarityScore: number; // 0-1 cosine similarity
  distance: number; // 0-1 euclidean or l2 distance
  tokenDistance?: number;
  semanticSimilarity?: number; // Additional semantic metric
}

interface RelevanceAnalysisProps {
  metrics: RelevanceMetrics;
  showDetailed?: boolean;
  onToggleDetailed?: (show: boolean) => void;
}

/**
 * Display detailed relevance analysis
 */
export const RelevanceAnalysis = memo(
  ({ metrics, showDetailed = false, onToggleDetailed }: RelevanceAnalysisProps) => {
    const [isExpanded, setIsExpanded] = useState(showDetailed);

    const getSimilarityColor = (score: number) => {
      if (score > 0.8) return "text-green-600";
      if (score > 0.6) return "text-blue-600";
      if (score > 0.4) return "text-yellow-600";
      return "text-red-600";
    };

    const getSimilarityLabel = (score: number) => {
      if (score > 0.8) return "Highly Relevant";
      if (score > 0.6) return "Relevant";
      if (score > 0.4) return "Somewhat Relevant";
      return "Weakly Relevant";
    };

    const getDistanceLabel = (distance: number) => {
      if (distance < 0.2) return "Very Close";
      if (distance < 0.4) return "Close";
      if (distance < 0.6) return "Moderate";
      return "Far";
    };

    return (
      <div className="border border-vscode-widget-border rounded p-2 bg-vscode-list-hoverBackground">
        {/* Summary */}
        <button
          onClick={() => {
            setIsExpanded(!isExpanded);
            onToggleDetailed?.(!isExpanded);
          }}
          className="w-full flex items-center justify-between px-2 py-1 hover:bg-vscode-list-activeSelectionBackground rounded transition-colors"
        >
          <div className="flex items-center gap-2">
            {isExpanded ? <EyeOff size={12} /> : <Eye size={12} />}
            <span className="text-[10px] font-medium text-vscode-editor-foreground">
              Relevance Details
            </span>
          </div>
          <ChevronDown
            size={12}
            className={`transform transition-transform ${isExpanded ? "rotate-180" : ""}`}
          />
        </button>

        {/* Quick summary */}
        <div className="mt-1 text-[10px] text-vscode-descriptionForeground">
          <div className="flex items-center justify-between">
            <span className={`font-bold ${getSimilarityColor(metrics.similarityScore)}`}>
              {getSimilarityLabel(metrics.similarityScore)}
            </span>
            <span className="opacity-60">
              {Math.round(metrics.similarityScore * 100)}%
            </span>
          </div>

          {/* Progress bar */}
          <div className="mt-1 h-1 bg-vscode-widget-border rounded-full overflow-hidden">
            <div
              className={`h-full transition-all ${getSimilarityColor(metrics.similarityScore).replace("text-", "bg-")}`}
              style={{ width: `${metrics.similarityScore * 100}%` }}
            />
          </div>
        </div>

        {/* Detailed metrics */}
        {isExpanded && (
          <div className="mt-2 space-y-1 pt-2 border-t border-vscode-widget-border text-[9px]">
            <div className="flex items-center justify-between">
              <span className="text-vscode-descriptionForeground">Cosine Similarity:</span>
              <span className="font-mono font-bold">
                {metrics.similarityScore.toFixed(4)}
              </span>
            </div>

            <div className="flex items-center justify-between">
              <span className="text-vscode-descriptionForeground">Vector Distance:</span>
              <span className="font-mono">{getDistanceLabel(metrics.distance)}</span>
            </div>

            <div className="flex items-center justify-between">
              <span className="text-vscode-descriptionForeground">L2 Distance:</span>
              <span className="font-mono font-bold">{metrics.distance.toFixed(4)}</span>
            </div>

            {metrics.tokenDistance !== undefined && (
              <div className="flex items-center justify-between">
                <span className="text-vscode-descriptionForeground">Token Distance:</span>
                <span className="font-mono">{metrics.tokenDistance}</span>
              </div>
            )}

            {metrics.semanticSimilarity !== undefined && (
              <div className="flex items-center justify-between">
                <span className="text-vscode-descriptionForeground">Semantic Match:</span>
                <span className="font-mono">
                  {Math.round(metrics.semanticSimilarity * 100)}%
                </span>
              </div>
            )}
          </div>
        )}
      </div>
    );
  }
);

RelevanceAnalysis.displayName = "RelevanceAnalysis";
