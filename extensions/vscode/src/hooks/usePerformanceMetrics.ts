/**
 * usePerformanceMetrics Hook
 * React hook for tracking performance metrics
 */

import { useCallback, useRef, useEffect, useState } from 'react';
import { PerformanceMonitor } from '../monitoring/performanceMonitor';
import { MemoryMonitor } from '../monitoring/memoryMonitor';
import { EventTracker } from '../analytics/eventTracker';
import { PerformanceMetrics } from '../types/analytics';

export interface PerformanceData {
  responseTime: number;
  memoryUsage: number;
  memoryUsagePercent: number;
  cpuUsage: number;
  isHealthy: boolean;
}

export function usePerformanceMetrics(sessionId: string) {
  const performanceMonitorRef = useRef<PerformanceMonitor | null>(null);
  const memoryMonitorRef = useRef<MemoryMonitor | null>(null);
  const eventTrackerRef = useRef<EventTracker | null>(null);
  const [performanceData, setPerformanceData] = useState<PerformanceData>({
    responseTime: 0,
    memoryUsage: 0,
    memoryUsagePercent: 0,
    cpuUsage: 0,
    isHealthy: true,
  });

  // Initialize monitors
  useEffect(() => {
    if (!performanceMonitorRef.current) {
      eventTrackerRef.current = new EventTracker(sessionId);
      performanceMonitorRef.current = new PerformanceMonitor(
        sessionId,
        eventTrackerRef.current,
      );
      memoryMonitorRef.current = new MemoryMonitor(sessionId, eventTrackerRef.current);

      // Start memory monitoring
      memoryMonitorRef.current.startAutoMonitoring(5000);
    }

    return () => {
      // Cleanup
      memoryMonitorRef.current?.stopAutoMonitoring();
    };
  }, [sessionId]);

  // Start measuring operation
  const startMeasurement = useCallback((label: string): string => {
    if (!performanceMonitorRef.current) {
      throw new Error('Performance monitor not initialized');
    }
    return performanceMonitorRef.current.startMeasurement(label);
  }, []);

  // End measurement
  const endMeasurement = useCallback((markId: string): number => {
    if (!performanceMonitorRef.current) {
      throw new Error('Performance monitor not initialized');
    }
    return performanceMonitorRef.current.endMeasurement(markId);
  }, []);

  // Record metrics
  const recordMetrics = useCallback(
    (
      responseTime: number,
      memoryUsage: number,
      cpuUsage: number,
    ): void => {
      if (!performanceMonitorRef.current) return;

      performanceMonitorRef.current.recordPoint(responseTime, memoryUsage, cpuUsage);

      // Update state
      const memoryPercent = memoryMonitorRef.current?.getCurrentUsagePercent() ?? 0;
      setPerformanceData({
        responseTime,
        memoryUsage,
        memoryUsagePercent: memoryPercent * 100,
        cpuUsage,
        isHealthy: responseTime < 5000 && memoryPercent < 0.85 && cpuUsage < 80,
      });
    },
    [],
  );

  // Get performance summary
  const getMetricsSummary = useCallback((): PerformanceMetrics => {
    if (!performanceMonitorRef.current) {
      return {
        responseTime: 0,
        tokenInput: 0,
        tokenOutput: 0,
        costUSD: 0,
        cacheHitRate: 0,
        errorRate: 0,
        memoryUsageMB: 0,
        cpuUsagePercent: 0,
      };
    }
    return performanceMonitorRef.current.getSummary();
  }, []);

  // Get memory health
  const getMemoryHealth = useCallback(
    () => {
      if (!memoryMonitorRef.current) {
        return { status: 'healthy', usagePercent: 0, message: 'N/A' };
      }
      return memoryMonitorRef.current.getHealthStatus();
    },
    [],
  );

  // Get memory trend
  const getMemoryTrend = useCallback(
    (lastN: number = 50) => {
      if (!memoryMonitorRef.current) return [];
      return memoryMonitorRef.current.getMemoryTrend(lastN);
    },
    [],
  );

  // Detect memory leak
  const detectMemoryLeak = useCallback(
    () => {
      if (!memoryMonitorRef.current) {
        return { detected: false, trend: 0, recommendation: 'Not enough data' };
      }
      return memoryMonitorRef.current.detectMemoryLeak();
    },
    [],
  );

  // Get average response time
  const getAverageResponseTime = useCallback(
    (lastN: number = 100) => {
      if (!performanceMonitorRef.current) return 0;
      return performanceMonitorRef.current.getAverageResponseTime(lastN);
    },
    [],
  );

  // Clear old measurements
  const clearOldMeasurements = useCallback((keepLastN: number = 1000) => {
    if (!performanceMonitorRef.current) return;
    performanceMonitorRef.current.clearOldMeasurements(keepLastN);
  }, []);

  return {
    performanceData,
    startMeasurement,
    endMeasurement,
    recordMetrics,
    getMetricsSummary,
    getMemoryHealth,
    getMemoryTrend,
    detectMemoryLeak,
    getAverageResponseTime,
    clearOldMeasurements,
  };
}
