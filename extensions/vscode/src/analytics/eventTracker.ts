/**
 * Event Tracker
 * Tracks and records analytics events
 */

import { v4 as uuidv4 } from 'uuid';
import {
  AnalyticsEvent,
  EventType,
  ChatEvent,
  RAGEvent,
  ProviderEvent,
  UIActionEvent,
  Alert,
} from '../types/analytics';

export class EventTracker {
  private sessionId: string;
  private eventStack: AnalyticsEvent[] = [];
  private eventTimestamps: Map<string, number> = new Map();

  constructor(sessionId: string) {
    this.sessionId = sessionId;
  }

  /**
   * Create a new analytics event
   */
  createEvent(
    type: EventType,
    metadata: Record<string, unknown>,
    sessionId: string,
    userId?: string,
  ): AnalyticsEvent {
    const event: AnalyticsEvent = {
      id: uuidv4(),
      type,
      timestamp: Date.now(),
      metadata,
      sessionId,
      userId,
    };

    this.eventStack.push(event);
    return event;
  }

  /**
   * Create a chat event
   */
  createChatEvent(
    type: 'send' | 'receive' | 'error',
    data: {
      messageLength: number;
      tokenCount?: number;
      provider?: string;
      model?: string;
      responseTime?: number;
      error?: string;
    },
  ): ChatEvent {
    const eventType: EventType = type === 'error' ? 'chat.error' : `chat.${type}` as EventType;
    const event: ChatEvent = {
      id: uuidv4(),
      type: eventType,
      timestamp: Date.now(),
      metadata: data,
      sessionId: this.sessionId,
    };

    this.eventStack.push(event);
    return event;
  }

  /**
   * Create a RAG event
   */
  createRAGEvent(
    type: 'search' | 'retrieve' | 'rank',
    data: {
      query: string;
      resultCount: number;
      relevanceScore?: number;
      executionTime: number;
      cacheHit?: boolean;
    },
  ): RAGEvent {
    const eventType: EventType = `rag.${type}` as EventType;
    const event: RAGEvent = {
      id: uuidv4(),
      type: eventType,
      timestamp: Date.now(),
      metadata: data,
      sessionId: this.sessionId,
    };

    this.eventStack.push(event);
    return event;
  }

  /**
   * Create a provider event
   */
  createProviderEvent(
    type: 'switch' | 'fallback' | 'error',
    data: {
      provider: string;
      previousProvider?: string;
      reason?: string;
      error?: string;
    },
  ): ProviderEvent {
    const eventType: EventType = `provider.${type}` as EventType;
    const event: ProviderEvent = {
      id: uuidv4(),
      type: eventType,
      timestamp: Date.now(),
      metadata: data,
      sessionId: this.sessionId,
    };

    this.eventStack.push(event);
    return event;
  }

  /**
   * Create a UI action event
   */
  createUIActionEvent(data: {
    action: string;
    component: string;
    targetId?: string;
    value?: unknown;
  }): UIActionEvent {
    const event: UIActionEvent = {
      id: uuidv4(),
      type: 'ui.action',
      timestamp: Date.now(),
      metadata: data,
      sessionId: this.sessionId,
    };

    this.eventStack.push(event);
    return event;
  }

  /**
   * Track event duration
   */
  startTrackingDuration(eventId: string): void {
    this.eventTimestamps.set(eventId, Date.now());
  }

  /**
   * End duration tracking and update event
   */
  endTrackingDuration(eventId: string, event: AnalyticsEvent): AnalyticsEvent {
    const startTime = this.eventTimestamps.get(eventId);
    if (startTime) {
      event.duration = Date.now() - startTime;
      this.eventTimestamps.delete(eventId);
    }
    return event;
  }

  /**
   * Log an alert as an event
   */
  logAlert(alert: Alert): void {
    this.createEvent('system.warning', {
      alertId: alert.id,
      ruleId: alert.ruleId,
      severity: alert.severity,
      title: alert.title,
      message: alert.message,
    }, this.sessionId);
  }

  /**
   * Log a system error
   */
  logSystemError(error: Error | string, context?: Record<string, unknown>): void {
    const message = typeof error === 'string' ? error : error.message;
    this.createEvent('system.error', {
      message,
      stack: error instanceof Error ? error.stack : undefined,
      context,
    }, this.sessionId);
  }

  /**
   * Log a system warning
   */
  logSystemWarning(message: string, context?: Record<string, unknown>): void {
    this.createEvent('system.warning', {
      message,
      context,
    }, this.sessionId);
  }

  /**
   * Log cache hit
   */
  logCacheHit(key: string, hitSize?: number): void {
    this.createEvent('cache.hit', {
      key,
      hitSize,
      timestamp: Date.now(),
    }, this.sessionId);
  }

  /**
   * Log cache miss
   */
  logCacheMiss(key: string): void {
    this.createEvent('cache.miss', {
      key,
      timestamp: Date.now(),
    }, this.sessionId);
  }

  /**
   * Log memory warning
   */
  logMemoryWarning(usagePercent: number, usageMB: number): void {
    this.createEvent('memory.warning', {
      usagePercent,
      usageMB,
    }, this.sessionId);
  }

  /**
   * Log rate limit warning
   */
  logRateLimitWarning(provider: string, remaining: number, limit: number): void {
    this.createEvent('rate.limit.warning', {
      provider,
      remaining,
      limit,
      usagePercent: ((limit - remaining) / limit) * 100,
    }, this.sessionId);
  }

  /**
   * Log token usage
   */
  logTokenUsage(
    provider: string,
    inputTokens: number,
    outputTokens: number,
    costUSD: number,
  ): void {
    this.createEvent('token.usage', {
      provider,
      inputTokens,
      outputTokens,
      totalTokens: inputTokens + outputTokens,
      costUSD,
    }, this.sessionId);
  }

  /**
   * Get all tracked events
   */
  getEvents(): AnalyticsEvent[] {
    return [...this.eventStack];
  }

  /**
   * Get events of a specific type
   */
  getEventsByType(type: EventType): AnalyticsEvent[] {
    return this.eventStack.filter((e) => e.type === type);
  }

  /**
   * Get events within a time range
   */
  getEventsByTimeRange(startTime: number, endTime: number): AnalyticsEvent[] {
    return this.eventStack.filter(
      (e) => e.timestamp >= startTime && e.timestamp <= endTime,
    );
  }

  /**
   * Clear event stack
   */
  clear(): void {
    this.eventStack = [];
    this.eventTimestamps.clear();
  }

  /**
   * Get event count
   */
  getEventCount(): number {
    return this.eventStack.length;
  }

  /**
   * Get event count by type
   */
  getEventCountByType(): Record<EventType, number> {
    const counts: Record<EventType, number> = {} as Record<EventType, number>;
    for (const event of this.eventStack) {
      counts[event.type] = (counts[event.type] || 0) + 1;
    }
    return counts;
  }

  /**
   * Get last event
   */
  getLastEvent(): AnalyticsEvent | undefined {
    return this.eventStack[this.eventStack.length - 1];
  }

  /**
   * Get events for session
   */
  getSessionEvents(): AnalyticsEvent[] {
    return this.eventStack.filter((e) => e.sessionId === this.sessionId);
  }
}
