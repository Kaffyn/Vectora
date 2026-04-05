'use client'

import { useUIStore } from '@/store/uiStore'

const TAB_TITLES = {
  chat: 'Chat',
  codigo: 'Código',
  index: 'Index',
  manager: 'Manager',
}

export default function Header() {
  const { activeTab } = useUIStore()

  return (
    <header className="h-16 border-b border-zinc-800 bg-zinc-900 px-6 flex items-center justify-between">
      <h2 className="text-lg font-semibold">
        {TAB_TITLES[activeTab as keyof typeof TAB_TITLES]}
      </h2>
      <div className="text-sm text-zinc-500">v2.0</div>
    </header>
  )
}
