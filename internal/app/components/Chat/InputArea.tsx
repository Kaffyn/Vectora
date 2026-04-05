'use client'

import { useRef, useState } from 'react'
import { Send } from 'lucide-react'

export default function InputArea({
  onSend,
  disabled = false,
}: {
  onSend: (query: string) => void
  disabled?: boolean
}) {
  const [value, setValue] = useState('')
  const textareaRef = useRef<HTMLTextAreaElement>(null)

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      handleSend()
    }
  }

  const handleSend = () => {
    if (!value.trim() || disabled) return
    onSend(value)
    setValue('')
    if (textareaRef.current) {
      textareaRef.current.style.height = 'auto'
    }
  }

  const handleInputChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    setValue(e.target.value)
    if (textareaRef.current) {
      textareaRef.current.style.height = 'auto'
      textareaRef.current.style.height = `${Math.min(textareaRef.current.scrollHeight, 200)}px`
    }
  }

  return (
    <div className="border-t border-zinc-800 p-6 bg-zinc-900">
      <div className="flex gap-3 items-end">
        <textarea
          ref={textareaRef}
          value={value}
          onChange={handleInputChange}
          onKeyDown={handleKeyDown}
          disabled={disabled}
          className="flex-1 px-4 py-3 rounded-lg bg-zinc-800 text-zinc-50 border border-zinc-700 resize-none focus:outline-none focus:ring-2 focus:ring-emerald-500"
          rows={1}
          placeholder="Faça uma pergunta..."
        />
        <button
          onClick={handleSend}
          disabled={!value.trim() || disabled}
          className="px-4 py-3 bg-emerald-600 hover:bg-emerald-700 text-white rounded-lg disabled:opacity-50"
        >
          <Send className="w-5 h-5" />
        </button>
      </div>
    </div>
  )
}
