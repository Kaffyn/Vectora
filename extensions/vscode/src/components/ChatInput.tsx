import React, { useState, useRef, useEffect } from "react";
import { Send, Square, Command, Mic, ArrowLeft, FileText, ListChecks } from "lucide-react";
import { cn } from "@utils/cn";
import { ContextSelector } from "./selectors/ContextSelector";
import { ModeSelector, MODES } from "./selectors/ModeSelector";
import { ModelSelector, MODELS } from "./selectors/ModelSelector";
import { PolicySelector, POLICIES } from "./selectors/PolicySelector";

interface ChatInputProps {
  onSend: (text: string) => void;
  onCancel?: () => void;
  isStreaming: boolean;
}

export const ChatInput: React.FC<ChatInputProps> = ({ onSend, onCancel, isStreaming }) => {
  const [text, setText] = useState("");
  const [selectedModel, setSelectedModel] = useState(MODELS[2]);
  const [selectedMode, setSelectedMode] = useState(MODES[0]);
  const [selectedPolicy, setSelectedPolicy] = useState(POLICIES[0]);

  const textareaRef = useRef<HTMLTextAreaElement>(null);

  useEffect(() => {
    const textarea = textareaRef.current;
    if (textarea) {
      textarea.style.height = "auto";
      textarea.style.height = `${Math.min(textarea.scrollHeight, 180)}px`;
    }
  }, [text]);

  const handleSend = () => {
    if (text.trim() && !isStreaming) {
      onSend(text);
      setText("");
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  };

  return (
    <div className="flex flex-col gap-3 px-4 pb-4">
      {/* Status Row - Above the input box */}
      <div className="flex items-center justify-between px-1 mb-1">
        <div className="flex items-center gap-2">
          <button className="p-1 rounded-md hover:bg-vscode-toolbar-hoverBackground text-vscode-descriptionForeground transition-colors opacity-60 hover:opacity-100">
            <ArrowLeft size={14} />
          </button>
          <div className="flex items-center gap-1.5 px-2 py-0.5 rounded bg-vscode-badge-background/5 text-vscode-descriptionForeground">
            <FileText size={12} className="opacity-70" />
            <span className="text-[11px] font-medium">0 Files With Changes</span>
          </div>
        </div>

        <button className="flex items-center gap-1.5 px-2.5 py-1 rounded-lg bg-vscode-button-secondaryBackground/10 hover:bg-vscode-button-secondaryBackground text-vscode-button-secondaryForeground text-[11px] font-bold transition-all active:scale-95 shadow-sm border border-vscode-border/10">
          <ListChecks size={14} className="opacity-80" />
          <span>Review Changes</span>
        </button>
      </div>

      <div
        className={cn(
          "flex flex-col bg-vscode-input-background border border-vscode-input-border rounded-xl transition-all focus-within:border-vscode-focusBorder shadow-lg overflow-hidden",
          isStreaming && "opacity-90",
        )}
      >
        <textarea
          ref={textareaRef}
          rows={1}
          value={text}
          onChange={(e) => setText(e.target.value)}
          onKeyDown={handleKeyDown}
          placeholder="Ask anything, @ to mention, / for workflows"
          className="w-full bg-transparent border-none outline-none resize-none text-[13px] px-4 py-3 min-h-[44px] max-h-[180px] text-vscode-input-foreground placeholder:text-vscode-input-placeholderForeground/50 scrollbar-none font-medium"
          disabled={isStreaming}
        />

        <div className="flex items-center justify-between px-2 pb-2 h-10 select-none">
          <div className="flex items-center gap-1">
            <ContextSelector />
            <ModeSelector selectedMode={selectedMode} onSelect={setSelectedMode} />
            <ModelSelector selectedModel={selectedModel} onSelect={setSelectedModel} />
            <PolicySelector selectedPolicy={selectedPolicy} onSelect={setSelectedPolicy} />
          </div>

          <div className="flex items-center gap-2">
            <button className="p-1.5 rounded-md hover:bg-vscode-toolbar-hoverBackground text-vscode-descriptionForeground transition-colors">
              <Mic size={14} />
            </button>
            {isStreaming ? (
              <button
                onClick={onCancel}
                className="w-7 h-7 flex items-center justify-center rounded-full bg-vscode-testing-iconErrored text-white hover:brightness-110 active:scale-95 transition-all shadow-md"
                title="Stop generation"
              >
                <Square size={10} fill="currentColor" />
              </button>
            ) : (
              <button
                onClick={handleSend}
                disabled={!text.trim()}
                className="w-7 h-7 flex items-center justify-center rounded-full bg-vscode-button-background text-vscode-button-foreground hover:bg-vscode-button-hoverBackground disabled:opacity-30 disabled:grayscale transition-all shadow-md active:scale-95"
              >
                <Send size={12} strokeWidth={2.5} />
              </button>
            )}
          </div>
        </div>
      </div>

      <div className="flex items-center justify-center gap-4 px-2 select-none">
        <div className="flex items-center gap-1.5 text-vscode-descriptionForeground opacity-40 hover:opacity-100 transition-opacity">
          <Command size={10} />
          <span className="text-[9px] font-bold uppercase tracking-tighter">Enter to send</span>
        </div>
        <div className="w-1 h-1 rounded-full bg-vscode-border/20 opacity-20" />
        <div className="flex items-center gap-1.5 text-vscode-descriptionForeground opacity-40 hover:opacity-100 transition-opacity">
          <span className="text-[9px] font-bold uppercase tracking-tighter">Shift + Enter for multiline</span>
        </div>
      </div>
    </div>
  );
};
