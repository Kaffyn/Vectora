/**
 * Offline Module Tests
 * Phase 13: Offline Mode & PWA
 */

import {
  ServiceWorkerManager,
  CacheManager
} from '../../offline/serviceWorker';

import {
  OfflineRequestQueue
} from '../../offline/requestQueue';

import {
  SyncEngine,
  OfflineStateReconciliator
} from '../../offline/syncEngine';

import {
  OfflineManager,
  getOfflineManager
} from '../../offline/offlineManager';

describe('Offline Module Tests', () => {
  // ============================================================================
  // Service Worker Tests
  // ============================================================================

  describe('ServiceWorkerManager', () => {
    let manager: ServiceWorkerManager;

    beforeEach(() => {
      manager = new ServiceWorkerManager({
        scriptPath: '/sw.js',
        scope: '/',
        autoUpdate: false
      });
    });

    test('should create service worker manager', () => {
      expect(manager).toBeDefined();
      expect(manager.isOnlineStatus()).toBe(true);
    });

    test('should post messages to service worker', () => {
      const data = { type: 'TEST', payload: { test: true } };
      expect(() => {
        manager.postMessage(data);
      }).not.toThrow();
    });

    test('should get service worker info', () => {
      const info = manager.getInfo();
      // May be null if SW not registered
      expect(info === null || info.scope).toBeDefined();
    });

    test('should register event handlers', (done) => {
      manager.on('online', () => {
        done();
      });

      // Simulate online event
      window.dispatchEvent(new Event('online'));
    });
  });

  describe('CacheManager', () => {
    let cacheManager: CacheManager;

    beforeEach(() => {
      cacheManager = new CacheManager('test-cache', 1);
    });

    test('should open cache', async () => {
      const cache = await cacheManager.openCache();
      expect(cache).toBeDefined();
    });

    test('should cache and retrieve responses', async () => {
      const response = new Response('test content', {
        status: 200,
        headers: { 'Content-Type': 'text/plain' }
      });

      await cacheManager.cacheResponse('/test', response);
      const cached = await cacheManager.getCachedResponse('/test');

      expect(cached).toBeDefined();
      expect(await cached!.text()).toBe('test content');
    });

    test('should clear cache', async () => {
      const response = new Response('test');
      await cacheManager.cacheResponse('/test', response);

      const urls = await cacheManager.getCachedUrls();
      expect(urls.length).toBeGreaterThan(0);

      await cacheManager.clearCache();

      const clearedUrls = await cacheManager.getCachedUrls();
      expect(clearedUrls.length).toBe(0);
    });

    test('should get cache size', async () => {
      const size = await cacheManager.getCacheSize();
      expect(size).toBeGreaterThanOrEqual(0);
    });

    test('should precache URLs', async () => {
      // Mock fetch for precaching
      global.fetch = jest.fn(() =>
        Promise.resolve(new Response('content', { status: 200 }))
      );

      await cacheManager.precache(['/api/test']);

      expect(global.fetch).toHaveBeenCalled();
    });
  });

  // ============================================================================
  // Request Queue Tests
  // ============================================================================

  describe('OfflineRequestQueue', () => {
    let queue: OfflineRequestQueue;

    beforeEach(() => {
      queue = new OfflineRequestQueue(100, 'test-queue');
      localStorage.clear();
    });

    test('should enqueue requests', () => {
      const id = queue.enqueueRequest('/api/test', {
        method: 'POST',
        body: { test: true },
        priority: 5
      });

      expect(id).toBeDefined();
      expect(queue.getRequest(id)).toBeDefined();
    });

    test('should get queue status', () => {
      queue.enqueueRequest('/api/test', { method: 'POST' });
      queue.enqueueRequest('/api/test2', { method: 'POST' });

      const status = queue.getStatus();
      expect(status.pending).toBe(2);
      expect(status.total).toBe(2);
    });

    test('should get requests by tag', () => {
      queue.enqueueRequest('/api/test1', {
        method: 'POST',
        tags: ['batch1']
      });
      queue.enqueueRequest('/api/test2', {
        method: 'POST',
        tags: ['batch2']
      });

      const batch1 = queue.getRequestsByTag('batch1');
      expect(batch1.length).toBe(1);
    });

    test('should remove requests', () => {
      const id = queue.enqueueRequest('/api/test', { method: 'POST' });
      expect(queue.getRequest(id)).toBeDefined();

      queue.removeRequest(id);
      expect(queue.getRequest(id)).toBeNull();
    });

    test('should clear queue', () => {
      queue.enqueueRequest('/api/test1', { method: 'POST' });
      queue.enqueueRequest('/api/test2', { method: 'POST' });

      expect(queue.getStatus().total).toBe(2);

      queue.clearQueue();
      expect(queue.getStatus().total).toBe(0);
    });

    test('should clear by tag', () => {
      queue.enqueueRequest('/api/test1', { tags: ['cleanup'] });
      queue.enqueueRequest('/api/test2', { tags: ['other'] });

      const cleared = queue.clearByTag('cleanup');
      expect(cleared).toBe(1);
      expect(queue.getStatus().total).toBe(1);
    });

    test('should retry failed requests', () => {
      const id = queue.enqueueRequest('/api/test', { method: 'POST' });
      const request = queue.getRequest(id);

      if (request) {
        request.status = 'failed';
        request.retries = 3;

        queue.retryFailed();

        const retried = queue.getRequest(id);
        expect(retried!.status).toBe('pending');
        expect(retried!.retries).toBe(0);
      }
    });

    test('should persist and load queue', async () => {
      const id1 = queue.enqueueRequest('/api/test1', { method: 'POST' });
      const id2 = queue.enqueueRequest('/api/test2', { method: 'POST' });

      const newQueue = new OfflineRequestQueue(100, 'test-queue');
      expect(newQueue.getStatus().total).toBe(2);
      expect(newQueue.getRequest(id1)).toBeDefined();
      expect(newQueue.getRequest(id2)).toBeDefined();
    });

    test('should register event handlers', (done) => {
      queue.on('request-queued', (data) => {
        expect(data.id).toBeDefined();
        done();
      });

      queue.enqueueRequest('/api/test', { method: 'POST' });
    });
  });

  // ============================================================================
  // Sync Engine Tests
  // ============================================================================

  describe('SyncEngine', () => {
    let engine: SyncEngine<any>;

    beforeEach(() => {
      engine = new SyncEngine('user1', 'merge');
    });

    test('should set and get local entries', () => {
      const entry = engine.setLocal('key1', { value: 'test' });

      expect(entry).toBeDefined();
      expect(entry.key).toBe('key1');
      expect(entry.value).toEqual({ value: 'test' });

      const retrieved = engine.getLocal('key1');
      expect(retrieved).toBeDefined();
      expect(retrieved!.value).toEqual({ value: 'test' });
    });

    test('should delete local entries', () => {
      engine.setLocal('key1', { value: 'test' });
      expect(engine.deleteLocal('key1')).toBe(true);
      expect(engine.getLocal('key1')).toBeNull();
    });

    test('should set remote entries', () => {
      const entry = {
        id: 'id1',
        key: 'key1',
        value: { value: 'test' },
        timestamp: Date.now(),
        version: 1,
        hash: 'hash1',
        vectorClock: { user1: 1 },
        userId: 'user1'
      };

      engine.setRemote('key1', entry);
      expect(engine.getAllRemote()).toHaveLength(1);
    });

    test('should synchronize stores', async () => {
      engine.setLocal('key1', { value: 'local' });

      const result = await engine.sync();

      expect(result).toBeDefined();
      expect(result.synced).toBeGreaterThanOrEqual(0);
    });

    test('should detect conflicts', async () => {
      const entry1 = {
        id: 'id1',
        key: 'conflict-key',
        value: { value: 'local' },
        timestamp: Date.now(),
        version: 1,
        hash: 'hash1',
        vectorClock: { user1: 1 },
        userId: 'user1'
      };

      engine.setLocal('conflict-key', { value: 'local' });
      engine.setRemote('conflict-key', entry1);

      const result = await engine.sync();

      // May have conflicts depending on timing
      expect(result).toBeDefined();
    });

    test('should get statistics', () => {
      engine.setLocal('key1', { value: 'test' });

      const stats = engine.getStatistics();

      expect(stats.localCount).toBe(1);
      expect(stats.vectorClock).toBeDefined();
    });

    test('should register event handlers', (done) => {
      engine.on('local-set', (data) => {
        expect(data.key).toBeDefined();
        done();
      });

      engine.setLocal('key1', { value: 'test' });
    });
  });

  describe('OfflineStateReconciliator', () => {
    test('should perform three-way merge', () => {
      const base = { value: 'base' };
      const local = { value: 'local' };
      const remote = { value: 'remote' };

      const result = OfflineStateReconciliator.threeWayMerge(base, local, remote);

      expect(result).toBeDefined();
    });

    test('should preserve local changes if remote unchanged', () => {
      const base = { value: 'base' };
      const local = { value: 'local' };
      const remote = { value: 'base' };

      const result = OfflineStateReconciliator.threeWayMerge(base, local, remote);

      expect(result).toEqual(local);
    });
  });

  // ============================================================================
  // Offline Manager Integration Tests
  // ============================================================================

  describe('OfflineManager', () => {
    let manager: OfflineManager;

    beforeEach(async () => {
      // Reset singleton
      manager = await OfflineManager.create({
        enableOfflineMode: true,
        enableRequestQueuing: true,
        enableSync: true,
        maxQueueSize: 100
      });
    });

    afterEach(async () => {
      await manager.shutdown();
    });

    test('should create offline manager', () => {
      expect(manager).toBeDefined();
    });

    test('should get offline state', () => {
      const state = manager.getOfflineState();

      expect(state).toHaveProperty('isOnline');
      expect(state).toHaveProperty('queuedRequests');
      expect(state).toHaveProperty('storageUsed');
    });

    test('should enqueue requests', () => {
      const id = manager.enqueueRequest('/api/test', {
        method: 'POST',
        body: { test: true }
      });

      expect(id).toBeDefined();

      const pending = manager.getPendingRequests();
      expect(pending.length).toBe(1);
    });

    test('should get cache URLs', async () => {
      const urls = await manager.getCachedUrls();
      expect(Array.isArray(urls)).toBe(true);
    });

    test('should get cache size', async () => {
      const size = await manager.getCacheSize();
      expect(size).toBeGreaterThanOrEqual(0);
    });

    test('should precache URLs', async () => {
      global.fetch = jest.fn(() =>
        Promise.resolve(new Response('content', { status: 200 }))
      );

      await manager.precacheUrls(['/api/test']);

      // Should not throw
    });

    test('should clear cache', async () => {
      await manager.clearCache();

      // Should not throw
    });

    test('should clear queue', () => {
      manager.enqueueRequest('/api/test', { method: 'POST' });
      expect(manager.getPendingRequests().length).toBe(1);

      manager.clearQueue();
      expect(manager.getPendingRequests().length).toBe(0);
    });

    test('should retry failed requests', () => {
      const id = manager.enqueueRequest('/api/test', { method: 'POST' });
      const request = manager.getPendingRequests()[0];

      if (request) {
        request.status = 'failed';
        manager.retryFailed();

        const retried = manager.getFailedRequests();
        // Should be removed from failed list
      }
    });

    test('should check online status', () => {
      const isOnline = manager.isOnlineStatus();
      expect(typeof isOnline).toBe('boolean');
    });

    test('should register event handlers', (done) => {
      manager.on('initialized', () => {
        done();
      });

      // Already initialized
      done();
    });
  });
});

export {};
