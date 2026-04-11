import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";
import rehypeHighlight from "rehype-highlight";
import { User, Cpu, CheckCircle2, Loader2, AlertCircle } from "lucide-react";
import type { Message, ToolCall } from "../types";
import { cn } from "@utils/cn";

interface ChatMessageProps {
  message: Message;
}

export const ChatMessage: React.FC<ChatMessageProps> = ({ message }) => {
  const isUser = message.role === "user";

  return (
    <div
      className={cn(
        "flex w-full gap-4 py-3 border-b border-vscode-border/10 transition-all duration-300 animate-in fade-in slide-in-from-bottom-3",
        isUser ? "bg-transparent" : "bg-vscode-editor-background/50 backdrop-blur-sm px-2 rounded-lg my-1",
      )}
    >
      <div className="flex-grow min-w-0 space-y-1">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-2">
            <span className="text-[10px] font-medium text-vscode-descriptionForeground">
              {new Date(message.timestamp).toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" })}
            </span>
          </div>
        </div>

        <div
          className={cn(
            "markdown-content text-[13px] leading-relaxed break-words overflow-x-auto selection:bg-vscode-editor-selectionBackground",
            isUser ? "text-vscode-editor-foreground font-medium" : "text-vscode-editor-foreground opacity-90",
          )}
        >
          <ReactMarkdown
            remarkPlugins={[remarkGfm]}
            rehypePlugins={[rehypeHighlight]}
            components={{
              pre: ({ node, ...props }) => (
                <div className="relative group my-4 rounded-lg overflow-hidden border border-vscode-widget-border shadow-sm">
                  <div className="flex items-center justify-between px-3 py-1 bg-vscode-editorWidget-background border-b border-vscode-widget-border">
                    <span className="text-[9px] font-bold text-vscode-descriptionForeground uppercase tracking-tighter">
                      Code
                    </span>
                  </div>
                  <pre
                    {...props}
                    className="bg-vscode-editor-background p-4 overflow-x-auto font-mono text-[11px] leading-relaxed scrollbar-thin"
                  />
                </div>
              ),
              code: ({ node, ...props }) => {
                const isInline = !props.className;
                return isInline ? (
                  <code
                    className="bg-vscode-editor-selectionBackground/30 px-1 py-0.5 rounded text-vscode-editor-foreground font-mono text-[12px]"
                    {...props}
                  />
                ) : (
                  <code {...props} />
                );
              },
            }}
          >
            {message.content}
          </ReactMarkdown>
        </div>

        {message.tools && message.tools.length > 0 && (
          <div className="flex flex-wrap gap-2 mt-2 pt-2 border-t border-vscode-border/5">
            {message.tools.map((tool) => (
              <ToolBadge key={tool.id} tool={tool} />
            ))}
          </div>
        )}
      </div>
    </div>
  );
};

const ToolBadge: React.FC<{ tool: ToolCall }> = ({ tool }) => {
  const getIcon = () => {
    switch (tool.status) {
      case "completed":
        return <CheckCircle2 size={11} className="text-vscode-testing-iconPassed" />;
      case "in_progress":
        return <Loader2 size={11} className="animate-spin text-vscode-button-background" />;
      case "failed":
        return <AlertCircle size={11} className="text-vscode-testing-iconErrored" />;
    }
  };

  return (
    <div className="group flex items-center gap-1.5 px-2 py-0.5 rounded border border-vscode-widget-border bg-vscode-editorWidget-background text-[10px] hover:border-vscode-focusBorder transition-all cursor-default">
      {getIcon()}
      <span className="text-vscode-descriptionForeground group-hover:text-vscode-editor-foreground transition-colors font-medium">
        {tool.title}
      </span>
    </div>
  );
};
