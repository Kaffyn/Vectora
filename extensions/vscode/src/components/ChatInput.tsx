import React, { useState, useRef, useEffect } from 'react';
import { Send, Square, Command, Plus, ChevronDown, Mic, Sparkles, Image as ImageIcon, AtSign, Zap, Shield, AlertTriangle, Terminal, Edit3 } from 'lucide-react';
import { cn } from '../utils/cn';
import { Dropdown, DropdownItem } from './Dropdown';

interface ChatInputProps {
    onSend: (text: string) => void;
    onCancel?: () => void;
    isStreaming: boolean;
}

type ModelItem = { id: string; name: string; info?: string; isNew?: boolean; isThinking?: boolean; group: string };
type ModeItem = { id: string; name: string; description: string };
type PolicyItem = { id: string; name: string; description: string; icon: React.ReactNode };

const MODELS: ModelItem[] = [
    { id: 'gemini-3.1-pro-high', name: 'Gemini 3.1 Pro (High)', info: 'High', isNew: true, group: 'Google' },
    { id: 'gemini-3.1-pro-low', name: 'Gemini 3.1 Pro (Low)', info: 'Low', isNew: true, group: 'Google' },
    { id: 'gemini-3-flash', name: 'Gemini 3 Flash', group: 'Google' },
    { id: 'claude-sonnet-4.6', name: 'Claude Sonnet 4.6', info: 'Thinking', isThinking: true, group: 'Anthropic' },
    { id: 'claude-opus-4.6', name: 'Claude Opus 4.6', info: 'Thinking', isThinking: true, group: 'Anthropic' },
    { id: 'gpt-oss-120b', name: 'GPT-OSS 120B', info: 'Medium', group: 'Open Source' },
];

const MODES: ModeItem[] = [
    { id: 'planning', name: 'Planning', description: 'Agent can plan before executing tasks. Use for deep research, complex tasks, or collaborative work' },
    { id: 'fast', name: 'Fast', description: 'Agent will execute tasks directly. Use for simple tasks that can be completed faster' },
];

const POLICIES: PolicyItem[] = [
    { id: 'ask', name: 'Ask before edits', description: 'Based on diffs, where you must accept changes', icon: <Edit3 size={14} /> },
    { id: 'automatic', name: 'Edit automatically', description: 'Can edit files, but terminal/search/embed need auth', icon: <Terminal size={14} /> },
    { id: 'yolo', name: 'YOLO', description: 'Full autonomous mode, everything allowed', icon: <Zap size={14} className="text-yellow-500" /> },
];

export const ChatInput: React.FC<ChatInputProps> = ({ onSend, onCancel, isStreaming }) => {
    const [text, setText] = useState('');
    const [selectedModel, setSelectedModel] = useState(MODELS[2]);
    const [selectedMode, setSelectedMode] = useState(MODES[0]);
    const [selectedPolicy, setSelectedPolicy] = useState(POLICIES[0]);
    
    const textareaRef = useRef<HTMLTextAreaElement>(null);

    useEffect(() => {
        const textarea = textareaRef.current;
        if (textarea) {
            textarea.style.height = 'auto';
            textarea.style.height = `${Math.min(textarea.scrollHeight, 180)}px`;
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

            <div className={cn(
                "flex flex-col bg-vscode-input-background border border-vscode-input-border rounded-xl transition-all focus-within:border-vscode-focusBorder shadow-lg overflow-hidden",
                isStreaming && "opacity-90"
            )}>
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
                        {/* Context Menu */}
                        <Dropdown 
                            trigger={
                                <div className="p-1.5 rounded-md hover:bg-vscode-toolbar-hoverBackground text-vscode-descriptionForeground transition-colors">
                                    <Plus size={14} />
                                </div>
                            }
                        >
                            <div className="px-3 py-1.5 text-[10px] font-bold text-vscode-descriptionForeground/60 uppercase">Add context</div>
                            <DropdownItem onClick={() => {}} className="flex-row items-center gap-2">
                                <ImageIcon size={14} /> <span>Media</span>
                            </DropdownItem>
                            <DropdownItem onClick={() => {}} className="flex-row items-center gap-2">
                                <AtSign size={14} /> <span>Mentions</span>
                            </DropdownItem>
                            <DropdownItem onClick={() => {}} className="flex-row items-center gap-2 pb-2">
                                <Edit3 size={14} /> <span>Workflows</span>
                            </DropdownItem>
                        </Dropdown>

                        {/* Mode Selector */}
                        <Dropdown 
                            trigger={
                                <div className="flex items-center gap-1 px-2 py-1 rounded-md hover:bg-vscode-toolbar-hoverBackground cursor-pointer text-vscode-descriptionForeground group">
                                     <span className="text-[11px] font-bold group-hover:text-vscode-editor-foreground transition-colors">{selectedMode.name}</span>
                                     <ChevronDown size={12} className="opacity-50" />
                                </div>
                            }
                            contentClassName="min-w-[280px]"
                        >
                            {MODES.map(mode => (
                                <DropdownItem key={mode.id} onClick={() => setSelectedMode(mode)} active={selectedMode.id === mode.id} className="py-2.5">
                                    <span className="font-bold">{mode.name}</span>
                                    <span className="opacity-60 text-[10px] leading-tight mt-0.5">{mode.description}</span>
                                </DropdownItem>
                            ))}
                        </Dropdown>

                        {/* Model Selector */}
                        <Dropdown 
                            trigger={
                                <div className="flex items-center gap-1 px-2 py-1 rounded-md hover:bg-vscode-toolbar-hoverBackground cursor-pointer text-vscode-descriptionForeground group">
                                     <Sparkles size={12} className="text-vscode-button-background" />
                                     <span className="text-[11px] font-bold group-hover:text-vscode-editor-foreground transition-colors">{selectedModel.name}</span>
                                     <ChevronDown size={12} className="opacity-50" />
                                </div>
                            }
                            contentClassName="min-w-[240px]"
                        >
                            <div className="px-3 py-1.5 text-[10px] font-bold text-vscode-descriptionForeground/60 uppercase">Model</div>
                            {MODELS.map(model => (
                                <DropdownItem key={model.id} onClick={() => setSelectedModel(model)} active={selectedModel.id === model.id} className="flex-row items-center justify-between">
                                    <div className="flex items-center gap-2">
                                        <span className="font-medium">{model.name}</span>
                                        {model.isThinking && <AlertTriangle size={10} className="opacity-60" />}
                                    </div>
                                    <div className="flex items-center gap-1">
                                        {model.isNew && <span className="text-[8px] bg-vscode-badge-background px-1 rounded opacity-80 uppercase font-black">New</span>}
                                    </div>
                                </DropdownItem>
                            ))}
                        </Dropdown>

                        {/* Policy Selector */}
                        <Dropdown 
                            trigger={
                                <div className="flex items-center gap-1 px-2 py-1 rounded-md hover:bg-vscode-toolbar-hoverBackground cursor-pointer text-vscode-descriptionForeground group" title="Safety Policy">
                                     {selectedPolicy.icon}
                                     <ChevronDown size={12} className="opacity-50" />
                                </div>
                            }
                            contentClassName="min-w-[280px]"
                            align="right"
                        >
                            <div className="px-3 py-1.5 text-[10px] font-bold text-vscode-descriptionForeground/60 uppercase">Safety Policy</div>
                            {POLICIES.map(policy => (
                                <DropdownItem key={policy.id} onClick={() => setSelectedPolicy(policy)} active={selectedPolicy.id === policy.id} className="py-2.5">
                                    <div className="flex items-center gap-2">
                                        {policy.icon}
                                        <span className="font-bold">{policy.name}</span>
                                    </div>
                                    <span className="opacity-60 text-[10px] leading-tight mt-0.5">{policy.description}</span>
                                </DropdownItem>
                            ))}
                        </Dropdown>
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
