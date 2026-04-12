/**
 * Offline Request Queue Module
 * Phase 13: Offline Mode & PWA
 *
 * Provides:
 * - Persistent request queuing
 * - Retry with exponential backoff
 * - Request prioritization
 * - Conflict detection
 * - State reconciliation
 */

/**
 * Queued request
 */
export interface QueuedRequest {
  id: string;
  timestamp: number;
  url: string;
  method: 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH';
  headers?: Record<string, string>;
  body?: any;
  priority: number; // 0 = lowest, 10 = highest
  retries: number;
  maxRetries: number;
  status: 'pending' | 'retrying' | 'failed' | 'completed';
  error?: string;
  tags?: string[];
  dependencies?: string[]; // IDs of requests that must complete first
}

/**
 * Request queue result
 */
export interface QueueResult {
  id: string;
  success: boolean;
  response?: any;
  error?: string;
  retryCount: number;
}

/**
 * OfflineRequestQueue - Manages queued requests for offline scenarios
 */
export class OfflineRequestQueue {
  private queue: Map<string, QueuedRequest> = new Map();
  private processing: boolean = false;
  private maxQueueSize: number;
  private storageKey: string;
  private eventHandlers: Map<string, Function[]> = new Map();
  private isOnline: boolean = navigator.onLine;

  constructor(maxQueueSize: number = 1000, storageKey: string = 'vectora-request-queue') {
    this.maxQueueSize = maxQueueSize;
    this.storageKey = storageKey;

    // Track online/offline status
    window.addEventListener('online', () => this.handleOnline());
    window.addEventListener('offline', () => this.handleOffline());

    // Load persisted queue
    this.loadFromStorage();
  }

  /**
   * Enqueue a request
   */
  enqueueRequest(
    url: string,
    options: {
      method?: 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH';
      headers?: Record<string, string>;
      body?: any;
      priority?: number;
      maxRetries?: number;
      tags?: string[];
      dependencies?: string[];
    } = {}
  ): string {
    if (this.queue.size >= this.maxQueueSize) {
      throw new Error('Request queue is full');
    }

    const id = this.generateId();
    const request: QueuedRequest = {
      id,
      timestamp: Date.now(),
      url,
      method: (options.method || 'POST') as any,
      headers: options.headers,
      body: options.body,
      priority: options.priority || 5,
      retries: 0,
      maxRetries: options.maxRetries || 3,
      status: 'pending',
      tags: options.tags,
      dependencies: options.dependencies
    };

    this.queue.set(id, request);
    this.persistQueue();
    this.emit('request-queued', request);

    return id;
  }

  /**
   * Process queued requests (when online)
   */
  async processQueue(): Promise<QueueResult[]> {
    if (this.processing || !this.isOnline) {
      return [];
    }

    this.processing = true;
    const results: QueueResult[] = [];

    try {
      // Sort by priority (highest first) and timestamp (oldest first)
      const sorted = Array.from(this.queue.values())
        .filter(r => r.status === 'pending')
        .sort((a, b) => {
          if (b.priority !== a.priority) {
            return b.priority - a.priority;
          }
          return a.timestamp - b.timestamp;
        });

      for (const request of sorted) {
        // Check if dependencies are met
        if (!this.areDependenciesMet(request)) {
          continue;
        }

        const result = await this.processRequest(request);
        results.push(result);

        if (result.success) {
          this.queue.delete(request.id);
        }
      }

      this.persistQueue();
    } finally {
      this.processing = false;
    }

    return results;
  }

  /**
   * Process single request with retry logic
   */
  private async processRequest(request: QueuedRequest): Promise<QueueResult> {
    try {
      request.status = 'retrying';
      this.emit('request-processing', request);

      const response = await this.executeRequest(request);

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}`);
      }

      const data = await this.parseResponse(response);

      request.status = 'completed';
      this.emit('request-completed', { request, response: data });

      return {
        id: request.id,
        success: true,
        response: data,
        retryCount: request.retries
      };
    } catch (error) {
      request.retries++;
      request.error = String(error);

      if (request.retries >= request.maxRetries) {
        request.status = 'failed';
        this.emit('request-failed', { request, error });

        return {
          id: request.id,
          success: false,
          error: String(error),
          retryCount: request.retries
        };
      } else {
        // Exponential backoff
        const delay = Math.pow(2, request.retries) * 1000;
        await new Promise(resolve => setTimeout(resolve, delay));

        // Retry
        return this.processRequest(request);
      }
    }
  }

  /**
   * Execute HTTP request
   */
  private async executeRequest(request: QueuedRequest): Promise<Response> {
    const options: RequestInit = {
      method: request.method,
      headers: {
        'Content-Type': 'application/json',
        ...request.headers
      }
    };

    if (request.body) {
      options.body = typeof request.body === 'string'
        ? request.body
        : JSON.stringify(request.body);
    }

    return fetch(request.url, options);
  }

  /**
   * Parse response
   */
  private async parseResponse(response: Response): Promise<any> {
    const contentType = response.headers.get('content-type');

    if (contentType?.includes('application/json')) {
      return response.json();
    } else if (contentType?.includes('text')) {
      return response.text();
    } else {
      return response.blob();
    }
  }

  /**
   * Check if request dependencies are met
   */
  private areDependenciesMet(request: QueuedRequest): boolean {
    if (!request.dependencies || request.dependencies.length === 0) {
      return true;
    }

    for (const depId of request.dependencies) {
      const dep = this.queue.get(depId);
      if (!dep || dep.status !== 'completed') {
        return false;
      }
    }

    return true;
  }

  /**
   * Get queue status
   */
  getStatus(): {
    total: number;
    pending: number;
    retrying: number;
    failed: number;
    completed: number;
    isProcessing: boolean;
  } {
    let pending = 0;
    let retrying = 0;
    let failed = 0;
    let completed = 0;

    for (const request of this.queue.values()) {
      switch (request.status) {
        case 'pending':
          pending++;
          break;
        case 'retrying':
          retrying++;
          break;
        case 'failed':
          failed++;
          break;
        case 'completed':
          completed++;
          break;
      }
    }

    return {
      total: this.queue.size,
      pending,
      retrying,
      failed,
      completed,
      isProcessing: this.processing
    };
  }

  /**
   * Get request by ID
   */
  getRequest(id: string): QueuedRequest | null {
    return this.queue.get(id) || null;
  }

  /**
   * Get requests by tag
   */
  getRequestsByTag(tag: string): QueuedRequest[] {
    return Array.from(this.queue.values())
      .filter(r => r.tags?.includes(tag));
  }

  /**
   * Get pending requests
   */
  getPendingRequests(): QueuedRequest[] {
    return Array.from(this.queue.values())
      .filter(r => r.status === 'pending');
  }

  /**
   * Get failed requests
   */
  getFailedRequests(): QueuedRequest[] {
    return Array.from(this.queue.values())
      .filter(r => r.status === 'failed');
  }

  /**
   * Remove request from queue
   */
  removeRequest(id: string): boolean {
    const removed = this.queue.delete(id);
    if (removed) {
      this.persistQueue();
    }
    return removed;
  }

  /**
   * Clear queue
   */
  clearQueue(): void {
    this.queue.clear();
    this.persistQueue();
  }

  /**
   * Clear by tag
   */
  clearByTag(tag: string): number {
    let count = 0;
    for (const [id, request] of this.queue.entries()) {
      if (request.tags?.includes(tag)) {
        this.queue.delete(id);
        count++;
      }
    }
    if (count > 0) {
      this.persistQueue();
    }
    return count;
  }

  /**
   * Retry failed requests
   */
  retryFailed(): void {
    for (const request of this.queue.values()) {
      if (request.status === 'failed') {
        request.status = 'pending';
        request.retries = 0;
        request.error = undefined;
      }
    }
    this.persistQueue();
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
   * Save queue to localStorage
   */
  private persistQueue(): void {
    try {
      const data = Array.from(this.queue.values());
      localStorage.setItem(this.storageKey, JSON.stringify(data));
    } catch (error) {
      console.error('Failed to persist queue:', error);
    }
  }

  /**
   * Load queue from localStorage
   */
  private loadFromStorage(): void {
    try {
      const data = localStorage.getItem(this.storageKey);
      if (data) {
        const requests: QueuedRequest[] = JSON.parse(data);
        for (const request of requests) {
          this.queue.set(request.id, request);
        }
      }
    } catch (error) {
      console.error('Failed to load queue:', error);
    }
  }

  /**
   * Handle online event
   */
  private handleOnline(): void {
    this.isOnline = true;
    this.emit('online', null);
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
   * Generate unique ID
   */
  private generateId(): string {
    return `${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
  }
}

export default OfflineRequestQueue;
