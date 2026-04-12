/**
 * FilePreviewPopup - File Content Preview on Hover
 *
 * Shows file snippet preview when hovering over RAG search results.
 */

import React, { memo, useState, useRef, useEffect } from "react";
import { ChevronDown } from "lucide-react";
import CodeBlock from "../common/CodeBlock";

interface FilePreviewPopupProps {
  filename: string;
  content: string;
  language?: string;
  maxLines?: number;
  isVisible: boolean;
}

/**
 * Popup preview of file content
 */
export const FilePreviewPopup = memo(
  ({
    filename,
    content,
    language,
    maxLines = 10,
    isVisible,
  }: FilePreviewPopupProps) => {
    const [expandedLines, setExpandedLines] = useState(maxLines);
    const popupRef = useRef<HTMLDivElement>(null);

    if (!isVisible) {
      return null;
    }

    const lines = content.split("\n");
    const isLongContent = lines.length > maxLines;
    const displayContent = lines.slice(0, expandedLines).join("\n");
    const remainingLines = Math.max(0, lines.length - expandedLines);

    return (
      <div
        ref={popupRef}
        className="absolute z-50 top-full mt-2 bg-vscode-editor-background border border-vscode-widget-border rounded shadow-xl max-w-md"
      >
        {/* Header */}
        <div className="px-3 py-2 border-b border-vscode-widget-border flex items-center justify-between">
          <span className="text-xs font-mono text-vscode-descriptionForeground truncate">
            {filename}
          </span>
          <span className="text-[10px] text-vscode-descriptionForeground opacity-60">
            {lines.length} lines
          </span>
        </div>

        {/* Content */}
        <div className="max-h-48 overflow-y-auto">
          <CodeBlock code={displayContent} language={language} />
        </div>

        {/* Show more button */}
        {isLongContent && remainingLines > 0 && (
          <button
            onClick={() => setExpandedLines((prev) => prev + maxLines)}
            className="w-full px-3 py-2 text-[10px] text-vscode-descriptionForeground hover:bg-vscode-list-hoverBackground border-t border-vscode-widget-border flex items-center justify-center gap-1 transition-colors"
          >
            <ChevronDown size={10} />
            <span>Show {Math.min(remainingLines, maxLines)} more lines</span>
          </button>
        )}
      </div>
    );
  }
);

FilePreviewPopup.displayName = "FilePreviewPopup";
