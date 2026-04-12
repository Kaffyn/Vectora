/**
 * Service Worker Module - Progressive Web App Support
 * Phase 13: Offline Mode & PWA
 *
 * Provides:
 * - Service worker registration
 * - Cache management strategies
 * - Offline detection
 * - Background sync support
 * - Push notification handling
 */

/**
 * Service worker configuration
 */
export interface ServiceWorkerConfig {
  scriptPath?: string;
  scope?: string;
  updateCheckInterval?: number;
  autoUpdate?: boolean;
  cacheName?: string;
  cacheVersion?: number;
}

/**
 * Cache strategy type
 */
export type CacheStrategy = 'network-first' | 'cache-first' | 'stale-while-revalidate';

/**
 * Service worker instance information
 */
export interface ServiceWorkerInfo {
  registration: ServiceWorkerRegistration;
  active: ServiceWorker | null;
  installing: ServiceWorker | null;
  waiting: ServiceWorker | null;
  scope: string;
  updateTime: number;
}

/**
 * ServiceWorkerManager - Manages service worker lifecycle
 */
export class ServiceWorkerManager {
  private config: Required<ServiceWorkerConfig>;
  private registration: ServiceWorkerRegistration | null = null;
  private isOnline: boolean = navigator.onLine;
  private eventHandlers: Map<string, Function[]> = new Map();
  private updateCheckInterval: NodeJS.Timer | null = null;

  constructor(config?: ServiceWorkerConfig) {
    this.config = {
      scriptPath: config?.scriptPath || '/sw.js',
      scope: config?.scope || '/',
      updateCheckInterval: config?.updateCheckInterval || 60000, // 1 minute
      autoUpdate: config?.autoUpdate ?? true,
      cacheName: config?.cacheName || 'vectora-cache',
      cacheVersion: config?.cacheVersion || 1
    };

    // Track online/offline status
    window.addEventListener('online', () => this.handleOnline());
    window.addEventListener('offline', () => this.handleOffline());
  }

  /**
   * Register service worker
   */
  async register(): Promise<ServiceWorkerRegistration | null> {
    try {
      if (!('serviceWorker' in navigator)) {
        console.warn('Service Workers not supported');
        return null;
      }

      this.registration = await navigator.serviceWorker.register(this.config.scriptPath, {
        scope: this.config.scope
      });

      console.log('Service Worker registered:', this.registration.scope);

      // Setup update checking
      if (this.config.autoUpdate) {
        this.startUpdateCheck();
      }

      // Handle updates
      this.registration.addEventListener('updatefound', () => this.handleUpdateFound());

      // Listen for messages from SW
      navigator.serviceWorker.addEventListener('message', (event) => {
        this.handleMessage(event.data);
      });

      this.emit('registered', this.registration);
      return this.registration;
    } catch (error) {
      console.error('Service Worker registration failed:', error);
      this.emit('registration-failed', error);
      return null;
    }
  }

  /**
   * Unregister service worker
   */
  async unregister(): Promise<boolean> {
    try {
      if (!this.registration) {
        return false;
      }

      const success = await this.registration.unregister();
      if (success) {
        this.registration = null;
        this.stopUpdateCheck();
        this.emit('unregistered', null);
      }
      return success;
    } catch (error) {
      console.error('Service Worker unregister failed:', error);
      return false;
    }
  }

  /**
   * Check for updates
   */
  async checkForUpdates(): Promise<boolean> {
    try {
      if (!this.registration) {
        return false;
      }

      await this.registration.update();
      return true;
    } catch (error) {
      console.error('Update check failed:', error);
      return false;
    }
  }

  /**
   * Skip waiting and activate new SW
   */
  skipWaiting(): void {
    if (!this.registration?.waiting) {
      return;
    }

    this.registration.waiting.postMessage({ type: 'SKIP_WAITING' });
  }

  /**
   * Claim clients (activate new SW immediately)
   */
  async claimClients(): Promise<void> {
    if (!this.registration?.active) {
      return;
    }

    this.registration.active.postMessage({ type: 'CLIENTS_CLAIM' });
  }

  /**
   * Get service worker info
   */
  getInfo(): ServiceWorkerInfo | null {
    if (!this.registration) {
      return null;
    }

    return {
      registration: this.registration,
      active: this.registration.active,
      installing: this.registration.installing,
      waiting: this.registration.waiting,
      scope: this.registration.scope,
      updateTime: Date.now()
    };
  }

  /**
   * Post message to service worker
   */
  postMessage(data: any): void {
    if (!this.registration?.active) {
      return;
    }

    this.registration.active.postMessage(data);
  }

  /**
   * Check online status
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
   * Start periodic update check
   */
  private startUpdateCheck(): void {
    this.updateCheckInterval = setInterval(() => {
      this.checkForUpdates();
    }, this.config.updateCheckInterval);
  }

  /**
   * Stop periodic update check
   */
  private stopUpdateCheck(): void {
    if (this.updateCheckInterval) {
      clearInterval(this.updateCheckInterval);
      this.updateCheckInterval = null;
    }
  }

  /**
   * Handle online event
   */
  private handleOnline(): void {
    this.isOnline = true;
    this.emit('online', null);
  }

  /**
   * Handle offline event
   */
  private handleOffline(): void {
    this.isOnline = false;
    this.emit('offline', null);
  }

  /**
   * Handle update found
   */
  private handleUpdateFound(): void {
    this.emit('update-found', this.registration?.waiting);
  }

  /**
   * Handle message from service worker
   */
  private handleMessage(data: any): void {
    const { type, payload } = data;
    this.emit(`message:${type}`, payload);
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

/**
 * Cache manager for offline support
 */
export class CacheManager {
  private cacheName: string;
  private cacheVersion: number;

  constructor(cacheName: string = 'vectora-cache', cacheVersion: number = 1) {
    this.cacheName = `${cacheName}-v${cacheVersion}`;
    this.cacheVersion = cacheVersion;
  }

  /**
   * Open cache
   */
  async openCache(): Promise<Cache> {
    return caches.open(this.cacheName);
  }

  /**
   * Cache a request/response pair
   */
  async cacheResponse(request: Request | string, response: Response): Promise<void> {
    const cache = await this.openCache();
    await cache.put(request, response);
  }

  /**
   * Get cached response
   */
  async getCachedResponse(request: Request | string): Promise<Response | null> {
    const cache = await this.openCache();
    return cache.match(request) || null;
  }

  /**
   * Delete from cache
   */
  async deleteFromCache(request: Request | string): Promise<boolean> {
    const cache = await this.openCache();
    return cache.delete(request);
  }

  /**
   * Clear entire cache
   */
  async clearCache(): Promise<void> {
    const cache = await this.openCache();
    const requests = await cache.keys();
    for (const request of requests) {
      await cache.delete(request);
    }
  }

  /**
   * Get all cached URLs
   */
  async getCachedUrls(): Promise<string[]> {
    const cache = await this.openCache();
    const requests = await cache.keys();
    return requests.map(r => r.url);
  }

  /**
   * Get cache size
   */
  async getCacheSize(): Promise<number> {
    const cache = await this.openCache();
    const requests = await cache.keys();
    let size = 0;

    for (const request of requests) {
      const response = await cache.match(request);
      if (response) {
        const blob = await response.blob();
        size += blob.size;
      }
    }

    return size;
  }

  /**
   * Clean old cache versions
   */
  async cleanOldVersions(): Promise<void> {
    const cacheNames = await caches.keys();
    for (const name of cacheNames) {
      if (!name.includes(String(this.cacheVersion))) {
        await caches.delete(name);
      }
    }
  }

  /**
   * Precache specific URLs
   */
  async precache(urls: string[]): Promise<void> {
    const cache = await this.openCache();

    for (const url of urls) {
      try {
        const response = await fetch(url);
        if (response.ok) {
          await cache.put(url, response);
        }
      } catch (error) {
        console.warn(`Failed to precache ${url}:`, error);
      }
    }
  }

  /**
   * Match cache with strategy
   */
  async match(request: Request, strategy: CacheStrategy): Promise<Response | null> {
    switch (strategy) {
      case 'cache-first':
        return this.cacheFirst(request);
      case 'network-first':
        return this.networkFirst(request);
      case 'stale-while-revalidate':
        return this.staleWhileRevalidate(request);
      default:
        return null;
    }
  }

  /**
   * Cache first strategy
   */
  private async cacheFirst(request: Request): Promise<Response | null> {
    const cached = await this.getCachedResponse(request);
    if (cached) {
      return cached;
    }

    try {
      const response = await fetch(request);
      if (response.ok) {
        await this.cacheResponse(request, response.clone());
      }
      return response;
    } catch (error) {
      return null;
    }
  }

  /**
   * Network first strategy
   */
  private async networkFirst(request: Request): Promise<Response | null> {
    try {
      const response = await fetch(request);
      if (response.ok) {
        await this.cacheResponse(request, response.clone());
      }
      return response;
    } catch (error) {
      return this.getCachedResponse(request);
    }
  }

  /**
   * Stale while revalidate strategy
   */
  private async staleWhileRevalidate(request: Request): Promise<Response | null> {
    const cached = await this.getCachedResponse(request);

    const fetchPromise = fetch(request)
      .then(response => {
        if (response.ok) {
          this.cacheResponse(request, response.clone());
        }
        return response;
      })
      .catch(() => null);

    return cached || fetchPromise;
  }
}

export default ServiceWorkerManager;
