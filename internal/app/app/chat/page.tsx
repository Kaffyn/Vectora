'use client'

import { useState, useEffect } from 'react'
import { useChatStore } from '@/store/chatStore'
import ChatFeed from '@/components/Chat/ChatFeed'
import InputArea from '@/components/Chat/InputArea'

export default function ChatPage() {
  const { messages, addMessage } = useChatStore()
  const [isLoading, setIsLoading] = useState(false)

  const handleSendQuery = async (query: string) => {
    // Add user message
    addMessage({
      id: crypto.randomUUID(),
      role: 'user',
      content: query,
      timestamp: new Date(),
    })

    setIsLoading(true)
    // Simula resposta do LLM
    setTimeout(() => {
      addMessage({
        id: crypto.randomUUID(),
        role: 'assistant',
        content: 'Resposta do LLM será aqui...',
        timestamp: new Date(),
      })
      setIsLoading(false)
    }, 1000)
  }

  return (
    <div className="flex-1 flex flex-col overflow-hidden">
      <ChatFeed messages={messages} isLoading={isLoading} />
      <InputArea onSend={handleSendQuery} disabled={isLoading} />
    </div>
  )
}
