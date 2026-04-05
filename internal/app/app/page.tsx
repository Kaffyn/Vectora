'use client'

import { useEffect, useState } from 'react'
import Sidebar from '@/components/Common/Sidebar'
import Header from '@/components/Common/Header'
import ChatPage from '@/app/chat/page'

export default function Home() {
  const [mounted, setMounted] = useState(false)

  useEffect(() => {
    setMounted(true)
  }, [])

  if (!mounted) return null

  return (
    <div className="flex h-screen">
      <Sidebar />
      <div className="flex-1 flex flex-col">
        <Header />
        <ChatPage />
      </div>
    </div>
  )
}
