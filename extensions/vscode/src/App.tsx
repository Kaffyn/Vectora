import { useState, useEffect, useRef, useCallback } from 'react';
import { ChatHeader } from './components/ChatHeader';
import { ChatMessage } from './components/ChatMessage';
import { ChatInput } from './components/ChatInput';
import type { Message, ToolCall, Role } from './types';
import { vscode } from './utils/vscode';

function App() {
    const [messages, setMessages] = useState<Message[]>([]);
    const [isStreaming, setIsStreaming] = useState(false);
    const messagesEndRef = useRef<HTMLDivElement>(null);

    const scrollToBottom = useCallback(() => {
        messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
    }, []);

    useEffect(() => {
        scrollToBottom();
    }, [messages, scrollToBottom]);

    useEffect(() => {
        const handleMessage = (event: MessageEvent) => {
            const data = event.data;
            switch (data.type) {
                case 'set_messages':
                    setMessages(data.messages);
                    break;
                case 'inject_message':
                    handleSendMessage(data.text);
                    break;
                case 'user_message':
                    addMessage('user', data.text, data.id);
                    setIsStreaming(true);
                    break;
                case 'agent_chunk':
                    appendAgentChunk(data.id, data.text);
                    break;
                case 'tool_call':
                    handleToolCall(data);
                    break;
                case 'tool_call_update':
                    updateToolStatus(data.toolCallId, data.status);
                    break;
                case 'stream_end':
                    setIsStreaming(false);
                    break;
                case 'clear_chat':
                    setMessages([]);
                    break;
                case 'error':
                    addMessage('agent', `❌ Error: ${data.message}`);
                    setIsStreaming(false);
                    break;
            }
        };

        window.addEventListener('message', handleMessage);
        return () => window.removeEventListener('message', handleMessage);
    }, []);

    const addMessage = (role: Role, content: string, id?: string) => {
        const newMessage: Message = {
            id: id || Date.now().toString(),
            role,
            content,
            timestamp: Date.now(),
        };
        setMessages(prev => [...prev, newMessage]);
    };

    const appendAgentChunk = (_sessionId: string, text: string) => {
        setMessages(prev => {
            const last = prev[prev.length - 1];
            if (last && last.role === 'agent') {
                const updated = { ...last, content: last.content + text };
                return [...prev.slice(0, -1), updated];
            } else {
                return [...prev, {
                    id: Date.now().toString(),
                    role: 'agent',
                    content: text,
                    timestamp: Date.now()
                }];
            }
        });
    };

    const handleToolCall = (data: any) => {
        setMessages(prev => {
            const last = prev[prev.length - 1];
            if (last && last.role === 'agent') {
                const tool: ToolCall = {
                    id: data.toolCallId,
                    title: data.title,
                    status: data.status || 'in_progress'
                };
                const tools = last.tools ? [...last.tools, tool] : [tool];
                return [...prev.slice(0, -1), { ...last, tools }];
            }
            return prev;
        });
    };

    const updateToolStatus = (toolCallId: string, status: ToolCall['status']) => {
        setMessages(prev => {
            return prev.map(msg => {
                if (msg.role === 'agent' && msg.tools) {
                    return {
                        ...msg,
                        tools: msg.tools.map(t => t.id === toolCallId ? { ...t, status } : t)
                    };
                }
                return msg;
            });
        });
    };

    const handleSendMessage = (text: string) => {
        if (!text.trim() || isStreaming) return;
        vscode.postMessage({ type: 'send', text });
    };

    const handleCancel = () => {
        vscode.postMessage({ type: 'cancel' });
    };

    return (
        <div className="flex flex-col h-screen bg-vscode-bg text-vscode-fg overflow-hidden selection:bg-vscode-accent/30 Selection:text-vscode-accentFg font-sans">
            <ChatHeader />
            
            <div className="flex-grow overflow-y-auto scrollbar-none scroll-smooth pb-8 pt-2">
                <div className="max-w-3xl mx-auto px-4 space-y-2">
                    {messages.length === 0 ? (
                        <div className="h-[65vh] flex flex-col items-center justify-center space-y-4 animate-in fade-in duration-1000">
                             <div className="w-14 h-14 rounded-xl bg-vscode-widget-shadow/10 flex items-center justify-center border border-vscode-widget-border contrast-125">
                                <img src="https://raw.githubusercontent.com/Kaffyn/Vectora/main/media/icon.png" className="w-8 h-8 opacity-80" alt="Vectora" />
                             </div>
                             <div className="text-center">
                                <h2 className="text-sm font-bold tracking-tight text-vscode-editor-foreground">Vectora AI</h2>
                                <p className="text-[11px] text-vscode-descriptionForeground font-medium mt-1">Ready to assist Master with engineering questions</p>
                             </div>
                        </div>
                    ) : (
                        messages.map((msg: Message) => (
                            <ChatMessage key={msg.id} message={msg} />
                        ))
                    )}
                    <div ref={messagesEndRef} className="h-4" />
                </div>
            </div>

            <div className="p-4 bg-gradient-to-t from-vscode-bg via-vscode-bg to-transparent">
                <ChatInput 
                    onSend={handleSendMessage} 
                    onCancel={handleCancel}
                    isStreaming={isStreaming} 
                />
            </div>
        </div>
    );
}

export default App;
