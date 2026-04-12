/**
 * Vectora State Context
 * Minimal, focused state management para extensão VS Code
 */

import React, { createContext, useContext, ReactNode, useState, useCallback, useEffect } from "react";
import type { ExtensionState, Provider, SessionUpdate, ChatMessage, VectoraSettings } from "@/types";

/**
 * Contexto de estado global Vectora - minimalista
 */
const VectoraStateContext = createContext<ExtensionState | null>(null);

/**
 * Provider para estado Vectora
 */
export function VectoraStateProvider({ children }: { children: ReactNode }) {
  const [state, setState] = useState<ExtensionState>(() => ({
    coreStatus: "stopped",
    isConnected: false,
    selectedProvider: null,
    selectedModel: null,
    workspace: null,
    currentSessionId: null,
    isProcessing: false,
    availableProviders: [],
    availableModels: [],
    error: null,
    workspaceEmbeddingStatus: {
      isIndexing: false,
      progress: 0,
      vectorCount: 0,
    },
  }));

  return (
    <VectoraStateContext.Provider value={state}>
      {children}
    </VectoraStateContext.Provider>
  );
}

/**
 * Hook para acessar estado Vectora
 */
export function useVectoraState(): ExtensionState {
  const context = useContext(VectoraStateContext);
  if (!context) {
    throw new Error("useVectoraState must be used within VectoraStateProvider");
  }
  return context;
}

/**
 * Hook para gerenciar conexão com Core
 */
export function useCoreConnection() {
  const state = useVectoraState();
  const [isConnecting, setIsConnecting] = useState(false);

  const connect = useCallback(async () => {
    if (state.isConnected) return;
    setIsConnecting(true);
    try {
      // TODO: Chamar client.connect() quando implementado
      console.log("Connecting to Vectora Core...");
    } finally {
      setIsConnecting(false);
    }
  }, [state.isConnected]);

  const disconnect = useCallback(async () => {
    if (!state.isConnected) return;
    // TODO: Chamar client.disconnect()
    console.log("Disconnecting from Vectora Core...");
  }, [state.isConnected]);

  return {
    isConnected: state.isConnected,
    isConnecting,
    coreStatus: state.coreStatus,
    connect,
    disconnect,
  };
}

/**
 * Hook para gerenciar sessão ativa
 */
export function useSession() {
  const state = useVectoraState();
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [isLoading, setIsLoading] = useState(false);

  const createSession = useCallback(async (provider: string, model: string) => {
    // TODO: Chamar client.createSession()
    console.log(`Creating session with ${provider}/${model}`);
  }, []);

  const sendMessage = useCallback(async (content: string) => {
    if (!state.currentSessionId) {
      throw new Error("No active session");
    }

    setIsLoading(true);
    try {
      // TODO: Chamar client.prompt()
      console.log("Sending message:", content);
    } finally {
      setIsLoading(false);
    }
  }, [state.currentSessionId]);

  const cancelSession = useCallback(async () => {
    if (state.currentSessionId) {
      // TODO: Chamar client.cancelSession()
      console.log("Canceling session:", state.currentSessionId);
    }
  }, [state.currentSessionId]);

  return {
    sessionId: state.currentSessionId,
    messages,
    isLoading,
    provider: state.selectedProvider,
    model: state.selectedModel,
    createSession,
    sendMessage,
    cancelSession,
  };
}

/**
 * Hook para gerenciar seleção de provider
 */
export function useProviderSelection() {
  const state = useVectoraState();

  const selectProvider = useCallback((providerId: string, modelId?: string) => {
    // TODO: Atualizar estado
    console.log(`Selected provider: ${providerId}, model: ${modelId}`);
  }, []);

  return {
    providers: state.availableProviders,
    selectedProvider: state.selectedProvider,
    selectedModel: state.selectedModel,
    selectProvider,
  };
}

/**
 * Hook para gerenciar embedding/RAG
 */
export function useEmbedding() {
  const state = useVectoraState();

  const startIndexing = useCallback(async () => {
    // TODO: Chamar IPC para começar indexing
    console.log("Starting workspace indexing...");
  }, []);

  const stopIndexing = useCallback(async () => {
    // TODO: Chamar IPC para parar indexing
    console.log("Stopping workspace indexing...");
  }, []);

  return {
    isIndexing: state.workspaceEmbeddingStatus?.isIndexing || false,
    progress: state.workspaceEmbeddingStatus?.progress || 0,
    vectorCount: state.workspaceEmbeddingStatus?.vectorCount || 0,
    startIndexing,
    stopIndexing,
  };
}
