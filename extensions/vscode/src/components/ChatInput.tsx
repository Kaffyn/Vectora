import React, { useState, useRef, useEffect } from 'react';
import { Send, Square, Command } from 'lucide-react';

interface ChatInputProps {
    onSend: (text: string) => void;
    onCancel?: () => void;
    isStreaming: boolean;
}

export const ChatInput: React.FC<ChatInputProps> = ({ onSend, onCancel, isStreaming }) => {
    const [text, setText] = useState('');
    const textareaRef = useRef<HTMLTextAreaElement>(null);

    // Auto-resize textarea
    useEffect(() => {
        const textarea = textareaRef.current;
        if (textarea) {
            textarea.style.height = 'auto';
            textarea.style.height = `${Math.min(textarea.scrollHeight, 200)}px`;
        }
    }, [text]);

    const handleSend = () => {
        if (text.trim() && !isStreaming) {
            onSend(text);
            setText('');
        }
    };

    const handleKeyDown = (e: React.KeyboardEvent) => {
        if (e.key === 'Enter' && !e.shiftKey) {
            e.preventDefault();
            handleSend();
        }
    };

    return (
        <div className="flex flex-col gap-2">
            <div className="relative group transition-all duration-500">
                <div className="absolute -inset-0.5 bg-gradient-to-r from-vscode-accent to-purple-600 rounded-xl blur opacity-20 group-focus-within:opacity-40 transition duration-1000 group-focus-within:duration-200"></div>
                <div className="relative flex items-end gap-2 bg-vscode-inputBg/80 backdrop-blur-xl border border-vscode-border/40 rounded-xl px-4 py-3 transition-all focus-within:border-vscode-accent/50 shadow-2xl">
                    <textarea
                        ref={textareaRef}
                        rows={1}
                        value={text}
                        onChange={(e) => setText(e.target.value)}
                        onKeyDown={handleKeyDown}
                        placeholder={isStreaming ? "Vectora is solving..." : "Ask anything to Vectora..."}
                        className="flex-grow bg-transparent border-none outline-none resize-none text-[13px] py-1 max-h-[200px] text-vscode-inputFg placeholder:text-vscode-fg/30 scrollbar-none font-medium"
                        disabled={isStreaming}
                    />
                    
                    <div className="flex flex-col justify-end h-full py-0.5">
                        {isStreaming ? (
                            <button
                                onClick={onCancel}
                                className="p-2 rounded-lg bg-red-500/10 text-rose-500 hover:bg-red-500/20 transition-all hover:scale-110 active:scale-95 border border-rose-500/20"
                                title="Stop generation"
                            >
                                <Square size={14} fill="currentColor" strokeWidth={3} />
                            </button>
                        ) : (
                            <button
                                onClick={handleSend}
                                disabled={!text.trim()}
                                className="p-2 rounded-lg bg-vscode-accent text-vscode-accentFg hover:brightness-110 disabled:opacity-20 disabled:grayscale transition-all shadow-lg active:scale-95 border border-vscode-accent/50"
                            >
                                <Send size={14} strokeWidth={3} />
                            </button>
                        )}
                    </div>
                </div>
            </div>
            
            <div className="flex items-center justify-center gap-4 px-2 select-none">
                <div className="flex items-center gap-1.5 opacity-50 hover:opacity-80 transition-opacity">
                    <Command size={10} />
                    <span className="text-[9px] font-bold uppercase tracking-tighter text-vscode-fg">Enter to send</span>
                </div>
                <div className="w-1 h-1 rounded-full bg-vscode-fg/20" />
                <div className="flex items-center gap-1.5 opacity-50 hover:opacity-80 transition-opacity">
                    <span className="text-[9px] font-bold uppercase tracking-tighter text-vscode-fg">Shift + Enter for multiline</span>
                </div>
            </div>
        </div>
    );
};
