/**
 * Cache Manager Module - Multi-tier Caching & Memory Management
 * Phase 15: Performance Final Polish
 *
 * Provides:
 * - LRU cache implementation
 * - Memory management
 * - Cache invalidation strategies
 * - Performance metrics
 */

/**
 * Cache entry
 */
interface CacheEntry<T> {
  value: T;
  timestamp: number;
  accessCount: number;
  lastAccess: number;
  ttl?: number;
}

/**
 * Cache statistics
 */
export interface CacheStats {
  hits: number;
  misses: number;
  evictions: number;
  hitRate: number;
  memoryUsed: number;
  itemsCount: number;
}

/**
 * LRU Cache implementation
 */
export class LRUCache<K, V> {
  private cache: Map<K, CacheEntry<V>> = new Map();
  private maxSize: number;
  private maxMemory: number;
  private stats = {
    hits: 0,
    misses: 0,
    evictions: 0
  };

  constructor(maxSize: number = 100, maxMemoryMb: number = 50) {
    this.maxSize = maxSize;
    this.maxMemory = maxMemoryMb * 1024 * 1024;
  }

  /**
   * Get value from cache
   */
  get(key: K): V | undefined {
    const entry = this.cache.get(key);

    if (!entry) {
      this.stats.misses++;
      return undefined;
    }

    // Check if expired
    if (entry.ttl && Date.now() > entry.timestamp + entry.ttl) {
      this.cache.delete(key);
      this.stats.misses++;
      return undefined;
    }

    // Update access info
    entry.accessCount++;
    entry.lastAccess = Date.now();

    this.stats.hits++;

    return entry.value;
  }

  /**
   * Set value in cache
   */
  set(key: K, value: V, ttlMs?: number): void {
    // Remove existing entry
    if (this.cache.has(key)) {
      this.cache.delete(key);
    }

    // Check size before adding
    if (this.cache.size >= this.maxSize) {
      this.evictLRU();
    }

    const entry: CacheEntry<V> = {
      value,
      timestamp: Date.now(),
      accessCount: 0,
      lastAccess: Date.now(),
      ttl: ttlMs
    };

    this.cache.set(key, entry);

    // Check memory usage
    this.checkMemory();
  }

  /**
   * Check if key exists
   */
  has(key: K): boolean {
    return this.cache.has(key);
  }

  /**
   * Delete key
   */
  delete(key: K): boolean {
    return this.cache.delete(key);
  }

  /**
   * Clear all entries
   */
  clear(): void {
    this.cache.clear();
  }

  /**
   * Get cache size
   */
  getSize(): number {
    return this.cache.size;
  }

  /**
   * Get all keys
   */
  getKeys(): K[] {
    return Array.from(this.cache.keys());
  }

  /**
   * Get statistics
   */
  getStats(): CacheStats {
    const totalRequests = this.stats.hits + this.stats.misses;

    return {
      hits: this.stats.hits,
      misses: this.stats.misses,
      evictions: this.stats.evictions,
      hitRate: totalRequests > 0 ? this.stats.hits / totalRequests : 0,
      memoryUsed: this.estimateMemoryUsage(),
      itemsCount: this.cache.size
    };
  }

  /**
   * Reset statistics
   */
  resetStats(): void {
    this.stats = { hits: 0, misses: 0, evictions: 0 };
  }

  /**
   * Evict least recently used entry
   */
  private evictLRU(): void {
    let lruKey: K | null = null;
    let lruTime = Infinity;

    for (const [key, entry] of this.cache.entries()) {
      if (entry.lastAccess < lruTime) {
        lruTime = entry.lastAccess;
        lruKey = key;
      }
    }

    if (lruKey !== null) {
      this.cache.delete(lruKey);
      this.stats.evictions++;
    }
  }

  /**
   * Check memory usage and evict if needed
   */
  private checkMemory(): void {
    const memoryUsed = this.estimateMemoryUsage();

    if (memoryUsed > this.maxMemory) {
      // Evict entries until memory is under limit
      while (this.cache.size > 0 && this.estimateMemoryUsage() > this.maxMemory) {
        this.evictLRU();
      }
    }
  }

  /**
   * Estimate memory usage
   */
  private estimateMemoryUsage(): number {
    let total = 0;

    for (const entry of this.cache.values()) {
      const str = JSON.stringify(entry.value);
      total += str.length * 2; // Rough estimate: 2 bytes per character
    }

    return total;
  }
}

/**
 * Multi-tier cache (Memory -> IndexedDB -> Service Worker)
 */
export class MultiTierCache<K, V> {
  private memory: LRUCache<K, V>;
  private indexedDB: IDBDatabase | null = null;
  private storeName: string;
  private eventHandlers: Map<string, Function[]> = new Map();

  constructor(
    storeName: string = 'cache-store',
    memorySize: number = 100,
    memoryMaxMb: number = 50
  ) {
    this.memory = new LRUCache(memorySize, memoryMaxMb);
    this.storeName = storeName;
  }

  /**
   * Initialize IndexedDB
   */
  async initIndexedDB(): Promise<void> {
    return new Promise((resolve, reject) => {
      const request = indexedDB.open('vectora-cache', 1);

      request.onerror = () => reject(request.error);
      request.onsuccess = () => {
        this.indexedDB = request.result;
        resolve();
      };

      request.onupgradeneeded = (event) => {
        const db = (event.target as IDBOpenDBRequest).result;
        if (!db.objectStoreNames.contains(this.storeName)) {
          db.createObjectStore(this.storeName, { keyPath: 'key' });
        }
      };
    });
  }

  /**
   * Get value from cache hierarchy
   */
  async get(key: K): Promise<V | undefined> {
    // Try memory first
    let value = this.memory.get(key);
    if (value !== undefined) {
      this.emit('cache-hit-memory', { key });
      return value;
    }

    // Try IndexedDB
    if (this.indexedDB) {
      value = await this.getFromIndexedDB(key);
      if (value !== undefined) {
        // Move to memory cache
        this.memory.set(key, value);
        this.emit('cache-hit-idb', { key });
        return value;
      }
    }

    this.emit('cache-miss', { key });
    return undefined;
  }

  /**
   * Set value in cache hierarchy
   */
  async set(key: K, value: V, ttlMs?: number): Promise<void> {
    // Set in memory
    this.memory.set(key, value, ttlMs);

    // Set in IndexedDB for persistence
    if (this.indexedDB) {
      await this.setInIndexedDB(key, value, ttlMs);
    }

    this.emit('cache-set', { key });
  }

  /**
   * Get from IndexedDB
   */
  private getFromIndexedDB(key: K): Promise<V | undefined> {
    return new Promise((resolve, reject) => {
      if (!this.indexedDB) {
        resolve(undefined);
        return;
      }

      const transaction = this.indexedDB.transaction([this.storeName], 'readonly');
      const store = transaction.objectStore(this.storeName);
      const request = store.get(String(key));

      request.onerror = () => reject(request.error);
      request.onsuccess = () => {
        const entry = request.result;

        if (!entry) {
          resolve(undefined);
          return;
        }

        // Check TTL
        if (entry.ttl && Date.now() > entry.timestamp + entry.ttl) {
          this.deleteFromIndexedDB(key);
          resolve(undefined);
          return;
        }

        resolve(entry.value);
      };
    });
  }

  /**
   * Set in IndexedDB
   */
  private setInIndexedDB(key: K, value: V, ttlMs?: number): Promise<void> {
    return new Promise((resolve, reject) => {
      if (!this.indexedDB) {
        resolve();
        return;
      }

      const transaction = this.indexedDB.transaction([this.storeName], 'readwrite');
      const store = transaction.objectStore(this.storeName);

      const entry = {
        key: String(key),
        value,
        timestamp: Date.now(),
        ttl: ttlMs
      };

      const request = store.put(entry);

      request.onerror = () => reject(request.error);
      request.onsuccess = () => resolve();
    });
  }

  /**
   * Delete from IndexedDB
   */
  private deleteFromIndexedDB(key: K): Promise<void> {
    return new Promise((resolve, reject) => {
      if (!this.indexedDB) {
        resolve();
        return;
      }

      const transaction = this.indexedDB.transaction([this.storeName], 'readwrite');
      const store = transaction.objectStore(this.storeName);
      const request = store.delete(String(key));

      request.onerror = () => reject(request.error);
      request.onsuccess = () => resolve();
    });
  }

  /**
   * Clear all caches
   */
  async clearAll(): Promise<void> {
    this.memory.clear();

    if (this.indexedDB) {
      return new Promise((resolve, reject) => {
        const transaction = this.indexedDB!.transaction([this.storeName], 'readwrite');
        const store = transaction.objectStore(this.storeName);
        const request = store.clear();

        request.onerror = () => reject(request.error);
        request.onsuccess = () => resolve();
      });
    }
  }

  /**
   * Get memory stats
   */
  getMemoryStats(): CacheStats {
    return this.memory.getStats();
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

export default LRUCache;
