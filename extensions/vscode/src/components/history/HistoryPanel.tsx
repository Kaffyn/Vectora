/**
 * HistoryPanel - Chat Session History Display
 *
 * Shows list of past conversations with search, delete, and restore options.
 */

import React, { memo, useState, useEffect } from "react";
import { Search, Trash2, RotateCcw, X } from "lucide-react";
import { useSessionHistory } from "@/hooks/useSessionHistory";
import { Button } from "@/components/ui";
import { cn } from "@/lib/utils";

interface HistoryPanelProps {
  onSelectSession?: (sessionId: string) => void;
  onClose?: () => void;
}

/**
 * Displays and manages chat session history
 */
export const HistoryPanel = memo(({ onSelectSession, onClose }: HistoryPanelProps) => {
  const { list, search, remove, clear, stats } = useSessionHistory();
  const [searchQuery, setSearchQuery] = useState("");
  const [sessions, setSessions] = useState<ReturnType<typeof list>>([]);

  /**
   * Load sessions on mount and when search changes
   */
  useEffect(() => {
    if (searchQuery.trim()) {
      // Search mode
      const results = search(searchQuery).map((session) => ({
        id: session.id,
        title: session.title,
        provider: session.provider,
        model: session.model,
        messageCount: session.messages.length,
        lastModified: session.messages[session.messages.length - 1]?.timestamp
          ? new Date(session.messages[session.messages.length - 1].timestamp).getTime()
          : Date.now(),
      }));
      setSessions(results);
    } else {
      // List mode
      setSessions(list());
    }
  }, [searchQuery, list, search]);

  const handleDelete = (sessionId: string, e: React.MouseEvent) => {
    e.stopPropagation();
    if (confirm("Delete this session?")) {
      remove(sessionId);
      setSessions((prev) => prev.filter((s) => s.id !== sessionId));
    }
  };

  const handleClearAll = () => {
    if (confirm("Delete all history? This cannot be undone.")) {
      clear();
      setSessions([]);
    }
  };

  const formatDate = (timestamp: number) => {
    const date = new Date(timestamp);
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffMins = Math.floor(diffMs / 60000);
    const diffHours = Math.floor(diffMs / 3600000);
    const diffDays = Math.floor(diffMs / 86400000);

    if (diffMins < 1) return "Just now";
    if (diffMins < 60) return `${diffMins}m ago`;
    if (diffHours < 24) return `${diffHours}h ago`;
    if (diffDays < 7) return `${diffDays}d ago`;

    return date.toLocaleDateString();
  };

  return (
    <div className="flex flex-col h-full bg-vscode-sideBar-background">
      {/* Header */}
      <div className="px-4 py-3 border-b border-vscode-widget-border flex items-center justify-between">
        <h2 className="text-sm font-bold text-vscode-editor-foreground">Chat History</h2>
        {onClose && (
          <button
            onClick={onClose}
            className="p-1 hover:bg-vscode-list-hoverBackground rounded transition-colors"
          >
            <X size={14} />
          </button>
        )}
      </div>

      {/* Search */}
      <div className="px-4 py-2 border-b border-vscode-widget-border">
        <div className="flex items-center gap-2 px-2 py-1.5 bg-vscode-input-background border border-vscode-widget-border rounded">
          <Search size={12} className="text-vscode-descriptionForeground" />
          <input
            type="text"
            placeholder="Search conversations..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="flex-1 bg-transparent text-xs text-vscode-editor-foreground placeholder-vscode-descriptionForeground outline-none"
          />
          {searchQuery && (
            <button
              onClick={() => setSearchQuery("")}
              className="text-vscode-descriptionForeground hover:text-vscode-editor-foreground"
            >
              <X size={12} />
            </button>
          )}
        </div>
      </div>

      {/* Sessions List */}
      <div className="flex-1 overflow-y-auto">
        {sessions.length === 0 ? (
          <div className="flex flex-col items-center justify-center h-full gap-2 text-center p-4">
            <p className="text-xs text-vscode-descriptionForeground">
              {searchQuery ? "No results found" : "No history yet"}
            </p>
            <p className="text-[10px] text-vscode-descriptionForeground opacity-60">
              Start a conversation to see it here
            </p>
          </div>
        ) : (
          <div className="space-y-1 p-2">
            {sessions.map((session) => (
              <div
                key={session.id}
                onClick={() => onSelectSession?.(session.id)}
                className={cn(
                  "px-3 py-2 rounded text-xs cursor-pointer transition-colors",
                  "hover:bg-vscode-list-hoverBackground group"
                )}
              >
                <div className="flex items-start justify-between gap-2">
                  <div className="flex-1 min-w-0">
                    <div className="font-medium text-vscode-editor-foreground truncate">
                      {session.title || "Untitled"}
                    </div>
                    <div className="text-[10px] text-vscode-descriptionForeground mt-0.5">
                      {session.provider} • {session.messageCount} messages
                    </div>
                    <div className="text-[9px] text-vscode-descriptionForeground opacity-60 mt-1">
                      {formatDate(session.lastModified)}
                    </div>
                  </div>
                  <button
                    onClick={(e) => handleDelete(session.id, e)}
                    className="p-1 opacity-0 group-hover:opacity-100 hover:bg-red-500 hover:bg-opacity-20 rounded transition-all text-red-400"
                  >
                    <Trash2 size={12} />
                  </button>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Footer */}
      <div className="px-4 py-3 border-t border-vscode-widget-border space-y-2">
        <div className="text-[10px] text-vscode-descriptionForeground">
          {stats.sessionCount} sessions • {(stats.storageSize / 1024).toFixed(1)} KB
        </div>
        {sessions.length > 0 && (
          <Button
            variant="secondary"
            className="w-full text-xs h-7"
            onClick={handleClearAll}
          >
            <Trash2 size={12} className="mr-1" />
            Clear All
          </Button>
        )}
      </div>
    </div>
  );
});

HistoryPanel.displayName = "HistoryPanel";
