/**
 * useStreamingContent - Hook for managing streaming content updates
 *
 * Handles smooth content streaming with debouncing and animation.
 */

import { useState, useCallback, useRef, useEffect } from "react";

interface StreamingState {
  content: string;
  isStreaming: boolean;
  tokenCount: number;
  duration: number; // milliseconds
}

/**
 * Custom hook for managing streaming content
 */
export function useStreamingContent(initialContent = "") {
  const [state, setState] = useState<StreamingState>({
    content: initialContent,
    isStreaming: false,
    tokenCount: 0,
    duration: 0,
  });

  const startTimeRef = useRef<number>(0);
  const updateIntervalRef = useRef<NodeJS.Timeout>();

  /**
   * Start streaming
   */
  const startStreaming = useCallback(() => {
    startTimeRef.current = Date.now();
    setState((prev) => ({
      ...prev,
      isStreaming: true,
      content: "",
      tokenCount: 0,
      duration: 0,
    }));

    // Update duration every 100ms
    updateIntervalRef.current = setInterval(() => {
      setState((prev) => ({
        ...prev,
        duration: Date.now() - startTimeRef.current,
      }));
    }, 100);
  }, []);

  /**
   * Append chunk to streaming content
   */
  const appendChunk = useCallback((chunk: string) => {
    setState((prev) => {
      const newContent = prev.content + chunk;
      return {
        ...prev,
        content: newContent,
        tokenCount: Math.ceil(newContent.length / 4), // Rough token estimate
      };
    });
  }, []);

  /**
   * Stop streaming
   */
  const stopStreaming = useCallback(() => {
    if (updateIntervalRef.current) {
      clearInterval(updateIntervalRef.current);
    }
    setState((prev) => ({
      ...prev,
      isStreaming: false,
      duration: Date.now() - startTimeRef.current,
    }));
  }, []);

  /**
   * Reset content
   */
  const reset = useCallback(() => {
    if (updateIntervalRef.current) {
      clearInterval(updateIntervalRef.current);
    }
    setState({
      content: initialContent,
      isStreaming: false,
      tokenCount: 0,
      duration: 0,
    });
  }, [initialContent]);

  /**
   * Cleanup on unmount
   */
  useEffect(() => {
    return () => {
      if (updateIntervalRef.current) {
        clearInterval(updateIntervalRef.current);
      }
    };
  }, []);

  return {
    ...state,
    startStreaming,
    appendChunk,
    stopStreaming,
    reset,
  };
}
