/**
 * StreamingMessageDisplay - Real-time Message Streaming
 *
 * Displays AI response with token-by-token streaming animation.
 */

import React, { memo, useEffect, useRef } from "react";
import { Zap } from "lucide-react";
import MarkdownBlock from "../common/MarkdownBlock";

interface StreamingMessageDisplayProps {
  content: string;
  isStreaming: boolean;
  thinkingContent?: string;
  showThinking?: boolean;
  onToggleThinking?: (show: boolean) => void;
}

/**
 * Display streaming message with smooth animation
 */
export const StreamingMessageDisplay = memo(
  ({
    content,
    isStreaming,
    thinkingContent,
    showThinking = false,
    onToggleThinking,
  }: StreamingMessageDisplayProps) => {
    const contentRef = useRef<HTMLDivElement>(null);

    /**
     * Auto-scroll to bottom when content changes
     */
    useEffect(() => {
      if (contentRef.current && isStreaming) {
        contentRef.current.scrollIntoView({ behavior: "smooth", block: "end" });
      }
    }, [content, isStreaming]);

    return (
      <div className="space-y-2">
        {/* Thinking content (collapsible) */}
        {thinkingContent && (
          <div className="border border-amber-500 border-opacity-20 bg-amber-500 bg-opacity-5 rounded p-3">
            <button
              onClick={() => onToggleThinking?.(!showThinking)}
              className="flex items-center gap-2 text-xs font-medium text-amber-700 hover:text-amber-900 transition-colors w-full"
            >
              <Zap size={12} />
              <span>Thinking Process {showThinking ? "▼" : "▶"}</span>
            </button>

            {showThinking && (
              <div className="mt-2 text-xs text-amber-700 whitespace-pre-wrap max-h-48 overflow-y-auto font-mono">
                {thinkingContent}
              </div>
            )}
          </div>
        )}

        {/* Main content */}
        <div
          ref={contentRef}
          className={`${isStreaming ? "animate-pulse" : ""}`}
        >
          <MarkdownBlock content={content || (isStreaming ? "▌" : "")} />
        </div>

        {/* Streaming indicator */}
        {isStreaming && (
          <div className="flex items-center gap-2 text-[10px] text-vscode-descriptionForeground">
            <span className="inline-block w-1.5 h-1.5 bg-green-500 rounded-full animate-pulse" />
            <span>Receiving response...</span>
          </div>
        )}
      </div>
    );
  }
);

StreamingMessageDisplay.displayName = "StreamingMessageDisplay";
