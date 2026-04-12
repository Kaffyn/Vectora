/**
 * ToolCallProgress - Display Tool Execution Status
 *
 * Shows real-time progress of tool calls with status indicators.
 */

import React, { memo } from "react";
import { Zap, CheckCircle, AlertCircle, Clock } from "lucide-react";

export type ToolStatus = "pending" | "running" | "completed" | "error";

export interface ToolCallProgressProps {
  toolName: string;
  status: ToolStatus;
  progress?: number; // 0-100 for running status
  error?: string;
  result?: string;
  duration?: number; // milliseconds
}

/**
 * Display tool call execution progress
 */
export const ToolCallProgress = memo(
  ({ toolName, status, progress, error, result, duration }: ToolCallProgressProps) => {
    const getStatusIcon = () => {
      switch (status) {
        case "pending":
          return <Clock size={14} className="text-blue-500 animate-bounce" />;
        case "running":
          return <Zap size={14} className="text-yellow-500 animate-pulse" />;
        case "completed":
          return <CheckCircle size={14} className="text-green-500" />;
        case "error":
          return <AlertCircle size={14} className="text-red-500" />;
      }
    };

    const getStatusText = () => {
      switch (status) {
        case "pending":
          return "Preparing...";
        case "running":
          return progress !== undefined ? `Running... ${progress}%` : "Running...";
        case "completed":
          return "Completed";
        case "error":
          return "Error";
      }
    };

    const getStatusColor = () => {
      switch (status) {
        case "pending":
          return "text-blue-500";
        case "running":
          return "text-yellow-500";
        case "completed":
          return "text-green-500";
        case "error":
          return "text-red-500";
      }
    };

    const formatDuration = (ms: number) => {
      if (ms < 1000) return `${Math.round(ms)}ms`;
      return `${(ms / 1000).toFixed(2)}s`;
    };

    return (
      <div className={`border border-l-4 rounded p-3 mb-2 ${
        status === "error" ? "border-red-500 bg-red-500 bg-opacity-5" :
        status === "completed" ? "border-green-500 bg-green-500 bg-opacity-5" :
        status === "running" ? "border-yellow-500 bg-yellow-500 bg-opacity-5" :
        "border-blue-500 bg-blue-500 bg-opacity-5"
      }`}>
        <div className="flex items-start gap-3">
          {getStatusIcon()}

          <div className="flex-1">
            <div className="flex items-center gap-2">
              <span className="font-mono text-xs font-bold text-vscode-editor-foreground">
                {toolName}
              </span>
              <span className={`text-[10px] font-medium ${getStatusColor()}`}>
                {getStatusText()}
              </span>
              {duration && status === "completed" && (
                <span className="text-[10px] text-vscode-descriptionForeground ml-auto">
                  {formatDuration(duration)}
                </span>
              )}
            </div>

            {/* Progress bar for running status */}
            {status === "running" && progress !== undefined && (
              <div className="mt-2 h-1.5 bg-vscode-widget-border rounded-full overflow-hidden">
                <div
                  className="h-full bg-yellow-500 transition-all duration-300"
                  style={{ width: `${progress}%` }}
                />
              </div>
            )}

            {/* Error message */}
            {status === "error" && error && (
              <div className="mt-2 text-[10px] text-red-600 font-mono">
                {error}
              </div>
            )}

            {/* Result preview */}
            {status === "completed" && result && (
              <div className="mt-2 text-[10px] text-vscode-descriptionForeground max-h-16 overflow-hidden">
                <span className="opacity-60">Result: </span>
                <span className="font-mono">{result.substring(0, 100)}{result.length > 100 ? "..." : ""}</span>
              </div>
            )}
          </div>
        </div>
      </div>
    );
  }
);

ToolCallProgress.displayName = "ToolCallProgress";
