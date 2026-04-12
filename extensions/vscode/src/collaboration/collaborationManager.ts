/**
 * Collaboration Manager - Unified Real-time Collaboration
 * Phase 14: Collaborative Features
 *
 * Provides:
 * - Multi-user workspace management
 * - Real-time editing with OT
 * - Presence and activity tracking
 * - Workspace synchronization
 */

import { WebSocketManager, type WebSocketConfig } from './websocketManager';
import { PresenceManager, type UserPresence } from './presence';
import { OperationalTransformer, CollaborativeDocument, type Operation, type ChangeSet } from './operationalTransform';
import { ActivityFeedManager, type Activity } from './activityFeed';

/**
 * Collaboration configuration
 */
export interface CollaborationConfig {
  websocketUrl: string;
  enableRealtimeEditing?: boolean;
  enablePresenceTracking?: boolean;
  enableActivityLogging?: boolean;
  maxDocumentSize?: number;
  autoSaveInterval?: number;
}

/**
 * Collaborative workspace
 */
export interface CollaborativeWorkspace {
  id: string;
  name: string;
  ownerId: string;
  members: UserPresence[];
  documents: Map<string, CollaborativeDocument>;
  createdAt: number;
  updatedAt: number;
}

/**
 * CollaborationManager - Main collaboration orchestrator
 */
export class CollaborationManager {
  private websocket: WebSocketManager | null = null;
  private presenceManager: PresenceManager;
  private documents: Map<string, CollaborativeDocument> = new Map();
  private activityFeed: ActivityFeedManager;
  private workspace: CollaborativeWorkspace | null = null;

  private config: Required<CollaborationConfig>;
  private userId: string = '';
  private sessionId: string = '';
  private eventHandlers: Map<string, Function[]> = new Map();
  private autoSaveInterval: NodeJS.Timer | null = null;

  private constructor(config: CollaborationConfig) {
    this.config = {
      websocketUrl: config.websocketUrl,
      enableRealtimeEditing: config.enableRealtimeEditing ?? true,
      enablePresenceTracking: config.enablePresenceTracking ?? true,
      enableActivityLogging: config.enableActivityLogging ?? true,
      maxDocumentSize: config.maxDocumentSize ?? 10 * 1024 * 1024,
      autoSaveInterval: config.autoSaveInterval ?? 30000
    };

    this.presenceManager = new PresenceManager();
    this.activityFeed = new ActivityFeedManager();
  }

  /**
   * Create and initialize collaboration manager
   */
  static async create(config: CollaborationConfig): Promise<CollaborationManager> {
    const manager = new CollaborationManager(config);
    await manager.initialize();
    return manager;
  }

  /**
   * Initialize collaboration manager
   */
  private async initialize(): Promise<void> {
    try {
      // Setup WebSocket
      this.websocket = new WebSocketManager({
        url: this.config.websocketUrl,
        reconnect: true,
        reconnectAttempts: 5
      });

      // Setup message handlers
      this.setupMessageHandlers();

      this.emit('initialized', null);
    } catch (error) {
      console.error('Failed to initialize collaboration manager:', error);
      this.emit('initialization-error', error);
    }
  }

  /**
   * Connect to collaboration session
   */
  async connect(userId: string, sessionId: string, workspaceId: string): Promise<void> {
    this.userId = userId;
    this.sessionId = sessionId;

    try {
      if (!this.websocket) {
        throw new Error('WebSocket not initialized');
      }

      await this.websocket.connect(userId, sessionId);

      // Join workspace
      await this.websocket.sendMessage('workspace:join', {
        workspaceId,
        userId,
        sessionId
      });

      this.emit('connected', { userId, workspaceId });
    } catch (error) {
      console.error('Failed to connect:', error);
      this.emit('connection-error', error);
      throw error;
    }
  }

  /**
   * Disconnect from collaboration
   */
  disconnect(): void {
    this.stopAutoSave();

    if (this.websocket) {
      this.websocket.disconnect();
    }

    this.presenceManager.clear();
    this.documents.clear();

    this.emit('disconnected', null);
  }

  /**
   * Create or open document
   */
  createDocument(documentId: string, initialContent: string = ''): CollaborativeDocument {
    let doc = this.documents.get(documentId);

    if (!doc) {
      doc = new CollaborativeDocument(initialContent);
      this.documents.set(documentId, doc);
    }

    return doc;
  }

  /**
   * Get document
   */
  getDocument(documentId: string): CollaborativeDocument | null {
    return this.documents.get(documentId) || null;
  }

  /**
   * Apply operations to document
   */
  async applyOperations(documentId: string, operations: Operation[]): Promise<void> {
    const doc = this.getDocument(documentId);
    if (!doc) {
      throw new Error(`Document ${documentId} not found`);
    }

    // Create change set
    const changeSet = doc.createChangeSet(operations, this.userId);

    // Apply locally
    doc.applyChangeSet(changeSet);

    // Log activity
    if (this.config.enableActivityLogging) {
      this.activityFeed.logActivity({
        type: 'edit',
        userId: this.userId,
        username: this.userId,
        content: `Edited ${documentId}`,
        resourceId: documentId,
        resourceType: 'document'
      });
    }

    // Send to server
    if (this.websocket?.isConnected()) {
      await this.websocket.sendMessage('document:edit', {
        documentId,
        changeSet
      });
    }

    this.emit('document-changed', { documentId, changeSet });
  }

  /**
   * Update presence
   */
  updatePresence(presence: Omit<UserPresence, 'lastSeen'>): void {
    const fullPresence: UserPresence = {
      ...presence,
      lastSeen: Date.now()
    };

    this.presenceManager.setPresence(fullPresence);

    // Send to server
    if (this.websocket?.isConnected()) {
      this.websocket.sendMessage('presence:update', fullPresence);
    }

    this.emit('presence-updated', fullPresence);
  }

  /**
   * Get all presences
   */
  getPresences(): UserPresence[] {
    return this.presenceManager.getAllPresences();
  }

  /**
   * Send chat message
   */
  async sendMessage(text: string): Promise<void> {
    if (!this.websocket?.isConnected()) {
      throw new Error('Not connected');
    }

    await this.websocket.sendMessage('chat:message', {
      text,
      userId: this.userId,
      timestamp: Date.now()
    });

    // Log activity
    if (this.config.enableActivityLogging) {
      this.activityFeed.logActivity({
        type: 'chat',
        userId: this.userId,
        username: this.userId,
        content: text
      });
    }

    this.emit('message-sent', { text, timestamp: Date.now() });
  }

  /**
   * Get activity feed
   */
  getActivityFeed(limit: number = 50): Activity[] {
    return this.activityFeed.getRecentActivities(limit);
  }

  /**
   * Get activity statistics
   */
  getActivityStatistics() {
    return this.activityFeed.getStatistics();
  }

  /**
   * Start auto-save
   */
  startAutoSave(documentId: string, interval?: number): void {
    this.stopAutoSave();

    const saveInterval = interval || this.config.autoSaveInterval;

    this.autoSaveInterval = setInterval(() => {
      const doc = this.getDocument(documentId);
      if (doc) {
        this.emit('auto-save', { documentId, content: doc.getContent() });
      }
    }, saveInterval);
  }

  /**
   * Stop auto-save
   */
  private stopAutoSave(): void {
    if (this.autoSaveInterval) {
      clearInterval(this.autoSaveInterval);
      this.autoSaveInterval = null;
    }
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
   * Setup WebSocket message handlers
   */
  private setupMessageHandlers(): void {
    if (!this.websocket) return;

    // Handle document changes from other users
    this.websocket.on('document:edit', (message: any) => {
      const { documentId, changeSet } = message.data;
      const doc = this.getDocument(documentId);

      if (doc) {
        doc.applyChangeSet(changeSet);
        this.emit('remote-edit', { documentId, changeSet });
      }
    });

    // Handle presence updates
    this.websocket.on('presence:update', (message: any) => {
      const presence = message.data as UserPresence;
      if (presence.userId !== this.userId) {
        this.presenceManager.setPresence(presence);
        this.emit('remote-presence-update', presence);
      }
    });

    // Handle chat messages
    this.websocket.on('chat:message', (message: any) => {
      const { text, userId, timestamp } = message.data;

      if (userId !== this.userId) {
        this.activityFeed.logActivity({
          type: 'chat',
          userId,
          username: userId,
          content: text
        });
      }

      this.emit('remote-message', { text, userId, timestamp });
    });

    // Handle workspace updates
    this.websocket.on('workspace:update', (message: any) => {
      this.workspace = message.data;
      this.emit('workspace-updated', this.workspace);
    });
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

export default CollaborationManager;
