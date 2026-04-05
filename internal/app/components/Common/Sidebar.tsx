'use client'

import { useUIStore } from '@/store/uiStore'
import { MessageSquare, Code2, Database, Settings } from 'lucide-react'

const TABS = [
  { id: 'chat', label: 'Chat', icon: MessageSquare },
  { id: 'codigo', label: 'Código', icon: Code2 },
  { id: 'index', label: 'Index', icon: Database },
  { id: 'manager', label: 'Manager', icon: Settings },
]

export default function Sidebar() {
  const { activeTab, setActiveTab } = useUIStore()

  return (
    <aside className="w-64 bg-zinc-900 border-r border-zinc-800 flex flex-col">
      <div className="p-4 border-b border-zinc-800">
        <h1 className="text-xl font-bold text-emerald-500">Vectora</h1>
      </div>

      <nav className="flex-1 p-4 space-y-2">
        {TABS.map((tab) => {
          const Icon = tab.icon
          const isActive = activeTab === tab.id

          return (
            <button
              key={tab.id}
              onClick={() => setActiveTab(tab.id as any)}
              className={`w-full flex items-center gap-3 px-4 py-3 rounded-lg transition-colors ${
                isActive
                  ? 'bg-emerald-500/20 text-emerald-400 border border-emerald-500/50'
                  : 'text-zinc-400 hover:text-zinc-300 hover:bg-zinc-800'
              }`}
            >
              <Icon className="w-5 h-5" />
              <span>{tab.label}</span>
            </button>
          )
        })}
      </nav>

      <div className="border-t border-zinc-800 p-4">
        <div className="flex items-center gap-2 text-xs text-zinc-500">
          <div className="w-2 h-2 bg-emerald-500 rounded-full" />
          Daemon Conectado
        </div>
      </div>
    </aside>
  )
}
