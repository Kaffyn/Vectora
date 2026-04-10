import React from 'react';
import { Plus, History, Settings, X, Sparkles } from 'lucide-react';
import { vscode } from '../utils/vscode';

export const ChatHeader: React.FC = () => {
    return (
        <div className="flex items-center justify-between px-4 py-3 border-b border-vscode-border/20 bg-vscode-bg/60 backdrop-blur-xl sticky top-0 z-20 shadow-sm">
            <div className="flex items-center gap-2.5 group cursor-default">
                <div className="w-6 h-6 rounded-lg bg-vscode-accent flex items-center justify-center shadow-lg shadow-vscode-accent/30 group-hover:scale-110 transition-transform duration-500">
                    <Sparkles size={14} className="text-vscode-accentFg animate-pulse" />
                </div>
                <div className="flex flex-col -space-y-1">
                    <h1 className="text-[10px] font-black uppercase tracking-[3px] opacity-90 text-vscode-accent">
                        Vectora
                    </h1>
                    <span className="text-[8px] font-bold opacity-30 uppercase tracking-widest">
                        Advanced AI
                    </span>
                </div>
            </div>

            <div className="flex items-center gap-1 opacity-40 hover:opacity-100 transition-opacity duration-300">
                <button 
                    onClick={() => vscode.postMessage({ type: 'clear' })}
                    className="p-1.5 rounded-lg hover:bg-vscode-inputBg hover:text-vscode-accent transition-all active:scale-90"
                    title="Clear Session"
                >
                    <Plus size={15} strokeWidth={2.5} />
                </button>
                <button 
                    onClick={() => vscode.postMessage({ type: 'show_history' })}
                    className="p-1.5 rounded-lg hover:bg-vscode-inputBg hover:text-vscode-accent transition-all active:scale-90"
                    title="History"
                >
                    <History size={15} strokeWidth={2.5} />
                </button>
                <div className="w-[1px] h-3.5 bg-vscode-border/30 mx-0.5" />
                <button 
                    onClick={() => vscode.postMessage({ type: 'open_settings' })}
                    className="p-1.5 rounded-lg hover:bg-vscode-inputBg hover:text-vscode-accent transition-all active:scale-90"
                    title="Configure Vectora"
                >
                    <Settings size={15} strokeWidth={2.5} />
                </button>
            </div>
        </div>
    );
};
