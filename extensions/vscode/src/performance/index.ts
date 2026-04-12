/**
 * Performance Module - Index (Barrel Export)
 * Phase 15: Performance Final Polish
 *
 * Exports all performance utilities and services
 */

// Code Splitting & Lazy Loading
export {
  CodeSplitter,
  LazyComponentLoader,
  RouteCodeSplitter,
  type ModuleMetadata,
  type SplitStrategy
} from './codeSpitting';

// Caching & Memory Management
export {
  LRUCache,
  MultiTierCache,
  type CacheStats
} from './cacheManager';

// Performance Monitoring
export {
  PerformanceMonitor,
  getPerformanceMonitor,
  type PerformanceMetric,
  type WebVitals,
  type PerformanceBudget
} from './performanceMonitor';

export default {
  CodeSplitter,
  LazyComponentLoader,
  RouteCodeSplitter,
  LRUCache,
  MultiTierCache,
  PerformanceMonitor
};
