'use client'

import { useEffect, useRef } from 'react'
import MessageBubble from './MessageBubble'

interface Message {
  id: string
  role: string
  content: string
  timestamp: Date
}

interface Props {
  messages: Message[]
  isLoading: boolean
}

export default function ChatFeed({ messages, isLoading }: Props) {
  const scrollRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight
    }
  }, [messages])

  return (
    <div ref={scrollRef} className="flex-1 overflow-y-auto p-6 space-y-4">
      {messages.length === 0 ? (
        <div className="flex items-center justify-center h-full text-zinc-500">
          <p>Comece digitando uma pergunta...</p>
        </div>
      ) : (
        <>
          {messages.map((msg) => (
            <MessageBubble key={msg.id} message={msg} />
          ))}
          {isLoading && (
            <div className="flex justify-start">
              <div className="bg-zinc-800 rounded-lg px-4 py-3 animate-pulse">
                <div className="w-32 h-4 bg-zinc-700 rounded"></div>
              </div>
            </div>
          )}
        </>
      )}
    </div>
  )
}
