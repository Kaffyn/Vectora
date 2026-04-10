import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import rehypeHighlight from 'rehype-highlight';
import { User, Cpu, CheckCircle2, Loader2, AlertCircle } from 'lucide-react';
import type { Message, ToolCall } from '../types';
import { clsx, type ClassValue } from 'clsx';
import { twMerge } from 'tailwind-merge';

function cn(...inputs: ClassValue[]) {
    return twMerge(clsx(inputs));
}

interface ChatMessageProps {
    message: Message;
}

export const ChatMessage: React.FC<ChatMessageProps> = ({ message }) => {
    const isUser = message.role === 'user';

    return (
        <div className={cn(
            "flex w-full gap-4 py-6 border-b border-vscode-border/20 transition-all duration-300 animate-in fade-in slide-in-from-bottom-3",
            isUser ? "bg-transparent" : "bg-vscode-bg/30 backdrop-blur-sm px-4 rounded-xl my-2 border border-vscode-border/10 shadow-sm"
        )}>
            <div className="flex-shrink-0 mt-1">
                {isUser ? (
                    <div className="w-9 h-9 rounded-xl bg-vscode-accent flex items-center justify-center text-vscode-accentFg shadow-md shadow-vscode-accent/20 border border-vscode-accent/50">
                        <User size={20} strokeWidth={2.5} />
                    </div>
                ) : (
                    <div className="w-9 h-9 rounded-xl bg-vscode-fg/5 flex items-center justify-center text-vscode-accent border border-vscode-border/50 shadow-inner">
                        <Cpu size={20} strokeWidth={2.5} />
                    </div>
                )}
            </div>

            <div className="flex-grow min-w-0 space-y-3">
                <div className="flex items-center justify-between">
                    <div className="flex items-center gap-2">
                        <span className={cn(
                            "text-[10px] font-black uppercase tracking-widest",
                            isUser ? "text-vscode-accent" : "text-vscode-fg/60"
                        )}>
                            {isUser ? 'Master' : 'Vectora Intelligence'}
                        </span>
                        <div className="w-1 h-1 rounded-full bg-vscode-fg/20" />
                        <span className="text-[10px] opacity-30 font-medium">
                            {new Date(message.timestamp).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                        </span>
                    </div>
                </div>

                <div className={cn(
                    "markdown-content text-[13px] leading-relaxed break-words overflow-x-auto selection:bg-vscode-accent/20",
                    isUser ? "text-vscode-fg/90" : "text-vscode-fg/80"
                )}>
                    <ReactMarkdown
                        remarkPlugins={[remarkGfm]}
                        rehypePlugins={[rehypeHighlight]}
                        components={{
                            pre: ({ node, ...props }) => (
                                <div className="relative group my-4 rounded-xl overflow-hidden border border-vscode-border/30 shadow-lg">
                                    <div className="flex items-center justify-between px-4 py-2 bg-vscode-inputBg/80 backdrop-blur-md border-b border-vscode-border/20">
                                        <span className="text-[10px] font-bold opacity-40 uppercase tracking-tighter">Code Snippet</span>
                                        <div className="flex gap-1.5 focus-within:opacity-100 opacity-0 group-hover:opacity-100 transition-opacity">
                                            <div className="w-2.5 h-2.5 rounded-full bg-red-500/50" />
                                            <div className="w-2.5 h-2.5 rounded-full bg-yellow-500/50" />
                                            <div className="w-2.5 h-2.5 rounded-full bg-green-500/50" />
                                        </div>
                                    </div>
                                    <pre {...props} className="bg-vscode-inputBg/40 p-5 overflow-x-auto font-mono text-[11px] leading-relaxed scrollbar-thin" />
                                </div>
                            ),
                            code: ({ node, ...props }) => {
                                const isInline = !props.className;
                                return isInline ? (
                                    <code className="bg-vscode-accent/10 px-1.5 py-0.5 rounded text-vscode-accent font-mono text-[12px] border border-vscode-accent/5" {...props} />
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
                    <div className="flex flex-wrap gap-2 mt-4 pt-4 border-t border-vscode-border/10">
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
            case 'completed': return <CheckCircle2 size={13} className="text-emerald-500" />;
            case 'in_progress': return <Loader2 size={13} className="animate-spin text-vscode-accent" />;
            case 'failed': return <AlertCircle size={13} className="text-rose-500" />;
        }
    };

    return (
        <div className="group flex items-center gap-2 px-3 py-1.5 rounded-lg bg-vscode-inputBg/60 backdrop-blur-sm border border-vscode-border/30 text-[11px] hover:border-vscode-accent/40 hover:bg-vscode-inputBg/80 transition-all cursor-default shadow-sm active:scale-95">
            {getIcon()}
            <span className="font-semibold text-vscode-fg/70 group-hover:text-vscode-fg transition-colors tracking-tight">
                {tool.title}
            </span>
        </div>
    );
};
