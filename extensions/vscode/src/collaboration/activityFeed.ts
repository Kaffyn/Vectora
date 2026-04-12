/**
 * Activity Feed Module - User Activity Tracking
 * Phase 14: Collaborative Features
 *
 * Provides:
 * - Activity logging
 * - Activity feed
 * - Timeline generation
 * - Activity filtering and search
 */

/**
 * Activity type
 */
export type ActivityType = 'edit' | 'chat' | 'mention' | 'collaboration' | 'file_change' | 'comment' | 'reaction' | 'presence';

/**
 * Activity entry
 */
export interface Activity {
  id: string;
  type: ActivityType;
  userId: string;
  username: string;
  timestamp: number;
  content: string;
  resourceId?: string;
  resourceType?: string;
  metadata?: Record<string, any>;
  relatedUsers?: string[];
}

/**
 * Activity statistics
 */
export interface ActivityStats {
  totalActivities: number;
  activitiesByType: Record<ActivityType, number>;
  activitiesByUser: Record<string, number>;
  mostActiveUsers: Array<{ userId: string; count: number }>;
  recentActivities: Activity[];
}

/**
 * Activity feed manager
 */
export class ActivityFeedManager {
  private activities: Activity[] = [];
  private maxActivities: number = 10000;
  private filters: Set<ActivityType> = new Set();
  private eventHandlers: Map<string, Function[]> = new Map();
  private activityIndex: Map<string, Activity> = new Map();

  constructor(maxActivities: number = 10000) {
    this.maxActivities = maxActivities;
    this.filters = new Set(['edit', 'chat', 'mention', 'collaboration', 'file_change', 'comment', 'reaction']);
  }

  /**
   * Log activity
   */
  logActivity(activity: Omit<Activity, 'id' | 'timestamp'>): Activity {
    const fullActivity: Activity = {
      ...activity,
      id: `activity-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`,
      timestamp: Date.now()
    };

    this.activities.push(fullActivity);
    this.activityIndex.set(fullActivity.id, fullActivity);

    // Keep activities under max size
    if (this.activities.length > this.maxActivities) {
      const removed = this.activities.shift();
      if (removed) {
        this.activityIndex.delete(removed.id);
      }
    }

    this.emit('activity-logged', fullActivity);

    return fullActivity;
  }

  /**
   * Get recent activities
   */
  getRecentActivities(limit: number = 50): Activity[] {
    return this.activities
      .filter(a => this.filters.has(a.type))
      .sort((a, b) => b.timestamp - a.timestamp)
      .slice(0, limit);
  }

  /**
   * Get activities by user
   */
  getActivitiesByUser(userId: string, limit: number = 50): Activity[] {
    return this.activities
      .filter(a => a.userId === userId && this.filters.has(a.type))
      .sort((a, b) => b.timestamp - a.timestamp)
      .slice(0, limit);
  }

  /**
   * Get activities by type
   */
  getActivitiesByType(type: ActivityType, limit: number = 50): Activity[] {
    return this.activities
      .filter(a => a.type === type)
      .sort((a, b) => b.timestamp - a.timestamp)
      .slice(0, limit);
  }

  /**
   * Get activities in time range
   */
  getActivitiesInRange(startTime: number, endTime: number): Activity[] {
    return this.activities.filter(a => a.timestamp >= startTime && a.timestamp <= endTime);
  }

  /**
   * Get activity by ID
   */
  getActivity(id: string): Activity | null {
    return this.activityIndex.get(id) || null;
  }

  /**
   * Search activities
   */
  search(query: string): Activity[] {
    const lowerQuery = query.toLowerCase();

    return this.activities.filter(a =>
      a.content.toLowerCase().includes(lowerQuery) ||
      a.username.toLowerCase().includes(lowerQuery) ||
      a.resourceId?.toLowerCase().includes(lowerQuery)
    );
  }

  /**
   * Get activity statistics
   */
  getStatistics(timeRangeMs: number = 24 * 60 * 60 * 1000): ActivityStats {
    const now = Date.now();
    const recentActivities = this.activities.filter(a => a.timestamp >= now - timeRangeMs);

    const stats: ActivityStats = {
      totalActivities: recentActivities.length,
      activitiesByType: {
        edit: 0,
        chat: 0,
        mention: 0,
        collaboration: 0,
        file_change: 0,
        comment: 0,
        reaction: 0,
        presence: 0
      },
      activitiesByUser: {},
      mostActiveUsers: [],
      recentActivities: recentActivities.slice(-10)
    };

    for (const activity of recentActivities) {
      stats.activitiesByType[activity.type]++;
      stats.activitiesByUser[activity.userId] = (stats.activitiesByUser[activity.userId] || 0) + 1;
    }

    // Get most active users
    stats.mostActiveUsers = Object.entries(stats.activitiesByUser)
      .map(([userId, count]) => ({ userId, count }))
      .sort((a, b) => b.count - a.count)
      .slice(0, 10);

    return stats;
  }

  /**
   * Set activity type filters
   */
  setFilters(types: ActivityType[]): void {
    this.filters.clear();
    types.forEach(type => this.filters.add(type));
  }

  /**
   * Add activity type filter
   */
  addFilter(type: ActivityType): void {
    this.filters.add(type);
  }

  /**
   * Remove activity type filter
   */
  removeFilter(type: ActivityType): void {
    this.filters.delete(type);
  }

  /**
   * Get active filters
   */
  getFilters(): ActivityType[] {
    return Array.from(this.filters);
  }

  /**
   * Clear all activities
   */
  clear(): void {
    this.activities = [];
    this.activityIndex.clear();
  }

  /**
   * Export activities as JSON
   */
  exportJSON(): string {
    return JSON.stringify(this.activities, null, 2);
  }

  /**
   * Get activity count
   */
  getCount(): number {
    return this.activities.length;
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

export default ActivityFeedManager;
