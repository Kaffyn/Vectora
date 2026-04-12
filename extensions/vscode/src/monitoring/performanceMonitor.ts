/**
 * Performance Monitor
 * Monitors application performance metrics
 */

import { PerformanceMetrics, Alert } from '../types/analytics';
import { EventTracker } from '../analytics/eventTracker';

export interface PerformancePoint {
  timestamp: number;
  responseTime: number;
  memoryUsage: number;
  cpuUsage: number;
}

export class PerformanceMonitor {
  private sessionId: string;
  private eventTracker: EventTracker;
  private measurements: PerformancePoint[] = [];
  private startMarkers: Map<string, number> = new Map();
  private thresholds: Record<string, number> = {
    responseTime: 5000, // 5 seconds
    memoryUsage: 500, // 500MB
    cpuUsage: 80, // 80%
  };

  constructor(sessionId: string, eventTracker: EventTracker) {
    this.sessionId = sessionId;
    this.eventTracker = eventTracker;
  }

  /**
   * Start measuring an operation
   */
  startMeasurement(label: string): string {
    const markId = `${label}-${Date.now()}`;
    this.startMarkers.set(markId, performance.now());
    return markId;
  }

  /**
   * End measurement and record the duration
   */
  endMeasurement(markId: string, metadata?: Record<string, unknown>): number {
    const startTime = this.startMarkers.get(markId);
    if (!startTime) {
      console.warn(`No start marker found for ${markId}`);
      return 0;
    }

    const duration = performance.now() - startTime;
    this.startMarkers.delete(markId);

    if (duration > this.thresholds.responseTime) {
      this.eventTracker.createEvent('system.warning', {
        message: `Slow operation: ${markId} took ${duration.toFixed(2)}ms`,
        duration,
        metadata,
      }, this.sessionId);
    }

    return duration;
  }

  /**
   * Record a performance point
   */
  recordPoint(
    responseTime: number,
    memoryUsage: number,
    cpuUsage: number,
  ): void {
    const point: PerformancePoint = {
      timestamp: Date.now(),
      responseTime,
      memoryUsage,
      cpuUsage,
    };

    this.measurements.push(point);

    // Check thresholds
    if (responseTime > this.thresholds.responseTime) {
      this.eventTracker.createEvent('system.warning', {
        message: 'High response time detected',
        responseTime,
      }, this.sessionId);
    }

    if (memoryUsage > this.thresholds.memoryUsage) {
      this.eventTracker.logMemoryWarning(
        (memoryUsage / 1000) * 100,
        memoryUsage,
      );
    }

    if (cpuUsage > this.thresholds.cpuUsage) {
      this.eventTracker.createEvent('system.warning', {
        message: 'High CPU usage detected',
        cpuUsage,
      }, this.sessionId);
    }
  }

  /**
   * Get average response time
   */
  getAverageResponseTime(lastN: number = 100): number {
    const relevant = this.measurements.slice(-lastN);
    if (relevant.length === 0) return 0;

    const sum = relevant.reduce((acc, p) => acc + p.responseTime, 0);
    return sum / relevant.length;
  }

  /**
   * Get peak memory usage
   */
  getPeakMemoryUsage(): number {
    if (this.measurements.length === 0) return 0;
    return Math.max(...this.measurements.map((p) => p.memoryUsage));
  }

  /**
   * Get average memory usage
   */
  getAverageMemoryUsage(lastN: number = 100): number {
    const relevant = this.measurements.slice(-lastN);
    if (relevant.length === 0) return 0;

    const sum = relevant.reduce((acc, p) => acc + p.memoryUsage, 0);
    return sum / relevant.length;
  }

  /**
   * Get peak CPU usage
   */
  getPeakCpuUsage(): number {
    if (this.measurements.length === 0) return 0;
    return Math.max(...this.measurements.map((p) => p.cpuUsage));
  }

  /**
   * Get average CPU usage
   */
  getAverageCpuUsage(lastN: number = 100): number {
    const relevant = this.measurements.slice(-lastN);
    if (relevant.length === 0) return 0;

    const sum = relevant.reduce((acc, p) => acc + p.cpuUsage, 0);
    return sum / relevant.length;
  }

  /**
   * Get performance summary
   */
  getSummary(): PerformanceMetrics {
    return {
      responseTime: this.getAverageResponseTime(),
      tokenInput: 0,
      tokenOutput: 0,
      costUSD: 0,
      cacheHitRate: 0,
      errorRate: 0,
      memoryUsageMB: this.getAverageMemoryUsage(),
      cpuUsagePercent: this.getAverageCpuUsage(),
    };
  }

  /**
   * Set performance threshold
   */
  setThreshold(metric: string, value: number): void {
    if (metric in this.thresholds) {
      this.thresholds[metric] = value;
    }
  }

  /**
   * Get all measurements
   */
  getMeasurements(): PerformancePoint[] {
    return [...this.measurements];
  }

  /**
   * Clear old measurements
   */
  clearOldMeasurements(keepLastN: number = 1000): void {
    if (this.measurements.length > keepLastN) {
      this.measurements = this.measurements.slice(-keepLastN);
    }
  }

  /**
   * Get measurements for time range
   */
  getMeasurementsByTimeRange(startTime: number, endTime: number): PerformancePoint[] {
    return this.measurements.filter(
      (m) => m.timestamp >= startTime && m.timestamp <= endTime,
    );
  }

  /**
   * Calculate percentile
   */
  calculatePercentile(metric: 'responseTime' | 'memoryUsage' | 'cpuUsage', percentile: number): number {
    const values = this.measurements
      .map((p) => p[metric])
      .sort((a, b) => a - b);

    if (values.length === 0) return 0;

    const index = Math.ceil((percentile / 100) * values.length) - 1;
    return values[Math.max(0, index)];
  }

  /**
   * Get performance trend
   */
  getPerformanceTrend(
    metric: 'responseTime' | 'memoryUsage' | 'cpuUsage',
    periodCount: number = 10,
  ): Array<{ timestamp: number; value: number }> {
    if (this.measurements.length < periodCount) {
      return this.measurements.map((p) => ({
        timestamp: p.timestamp,
        value: p[metric],
      }));
    }

    const chunkSize = Math.floor(this.measurements.length / periodCount);
    const trend: Array<{ timestamp: number; value: number }> = [];

    for (let i = 0; i < periodCount; i++) {
      const start = i * chunkSize;
      const end = (i + 1) * chunkSize;
      const chunk = this.measurements.slice(start, end);

      if (chunk.length > 0) {
        const avgValue = chunk.reduce((sum, p) => sum + p[metric], 0) / chunk.length;
        trend.push({
          timestamp: chunk[Math.floor(chunk.length / 2)].timestamp,
          value: avgValue,
        });
      }
    }

    return trend;
  }

  /**
   * Detect performance anomalies
   */
  detectAnomalies(): Array<{
    timestamp: number;
    metric: string;
    value: number;
    baseline: number;
    deviation: number;
  }> {
    const anomalies: Array<{
      timestamp: number;
      metric: string;
      value: number;
      baseline: number;
      deviation: number;
    }> = [];

    if (this.measurements.length < 10) return anomalies;

    const avgResponseTime = this.getAverageResponseTime();
    const avgMemory = this.getAverageMemoryUsage();
    const avgCpu = this.getAverageCpuUsage();

    // Standard deviation threshold (2 sigma)
    const stdThreshold = 2;

    for (const measurement of this.measurements.slice(-20)) {
      // Check response time
      const rtDeviation = Math.abs(measurement.responseTime - avgResponseTime) / avgResponseTime;
      if (rtDeviation > stdThreshold * 0.1) {
        anomalies.push({
          timestamp: measurement.timestamp,
          metric: 'responseTime',
          value: measurement.responseTime,
          baseline: avgResponseTime,
          deviation: rtDeviation,
        });
      }

      // Check memory
      const memDeviation = Math.abs(measurement.memoryUsage - avgMemory) / avgMemory;
      if (memDeviation > stdThreshold * 0.1) {
        anomalies.push({
          timestamp: measurement.timestamp,
          metric: 'memoryUsage',
          value: measurement.memoryUsage,
          baseline: avgMemory,
          deviation: memDeviation,
        });
      }

      // Check CPU
      const cpuDeviation = Math.abs(measurement.cpuUsage - avgCpu) / avgCpu;
      if (cpuDeviation > stdThreshold * 0.1) {
        anomalies.push({
          timestamp: measurement.timestamp,
          metric: 'cpuUsage',
          value: measurement.cpuUsage,
          baseline: avgCpu,
          deviation: cpuDeviation,
        });
      }
    }

    return anomalies;
  }

  /**
   * Reset monitor
   */
  reset(): void {
    this.measurements = [];
    this.startMarkers.clear();
  }
}
