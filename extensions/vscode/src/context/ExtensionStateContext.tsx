import React, { createContext, useContext, useState, useCallback, ReactNode } from 'react';

export interface ExtensionState {
  isConnected: boolean;
  coreStatus: 'starting' | 'running' | 'stopped' | 'error';
  selectedProvider: string | null;
  workspace: string | null;
  error: string | null;
}

interface ExtensionContextType {
  state: ExtensionState;
  setConnected: (connected: boolean) => void;
  setCoreStatus: (status: ExtensionState['coreStatus']) => void;
  setSelectedProvider: (provider: string | null) => void;
  setWorkspace: (workspace: string | null) => void;
  setError: (error: string | null) => void;
}

const ExtensionStateContext = createContext<ExtensionContextType | undefined>(undefined);

export const ExtensionStateContextProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
  const [state, setState] = useState<ExtensionState>({
    isConnected: false,
    coreStatus: 'stopped',
    selectedProvider: null,
    workspace: null,
    error: null,
  });

  const setConnected = useCallback((connected: boolean) => {
    setState(prev => ({ ...prev, isConnected: connected }));
  }, []);

  const setCoreStatus = useCallback((coreStatus: ExtensionState['coreStatus']) => {
    setState(prev => ({ ...prev, coreStatus }));
  }, []);

  const setSelectedProvider = useCallback((provider: string | null) => {
    setState(prev => ({ ...prev, selectedProvider: provider }));
  }, []);

  const setWorkspace = useCallback((workspace: string | null) => {
    setState(prev => ({ ...prev, workspace }));
  }, []);

  const setError = useCallback((error: string | null) => {
    setState(prev => ({ ...prev, error }));
  }, []);

  const value: ExtensionContextType = {
    state,
    setConnected,
    setCoreStatus,
    setSelectedProvider,
    setWorkspace,
    setError,
  };

  return (
    <ExtensionStateContext.Provider value={value}>
      {children}
    </ExtensionStateContext.Provider>
  );
};

export const useExtensionState = (): ExtensionContextType => {
  const context = useContext(ExtensionStateContext);
  if (context === undefined) {
    throw new Error('useExtensionState must be used within ExtensionStateContextProvider');
  }
  return context;
};
