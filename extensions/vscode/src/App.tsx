import { useState, useEffect, useRef } from 'react';
import { ChatHeader } from './components/ChatHeader';
import { ChatMessage } from './components/ChatMessage';
import { ChatInput } from './components/ChatInput';
import type { Message } from './types';
import { MOCK_MESSAGES } from './MockData';
import { vscode } from './utils/vscode';

function App() {
    const [messages, setMessages] = useState<Message[]>(MOCK_MESSAGES);
    const [isStreaming, setIsStreaming] = useState(false);
    const messagesEndRef = useRef<HTMLDivElement>(null);

    const scrollToBottom = () => {
        messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
    };

    useEffect(() => {
        scrollToBottom();
    }, [messages]);

    useEffect(() => {
        // Listen for messages from the extension host
        window.addEventListener('message', (event) => {
            const message = event.data;
            switch (message.type) {
                case 'inject_message':
                    // Real implementation would handle this
                    break;
                case 'set_messages':
                    setMessages(message.messages);
                    break;
            }
        });
    }, []);

    const handleSendMessage = (text: string) => {
        const newMessage: Message = {
            id: Date.now().toString(),
            role: 'user',
            content: text,
            timestamp: Date.now(),
        };
        
        setMessages((prev: Message[]) => [...prev, newMessage]);
        vscode.postMessage({ type: 'send_prompt', text });
        
        // Simulating auto-response for the Mock demo
        setIsStreaming(true);
        setTimeout(() => {
            const response: Message = {
                id: (Date.now() + 1).toString(),
                role: 'agent',
                content: "I've received your message in the professional React UI! Since we are in **Mock Mode**, I am just echoing that your input was: " + text,
                timestamp: Date.now(),
            };
            setMessages((prev: Message[]) => [...prev, response]);
            setIsStreaming(false);
        }, 1500);
    };

    const handleCancel = () => {
        setIsStreaming(false);
        vscode.postMessage({ type: 'cancel_generation' });
    };

    return (
        <div className="flex flex-col h-screen bg-vscode-bg text-vscode-fg overflow-hidden">
            <ChatHeader />
            
            <div className="flex-grow overflow-y-auto scrollbar-none scroll-smooth pb-4">
                <div className="max-w-3xl mx-auto px-2">
                    {messages.map((msg: Message) => (
                        <ChatMessage key={msg.id} message={msg} />
                    ))}
                    <div ref={messagesEndRef} />
                </div>
            </div>

            <ChatInput 
                onSend={handleSendMessage} 
                onCancel={handleCancel}
                isStreaming={isStreaming} 
            />
        </div>
    );
}

export default App;
