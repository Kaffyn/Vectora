/**
 * Offline Mode Module - Index (Barrel Export)
 * Phase 13: Offline Mode & PWA
 *
 * Exports all offline mode utilities and services
 */

// Service Worker & Cache Management
export {
  ServiceWorkerManager,
  CacheManager,
  type ServiceWorkerConfig,
  type CacheStrategy,
  type ServiceWorkerInfo
} from './serviceWorker';

// Request Queuing
export {
  OfflineRequestQueue,
  type QueuedRequest,
  type QueueResult
} from './requestQueue';

// Synchronization & Conflict Resolution
export {
  SyncEngine,
  OfflineStateReconciliator,
  type SyncEntry,
  type Conflict,
  type SyncResult,
  type ConflictStrategy
} from './syncEngine';

// Offline Manager (Main Orchestrator)
export {
  OfflineManager,
  getOfflineManager,
  type OfflineConfig,
  type OfflineState
} from './offlineManager';

export default {
  ServiceWorkerManager,
  CacheManager,
  OfflineRequestQueue,
  SyncEngine,
  OfflineStateReconciliator,
  OfflineManager
};
