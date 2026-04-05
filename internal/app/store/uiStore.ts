import { create } from 'zustand'

type TabType = 'chat' | 'codigo' | 'index' | 'manager'

interface UIState {
  activeTab: TabType
  setActiveTab: (tab: TabType) => void
  loading: boolean
  setLoading: (loading: boolean) => void
}

export const useUIStore = create<UIState>((set) => ({
  activeTab: 'chat',
  setActiveTab: (tab) => set({ activeTab: tab }),
  loading: false,
  setLoading: (loading) => set({ loading }),
}))
