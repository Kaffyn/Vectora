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
            "flex w-full gap-3 py-4 border-b border-vscode-border/30 animate-in fade-in slide-in-from-bottom-2",
            isUser ? "bg-transparent" : "bg-vscode-bg/50"
        )}>
            <div className="flex-shrink-0 mt-1">
                {isUser ? (
                    <div className="w-8 h-8 rounded-full bg-vscode-accent flex items-center justify-center text-vscode-accentFg">
                        <User size={18} />
                    </div>
                ) : (
                    <div className="w-8 h-8 rounded-full bg-vscode-fg/10 flex items-center justify-center text-vscode-fg/70">
                        <Cpu size={18} />
                    </div>
                )}
            </div>

            <div className="flex-grow min-w-0 space-y-3">
                <div className="flex items-center gap-2">
                    <span className="text-[11px] font-bold uppercase tracking-wider opacity-50">
                        {isUser ? 'You' : 'Vectora'}
                    </span>
                    <span className="text-[10px] opacity-30">
                        {new Date(message.timestamp).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                    </span>
                </div>

                <div className="markdown-content text-[13px] leading-relaxed break-words overflow-x-auto pr-4">
                    <ReactMarkdown
                        remarkPlugins={[remarkGfm]}
                        rehypePlugins={[rehypeHighlight]}
                        components={{
                            pre: ({ node, ...props }) => (
                                <div className="relative group my-3">
                                    <pre {...props} className="bg-vscode-inputBg/50 p-4 rounded-lg border border-vscode-border/50 overflow-x-auto font-mono text-xs" />
                                </div>
                            ),
                            code: ({ node, ...props }) => (
                                <code {...props} className="bg-vscode-inputBg/80 px-1 rounded text-vscode-accent font-mono text-[12px]" />
                            ),
                        }}
                    >
                        {message.content}
                    </ReactMarkdown>
                </div>

                {message.tools && message.tools.length > 0 && (
                    <div className="flex flex-wrap gap-2 mt-4 pt-4 border-t border-vscode-border/20">
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
            case 'completed': return <CheckCircle2 size={12} className="text-green-500" />;
            case 'in_progress': return <Loader2 size={12} className="animate-spin text-vscode-accent" />;
            case 'failed': return <AlertCircle size={12} className="text-red-500" />;
        }
    };

    return (
        <div className="flex items-center gap-2 px-2 py-1 rounded-md bg-vscode-inputBg border border-vscode-border/50 text-[11px] hover:border-vscode-accent/50 transition-colors shadow-sm">
            {getIcon()}
            <span className="opacity-80">{tool.title}</span>
        </div>
    );
};
