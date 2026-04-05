'use client'

interface Message {
  id: string
  role: string
  content: string
  timestamp: Date
}

export default function MessageBubble({ message }: { message: Message }) {
  const isUser = message.role === 'user'

  return (
    <div className={`flex ${isUser ? 'justify-end' : 'justify-start'}`}>
      <div
        className={`max-w-2xl rounded-lg px-4 py-3 ${
          isUser
            ? 'bg-emerald-600 text-white'
            : 'bg-zinc-800 text-zinc-50'
        }`}
      >
        <p>{message.content}</p>
        <div className="text-xs opacity-60 mt-2">
          {new Date(message.timestamp).toLocaleTimeString('pt-BR')}
        </div>
      </div>
    </div>
  )
}
