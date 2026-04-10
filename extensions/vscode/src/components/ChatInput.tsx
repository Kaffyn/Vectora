import React, { useState, useRef, useEffect } from 'react';
import { Send, Square } from 'lucide-react';

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
        <div className="p-4 bg-vscode-bg border-t border-vscode-border/50">
            <div className="relative group flex items-end gap-2 bg-vscode-inputBg border border-vscode-border rounded-xl px-3 py-2 transition-all focus-within:border-vscode-accent/50 focus-within:shadow-[0_0_0_1px_rgba(var(--vscode-button-background),0.1)]">
                <textarea
                    ref={textareaRef}
                    rows={1}
                    value={text}
                    onChange={(e) => setText(e.target.value)}
                    onKeyDown={handleKeyDown}
                    placeholder={isStreaming ? "Vectora is thinking..." : "Ask Vectora anything..."}
                    className="flex-grow bg-transparent border-none outline-none resize-none text-[13px] py-1 max-h-[200px] text-vscode-inputFg placeholder:opacity-50"
                    disabled={isStreaming}
                />
                
                {isStreaming ? (
                    <button
                        onClick={onCancel}
                        className="p-1.5 rounded-lg bg-vscode-accent/10 text-vscode-accent hover:bg-vscode-accent/20 transition-colors"
                        title="Cancel generation"
                    >
                        <Square size={16} fill="currentColor" />
                    </button>
                ) : (
                    <button
                        onClick={handleSend}
                        disabled={!text.trim()}
                        className="p-1.5 rounded-lg bg-vscode-accent text-vscode-accentFg hover:opacity-90 disabled:opacity-20 disabled:grayscale transition-all"
                    >
                        <Send size={16} />
                    </button>
                )}
            </div>
            <div className="mt-2 flex justify-center">
                <span className="text-[10px] opacity-30 italic">Press Enter to send, Shift+Enter for new line</span>
            </div>
        </div>
    );
};
