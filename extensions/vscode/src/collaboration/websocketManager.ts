/**
 * WebSocket Manager - Real-time Communication
 * Phase 14: Collaborative Features
 *
 * Provides:
 * - WebSocket connection management
 * - Message routing and handling
 * - Connection pooling
 * - Automatic reconnection
 * - Message queuing
 */

/**
 * WebSocket message
 */
export interface WebSocketMessage {
  id: string;
  type: string;
  timestamp: number;
  userId: string;
  sessionId: string;
  data: any;
  metadata?: Record<string, any>;
}

/**
 * WebSocket configuration
 */
export interface WebSocketConfig {
  url: string;
  reconnect?: boolean;
  reconnectDelay?: number;
  reconnectAttempts?: number;
  messageTimeout?: number;
  pingInterval?: number;
}

/**
 * Connection state
 */
export type ConnectionState = 'disconnected' | 'connecting' | 'connected' | 'reconnecting' | 'error';

/**
 * WebSocketManager - Manages WebSocket connections
 */
export class WebSocketManager {
  private ws: WebSocket | null = null;
  private config: Required<WebSocketConfig>;
  private state: ConnectionState = 'disconnected';
  private reconnectAttempts: number = 0;
  private messageHandlers: Map<string, Function[]> = new Map();
  private pendingMessages: WebSocketMessage[] = [];
  private messageId: number = 0;
  private pingInterval: NodeJS.Timer | null = null;
  private eventHandlers: Map<string, Function[]> = new Map();
  private userId: string = '';
  private sessionId: string = '';

  constructor(config: WebSocketConfig) {
    this.config = {
      url: config.url,
      reconnect: config.reconnect ?? true,
      reconnectDelay: config.reconnectDelay ?? 1000,
      reconnectAttempts: config.reconnectAttempts ?? 5,
      messageTimeout: config.messageTimeout ?? 30000,
      pingInterval: config.pingInterval ?? 30000
    };
  }

  /**
   * Connect to WebSocket server
   */
  async connect(userId: string, sessionId: string): Promise<void> {
    this.userId = userId;
    this.sessionId = sessionId;

    return new Promise((resolve, reject) => {
      try {
        this.setState('connecting');
        this.ws = new WebSocket(this.config.url);

        this.ws.onopen = () => {
          this.reconnectAttempts = 0;
          this.setState('connected');
          this.startPingInterval();
          this.flushPendingMessages();
          this.emit('connected', null);
          resolve();
        };

        this.ws.onmessage = (event) => {
          this.handleMessage(event.data);
        };

        this.ws.onerror = (event) => {
          this.setState('error');
          this.emit('error', event);
          reject(event);
        };

        this.ws.onclose = () => {
          this.stopPingInterval();
          if (this.state !== 'disconnected') {
            this.handleClose();
          }
        };
      } catch (error) {
        this.setState('error');
        reject(error);
      }
    });
  }

  /**
   * Disconnect from WebSocket
   */
  disconnect(): void {
    this.setState('disconnected');
    this.stopPingInterval();

    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }

    this.emit('disconnected', null);
  }

  /**
   * Send message
   */
  async sendMessage(type: string, data: any): Promise<string> {
    const id = `msg-${++this.messageId}`;

    const message: WebSocketMessage = {
      id,
      type,
      timestamp: Date.now(),
      userId: this.userId,
      sessionId: this.sessionId,
      data
    };

    if (this.state === 'connected') {
      this.ws!.send(JSON.stringify(message));
      this.emit('message-sent', message);
    } else {
      // Queue message for later
      this.pendingMessages.push(message);
      this.emit('message-queued', message);
    }

    return id;
  }

  /**
   * Subscribe to message type
   */
  on(messageType: string, handler: Function): void {
    if (!this.messageHandlers.has(messageType)) {
      this.messageHandlers.set(messageType, []);
    }
    this.messageHandlers.get(messageType)!.push(handler);
  }

  /**
   * Unsubscribe from message type
   */
  off(messageType: string, handler: Function): void {
    const handlers = this.messageHandlers.get(messageType);
    if (handlers) {
      const index = handlers.indexOf(handler);
      if (index > -1) {
        handlers.splice(index, 1);
      }
    }
  }

  /**
   * Register event handler
   */
  addEventListener(event: string, handler: Function): void {
    if (!this.eventHandlers.has(event)) {
      this.eventHandlers.set(event, []);
    }
    this.eventHandlers.get(event)!.push(handler);
  }

  /**
   * Unregister event handler
   */
  removeEventListener(event: string, handler: Function): void {
    const handlers = this.eventHandlers.get(event);
    if (handlers) {
      const index = handlers.indexOf(handler);
      if (index > -1) {
        handlers.splice(index, 1);
      }
    }
  }

  /**
   * Get connection state
   */
  getState(): ConnectionState {
    return this.state;
  }

  /**
   * Check if connected
   */
  isConnected(): boolean {
    return this.state === 'connected';
  }

  /**
   * Get pending message count
   */
  getPendingMessageCount(): number {
    return this.pendingMessages.length;
  }

  /**
   * Handle incoming message
   */
  private handleMessage(data: string): void {
    try {
      const message: WebSocketMessage = JSON.parse(data);
      this.emit('message-received', message);

      // Route to message type handlers
      const handlers = this.messageHandlers.get(message.type);
      if (handlers) {
        for (const handler of handlers) {
          try {
            handler(message);
          } catch (error) {
            console.error(`Error in message handler for ${message.type}:`, error);
          }
        }
      }
    } catch (error) {
      console.error('Failed to parse WebSocket message:', error);
    }
  }

  /**
   * Handle connection close
   */
  private handleClose(): void {
    if (this.config.reconnect && this.reconnectAttempts < this.config.reconnectAttempts) {
      this.setState('reconnecting');
      this.reconnectAttempts++;

      const delay = this.config.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1);

      setTimeout(() => {
        this.connect(this.userId, this.sessionId);
      }, delay);
    } else {
      this.setState('disconnected');
      this.emit('reconnection-failed', null);
    }
  }

  /**
   * Flush pending messages when connected
   */
  private flushPendingMessages(): void {
    while (this.pendingMessages.length > 0) {
      const message = this.pendingMessages.shift()!;
      this.ws!.send(JSON.stringify(message));
      this.emit('message-sent', message);
    }
  }

  /**
   * Start ping interval for keep-alive
   */
  private startPingInterval(): void {
    this.pingInterval = setInterval(() => {
      if (this.isConnected()) {
        this.sendMessage('ping', { timestamp: Date.now() });
      }
    }, this.config.pingInterval);
  }

  /**
   * Stop ping interval
   */
  private stopPingInterval(): void {
    if (this.pingInterval) {
      clearInterval(this.pingInterval);
      this.pingInterval = null;
    }
  }

  /**
   * Set connection state
   */
  private setState(state: ConnectionState): void {
    if (this.state !== state) {
      this.state = state;
      this.emit('state-changed', { state });
    }
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
 * MultiWebSocketManager - Manages multiple WebSocket connections
 */
export class MultiWebSocketManager {
  private connections: Map<string, WebSocketManager> = new Map();

  /**
   * Create connection
   */
  createConnection(id: string, config: WebSocketConfig): WebSocketManager {
    const manager = new WebSocketManager(config);
    this.connections.set(id, manager);
    return manager;
  }

  /**
   * Get connection
   */
  getConnection(id: string): WebSocketManager | null {
    return this.connections.get(id) || null;
  }

  /**
   * Remove connection
   */
  removeConnection(id: string): boolean {
    const manager = this.connections.get(id);
    if (manager) {
      manager.disconnect();
      this.connections.delete(id);
      return true;
    }
    return false;
  }

  /**
   * Get all connections
   */
  getAllConnections(): WebSocketManager[] {
    return Array.from(this.connections.values());
  }

  /**
   * Disconnect all
   */
  disconnectAll(): void {
    for (const manager of this.connections.values()) {
      manager.disconnect();
    }
    this.connections.clear();
  }

  /**
   * Get active connections count
   */
  getActiveConnectionsCount(): number {
    return Array.from(this.connections.values()).filter(m => m.isConnected()).length;
  }
}

export default WebSocketManager;
