/**
 * useSessionHistory - Custom hook for chat session history management
 *
 * Provides functions to save, load, and manage chat session history
 * with automatic persistence.
 */

import { useCallback, useState, useEffect } from "react";
import type { ChatSession } from "@/types/vectora";
import {
  saveSession,
  loadSession,
  listSessions,
  searchSessions,
  deleteSession,
  clearHistory,
  getStorageSize,
} from "@/utils/historyStorage";

export interface HistoryStats {
  sessionCount: number;
  storageSize: number;
}

/**
 * Hook for managing chat session history
 */
export function useSessionHistory() {
  const [stats, setStats] = useState<HistoryStats>({ sessionCount: 0, storageSize: 0 });

  /**
   * Update statistics
   */
  const updateStats = useCallback(() => {
    const sessions = listSessions();
    const storageSize = getStorageSize();
    setStats({
      sessionCount: sessions.length,
      storageSize,
    });
  }, []);

  /**
   * Initialize stats on mount
   */
  useEffect(() => {
    updateStats();
  }, [updateStats]);

  /**
   * Save current session
   */
  const save = useCallback(
    (session: ChatSession) => {
      try {
        saveSession(session);
        updateStats();
        return true;
      } catch (err) {
        console.error("Failed to save session:", err);
        return false;
      }
    },
    [updateStats]
  );

  /**
   * Load session by ID
   */
  const load = useCallback((sessionId: string) => {
    return loadSession(sessionId);
  }, []);

  /**
   * Get list of all sessions
   */
  const list = useCallback(() => {
    return listSessions();
  }, []);

  /**
   * Search sessions
   */
  const search = useCallback((query: string) => {
    return searchSessions(query);
  }, []);

  /**
   * Delete session
   */
  const remove = useCallback(
    (sessionId: string) => {
      const success = deleteSession(sessionId);
      if (success) {
        updateStats();
      }
      return success;
    },
    [updateStats]
  );

  /**
   * Clear all history
   */
  const clear = useCallback(() => {
    clearHistory();
    updateStats();
  }, [updateStats]);

  return {
    // State
    stats,

    // Actions
    save,
    load,
    list,
    search,
    remove,
    clear,
    updateStats,
  };
}
