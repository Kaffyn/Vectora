/**
 * useDashboard Hook
 * React hook for dashboard data and analytics
 */

import { useCallback, useRef, useEffect, useState } from 'react';
import { Analytics } from '../analytics/analytics';
import { AnalyticsConfigManager } from '../analytics/analyticsConfig';
import {
  DashboardData,
  PerformanceMetrics,
  HealthStatus,
  Alert,
  Recommendation,
  UsageMetrics,
} from '../types/analytics';

export interface DashboardState {
  loading: boolean;
  data: DashboardData | null;
  error: Error | null;
  lastUpdate: number;
}

export function useDashboard() {
  const analyticsRef = useRef<Analytics | null>(null);
  const configRef = useRef<AnalyticsConfigManager | null>(null);
  const [state, setState] = useState<DashboardState>({
    loading: true,
    data: null,
    error: null,
    lastUpdate: 0,
  });
  const refreshIntervalRef = useRef<NodeJS.Timer | null>(null);

  // Initialize analytics
  useEffect(() => {
    if (!analyticsRef.current) {
      const config = AnalyticsConfigManager.getInstance();
      configRef.current = config;

      const analyticsConfig = config.getConfig();
      analyticsRef.current = Analytics.getInstance(analyticsConfig);

      // Initial load
      loadDashboardData();
    }

    return () => {
      if (refreshIntervalRef.current) {
        clearInterval(refreshIntervalRef.current);
      }
    };
  }, []);

  // Load dashboard data
  const loadDashboardData = useCallback(
    async (startTime?: number, endTime?: number) => {
      try {
        setState((prev) => ({ ...prev, loading: true, error: null }));

        if (!analyticsRef.current) {
          throw new Error('Analytics not initialized');
        }

        const data = analyticsRef.current.getDashboardData(startTime, endTime);

        setState({
          loading: false,
          data,
          error: null,
          lastUpdate: Date.now(),
        });
      } catch (error) {
        const err = error instanceof Error ? error : new Error(String(error));
        setState({
          loading: false,
          data: null,
          error: err,
          lastUpdate: Date.now(),
        });
      }
    },
    [],
  );

  // Start auto-refresh
  const startAutoRefresh = useCallback((intervalMs: number = 30000) => {
    if (refreshIntervalRef.current) {
      clearInterval(refreshIntervalRef.current);
    }

    refreshIntervalRef.current = setInterval(() => {
      loadDashboardData();
    }, intervalMs);
  }, [loadDashboardData]);

  // Stop auto-refresh
  const stopAutoRefresh = useCallback(() => {
    if (refreshIntervalRef.current) {
      clearInterval(refreshIntervalRef.current);
      refreshIntervalRef.current = null;
    }
  }, []);

  // Get performance metrics
  const getPerformanceMetrics = useCallback(
    (startTime?: number, endTime?: number): PerformanceMetrics | null => {
      if (!analyticsRef.current || !state.data) return null;

      const start = startTime || state.data.period.start;
      const end = endTime || state.data.period.end;

      return analyticsRef.current.getMetrics(start, end);
    },
    [state.data],
  );

  // Get health status
  const getHealthStatus = useCallback((): HealthStatus | null => {
    if (!analyticsRef.current) return null;
    return analyticsRef.current.getHealth();
  }, []);

  // Get active alerts
  const getActiveAlerts = useCallback((): Alert[] => {
    if (!analyticsRef.current) return [];
    return analyticsRef.current.getActiveAlerts();
  }, []);

  // Get recommendations
  const getRecommendations = useCallback((): Recommendation[] => {
    if (!state.data) return [];
    return state.data.recommendations;
  }, [state.data]);

  // Get usage metrics
  const getUsageMetrics = useCallback((): UsageMetrics | null => {
    if (!state.data) return null;
    return state.data.usage;
  }, [state.data]);

  // Export data
  const exportData = useCallback(
    async (
      format: 'json' | 'csv' = 'json',
      startTime?: number,
      endTime?: number,
    ): Promise<string> => {
      if (!analyticsRef.current) {
        throw new Error('Analytics not initialized');
      }

      return analyticsRef.current.exportData(
        format,
        startTime || state.data?.period.start,
        endTime || state.data?.period.end,
      );
    },
    [state.data],
  );

  // Clear old data
  const clearOldData = useCallback((retentionDays?: number) => {
    if (!analyticsRef.current) return;
    analyticsRef.current.clearOldData(retentionDays);
  }, []);

  // Get overview summary
  const getOverviewSummary = useCallback(
    () => {
      if (!state.data) {
        return {
          totalChats: 0,
          totalRAGQueries: 0,
          totalTokens: 0,
          totalErrors: 0,
        };
      }
      return state.data.overview;
    },
    [state.data],
  );

  // Get top features
  const getTopFeatures = useCallback(
    () => {
      if (!state.data) return [];
      return state.data.topFeatures;
    },
    [state.data],
  );

  // Get cost breakdown
  const getCostBreakdown = useCallback(
    () => {
      if (!state.data) return [];
      return state.data.costBreakdown;
    },
    [state.data],
  );

  // Resolve alert
  const resolveAlert = useCallback((alertId: string) => {
    if (!analyticsRef.current) return;
    analyticsRef.current.resolveAlert(alertId);
  }, []);

  return {
    // State
    state,
    loading: state.loading,
    data: state.data,
    error: state.error,
    lastUpdate: state.lastUpdate,

    // Methods
    loadDashboardData,
    startAutoRefresh,
    stopAutoRefresh,
    getPerformanceMetrics,
    getHealthStatus,
    getActiveAlerts,
    getRecommendations,
    getUsageMetrics,
    exportData,
    clearOldData,
    getOverviewSummary,
    getTopFeatures,
    getCostBreakdown,
    resolveAlert,
  };
}
