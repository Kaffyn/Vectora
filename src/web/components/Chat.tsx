import React from "react";

interface Message {
  role: string;
  content: string;
}

interface ChatProps {
  messages: Message[];
}

export function Chat({ messages }: ChatProps) {
  if (messages.length === 0) {
    return (
      <div className="flex flex-col gap-4 p-8">
        <p className="text-muted-foreground text-center">No messages yet</p>
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-4 p-6 overflow-y-auto max-h-[80vh]">
      {messages.map((msg, i) => (
        <div
          key={i}
          className={`p-4 rounded-xl shadow-sm border animate-in fade-in slide-in-from-bottom-2 duration-300 ${
            msg.role === "user"
              ? "bg-primary/10 border-primary/20 self-end max-w-[80%]"
              : "bg-secondary/10 border-secondary/20 self-start max-w-[80%]"
          }`}
        >
          <p className="text-sm font-medium mb-1 opacity-70 uppercase tracking-tight">
            {msg.role}
          </p>
          <div className="text-base leading-relaxed">{msg.content}</div>
        </div>
      ))}
    </div>
  );
}

// Support function for tests
export function renderChat(messages: Message[]) {
  // Simple representation for tests since we are not using a full DOM mock here
  if (messages.length === 0) return "No messages yet";
  return messages.map((m) => m.content).join(" ");
}
