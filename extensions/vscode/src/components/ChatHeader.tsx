import React from "react";
import { Plus, History, Settings, X, Sparkles } from "lucide-react";
import { vscode } from "@utils/vscode";

export const ChatHeader: React.FC = () => {
  return (
    <div className="flex items-center justify-between px-4 py-2 border-b border-vscode-border/10 bg-vscode-editor-background sticky top-0 z-20 shadow-sm">
      <div className="flex items-center gap-2 group cursor-default">
        <div className="flex flex-col -space-y-1">
          <h1 className="text-[10px] font-black uppercase tracking-[2px] text-vscode-editor-foreground">Vectora</h1>
        </div>
      </div>

      <div className="flex items-center gap-1">
        <button
          onClick={() => vscode.postMessage({ type: "clear" })}
          className="p-1 px-2 rounded-md hover:bg-vscode-toolbar-hoverBackground text-vscode-descriptionForeground hover:text-vscode-editor-foreground transition-all active:scale-95 flex items-center gap-1.5"
          title="Clear Session"
        >
          <Plus size={13} strokeWidth={2.5} />
          <span className="text-[9px] font-bold uppercase tracking-tight">Clear</span>
        </button>
        <div className="w-[1px] h-3 bg-vscode-border/20 mx-1" />
        <button
          onClick={() => vscode.postMessage({ type: "show_history" })}
          className="p-1.5 rounded-md hover:bg-vscode-toolbar-hoverBackground text-vscode-descriptionForeground hover:text-vscode-editor-foreground transition-all active:scale-90"
          title="History"
        >
          <History size={14} strokeWidth={2.5} />
        </button>
        <button
          onClick={() => vscode.postMessage({ type: "open_settings" })}
          className="p-1.5 rounded-md hover:bg-vscode-toolbar-hoverBackground text-vscode-descriptionForeground hover:text-vscode-editor-foreground transition-all active:scale-90"
          title="Settings"
        >
          <Settings size={14} strokeWidth={2.5} />
        </button>
      </div>
    </div>
  );
};
