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
        <div className="flex flex-col gap-3">
            <div className="relative group transition-all duration-300">
                <div className="relative flex items-end gap-2 bg-vscode-input-background border border-vscode-input-border rounded-lg px-3 py-2 transition-all focus-within:border-vscode-focusBorder shadow-sm">
                    <textarea
                        ref={textareaRef}
                        rows={1}
                        value={text}
                        onChange={(e) => setText(e.target.value)}
                        onKeyDown={handleKeyDown}
                        placeholder={isStreaming ? "Vectora is solving..." : "Ask anything to Vectora..."}
                        className="flex-grow bg-transparent border-none outline-none resize-none text-[13px] py-1 max-h-[180px] text-vscode-input-foreground placeholder:text-vscode-input-placeholderForeground scrollbar-none font-medium"
                        disabled={isStreaming}
                    />
                    
                    <div className="flex flex-col justify-end h-full py-0.5">
                        {isStreaming ? (
                            <button
                                onClick={onCancel}
                                className="p-2 rounded-md bg-vscode-testing-iconErrored/10 text-vscode-testing-iconErrored hover:bg-vscode-testing-iconErrored/20 transition-all border border-vscode-testing-iconErrored/20"
                                title="Stop generation"
                            >
                                <Square size={12} fill="currentColor" strokeWidth={3} />
                            </button>
                        ) : (
                            <button
                                onClick={handleSend}
                                disabled={!text.trim()}
                                className="p-2 rounded-md bg-vscode-button-background text-vscode-button-foreground hover:bg-vscode-button-hoverBackground disabled:opacity-30 disabled:grayscale transition-all shadow-sm border border-vscode-button-border"
                            >
                                <Send size={12} strokeWidth={3} />
                            </button>
                        )}
                    </div>
                </div>
            </div>
            
            <div className="flex items-center justify-center gap-4 px-2 select-none">
                <div className="flex items-center gap-1.5 text-vscode-descriptionForeground opacity-70 hover:opacity-100 transition-opacity">
                    <Command size={10} />
                    <span className="text-[9px] font-bold uppercase tracking-tighter">Enter to send</span>
                </div>
                <div className="w-1 h-1 rounded-full bg-vscode-border/20" />
                <div className="flex items-center gap-1.5 text-vscode-descriptionForeground opacity-70 hover:opacity-100 transition-opacity">
                    <span className="text-[9px] font-bold uppercase tracking-tighter">Shift + Enter for multiline</span>
                </div>
            </div>
        </div>
    );
};
