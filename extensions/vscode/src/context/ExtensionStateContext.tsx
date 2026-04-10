import React, { createContext, useCallback, useContext, useEffect, useState } from "react"
import { vscode } from "../utils/vscode"

// Minimal types for the ported components
export interface ExtensionState {
    clineMessages: any[];
    apiConfiguration?: any;
    customModes?: any[];
    mode?: string;
    soundEnabled?: boolean;
    soundVolume?: number;
    reasoningBlockCollapsed?: boolean;
    taskHistory?: any[];
    shouldShowAnnouncement?: boolean;
}

export interface ExtensionStateContextType extends ExtensionState {
    didHydrateState: boolean;
    setMode: (mode: string) => void;
    setReasoningBlockCollapsed: (collapsed: boolean) => void;
}

export const ExtensionStateContext = createContext<ExtensionStateContextType | undefined>(undefined)

export const ExtensionStateContextProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
    const [state, setState] = useState<ExtensionState>({
        clineMessages: [],
        apiConfiguration: { apiProvider: 'gemini' },
        mode: 'engineer',
        soundEnabled: false,
        soundVolume: 0.5,
        reasoningBlockCollapsed: true,
        taskHistory: [],
    })
    const [didHydrateState, setDidHydrateState] = useState(false)

    const handleMessage = useCallback((event: MessageEvent) => {
        const message = event.data
        switch (message.type) {
            case 'state':
                setState(prev => ({ ...prev, ...message.state }))
                setDidHydrateState(true)
                break
            case 'user_message':
                setState(prev => ({
                    ...prev,
                    clineMessages: [...prev.clineMessages, { type: 'say', say: 'text', text: message.text, ts: Date.now() }]
                }))
                break
            case 'agent_chunk':
                setState(prev => {
                    const last = prev.clineMessages[prev.clineMessages.length - 1]
                    if (last && last.role === 'agent') {
                        const updated = { ...last, text: (last.text || '') + message.text }
                        return { ...prev, clineMessages: [...prev.clineMessages.slice(0, -1), updated] }
                    }
                    return {
                        ...prev,
                        clineMessages: [...prev.clineMessages, { type: 'say', say: 'text', text: message.text, ts: Date.now(), role: 'agent' }]
                    }
                })
                break
        }
    }, [])

    useEffect(() => {
        window.addEventListener("message", handleMessage)
        vscode.postMessage({ type: "webviewDidLaunch" })
        return () => window.removeEventListener("message", handleMessage)
    }, [handleMessage])

    const contextValue: ExtensionStateContextType = {
        ...state,
        didHydrateState,
        setMode: (mode) => setState(prev => ({ ...prev, mode })),
        setReasoningBlockCollapsed: (reasoningBlockCollapsed) => setState(prev => ({ ...prev, reasoningBlockCollapsed })),
    }

    return <ExtensionStateContext.Provider value={contextValue}>{children}</ExtensionStateContext.Provider>
}

export const useExtensionState = () => {
    const context = useContext(ExtensionStateContext)
    if (context === undefined) {
        throw new Error("useExtensionState must be used within an ExtensionStateContextProvider")
    }
    return context
}
