/**
 * useAnalytics Hook
 * React hook for analytics tracking
 */

import { useCallback, useRef, useEffect } from 'react';
import { Analytics } from '../analytics/analytics';
import { AnalyticsConfigManager } from '../analytics/analyticsConfig';
import { AnalyticsEvent, EventType } from '../types/analytics';

export interface UseAnalyticsOptions {
  trackingEnabled?: boolean;
  sessionId?: string;
  userId?: string;
}

export function useAnalytics(options?: UseAnalyticsOptions) {
  const analyticsRef = useRef<Analytics | null>(null);
  const configRef = useRef<AnalyticsConfigManager | null>(null);

  // Initialize analytics on first render
  useEffect(() => {
    if (!analyticsRef.current) {
      const config = AnalyticsConfigManager.getInstance();
      configRef.current = config;

      const analyticsConfig = config.getConfig();
      analyticsRef.current = Analytics.getInstance(analyticsConfig);

      // Log startup event
      if (config.isTrackingAllowed()) {
        analyticsRef.current.trackEvent('system.startup', {
          timestamp: Date.now(),
        });
      }
    }
  }, []);

  // Track event
  const trackEvent = useCallback(
    (type: EventType, metadata: Record<string, unknown>): AnalyticsEvent => {
      if (!analyticsRef.current) {
        throw new Error('Analytics not initialized');
      }

      if (!configRef.current?.isTrackingAllowed()) {
        // Return dummy event if tracking disabled
        return {
          id: 'dummy',
          type,
          timestamp: Date.now(),
          metadata,
          sessionId: '',
        };
      }

      return analyticsRef.current.trackEvent(type, metadata);
    },
    [],
  );

  // Track chat message
  const trackChat = useCallback(
    (data: {
      messageLength: number;
      tokenCount?: number;
      provider?: string;
      model?: string;
      responseTime?: number;
      error?: string;
    }) => {
      if (!analyticsRef.current) return;
      analyticsRef.current.trackChat(data);
    },
    [],
  );

  // Track RAG operation
  const trackRAG = useCallback(
    (data: {
      query: string;
      resultCount: number;
      relevanceScore?: number;
      executionTime: number;
      cacheHit?: boolean;
    }) => {
      if (!analyticsRef.current) return;
      analyticsRef.current.trackRAG(data);
    },
    [],
  );

  // Track provider event
  const trackProvider = useCallback(
    (data: {
      provider: string;
      previousProvider?: string;
      type: 'switch' | 'fallback' | 'error';
      reason?: string;
      error?: string;
    }) => {
      if (!analyticsRef.current) return;
      analyticsRef.current.trackProvider(data);
    },
    [],
  );

  // Get analytics instance
  const getAnalytics = useCallback((): Analytics => {
    if (!analyticsRef.current) {
      throw new Error('Analytics not initialized');
    }
    return analyticsRef.current;
  }, []);

  // Check if tracking is enabled
  const isTrackingEnabled = useCallback((): boolean => {
    return configRef.current?.isTrackingAllowed() ?? false;
  }, []);

  return {
    trackEvent,
    trackChat,
    trackRAG,
    trackProvider,
    getAnalytics,
    isTrackingEnabled,
  };
}
