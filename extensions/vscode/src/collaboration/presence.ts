/**
 * Presence Tracking Module - User Presence & Activity
 * Phase 14: Collaborative Features
 *
 * Provides:
 * - User presence tracking
 * - Cursor/selection position sharing
 * - Activity indicators
 * - User state synchronization
 */

/**
 * User presence information
 */
export interface UserPresence {
  userId: string;
  sessionId: string;
  username: string;
  status: 'online' | 'idle' | 'offline' | 'away';
  lastSeen: number;
  color: string;
  metadata?: Record<string, any>;
}

/**
 * Cursor position
 */
export interface CursorPosition {
  userId: string;
  line: number;
  column: number;
  fileId: string;
  timestamp: number;
}

/**
 * Selection range
 */
export interface SelectionRange {
  userId: string;
  fileId: string;
  startLine: number;
  startColumn: number;
  endLine: number;
  endColumn: number;
  timestamp: number;
}

/**
 * Presence manager
 */
export class PresenceManager {
  private presences: Map<string, UserPresence> = new Map();
  private cursors: Map<string, CursorPosition> = new Map();
  private selections: Map<string, SelectionRange> = new Map();
  private idleTimeout: number = 300000; // 5 minutes
  private updateInterval: NodeJS.Timer | null = null;
  private eventHandlers: Map<string, Function[]> = new Map();

  constructor(idleTimeoutMs: number = 300000) {
    this.idleTimeout = idleTimeoutMs;
  }

  /**
   * Set user presence
   */
  setPresence(presence: UserPresence): void {
    this.presences.set(presence.userId, {
      ...presence,
      lastSeen: Date.now()
    });

    this.emit('presence-updated', presence);
  }

  /**
   * Get user presence
   */
  getPresence(userId: string): UserPresence | null {
    return this.presences.get(userId) || null;
  }

  /**
   * Get all presences
   */
  getAllPresences(): UserPresence[] {
    return Array.from(this.presences.values());
  }

  /**
   * Update user status
   */
  updateStatus(userId: string, status: 'online' | 'idle' | 'offline' | 'away'): boolean {
    const presence = this.presences.get(userId);
    if (!presence) {
      return false;
    }

    presence.status = status;
    presence.lastSeen = Date.now();
    this.emit('status-changed', { userId, status });

    return true;
  }

  /**
   * Remove user presence
   */
  removePresence(userId: string): boolean {
    const removed = this.presences.delete(userId);
    if (removed) {
      this.emit('presence-removed', { userId });
    }
    return removed;
  }

  /**
   * Set cursor position
   */
  setCursorPosition(position: CursorPosition): void {
    this.cursors.set(position.userId, position);
    this.emit('cursor-moved', position);
  }

  /**
   * Get cursor position
   */
  getCursorPosition(userId: string): CursorPosition | null {
    return this.cursors.get(userId) || null;
  }

  /**
   * Get all cursor positions
   */
  getAllCursorPositions(): CursorPosition[] {
    return Array.from(this.cursors.values());
  }

  /**
   * Get cursor positions for file
   */
  getFileCursorPositions(fileId: string): CursorPosition[] {
    return Array.from(this.cursors.values()).filter(c => c.fileId === fileId);
  }

  /**
   * Set selection range
   */
  setSelectionRange(selection: SelectionRange): void {
    this.selections.set(selection.userId, selection);
    this.emit('selection-changed', selection);
  }

  /**
   * Get selection range
   */
  getSelectionRange(userId: string): SelectionRange | null {
    return this.selections.get(userId) || null;
  }

  /**
   * Get all selections
   */
  getAllSelections(): SelectionRange[] {
    return Array.from(this.selections.values());
  }

  /**
   * Get selections for file
   */
  getFileSelections(fileId: string): SelectionRange[] {
    return Array.from(this.selections.values()).filter(s => s.fileId === fileId);
  }

  /**
   * Start tracking user activity
   */
  startActivityTracking(userId: string, updateInterval: number = 10000): void {
    if (this.updateInterval) {
      clearInterval(this.updateInterval);
    }

    this.updateInterval = setInterval(() => {
      this.checkIdleUsers();
    }, updateInterval);
  }

  /**
   * Stop activity tracking
   */
  stopActivityTracking(): void {
    if (this.updateInterval) {
      clearInterval(this.updateInterval);
      this.updateInterval = null;
    }
  }

  /**
   * Check for idle users
   */
  private checkIdleUsers(): void {
    const now = Date.now();

    for (const [userId, presence] of this.presences.entries()) {
      if (presence.status === 'online' && now - presence.lastSeen > this.idleTimeout) {
        this.updateStatus(userId, 'idle');
      }
    }
  }

  /**
   * Get online users
   */
  getOnlineUsers(): UserPresence[] {
    return Array.from(this.presences.values()).filter(p => p.status === 'online' || p.status === 'idle');
  }

  /**
   * Get offline users
   */
  getOfflineUsers(): UserPresence[] {
    return Array.from(this.presences.values()).filter(p => p.status === 'offline');
  }

  /**
   * Get user count by status
   */
  getUserCountByStatus(): Record<string, number> {
    const counts = {
      online: 0,
      idle: 0,
      offline: 0,
      away: 0
    };

    for (const presence of this.presences.values()) {
      counts[presence.status]++;
    }

    return counts;
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
   * Clear all data
   */
  clear(): void {
    this.presences.clear();
    this.cursors.clear();
    this.selections.clear();
    this.stopActivityTracking();
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
 * User colors for presence visualization
 */
export class PresenceColorManager {
  private colors = [
    '#FF6B6B', '#4ECDC4', '#45B7D1', '#FFA07A', '#98D8C8',
    '#F7DC6F', '#BB8FCE', '#85C1E2', '#F8B88B', '#A9DFBF'
  ];

  private userColors: Map<string, string> = new Map();

  /**
   * Get color for user
   */
  getColor(userId: string): string {
    if (this.userColors.has(userId)) {
      return this.userColors.get(userId)!;
    }

    const color = this.colors[this.userColors.size % this.colors.length];
    this.userColors.set(userId, color);

    return color;
  }

  /**
   * Clear colors
   */
  clear(): void {
    this.userColors.clear();
  }
}

export default PresenceManager;
