/**
 * Performance Monitor Module - Real-time Performance Metrics
 * Phase 15: Performance Final Polish
 *
 * Provides:
 * - Web Vitals tracking (LCP, FID, CLS)
 * - Performance metrics collection
 * - Resource timing analysis
 * - Memory profiling
 * - Performance budgets
 */

/**
 * Performance metric
 */
export interface PerformanceMetric {
  name: string;
  value: number;
  unit: string;
  timestamp: number;
  threshold?: number;
}

/**
 * Web Vitals
 */
export interface WebVitals {
  lcp?: number; // Largest Contentful Paint
  fid?: number; // First Input Delay
  cls?: number; // Cumulative Layout Shift
  ttfb?: number; // Time to First Byte
  fcp?: number; // First Contentful Paint
}

/**
 * Performance budget
 */
export interface PerformanceBudget {
  metric: string;
  limit: number;
  unit: string;
}

/**
 * Performance monitor
 */
export class PerformanceMonitor {
  private metrics: Map<string, PerformanceMetric[]> = new Map();
  private webVitals: WebVitals = {};
  private budgets: Map<string, PerformanceBudget> = new Map();
  private observers: PerformanceObserver[] = [];
  private eventHandlers: Map<string, Function[]> = new Map();
  private startTime: number = Date.now();

  constructor() {
    this.initializeObservers();
  }

  /**
   * Initialize performance observers
   */
  private initializeObservers(): void {
    // Observe Largest Contentful Paint
    if ('PerformanceObserver' in window) {
      try {
        const paintObserver = new PerformanceObserver((list) => {
          for (const entry of list.getEntries()) {
            if (entry.name === 'first-contentful-paint') {
              this.webVitals.fcp = entry.startTime;
              this.recordMetric('fcp', entry.startTime, 'ms');
            } else if (entry.name === 'largest-contentful-paint') {
              this.webVitals.lcp = entry.startTime;
              this.recordMetric('lcp', entry.startTime, 'ms');
              this.emit('web-vital-lcp', { lcp: entry.startTime });
            }
          }
        });

        paintObserver.observe({
          entryTypes: ['paint', 'largest-contentful-paint']
        });

        this.observers.push(paintObserver);
      } catch (e) {
        // Browser doesn't support
      }
    }

    // Observe First Input Delay
    if ('PerformanceEventTiming' in window) {
      try {
        const fidObserver = new PerformanceObserver((list) => {
          for (const entry of list.getEntries()) {
            if (entry.interactionId) {
              const fid = (entry as any).processingStart - entry.startTime;
              this.webVitals.fid = fid;
              this.recordMetric('fid', fid, 'ms');
              this.emit('web-vital-fid', { fid });
            }
          }
        });

        fidObserver.observe({
          entryTypes: ['first-input', 'event']
        });

        this.observers.push(fidObserver);
      } catch (e) {
        // Browser doesn't support
      }
    }

    // Observe Cumulative Layout Shift
    if ('PerformanceObserver' in window) {
      try {
        const clsObserver = new PerformanceObserver((list) => {
          for (const entry of list.getEntries()) {
            if (!(entry as any).hadRecentInput) {
              const cls = (entry as any).value;
              this.webVitals.cls = (this.webVitals.cls || 0) + cls;
              this.emit('web-vital-cls', { cls: this.webVitals.cls });
            }
          }
        });

        clsObserver.observe({
          entryTypes: ['layout-shift']
        });

        this.observers.push(clsObserver);
      } catch (e) {
        // Browser doesn't support
      }
    }

    // Record TTFB (from navigation timing)
    if (performance.timing) {
      const ttfb = performance.timing.responseStart - performance.timing.navigationStart;
      this.webVitals.ttfb = ttfb;
      this.recordMetric('ttfb', ttfb, 'ms');
    }
  }

  /**
   * Record custom metric
   */
  recordMetric(name: string, value: number, unit: string = 'ms'): void {
    const metric: PerformanceMetric = {
      name,
      value,
      unit,
      timestamp: Date.now(),
      threshold: this.budgets.get(name)?.limit
    };

    if (!this.metrics.has(name)) {
      this.metrics.set(name, []);
    }

    this.metrics.get(name)!.push(metric);

    // Check budget
    const budget = this.budgets.get(name);
    if (budget && value > budget.limit) {
      this.emit('budget-exceeded', { metric: name, value, limit: budget.limit });
    }

    this.emit('metric-recorded', metric);
  }

  /**
   * Measure operation
   */
  measureOperation(name: string, fn: () => void): number {
    const startTime = performance.now();
    fn();
    const duration = performance.now() - startTime;

    this.recordMetric(name, duration, 'ms');

    return duration;
  }

  /**
   * Measure async operation
   */
  async measureAsyncOperation<T>(name: string, fn: () => Promise<T>): Promise<T> {
    const startTime = performance.now();
    const result = await fn();
    const duration = performance.now() - startTime;

    this.recordMetric(name, duration, 'ms');

    return result;
  }

  /**
   * Get web vitals
   */
  getWebVitals(): WebVitals {
    return { ...this.webVitals };
  }

  /**
   * Get metric history
   */
  getMetricHistory(name: string): PerformanceMetric[] {
    return this.metrics.get(name) || [];
  }

  /**
   * Get average metric
   */
  getAverageMetric(name: string): number | null {
    const metrics = this.metrics.get(name);

    if (!metrics || metrics.length === 0) {
      return null;
    }

    const sum = metrics.reduce((acc, m) => acc + m.value, 0);
    return sum / metrics.length;
  }

  /**
   * Set performance budget
   */
  setBudget(metric: string, limit: number, unit: string = 'ms'): void {
    this.budgets.set(metric, { metric, limit, unit });
  }

  /**
   * Check if budget is exceeded
   */
  isBudgetExceeded(metric: string): boolean {
    const budget = this.budgets.get(metric);
    if (!budget) {
      return false;
    }

    const average = this.getAverageMetric(metric);
    return average !== null && average > budget.limit;
  }

  /**
   * Get all budgets
   */
  getBudgets(): PerformanceBudget[] {
    return Array.from(this.budgets.values());
  }

  /**
   * Get performance report
   */
  getPerformanceReport() {
    const report = {
      webVitals: this.getWebVitals(),
      metrics: {} as Record<string, { average: number; count: number; min: number; max: number }>,
      budgets: Array.from(this.budgets.values()),
      budgetStatus: {} as Record<string, boolean>,
      uptime: Date.now() - this.startTime
    };

    for (const [name, metrics] of this.metrics.entries()) {
      if (metrics.length > 0) {
        const values = metrics.map(m => m.value);
        report.metrics[name] = {
          average: values.reduce((a, b) => a + b) / values.length,
          count: values.length,
          min: Math.min(...values),
          max: Math.max(...values)
        };

        report.budgetStatus[name] = !this.isBudgetExceeded(name);
      }
    }

    return report;
  }

  /**
   * Get memory metrics (if available)
   */
  getMemoryMetrics(): any {
    if (!performance.memory) {
      return null;
    }

    return {
      usedJSHeapSize: performance.memory.usedJSHeapSize,
      totalJSHeapSize: performance.memory.totalJSHeapSize,
      jsHeapSizeLimit: performance.memory.jsHeapSizeLimit,
      usagePercentage: (performance.memory.usedJSHeapSize / performance.memory.jsHeapSizeLimit) * 100
    };
  }

  /**
   * Get resource timing
   */
  getResourceTiming(resourceName?: string): any[] {
    if (!performance.getEntriesByType) {
      return [];
    }

    let resources = performance.getEntriesByType('resource') as PerformanceResourceTiming[];

    if (resourceName) {
      resources = resources.filter(r => r.name.includes(resourceName));
    }

    return resources.map(r => ({
      name: r.name,
      duration: r.duration,
      transferSize: r.transferSize,
      decodedBodySize: r.decodedBodySize,
      serverTiming: (r as any).serverTiming
    }));
  }

  /**
   * Clear metrics
   */
  clearMetrics(): void {
    this.metrics.clear();
  }

  /**
   * Register event handler
   */
  on(event: string, handler: Function): void {
    if (!this.eventHandlers.has(event)) {
      this.eventHandlers.set(event, []);
    }
    this.eventHandlers.get(event)!.push(handler);
  }

  /**
   * Unregister event handler
   */
  off(event: string, handler: Function): void {
    const handlers = this.eventHandlers.get(event);
    if (handlers) {
      const index = handlers.indexOf(handler);
      if (index > -1) {
        handlers.splice(index, 1);
      }
    }
  }

  /**
   * Cleanup
   */
  cleanup(): void {
    for (const observer of this.observers) {
      observer.disconnect();
    }
    this.observers = [];
  }

  /**
   * Emit event
   */
  private emit(event: string, data: any): void {
    const handlers = this.eventHandlers.get(event);
    if (handlers) {
      for (const handler of handlers) {
        try {
          handler(data);
        } catch (error) {
          console.error(`Error in event handler for ${event}:`, error);
        }
      }
    }
  }
}

// Singleton instance
let performanceMonitorInstance: PerformanceMonitor | null = null;

/**
 * Get or create performance monitor instance
 */
export function getPerformanceMonitor(): PerformanceMonitor {
  if (!performanceMonitorInstance) {
    performanceMonitorInstance = new PerformanceMonitor();
  }
  return performanceMonitorInstance;
}

export default PerformanceMonitor;
