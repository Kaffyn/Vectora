/**
 * Offline Manager - Unified Offline Mode Orchestration
 * Phase 13: Offline Mode & PWA
 *
 * Provides:
 * - Offline/online state management
 * - Request queuing and processing
 * - Data synchronization
 * - Local storage management
 * - Offline indicators
 */

import { ServiceWorkerManager, CacheManager } from './serviceWorker';
import { OfflineRequestQueue, type QueuedRequest } from './requestQueue';
import { SyncEngine, type SyncEntry } from './syncEngine';

/**
 * Offline configuration
 */
export interface OfflineConfig {
  enableOfflineMode?: boolean;
  enableServiceWorker?: boolean;
  enableRequestQueuing?: boolean;
  enableSync?: boolean;
  storageQuota?: number; // in bytes
  maxQueueSize?: number;
  cacheStrategy?: 'network-first' | 'cache-first' | 'stale-while-revalidate';
}

/**
 * Offline state
 */
export interface OfflineState {
  isOnline: boolean;
  queuedRequests: number;
  failedRequests: number;
  pendingSync: number;
  storageUsed: number;
  storageQuota: number;
  lastSyncTime?: number;
}

/**
 * OfflineManager - Main offline mode orchestrator
 */
export class OfflineManager {
  private serviceWorkerManager: ServiceWorkerManager | null = null;
  private cacheManager: CacheManager | null = null;
  private requestQueue: OfflineRequestQueue;
  private syncEngine: SyncEngine<any> | null = null;

  private config: Required<OfflineConfig>;
  private isOnline: boolean = navigator.onLine;
  private eventHandlers: Map<string, Function[]> = new Map();

  private constructor(config?: OfflineConfig) {
    this.config = {
      enableOfflineMode: config?.enableOfflineMode ?? true,
      enableServiceWorker: config?.enableServiceWorker ?? true,
      enableRequestQueuing: config?.enableRequestQueuing ?? true,
      enableSync: config?.enableSync ?? true,
      storageQuota: config?.storageQuota ?? 50 * 1024 * 1024, // 50MB
      maxQueueSize: config?.maxQueueSize ?? 1000,
      cacheStrategy: config?.cacheStrategy ?? 'network-first'
    };

    this.requestQueue = new OfflineRequestQueue(this.config.maxQueueSize);

    // Track online/offline status
    window.addEventListener('online', () => this.handleOnline());
    window.addEventListener('offline', () => this.handleOffline());

    // Setup event forwarding
    this.requestQueue.on('online', () => this.emit('online', null));
    this.requestQueue.on('offline', () => this.emit('offline', null));
  }

  /**
   * Initialize offline manager
   */
  static async create(config?: OfflineConfig): Promise<OfflineManager> {
    const manager = new OfflineManager(config);
    await manager.initialize();
    return manager;
  }

  /**
   * Initialize all components
   */
  private async initialize(): Promise<void> {
    try {
      // Initialize service worker
      if (this.config.enableServiceWorker) {
        this.serviceWorkerManager = new ServiceWorkerManager({
          scriptPath: '/sw.js',
          cacheVersion: 1,
          autoUpdate: true
        });
        await this.serviceWorkerManager.register();
        this.emit('service-worker-ready', null);
      }

      // Initialize cache manager
      if (this.config.enableServiceWorker) {
        this.cacheManager = new CacheManager('vectora-cache', 1);
        this.emit('cache-manager-ready', null);
      }

      // Initialize sync engine
      if (this.config.enableSync) {
        const userId = this.getUserId();
        this.syncEngine = new SyncEngine(userId, 'merge');
        this.emit('sync-engine-ready', null);
      }

      this.emit('initialized', null);
    } catch (error) {
      console.error('Failed to initialize offline manager:', error);
      this.emit('initialization-error', error);
    }
  }

  /**
   * Enqueue API request for later processing
   */
  enqueueRequest(
    url: string,
    options?: {
      method?: 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH';
      headers?: Record<string, string>;
      body?: any;
      priority?: number;
      maxRetries?: number;
      tags?: string[];
    }
  ): string {
    if (!this.config.enableRequestQueuing) {
      throw new Error('Request queuing is disabled');
    }

    const id = this.requestQueue.enqueueRequest(url, options);
    this.emit('request-queued', { id, url });

    return id;
  }

  /**
   * Process queued requests (when online)
   */
  async processQueue(): Promise<any[]> {
    if (!this.isOnline || !this.config.enableRequestQueuing) {
      return [];
    }

    this.emit('processing-queue', null);
    const results = await this.requestQueue.processQueue();
    this.emit('queue-processed', { count: results.length });

    return results;
  }

  /**
   * Synchronize data with server
   */
  async syncData(localEntries: SyncEntry<any>[]): Promise<void> {
    if (!this.isOnline || !this.syncEngine) {
      return;
    }

    this.emit('sync-started', null);

    try {
      // Add local entries
      for (const entry of localEntries) {
        this.syncEngine.setLocal(entry.key, entry.value);
      }

      // Perform sync
      const result = await this.syncEngine.sync();

      this.emit('sync-completed', { result, timestamp: Date.now() });
    } catch (error) {
      console.error('Sync failed:', error);
      this.emit('sync-error', error);
    }
  }

  /**
   * Get offline state
   */
  getOfflineState(): OfflineState {
    const queueStatus = this.requestQueue.getStatus();

    return {
      isOnline: this.isOnline,
      queuedRequests: queueStatus.pending,
      failedRequests: queueStatus.failed,
      pendingSync: this.syncEngine?.getConflicts().length || 0,
      storageUsed: this.getStorageUsed(),
      storageQuota: this.config.storageQuota,
      lastSyncTime: undefined
    };
  }

  /**
   * Cache response
   */
  async cacheResponse(request: Request | string, response: Response): Promise<void> {
    if (!this.cacheManager) {
      return;
    }

    await this.cacheManager.cacheResponse(request, response);
  }

  /**
   * Get cached response
   */
  async getCachedResponse(request: Request | string): Promise<Response | null> {
    if (!this.cacheManager) {
      return null;
    }

    return this.cacheManager.getCachedResponse(request);
  }

  /**
   * Precache URLs
   */
  async precacheUrls(urls: string[]): Promise<void> {
    if (!this.cacheManager) {
      return;
    }

    await this.cacheManager.precache(urls);
    this.emit('urls-precached', { count: urls.length });
  }

  /**
   * Clear cache
   */
  async clearCache(): Promise<void> {
    if (!this.cacheManager) {
      return;
    }

    await this.cacheManager.clearCache();
    this.emit('cache-cleared', null);
  }

  /**
   * Get cached URLs
   */
  async getCachedUrls(): Promise<string[]> {
    if (!this.cacheManager) {
      return [];
    }

    return this.cacheManager.getCachedUrls();
  }

  /**
   * Get cache size
   */
  async getCacheSize(): Promise<number> {
    if (!this.cacheManager) {
      return 0;
    }

    return this.cacheManager.getCacheSize();
  }

  /**
   * Get pending requests
   */
  getPendingRequests(): QueuedRequest[] {
    return this.requestQueue.getPendingRequests();
  }

  /**
   * Get failed requests
   */
  getFailedRequests(): QueuedRequest[] {
    return this.requestQueue.getFailedRequests();
  }

  /**
   * Retry failed requests
   */
  retryFailed(): void {
    this.requestQueue.retryFailed();
    this.emit('failed-requests-retried', null);
  }

  /**
   * Clear offline queue
   */
  clearQueue(): void {
    this.requestQueue.clearQueue();
    this.emit('queue-cleared', null);
  }

  /**
   * Get service worker info
   */
  getServiceWorkerInfo() {
    if (!this.serviceWorkerManager) {
      return null;
    }

    return this.serviceWorkerManager.getInfo();
  }

  /**
   * Check for service worker updates
   */
  async checkForUpdates(): Promise<boolean> {
    if (!this.serviceWorkerManager) {
      return false;
    }

    return this.serviceWorkerManager.checkForUpdates();
  }

  /**
   * Update service worker
   */
  updateServiceWorker(): void {
    if (!this.serviceWorkerManager) {
      return;
    }

    this.serviceWorkerManager.skipWaiting();
    this.emit('service-worker-updating', null);
  }

  /**
   * Get online status
   */
  isOnlineStatus(): boolean {
    return this.isOnline;
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
   * Handle online event
   */
  private handleOnline(): void {
    this.isOnline = true;
    this.emit('online', null);

    // Process queue when coming online
    this.processQueue();
  }

  /**
   * Handle offline event
   */
  private handleOffline(): void {
    this.isOnline = false;
    this.emit('offline', null);
  }

  /**
   * Get storage used
   */
  private getStorageUsed(): number {
    let total = 0;

    // Estimate localStorage usage
    try {
      for (const key in localStorage) {
        if (localStorage.hasOwnProperty(key)) {
          total += localStorage[key].length + key.length;
        }
      }
    } catch (error) {
      console.warn('Could not calculate storage usage:', error);
    }

    return total;
  }

  /**
   * Get user ID (for sync)
   */
  private getUserId(): string {
    let userId = localStorage.getItem('vectora-user-id');
    if (!userId) {
      userId = `user-${Math.random().toString(36).substr(2, 9)}`;
      localStorage.setItem('vectora-user-id', userId);
    }
    return userId;
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

  /**
   * Cleanup and shutdown
   */
  async shutdown(): Promise<void> {
    if (this.serviceWorkerManager) {
      await this.serviceWorkerManager.unregister();
    }
    this.requestQueue.clearQueue();
    this.emit('shutdown', null);
  }
}

// Singleton instance
let offlineManagerInstance: OfflineManager | null = null;

/**
 * Get or create offline manager instance
 */
export async function getOfflineManager(config?: OfflineConfig): Promise<OfflineManager> {
  if (!offlineManagerInstance) {
    offlineManagerInstance = await OfflineManager.create(config);
  }
  return offlineManagerInstance;
}

export default OfflineManager;
