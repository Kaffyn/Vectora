import React from 'react';
import { Plus, History, Settings, X, Cpu } from 'lucide-react';
import { vscode } from '../utils/vscode';

export const ChatHeader: React.FC = () => {
    return (
        <div className="flex items-center justify-between px-4 py-3 border-b border-vscode-border bg-vscode-bg/80 backdrop-blur-md sticky top-0 z-10">
            <div className="flex items-center gap-2">
                <div className="w-5 h-5 rounded-md bg-vscode-accent flex items-center justify-center">
                    <Cpu size={12} className="text-vscode-accentFg" />
                </div>
                <h1 className="text-[11px] font-bold uppercase tracking-[2px] opacity-80">
                    Vectora AI
                </h1>
            </div>

            <div className="flex items-center gap-1.5 grayscale opacity-60 hover:grayscale-0 hover:opacity-100 transition-all">
                <button 
                    onClick={() => vscode.postMessage({ type: 'new_chat' })}
                    className="p-1.5 rounded-md hover:bg-vscode-inputBg transition-colors"
                    title="New Chat"
                >
                    <Plus size={16} />
                </button>
                <button 
                    onClick={() => vscode.postMessage({ type: 'show_history' })}
                    className="p-1.5 rounded-md hover:bg-vscode-inputBg transition-colors"
                    title="History"
                >
                    <History size={16} />
                </button>
                <button 
                    onClick={() => vscode.postMessage({ type: 'open_settings' })}
                    className="p-1.5 rounded-md hover:bg-vscode-inputBg transition-colors"
                    title="Settings"
                >
                    <Settings size={16} />
                </button>
                <div className="w-[1px] h-4 bg-vscode-border/50 mx-1" />
                <button 
                    onClick={() => vscode.postMessage({ type: 'close_panel' })}
                    className="p-1.5 rounded-md hover:bg-vscode-inputBg transition-colors text-red-400/70 hover:text-red-400"
                    title="Hide Panel"
                >
                    <X size={16} />
                </button>
            </div>
        </div>
    );
};
